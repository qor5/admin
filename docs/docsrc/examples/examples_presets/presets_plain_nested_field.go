package examples_presets

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

type PlainNestedBody struct {
	gorm.Model
	Name  string
	Items NumberCards
}

type NumberCards []*NumberCard

type NumberCard struct {
	Name   string
	Number string
}

func (n NumberCards) Value() (driver.Value, error) {
	return json.Marshal(n)
}

func (n *NumberCards) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), n)
	case []byte:
		return json.Unmarshal(v, n)
	default:
		return errors.New("not supported")
	}
}

func PresetsPlainNestedFieldStruct(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(
		&PlainNestedBody{},
	)
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&PlainNestedBody{})
	cl = mb.Listing()
	ce = mb.Editing()
	return
}

func PresetsPlainNestedField(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsPlainNestedFieldStruct(b, db)
	ce.Creating()
	ed := mb.Editing("Items")
	fb := b.NewFieldsBuilder(presets.WRITE).Model(&NumberCard{}).Only("Name", "Number")

	ed.Field("Items").Nested(fb)
	return
}
