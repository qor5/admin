package activity

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/samber/lo"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func firstUpperWord(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(string([]rune(name)[0:1]))
}

func getModelName(v any) string {
	segs := strings.Split(reflect.TypeOf(v).String(), ".")
	return strings.TrimLeft(segs[len(segs)-1], "*")
}

func keysValue(v any, keys []string) string {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return ""
	}
	vals := []string{}
	for _, key := range keys {
		rvField := rv.FieldByName(key)
		if !rvField.IsValid() {
			continue
		}
		vals = append(vals, fmt.Sprint(rvField.Interface()))
	}
	return strings.Join(vals, ":")
}

func getPrimaryKeys(t reflect.Type) (keys []string) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primary") {
			keys = append(keys, t.Field(i).Name)
			continue
		}

		if t.Field(i).Type.Kind() == reflect.Ptr && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKeys(t.Field(i).Type.Elem())...)
		}

		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKeys(t.Field(i).Type)...)
		}
	}
	return
}

var (
	schemaParserCacheStore sync.Map
	schemaParserNamer      = schema.NamingStrategy{IdentifierMaxLength: 64}
)

func ParseGormPrimaryFieldNames(v any) ([]string, error) {
	s, err := schema.Parse(v, &schemaParserCacheStore, schemaParserNamer)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}
	return lo.Map(s.PrimaryFields, func(f *schema.Field, _ int) string {
		return f.Name
	}), nil
}

func ObjectID(obj any) string {
	var id string
	if slugger, ok := obj.(presets.SlugEncoder); ok {
		id = slugger.PrimarySlug()
	} else {
		v, err := reflectutils.Get(obj, "ID")
		if err == nil {
			id = fmt.Sprint(v)
		}
	}
	return id
}

func FetchOldWithSlug(db *gorm.DB, ref any, slug string) (any, bool) {
	if slug == "" {
		return FetchOld(db, ref)
	}

	var (
		rt  = reflect.Indirect(reflect.ValueOf(ref)).Type()
		old = reflect.New(rt).Interface()
	)

	if slugger, ok := ref.(presets.SlugDecoder); ok {
		cs := slugger.PrimaryColumnValuesBySlug(slug)
		for key, value := range cs {
			db = db.Where(fmt.Sprintf("%s = ?", key), value)
		}
	} else {
		db = db.Where("id = ?", slug)
	}

	if db.First(old).Error != nil {
		return nil, false
	}

	return old, true
}

func FetchOld(db *gorm.DB, ref any) (any, bool) {
	var (
		rtRef = reflect.Indirect(reflect.ValueOf(ref))
		old   = reflect.New(rtRef.Type()).Interface()
		sqls  []string
		vars  []any
	)

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(ref); err != nil {
		return nil, false
	}

	for _, dbName := range stmt.Schema.DBNames {
		if field := stmt.Schema.LookUpField(dbName); field != nil && field.PrimaryKey {
			if value, isZero := field.ValueOf(db.Statement.Context, rtRef); !isZero {
				sqls = append(sqls, fmt.Sprintf("%v = ?", dbName))
				vars = append(vars, value)
			}
		}
	}

	if len(sqls) == 0 || len(vars) == 0 || len(sqls) != len(vars) {
		return nil, false
	}

	if db.Where(strings.Join(sqls, " AND "), vars...).First(old).Error != nil {
		return nil, false
	}

	return old, true
}

func GetUnreadNotesCount(db *gorm.DB, creatorID string, modelName string, modelKeyses []string) (map[string]int64, error) {
	if creatorID == "" {
		return nil, errors.New("creatorID is required")
	}

	args := []any{
		ActionNote, creatorID,
	}

	var explictWhere string
	if modelName != "" {
		explictWhere = ` AND model_name = ?`
		args = append(args, modelName)
	}
	if len(modelKeyses) > 0 {
		explictWhere += ` AND model_keys IN (?)`
		args = append(args, modelKeyses)
	}

	args = append(args, ActionLastView, creatorID)

	if modelName != "" {
		args = append(args, modelName)
	}
	if len(modelKeyses) > 0 {
		args = append(args, modelKeyses)
	}

	raw := fmt.Sprintf(`
	WITH NoteRecords AS (
		SELECT model_name, model_keys, created_at
		FROM activity_logs
		WHERE action = ? AND creator_id <> ? AND deleted_at IS NULL
			%s
	),
	LastViewedAts AS (
		SELECT model_name, model_keys, MAX(updated_at) AS last_viewed_at
		FROM public.activity_logs
		WHERE action = ? AND creator_id = ? AND deleted_at IS NULL
			%s
		GROUP BY model_name, model_keys
	)
	SELECT
		n.model_name,
		n.model_keys,
		COUNT(*) AS unread_note_count
	FROM NoteRecords n
	LEFT JOIN LastViewedAts lva
		ON n.model_name = lva.model_name
		AND n.model_keys = lva.model_keys
	WHERE lva.last_viewed_at IS NULL
		OR n.created_at > lva.last_viewed_at
	GROUP BY n.model_name, n.model_keys;`, explictWhere, explictWhere)

	result := []struct {
		ModelName       string
		ModelKeys       string
		UnreadNoteCount int64
	}{}
	if err := db.Raw(raw, args...).Scan(&result).Error; err != nil {
		return nil, err
	}

	counts := map[string]int64{}
	for _, r := range result {
		counts[r.ModelName+"/"+r.ModelKeys] = r.UnreadNoteCount
	}
	return counts, nil
}
