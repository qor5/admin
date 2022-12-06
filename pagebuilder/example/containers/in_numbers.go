package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"

	"github.com/qor5/admin/presets"
	"github.com/qor5/web"

	"github.com/qor5/admin/pagebuilder"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type InNumbers struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Heading string
	Items   InNumbersItems
}

type InNumbersItem struct {
	Heading string
	Text    string
}

func (*InNumbers) TableName() string {
	return "container_in_numbers"
}

type InNumbersItems []*InNumbersItem

func (this InNumbersItems) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *InNumbersItems) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func RegisterInNumbersContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("InNumbers").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*InNumbers)
			return InNumbersBody(v, input)
		})
	mb := vb.Model(&InNumbers{})
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Heading", "Items")

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&InNumbersItem{}).Only("Heading", "Text")

	eb.Field("Items").Nested(fb, &presets.DisplayFieldInSorter{Field: "Heading"})

}

func InNumbersBody(data *InNumbers, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("InNumbers"))+"_%v", data.ID), data.AnchorID, "container-in_numbers container-corner",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, input.IsEditor, "",
		Div(
			H2(data.Heading).Class("container-in_numbers-heading"),
			InNumbersItemsBody(data.Items),
		).Class("container-wrapper"),
	)
	return
}

func InNumbersItemsBody(items []*InNumbersItem) HTMLComponent {
	inNumbersItemsDiv := Div().Class("container-in_numbers-grid")
	for _, i := range items {
		inNumbersItemsDiv.AppendChildren(
			Div(
				Div(
					H2(i.Heading).Class("container-in_numbers-item-title"),
					Div(Text(i.Text)).Class("container-in_numbers-item-description"),
				).Class("container-in_numbers-inner"),
			).Class("container-in_numbers-item"),
		)
	}
	return inNumbersItemsDiv
}
