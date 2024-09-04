package activity

import (
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
		reflect.TypeOf(time.Time{}): func(old, new any, prefixField string) []Diff {
			var oldString, newString string
			if !old.(time.Time).IsZero() {
				oldString = old.(time.Time).Format(timeFormat)
			}
			if !new.(time.Time).IsZero() {
				newString = new.(time.Time).Format(timeFormat)
			}
			if oldString != newString {
				return []Diff{
					{Field: prefixField, Old: oldString, New: newString},
				}
			}
			return []Diff{}
		},
		reflect.TypeOf(media_library.MediaBox{}): func(old, new any, prefixField string) (diffs []Diff) {
			oldMediaBox := old.(media_library.MediaBox)
			newMediaBox := new.(media_library.MediaBox)

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
type TypeHandler func(old, new any, prefixField string) []Diff

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

func (db *DiffBuilder) Diff(old, new any) ([]Diff, error) {
	err := db.diffLoop(reflect.Indirect(reflect.ValueOf(old)), reflect.Indirect(reflect.ValueOf(new)), "")
	return db.diffs, err
}

func (db *DiffBuilder) diffLoop(old, new reflect.Value, prefixField string) error {
	if old.Type() != new.Type() {
		return fmt.Errorf("old and new type mismatch: %v != %v", old.Type(), new.Type())
	}

	handleNil := func() bool {
		if old.IsNil() && new.IsNil() {
			return true
		}

		if old.IsNil() && !new.IsNil() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: "", New: fmt.Sprintf("%+v", new.Interface())})
			return true
		}

		if !old.IsNil() && new.IsNil() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%+v", old.Interface()), New: ""})
			return true
		}
		return false
	}

	switch new.Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Uintptr:
		return nil
	case reflect.Interface, reflect.Ptr:
		if handleNil() {
			return nil
		}
		return db.diffLoop(old.Elem(), new.Elem(), prefixField)
	case reflect.Struct:
		for i := 0; i < new.Type().NumField(); i++ {
			if !old.Field(i).CanInterface() {
				continue
			}

			field := new.Type().Field(i)

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
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), new.Field(i).Interface(), newPrefixField)...)
				continue
			}

			if f := db.mb.typeHandlers[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), new.Field(i).Interface(), newPrefixField)...)
				continue
			}
			err := db.diffLoop(old.Field(i), new.Field(i), newPrefixField)
			if err != nil {
				return err
			}

		}
	case reflect.Array, reflect.Slice:
		if new.Kind() == reflect.Slice {
			if handleNil() {
				return nil
			}
		}

		var (
			oldLen  = old.Len()
			newLen  = new.Len()
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
			err := db.diffLoop(old.Index(i), new.Index(i), newPrefixField)
			if err != nil {
				return err
			}
		}

		if added {
			for i := minLen; i < newLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", New: fmt.Sprintf("%+v", new.Index(i).Interface())})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.Index(i).Interface()), New: ""})
			}
		}
	case reflect.Map:
		if handleNil() {
			return nil
		}

		var (
			oldKeys     = old.MapKeys()
			newKeys     = new.MapKeys()
			sameKeys    = []reflect.Value{}
			addedKeys   = []reflect.Value{}
			deletedKeys = []reflect.Value{}
		)

		for _, oldKey := range oldKeys {
			if new.MapIndex(oldKey).IsValid() {
				sameKeys = append(sameKeys, oldKey)
			} else {
				deletedKeys = append(deletedKeys, oldKey)
			}
		}

		for _, newKey := range newKeys {
			if !old.MapIndex(newKey).IsValid() {
				addedKeys = append(addedKeys, newKey)
			}
		}

		for _, key := range sameKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			err := db.diffLoop(old.MapIndex(key), new.MapIndex(key), newPrefixField)
			if err != nil {
				return err
			}
		}

		for _, key := range addedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", New: fmt.Sprintf("%+v", new.MapIndex(key).Interface())})
		}

		for _, key := range deletedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.MapIndex(key).Interface()), New: ""})
		}
	default:
		if old.Interface() != new.Interface() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Interface()), New: fmt.Sprintf("%v", new.Interface())})
		}
	}
	return nil
}

func formatFieldByDot(prefix string, suffix string) string {
	if prefix == "" {
		return suffix
	}
	if suffix == "" {
		return prefix
	}
	return prefix + "." + suffix
}
