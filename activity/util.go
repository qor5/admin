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
	"gorm.io/gorm/schema"
)

func firstUpperWord(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(string([]rune(name)[0:1]))
}

func modelName(v any) string {
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

func objectID(obj any) string {
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
