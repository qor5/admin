package utils

import (
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

// ParseTagOption parse tag options to hash
func ParseTagOption(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}

func GetObjectName(obj interface{}) string {
	modelType := reflect.TypeOf(obj)
	modelstr := modelType.String()
	modelName := modelstr[strings.LastIndex(modelstr, ".")+1:]
	return inflection.Plural(strcase.ToKebab(modelName))
}
