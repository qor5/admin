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

func ParsePrimaryFields(v any) ([]*schema.Field, error) {
	s, err := ParseSchema(v)
	if err != nil {
		return nil, err
	}
	return s.PrimaryFields, nil
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
	fields, err := ParsePrimaryFields(v)
	if err == nil {
		return lo.Map(fields, func(f *schema.Field, _ int) string { return f.Name })
	}
	// parsePrimaryKeys is more compatible if some of the model's fields do not obey sql.Driver very well
	return parsePrimaryKeys(reflect.Indirect(reflect.ValueOf(v)).Type())
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
