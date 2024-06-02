package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func generateAssertions(prefix string, obj any, ignore func(prefix string) bool) (assertions []string) {
	if ignore != nil && ignore(prefix) {
		return
	}
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			assertions = append(assertions, fmt.Sprintf("assert.Nil(t, %s)", prefix))
			return
		}
		assertions = append(assertions, fmt.Sprintf("assert.NotNil(t, %s)", prefix))
		prefix = "*" + prefix

		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Invalid:
		assertions = append(assertions, fmt.Sprintf("assert.Nil(t, %s)", prefix))
		return
	case reflect.Interface:
		if v.IsNil() {
			assertions = append(assertions, fmt.Sprintf("assert.Nil(t, %s)", prefix))
		} else {
			assertions = append(assertions, fmt.Sprintf("assert.NotNil(t, %s)", prefix))
		}
		return
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.IsNil() || v.Len() == 0 {
			// also use Empty for more fault tolerance if is nil
			assertions = append(assertions, fmt.Sprintf("assert.Empty(t, %s)", prefix))
			return
		}
	}

	switch v.Kind() {
	case reflect.Int:
		assertions = append(assertions, fmt.Sprintf("assert.Equal(t, %v, %s)", v.Interface(), prefix))
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		assertions = append(assertions, fmt.Sprintf("assert.Equal(t, %s(%v), %s)", v.Kind().String(), v.Interface(), prefix))
	case reflect.String:
		if v.String() == "" {
			assertions = append(assertions, fmt.Sprintf("assert.Empty(t, %s)", prefix))
		} else {
			assertions = append(assertions, fmt.Sprintf("assert.Equal(t, %s, %s)", strconv.Quote(v.String()), prefix))
		}
	case reflect.Bool:
		if v.Bool() {
			assertions = append(assertions, fmt.Sprintf("assert.True(t, %s)", prefix))
		} else {
			assertions = append(assertions, fmt.Sprintf("assert.False(t, %s)", prefix))
		}
	case reflect.Slice, reflect.Array:
		prefix = strings.TrimPrefix(prefix, "*")
		assertions = append(assertions, fmt.Sprintf("assert.Len(t, %s, %d)", prefix, v.Len()))
		for j := 0; j < v.Len(); j++ {
			elem := v.Index(j).Elem()
			nestedAssertions := generateAssertions(prefix+fmt.Sprintf("[%d]", j), elem.Interface(), ignore)
			assertions = append(assertions, nestedAssertions...)
		}
	case reflect.Map:
		prefix = strings.TrimPrefix(prefix, "*")
		assertions = append(assertions, fmt.Sprintf("assert.Len(t, %s, %d)", prefix, v.Len()))
		for _, key := range v.MapKeys() {
			elem := v.MapIndex(key)
			nestedAssertions := generateAssertions(fmt.Sprintf("%s[%v]", prefix, key.Interface()), elem.Interface(), ignore)
			assertions = append(assertions, nestedAssertions...)
		}
	case reflect.Struct:
		prefix = strings.TrimPrefix(prefix, "*")
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)
			nestedAssertions := generateAssertions(prefix+"."+field.Name, fieldValue.Interface(), ignore)
			assertions = append(assertions, nestedAssertions...)
		}
	default:
		panic(fmt.Sprintf("unsupport kind %v for %s", v.Kind(), prefix))
	}
	return
}
