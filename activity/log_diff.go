package activity

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/qor5/admin/v3/media/media_library"
)

var (
	timeFormat = "2006-01-02 15:04:05 MST"

	// @snippet_begin(ActivityDefaultIgnoredFields)
	DefaultIgnoredFields = []string{"ID", "UpdatedAt", "DeletedAt", "CreatedAt"}
	// @snippet_end

	// @snippet_begin(ActivityDefaultTypeHandles)
	DefaultTypeHandles = map[reflect.Type]TypeHandler{
		reflect.TypeOf(time.Time{}): func(oldObj, newObj any, prefixField string) []Diff {
			var oldString, newString string
			if !oldObj.(time.Time).IsZero() {
				oldString = oldObj.(time.Time).Format(timeFormat)
			}
			if !newObj.(time.Time).IsZero() {
				newString = newObj.(time.Time).Format(timeFormat)
			}
			if oldString != newString {
				return []Diff{
					{Field: prefixField, Old: oldString, New: newString},
				}
			}
			return []Diff{}
		},
		reflect.TypeOf(media_library.MediaBox{}): func(oldObj, newObj any, prefixField string) (diffs []Diff) {
			oldMediaBox := oldObj.(media_library.MediaBox)
			newMediaBox := newObj.(media_library.MediaBox)

			if oldMediaBox.Url != newMediaBox.Url {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "Url"), Old: oldMediaBox.Url, New: newMediaBox.Url})
			}

			if oldMediaBox.Description != newMediaBox.Description {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "Description"), Old: oldMediaBox.Description, New: newMediaBox.Description})
			}

			if oldMediaBox.VideoLink != newMediaBox.VideoLink {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "VideoLink"), Old: oldMediaBox.VideoLink, New: newMediaBox.VideoLink})
			}

			return diffs
		},
	}
	// @snippet_end
)

// @snippet_begin(ActivityTypeHandle)
type TypeHandler func(oldObj, newObj any, prefixField string) []Diff

// @snippet_end

type Diff struct {
	Field string
	Old   string
	New   string
}

type DiffBuilder struct {
	mb    *ModelBuilder
	diffs []Diff
}

func NewDiffBuilder(mb *ModelBuilder) *DiffBuilder {
	return &DiffBuilder{
		mb: mb,
	}
}

func (db *DiffBuilder) Diff(oldObj, newObj any) ([]Diff, error) {
	err := db.diffLoop(reflect.Indirect(reflect.ValueOf(oldObj)), reflect.Indirect(reflect.ValueOf(newObj)), "")
	return db.diffs, err
}

func (db *DiffBuilder) diffLoop(oldObj, newObj reflect.Value, prefixField string) error {
	if oldObj.Type() != newObj.Type() {
		return fmt.Errorf("old and new type mismatch: %v != %v", oldObj.Type(), newObj.Type())
	}

	handleNil := func() bool {
		if oldObj.IsNil() && newObj.IsNil() {
			return true
		}

		if oldObj.IsNil() && !newObj.IsNil() {
			newVal, _ := json.Marshal(newObj.Interface())

			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: "", New: string(newVal)})
			return true
		}

		if !oldObj.IsNil() && newObj.IsNil() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%+v", oldObj.Interface()), New: ""})
			return true
		}
		return false
	}

	switch newObj.Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Uintptr:
		return nil
	case reflect.Interface, reflect.Ptr:
		if handleNil() {
			return nil
		}
		return db.diffLoop(oldObj.Elem(), newObj.Elem(), prefixField)
	case reflect.Struct:
		for i := 0; i < newObj.Type().NumField(); i++ {
			if !oldObj.Field(i).CanInterface() {
				continue
			}

			field := newObj.Type().Field(i)

			var needContinue bool
			for _, ignoredField := range DefaultIgnoredFields {
				if ignoredField == field.Name {
					needContinue = true
					continue
				}
			}

			for _, ignoredField := range db.mb.ignoredFields {
				if ignoredField == field.Name {
					needContinue = true
					continue
				}
			}

			if needContinue {
				continue
			}

			newPrefixField := formatFieldByDot(prefixField, field.Name)
			if f := DefaultTypeHandles[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(oldObj.Field(i).Interface(), newObj.Field(i).Interface(), newPrefixField)...)
				continue
			}

			if f := db.mb.typeHandlers[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(oldObj.Field(i).Interface(), newObj.Field(i).Interface(), newPrefixField)...)
				continue
			}
			err := db.diffLoop(oldObj.Field(i), newObj.Field(i), newPrefixField)
			if err != nil {
				return err
			}

		}
	case reflect.Array, reflect.Slice:
		if newObj.Kind() == reflect.Slice {
			if handleNil() {
				return nil
			}
		}

		var (
			oldLen  = oldObj.Len()
			newLen  = newObj.Len()
			minLen  int
			added   bool
			deleted bool
		)

		if oldLen > newLen {
			minLen = newLen
			deleted = true
		}

		if oldLen < newLen {
			minLen = oldLen
			added = true
		}

		if oldLen == newLen {
			minLen = oldLen
		}

		for i := 0; i < minLen; i++ {
			newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
			err := db.diffLoop(oldObj.Index(i), newObj.Index(i), newPrefixField)
			if err != nil {
				return err
			}
		}

		if added {
			for i := minLen; i < newLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", New: fmt.Sprintf("%+v", newObj.Index(i).Interface())})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", oldObj.Index(i).Interface()), New: ""})
			}
		}
	case reflect.Map:
		if handleNil() {
			return nil
		}

		var (
			oldKeys     = oldObj.MapKeys()
			newKeys     = newObj.MapKeys()
			sameKeys    = []reflect.Value{}
			addedKeys   = []reflect.Value{}
			deletedKeys = []reflect.Value{}
		)

		for _, oldKey := range oldKeys {
			if newObj.MapIndex(oldKey).IsValid() {
				sameKeys = append(sameKeys, oldKey)
			} else {
				deletedKeys = append(deletedKeys, oldKey)
			}
		}

		for _, newKey := range newKeys {
			if !oldObj.MapIndex(newKey).IsValid() {
				addedKeys = append(addedKeys, newKey)
			}
		}

		for _, key := range sameKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			err := db.diffLoop(oldObj.MapIndex(key), newObj.MapIndex(key), newPrefixField)
			if err != nil {
				return err
			}
		}

		for _, key := range addedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", New: fmt.Sprintf("%+v", newObj.MapIndex(key).Interface())})
		}

		for _, key := range deletedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", oldObj.MapIndex(key).Interface()), New: ""})
		}
	default:
		if oldObj.Interface() != newObj.Interface() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", oldObj.Interface()), New: fmt.Sprintf("%v", newObj.Interface())})
		}
	}
	return nil
}

func formatFieldByDot(prefix, suffix string) string {
	if prefix == "" {
		return suffix
	}
	if suffix == "" {
		return prefix
	}
	return prefix + "." + suffix
}
