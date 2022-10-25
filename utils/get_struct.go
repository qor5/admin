package utils

import "reflect"

func GetStruct(t reflect.Type) interface{} {
	if t.Kind() == reflect.Struct {
		return reflect.New(t).Interface()
	}
	return GetStruct(t.Elem())
}
