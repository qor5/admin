package exchange

import (
	"strings"

	"github.com/iancoleman/strcase"
)

type Meta struct {
	field        string
	snakeField   string
	columnHeader string
	primaryKey   bool

	setter func(record interface{}, metaValues MetaValues) error
	valuer func(record interface{}) (string, error)
}

func NewMeta(field string) *Meta {
	field = strings.TrimSpace(field)
	return &Meta{
		field:        field,
		columnHeader: field,
		snakeField:   strcase.ToSnake(field),
	}
}

// default is field name
func (m *Meta) Header(s string) *Meta {
	m.columnHeader = strings.TrimSpace(s)
	return m
}

func (m *Meta) PrimaryKey(b bool) *Meta {
	m.primaryKey = b
	return m
}

// set values to special fields
// e.g. time.Time, media, associated records
func (m *Meta) Setter(f func(record interface{}, metaValues MetaValues) error) *Meta {
	m.setter = f
	return m
}

// format values when exporting
func (m *Meta) Valuer(f func(record interface{}) (string, error)) *Meta {
	m.valuer = f
	return m
}
