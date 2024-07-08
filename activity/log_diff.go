package activity

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/qor5/admin/v3/media/media_library"
)

var (
	// @snippet_begin(ActivityDefaultIgnoredFields)
	DefaultIgnoredFields = []string{"ID", "UpdatedAt", "DeletedAt", "CreatedAt"}
	// @snippet_end

	// @snippet_begin(ActivityDefaultTypeHandles)
	DefaultTypeHandles = map[reflect.Type]TypeHandler{
		reflect.TypeOf(time.Time{}): func(old, now any, prefixField string) []Diff {
			oldString := old.(time.Time).Format(time.RFC3339)
			nowString := now.(time.Time).Format(time.RFC3339)
			if oldString != nowString {
				return []Diff{
					// TODO: 对于空的时间，是否应该给到空？再者这里是否能处理 *time.Time
					{Field: prefixField, Old: oldString, Now: nowString},
				}
			}
			return []Diff{}
		},
		reflect.TypeOf(media_library.MediaBox{}): func(old, now any, prefixField string) (diffs []Diff) {
			oldMediaBox := old.(media_library.MediaBox)
			nowMediaBox := now.(media_library.MediaBox)

			if oldMediaBox.Url != nowMediaBox.Url {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "Url"), Old: oldMediaBox.Url, Now: nowMediaBox.Url})
			}

			if oldMediaBox.Description != nowMediaBox.Description {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "Description"), Old: oldMediaBox.Description, Now: nowMediaBox.Description})
			}

			if oldMediaBox.VideoLink != nowMediaBox.VideoLink {
				diffs = append(diffs, Diff{Field: formatFieldByDot(prefixField, "VideoLink"), Old: oldMediaBox.VideoLink, Now: nowMediaBox.VideoLink})
			}

			return diffs
		},
	}
	// @snippet_end
)

// @snippet_begin(ActivityTypeHandle)
type TypeHandler func(old, now any, prefixField string) []Diff

// @snippet_end

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
		mb: mb,
	}
}

func (db *DiffBuilder) Diff(old, now any) ([]Diff, error) {
	err := db.diffLoop(reflect.Indirect(reflect.ValueOf(old)), reflect.Indirect(reflect.ValueOf(now)), "")
	return db.diffs, err
}

func (db *DiffBuilder) diffLoop(old, now reflect.Value, prefixField string) error {
	if old.Type() != now.Type() {
		return fmt.Errorf("old and now type mismatch: %v != %v", old.Type(), now.Type())
	}

	handleNil := func() bool {
		if old.IsNil() && now.IsNil() {
			return true
		}

		if old.IsNil() && !now.IsNil() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: "", Now: fmt.Sprintf("%+v", now.Interface())})
			return true
		}

		if !old.IsNil() && now.IsNil() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%+v", old.Interface()), Now: ""})
			return true
		}
		return false
	}

	switch now.Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Uintptr:
		return nil
	case reflect.Interface, reflect.Ptr:
		if handleNil() {
			return nil
		}
		return db.diffLoop(old.Elem(), now.Elem(), prefixField)
	case reflect.Struct:
		for i := 0; i < now.Type().NumField(); i++ {
			if !old.Field(i).CanInterface() {
				continue
			}

			field := now.Type().Field(i)

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
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), now.Field(i).Interface(), newPrefixField)...)
				continue
			}

			if f := db.mb.typeHandlers[field.Type]; f != nil {
				db.diffs = append(db.diffs, f(old.Field(i).Interface(), now.Field(i).Interface(), newPrefixField)...)
				continue
			}
			err := db.diffLoop(old.Field(i), now.Field(i), newPrefixField)
			if err != nil {
				return err
			}

		}
	case reflect.Array, reflect.Slice:
		if now.Kind() == reflect.Slice {
			if handleNil() {
				return nil
			}
		}

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
			newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
			err := db.diffLoop(old.Index(i), now.Index(i), newPrefixField)
			if err != nil {
				return err
			}
		}

		if added {
			for i := minLen; i < nowLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", Now: fmt.Sprintf("%+v", now.Index(i).Interface())})
			}
		}

		if deleted {
			for i := minLen; i < oldLen; i++ {
				newPrefixField := formatFieldByDot(prefixField, strconv.Itoa(i))
				db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.Index(i).Interface()), Now: ""})
			}
		}
	case reflect.Map:
		if handleNil() {
			return nil
		}

		var (
			oldKeys     = old.MapKeys()
			newKeys     = now.MapKeys()
			sameKeys    = []reflect.Value{}
			addedKeys   = []reflect.Value{}
			deletedKeys = []reflect.Value{}
		)

		// TODO: 既然是字典，还使用这种通过循环来判定的方式有点怪怪的
		for _, oldKey := range oldKeys {
			var find bool
			for _, newKey := range newKeys {
				if oldKey.Interface() == newKey.Interface() {
					find = true
					break
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
					break
				}
			}
			if !find {
				addedKeys = append(addedKeys, newKey)
			}
		}

		for _, key := range sameKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			err := db.diffLoop(old.MapIndex(key), now.MapIndex(key), newPrefixField)
			if err != nil {
				return err
			}
		}

		for _, key := range addedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: "", Now: fmt.Sprintf("%+v", now.MapIndex(key).Interface())})
		}

		for _, key := range deletedKeys {
			newPrefixField := formatFieldByDot(prefixField, key.String())
			db.diffs = append(db.diffs, Diff{Field: newPrefixField, Old: fmt.Sprintf("%+v", old.MapIndex(key).Interface()), Now: ""})
		}
	default:
		if old.Interface() != now.Interface() {
			db.diffs = append(db.diffs, Diff{Field: prefixField, Old: fmt.Sprintf("%v", old.Interface()), Now: fmt.Sprintf("%v", now.Interface())})
		}
	}
	return nil
}

func formatFieldByDot(prefix string, suffix string) string {
	if len(prefix) == 0 {
		return suffix
	}
	if len(suffix) == 0 {
		return prefix
	}
	return prefix + "." + suffix
}
