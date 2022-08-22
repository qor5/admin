package login

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/sunfmin/reflectutils"
)

type PrimaryFielder interface {
	PrimaryField() string
}

func primaryField(obj interface{}) string {
	f := "ID"
	if v, ok := obj.(PrimaryFielder); ok {
		f = v.PrimaryField()
	}
	return f
}

func snakePrimaryField(obj interface{}) string {
	return strcase.ToSnake(primaryField(obj))
}

func objectID(obj interface{}) string {
	return fmt.Sprint(reflectutils.MustGet(obj, primaryField(obj)))
}
