package activity

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func findDeletedOldByWhere(db *gorm.DB) (interface{}, bool) {
	var (
		old  = reflect.New(reflect.Indirect(reflect.ValueOf(db.Statement.Dest)).Type()).Interface()
		sqls []string
		vars []interface{}
	)
	if where, ok := db.Statement.Clauses["WHERE"].Expression.(clause.Where); ok {
		for _, e := range where.Exprs {
			if expr, ok := e.(clause.Expr); ok {
				sqls = append(sqls, expr.SQL)
				vars = append(vars, expr.Vars...)
			}
		}
	}

	if len(sqls) == 0 || len(vars) == 0 || len(sqls) != len(vars) {
		return nil, false
	}

	if GlobalDB.Where(strings.Join(sqls, " AND "), vars...).First(old).Error != nil {
		return nil, false
	}
	return old, true
}

func findOld(obj interface{}) (interface{}, bool) {
	var (
		objValue = reflect.Indirect(reflect.ValueOf(obj))
		old      = reflect.New(objValue.Type()).Interface()
		sqls     []string
		vars     []interface{}
	)

	stmt := &gorm.Statement{DB: GlobalDB}
	if err := stmt.Parse(obj); err != nil {
		return nil, false
	}

	for _, dbName := range stmt.Schema.DBNames {
		if field := stmt.Schema.LookUpField(dbName); field != nil && field.PrimaryKey {
			if value, isZero := field.ValueOf(objValue); !isZero {
				sqls = append(sqls, fmt.Sprintf("%v = ?", dbName))
				vars = append(vars, value)
			}
		}
	}

	if len(sqls) == 0 || len(vars) == 0 || len(sqls) != len(vars) {
		return nil, false
	}

	if GlobalDB.Where(strings.Join(sqls, " AND "), vars...).First(old).Error != nil {
		return nil, false
	}

	return old, true
}

func ContextWithCreator(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, CreatorContextKey, name)
}

func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, DBContextKey, db)
}
