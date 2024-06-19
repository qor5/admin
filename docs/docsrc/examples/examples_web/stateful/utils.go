package stateful

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
)

func Copy(dst, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func Clone[T any](src T) (T, error) {
	var dst T
	if err := Copy(&dst, src); err != nil {
		return dst, err
	}
	return dst, nil
}

func MustClone[T any](src T) T {
	dst, err := Clone(src)
	if err != nil {
		panic(err)
	}
	return dst
}

func PrettyJSONString(v interface{}) (r string) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	r = string(b)
	return
}

func GetFuncName(f any) string {
	rv := reflect.ValueOf(f)
	if rv.Kind() != reflect.Func {
		panic("not a function")
	}
	funcName := runtime.FuncForPC(rv.Pointer()).Name()
	index := strings.LastIndex(funcName, ".")
	if index != -1 {
		funcName = funcName[index+1:]
	}
	funcName = strings.Split(funcName, "-")[0]
	return funcName
}
