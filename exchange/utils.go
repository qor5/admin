package exchange

import (
	"errors"
	"reflect"
	"strconv"
)

func setValueFromString(v reflect.Value, strVal string) error {
	if strVal == "" {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(strVal, 0, 64)
		if err != nil {
			return err
		}
		if v.OverflowInt(val) {
			return errors.New("Int value too big: " + strVal)
		}
		v.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(strVal, 0, 64)
		if err != nil {
			return err
		}
		if v.OverflowUint(val) {
			return errors.New("UInt value too big: " + strVal)
		}
		v.SetUint(val)
	case reflect.Float32:
		val, err := strconv.ParseFloat(strVal, 32)
		if err != nil {
			return err
		}
		v.SetFloat(val)
	case reflect.Float64:
		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return err
		}
		v.SetFloat(val)
	case reflect.String:
		v.SetString(strVal)
	case reflect.Bool:
		val, err := strconv.ParseBool(strVal)
		if err != nil {
			return err
		}
		v.SetBool(val)
	case reflect.Ptr:
		v.Set(reflect.New(v.Type().Elem()))
		return setValueFromString(v.Elem(), strVal)
	default:
		return errors.New("Unsupported kind: " + v.Kind().String())
	}
	return nil
}

func validateResourceAndMetas(r interface{}, metas []*Meta) error {
	if r == nil {
		return errors.New("resource is nil")
	}
	rt := reflect.TypeOf(r)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		return errors.New("resource is not ptr to struct")
	}

	if len(metas) == 0 {
		return errors.New("no metas")
	}

	hasPrimaryKey := false
	for i, _ := range metas {
		m := metas[i]
		if m.field == "" {
			return errors.New("field name is empty")
		}
		if m.columnHeader == "" {
			return errors.New("header is empty")
		}
		if m.primaryKey {
			hasPrimaryKey = true
		}
	}
	if !hasPrimaryKey {
		return errors.New("must have primary key field meta")
	}

	return nil
}
