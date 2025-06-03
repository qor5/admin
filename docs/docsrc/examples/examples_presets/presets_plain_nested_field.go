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
	if value == nil {
		// Insert default data when value is nil
		*n = NumberCards{&NumberCard{Name: "Default", Number: "0"}}
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			*n = NumberCards{&NumberCard{Name: "Default", Number: "0"}}
			return nil
		}
		if err := json.Unmarshal([]byte(v), n); err != nil {
			return err
		}
	case []byte:
		if len(v) == 0 {
			*n = NumberCards{&NumberCard{Name: "Default", Number: "0"}}
			return nil
		}
		if err := json.Unmarshal(v, n); err != nil {
			return err
		}
	default:
		return errors.New("not supported")
	}

	// If after unmarshaling, the slice is nil or empty, insert default data
	if *n == nil || len(*n) == 0 {
		*n = NumberCards{&NumberCard{Name: "Default", Number: "0"}}
	}
	return nil
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

	// 修改嵌套字段处理
	fb := b.NewFieldsBuilder(presets.WRITE).Model(&NumberCard{}).Only("Name", "Number")
	ce.Field("Items").Nested(fb).PlainFieldBody().HideLabel()

	return
}
