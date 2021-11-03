package exchange

import (
	"strings"

	"github.com/iancoleman/strcase"
)

type MetaValues interface {
	Get(field string) (val string)
}

type Meta struct {
	field        string
	snakeField   string
	columnHeader string
	primaryKey   bool

	setter MetaSetter
	valuer MetaValuer
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

type MetaSetter func(record interface{}, value string, metaValues MetaValues) error

// set values to special fields
// e.g. time.Time, struct, associated records ...
func (m *Meta) Setter(f MetaSetter) *Meta {
	m.setter = f
	return m
}

type MetaValuer func(record interface{}) (string, error)

// format values when exporting
func (m *Meta) Valuer(f MetaValuer) *Meta {
	m.valuer = f
	return m
}
