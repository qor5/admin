package exchange

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm"
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

	ret := rt.Elem()
	for i, _ := range metas {
		m := metas[i]
		if m.field == "" {
			return errors.New("field name is empty")
		}
		if m.setter == nil && m.valuer == nil {
			_, ok := ret.FieldByName(m.field)
			if !ok {
				return fmt.Errorf("field %s not found", m.field)
			}
		}
		if m.columnHeader == "" {
			return errors.New("header is empty")
		}
		if m.primaryKey && (m.setter != nil || m.valuer != nil) {
			return fmt.Errorf("can not set setter/valuer on primaryKey meta")
		}
	}

	return nil
}

func preloadDB(db *gorm.DB, associations []string) *gorm.DB {
	if len(associations) == 0 {
		return db
	}

	ndb := db.Preload(associations[0])
	for i := 1; i < len(associations); i++ {
		ndb = ndb.Preload(associations[i])
	}
	return ndb
}

func getIndirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}

	return getIndirect(reflect.Indirect(v))
}

func getIndirectStruct(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Struct {
		return t
	}
	return getIndirectStruct(t.Elem())
}

func clearPrimaryKeyValue(v reflect.Value) {
	t := v.Type()
	if idf, ok := t.FieldByName("ID"); ok {
		if strings.Contains(idf.Tag.Get("gorm"), "primarykey") {
			v.FieldByName("ID").SetUint(0)
		}
	}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if !strings.Contains(ft.Tag.Get("gorm"), "primarykey") {
			continue
		}
		v.Field(i).Set(reflect.New(ft.Type).Elem())
	}
}

func getParamsNumbers(n *int, t reflect.Type, associations []string) {
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		isAssociation := false
		for _, a := range associations {
			if ft.Name == a {
				isAssociation = true
				break
			}
		}
		if isAssociation {
			continue
		}
		if ft.Type.Kind() == reflect.Struct && ft.Anonymous {
			getParamsNumbers(n, ft.Type, nil)
			continue
		}
		*n++
	}
}

func splitInterfaceSlice(s []interface{}, size int) [][]interface{} {
	groupsLen := int(math.Ceil(float64(len(s)) / float64(size)))
	groups := make([][]interface{}, groupsLen)

	idx := 0
	for i := 0; i < groupsLen; i++ {
		idx = i * size
		if i != groupsLen-1 {
			groups[i] = s[idx : idx+size]
		} else {
			groups[i] = s[idx:]
		}
	}

	return groups
}

func splitReflectSliceValue(s reflect.Value, size int) []reflect.Value {
	groupsLen := int(math.Ceil(float64(s.Len()) / float64(size)))
	groups := make([]reflect.Value, 0, groupsLen)

	idx := 0
	for i := 0; i < groupsLen; i++ {
		idx = i * size
		endIdx := idx + size
		if i == groupsLen-1 {
			endIdx = s.Len()
		}

		vs := reflect.New(reflect.SliceOf(s.Type().Elem())).Elem()
		for j := idx; j < endIdx; j++ {
			vs = reflect.Append(vs, s.Index(j))
		}
		groups = append(groups, vs)
	}

	return groups
}
