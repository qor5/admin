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