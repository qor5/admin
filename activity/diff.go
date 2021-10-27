package activity

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/qor/qor5/media/media_library"
)

var (
	DefaultIgnoredFields = []string{"ID", "UpdatedAt", "DeletedAt", "CreatedAt"}
	DefaultTypeHandles   = map[reflect.Type]TypeHandle{
		reflect.TypeOf(time.Time{}): func(old, now interface{}, prefixField string) []Diff {
			oldString := old.(time.Time).Format(time.RFC3339)
			nowString := now.(time.Time).Format(time.RFC3339)
			if oldString != nowString {
				return []Diff{
					{Field: prefixField, Old: oldString, Now: nowString},
				}
			}
			return []Diff{}
		},
		reflect.TypeOf(media_library.MediaBox{}): func(old, now interface{}, prefixField string) (diffs []Diff) {
			oldMediaBox := old.(media_library.MediaBox)
			nowMediaBox := now.(media_library.MediaBox)

			if oldMediaBox.Url != nowMediaBox.Url {
				diffs = append(diffs, Diff{Field: fmt.Sprintf("%s.Url", prefixField), Old: oldMediaBox.Url, Now: nowMediaBox.Url})
			}

			if oldMediaBox.Description != nowMediaBox.Description {
				diffs = append(diffs, Diff{Field: fmt.Sprintf("%s.Description", prefixField), Old: oldMediaBox.Description, Now: nowMediaBox.Description})
			}

			if oldMediaBox.VideoLink != nowMediaBox.VideoLink {
				diffs = append(diffs, Diff{Field: fmt.Sprintf("%s.VideoLink", prefixField), Old: oldMediaBox.VideoLink, Now: nowMediaBox.VideoLink})
			}

			return diffs
		},
	}
)

type TypeHandle func(old, now interface{}, prefixField string) []Diff

type Diff struct {
	Field string
	Old   string
	Now   string
}

type DiffBuilder struct {
	mb    *ModelBuilder
	diffs []Diff
}

func NewDiffBuilder(mb *ModelBuilder) *DiffBuilder {
	return &DiffBuilder{
		mb:    mb,
		diffs: []Diff{},
	}
}

func (db *DiffBuilder) Diff(old, now interface{}) ([]Diff, error) {
	err := db.diffLoop(reflect.ValueOf(old), reflect.ValueOf(now), "")
	return db.diffs, err
}

func (db *DiffBuilder) diffLoop(old, now reflect.Value, prefixField string) error {
	if old.Type() != now.Type() {
		return errors.New("the two types are not the same")
	}

	if old.Kind() == reflect.Invalid {
		return errors.New("the kind is invalid")
	}

	switch now.Kind() {
	case reflect.Bool:
		if old.Bool() != now.Bool() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Bool()), Now: fmt.Sprintf("%v", now.Bool())})
		}
	case reflect.String:
		if old.String() != now.String() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: old.String(), Now: now.String()})
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if old.Int() != now.Int() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Int()), Now: fmt.Sprintf("%v", now.Int())})
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if old.Uint() != now.Uint() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Uint()), Now: fmt.Sprintf("%v", now.Uint())})
		}
	case reflect.Float64, reflect.Float32:
		if old.Float() != now.Float() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Float()), Now: fmt.Sprintf("%v", now.Float())})
		}
	case reflect.Complex128, reflect.Complex64:
		if old.Complex() != now.Complex() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Complex()), Now: fmt.Sprintf("%v", now.Complex())})
		}
	case reflect.Ptr:
		return db.diffLoop(reflect.Indirect(old), reflect.Indirect(now), prefixField)
	case reflect.Struct:
		for i := 0; i < now.Type().NumField(); i++ {
			field := now.Type().Field(i)

			var needContinue bool
			for _, ignoredField := range DefaultIgnoredFields {
				if ignoredField == field.Name {
					needContinue = true
				}
			}

			for _, ignoredField := range db.mb.ignoredFields {
				if ignoredField == field.Name {
					needContinue = true
				}
			}

			if needContinue {
				continue
			}

			newPrefixField := fmt.Sprintf("%s.%s", prefixField, field.Name)
			if f := DefaultTypeHandles[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), now.Field(i).Interface(), newPrefixField)...)
				continue
			}

			if f := db.mb.typeHanders[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), now.Field(i).Interface(), newPrefixField)...)
				continue
			}
			err := db.diffLoop(old.Field(i), now.Field(i), newPrefixField)
			if err != nil {
				return err
			}

		}
	case reflect.Array, reflect.Slice:
		var (
			oldLen  = old.Len()
			nowLen  = now.Len()
			minLen  int
			added   bool
			deleted bool
		)

		if oldLen > nowLen {
			minLen = nowLen
			deleted = true
		}

		if oldLen < nowLen {
			minLen = oldLen
			added = true
		}

		if oldLen == nowLen {
			minLen = oldLen
		}

		for i := 0; i < minLen; i++ {
			newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
			err := db.diffLoop(old.Index(i), now.Index(i), newPrefixField)
			if err != nil {
				return err
			}
		}

		if added {
			for i := minLen; i < nowLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", Now: fmt.Sprintf("%+v", now.Index(i).Interface())})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := fmt.Sprintf("%s.%d", prefixField, i)
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.Index(i).Interface()), Now: ""})
			}
		}
	case reflect.Map:
		var (
			oldKeys = old.MapKeys()
			newKeys = now.MapKeys()

			sameKeys    = []reflect.Value{}
			addedKeys   = []reflect.Value{}
			deletedKeys = []reflect.Value{}
		)

		for _, oldKey := range oldKeys {
			var find bool
			for _, newKey := range newKeys {
				if oldKey.Interface() == newKey.Interface() {
					find = true
				}
			}
			if find {
				sameKeys = append(sameKeys, oldKey)
			}
			if !find {
				deletedKeys = append(deletedKeys, oldKey)
			}
		}

		for _, newKey := range newKeys {
			var find bool
			for _, oldKey := range oldKeys {
				if oldKey.Interface() == newKey.Interface() {
					find = true
				}
			}
			if !find {
				addedKeys = append(addedKeys, newKey)
			}
		}

		for _, key := range sameKeys {
			newPrefixField := fmt.Sprintf("%s.%v", prefixField, key)
			err := db.diffLoop(old.MapIndex(key), now.MapIndex(key), newPrefixField)
			if err != nil {
				return err
			}
		}

		for _, key := range addedKeys {
			newPrefixField := fmt.Sprintf("%s.%v", prefixField, key)
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", Now: fmt.Sprintf("%+v", now.MapIndex(key).Interface())})
		}

		for _, key := range deletedKeys {
			newPrefixField := fmt.Sprintf("%s.%v", prefixField, key)
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.MapIndex(key).Interface()), Now: ""})
		}
	}

	return nil
}
