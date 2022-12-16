package activity

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/qor5/admin/presets"
	"gorm.io/gorm"
)

func findOldWithSlug(obj interface{}, slug string, db *gorm.DB) (interface{}, bool) {
	if slug == "" {
		return findOld(obj, db)
	}

	var (
		objValue = reflect.Indirect(reflect.ValueOf(obj))
		old      = reflect.New(objValue.Type()).Interface()
	)

	if slugger, ok := obj.(interface{ PrimaryColumnValuesBySlug(slug string) [][]string }); ok {
		cs := slugger.PrimaryColumnValuesBySlug(slug)
		for _, cond := range cs {
			db = db.Where(fmt.Sprintf("%s = ?", cond[0]), cond[1])
		}
	} else {
		db = db.Where("id = ?", slug)
	}

	if db.First(old).Error != nil {
		return nil, false
	}

	return old, true
}

func findOld(obj interface{}, db *gorm.DB) (interface{}, bool) {
	var (
		objValue = reflect.Indirect(reflect.ValueOf(obj))
		old      = reflect.New(objValue.Type()).Interface()
		sqls     []string
		vars     []interface{}
	)

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(obj); err != nil {
		return nil, false
	}

	for _, dbName := range stmt.Schema.DBNames {
		if field := stmt.Schema.LookUpField(dbName); field != nil && field.PrimaryKey {
			if value, isZero := field.ValueOf(db.Statement.Context, objValue); !isZero {
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

// getPrimaryKey get primary keys from a model
func getPrimaryKey(t reflect.Type) (keys []string) {
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		if strings.Contains(t.Field(i).Tag.Get("gorm"), "primary") {
			keys = append(keys, t.Field(i).Name)
			continue
		}

		if t.Field(i).Type.Kind() == reflect.Ptr && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type.Elem())...)
		}

		if t.Field(i).Type.Kind() == reflect.Struct && t.Field(i).Anonymous {
			keys = append(keys, getPrimaryKey(t.Field(i).Type)...)
		}
	}
	return
}

func ContextWithCreator(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, CreatorContextKey, name)
}

func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBContextKey, db)
}

func getBasicModel(m interface{}) interface{} {
	if preset, ok := m.(*presets.ModelBuilder); ok {
		return preset.NewModel()
	}

	return m
}
