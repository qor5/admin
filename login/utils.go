package login

import "reflect"

func underlyingReflectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return underlyingReflectType(t.Elem())
	}
	return t
}
