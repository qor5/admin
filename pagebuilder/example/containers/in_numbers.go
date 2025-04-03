package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/presets"

	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
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
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Heading", "Items").ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		p := obj.(*InNumbers)
		for i, v := range p.Items {
			if v == nil {
				continue
			}
			if v.Heading == "" {
				err.FieldError(fmt.Sprintf("Items[%v].Heading", i), "Heading can`t Empty")
			}
		}
		return
	})

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&InNumbersItem{}).Only("Heading", "Text")
	eb.Field("Items").Nested(fb, &presets.DisplayFieldInSorter{Field: "Heading"})
}

func InNumbersBody(data *InNumbers, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		data.AnchorID, "container-in_numbers container-corner",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, "",
		Div(
			H2(data.Heading).Class("container-in_numbers-heading"),
			InNumbersItemsBody(data.Items),
		).Class("container-wrapper"),
		Script(`window.addEventListener("message",(event)=>{console.log(event.data)},false)`),
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
