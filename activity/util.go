package activity

import (
	"cmp"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func FirstUpperWord(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(string([]rune(name)[0:1]))
}

func ParseModelName(v any) string {
	segs := strings.Split(reflect.TypeOf(v).String(), ".")
	return strings.TrimLeft(segs[len(segs)-1], "*")
}

func KeysValue(v any, keys []string, sep string) string {
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return ""
	}
	vals := []string{}
	for _, key := range keys {
		rv := rv
	while:
		for {
			rt := rv.Type()
			rtField, ok := rt.FieldByName(key)
			if !ok {
				break while
			}
			rvField := rv.FieldByName(key)
			if !rvField.IsValid() {
				break while
			}
			if rtField.Anonymous {
				rvField = reflect.Indirect(rvField)
				rv = rvField
				continue
			}

			vals = append(vals, fmt.Sprint(rvField.Interface()))
			break while
		}
	}
	return strings.Join(vals, sep)
}

var (
	schemaParserCacheStore sync.Map
	schemaParserNamer      = schema.NamingStrategy{IdentifierMaxLength: 64}
)

func ParseSchema(v any) (*schema.Schema, error) {
	s, err := schema.Parse(v, &schemaParserCacheStore, schemaParserNamer)
	if err != nil {
		return nil, errors.Wrap(err, "parse schema")
	}
	return s, nil
}

func ParseSchemaWithDB(db *gorm.DB, v any) (*schema.Schema, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(v); err != nil {
		return nil, errors.Wrap(err, "parse schema with db")
	}
	return stmt.Schema, nil
}

func parsePrimaryKeys(t reflect.Type) (keys []string) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primary") {
			keys = append(keys, t.Field(i).Name)
			continue
		}

		if t.Field(i).Type.Kind() == reflect.Ptr && t.Field(i).Anonymous {
			keys = append(keys, parsePrimaryKeys(t.Field(i).Type.Elem())...)
		}

		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Anonymous {
			keys = append(keys, parsePrimaryKeys(t.Field(i).Type)...)
		}
	}
	return
}

func ParsePrimaryKeys(v any) []string {
	s, err := ParseSchema(v)
	if err == nil {
		return lo.Map(s.PrimaryFields, func(f *schema.Field, _ int) string { return f.Name })
	}
	// parsePrimaryKeys is more compatible if some of the model's fields do not obey sql.Driver very well
	return parsePrimaryKeys(reflect.Indirect(reflect.ValueOf(v)).Type())
}

const dbKeyTablePrefix = "__table_prefix__"

// ScopeWithTablePrefix set table prefix
// 1. Only scenarios where a Model is provided are supported
// 2. Previously Table(...) will be overwritten
func ScopeWithTablePrefix(tablePrefix string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if v, ok := db.Get(dbKeyTablePrefix); ok {
			if v.(string) != tablePrefix {
				panic(fmt.Sprintf("table prefix is already set to %s", v))
			} else {
				return db
			}
		}

		if tablePrefix == "" {
			return db
		}

		stmt := db.Statement
		model := cmp.Or(stmt.Model, stmt.Dest)
		if model == nil {
			return db
		}

		s, err := ParseSchemaWithDB(db, model)
		if err != nil {
			db.AddError(err)
			return db
		}
		return db.Set(dbKeyTablePrefix, tablePrefix).Table(tablePrefix + s.Table)
	}
}

const I18nActionLabelPrefix = "ActivityAction"

func getActionLabel(evCtx *web.EventContext, action string) string {
	msgr := i18n.MustGetModuleMessages(evCtx.R, I18nActivityKey, Messages_en_US).(*Messages)
	label := defaultActionLabels(msgr)[action]
	if label == "" {
		label = i18n.PT(evCtx.R, presets.ModelsI18nModuleKey, I18nActionLabelPrefix, action)
	}
	return label
}

func FetchOldWithSlug(db *gorm.DB, ref any, slug string) (any, error) {
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

	if err := db.First(old).Error; err != nil {
		return nil, errors.Wrap(err, "fetch old with slug")
	}

	return old, nil
}

func FetchOld(db *gorm.DB, ref any) (any, error) {
	var (
		rtRef = reflect.Indirect(reflect.ValueOf(ref))
		old   = reflect.New(rtRef.Type()).Interface()
		sqls  []string
		vars  []any
	)

	s, err := ParseSchemaWithDB(db, ref)
	if err != nil {
		return nil, err
	}
	for _, dbName := range s.DBNames {
		if field := s.LookUpField(dbName); field != nil && field.PrimaryKey {
			if value, isZero := field.ValueOf(db.Statement.Context, rtRef); !isZero {
				sqls = append(sqls, fmt.Sprintf("%v = ?", dbName))
				vars = append(vars, value)
			}
		}
	}

	if len(sqls) == 0 || len(vars) == 0 || len(sqls) != len(vars) {
		return nil, errors.New("no primary key found")
	}

	if err := db.Where(strings.Join(sqls, " AND "), vars...).First(old).Error; err != nil {
		return nil, errors.Wrap(err, "fetch old")
	}

	return old, nil
}
