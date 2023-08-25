package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/ui/vuetify"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ListContent struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Items             ListItems `sql:"type:text;"`
	BackgroundColor   string
	Link              string
	LinkText          string
	LinkDisplayOption string
}

type ListItem struct {
	HeadingIcon string
	Heading     string
	Text        string
	Link        string
	LinkText    string
}

func (*ListContent) TableName() string {
	return "container_list_content"
}

type ListItems []*ListItem

func (this ListItems) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *ListItems) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func RegisterListContentContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("ListContent").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ListContent)
			return ListContentBody(v, input)
		})
	mb := vb.Model(&ListContent{})
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "BackgroundColor", "Items", "Link", "LinkText", "LinkDisplayOption")
	eb.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VSelect().
			Items([]string{"white", "grey"}).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})
	eb.Field("LinkDisplayOption").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VSelect().
			Items([]string{"desktop", "mobile", "all"}).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&ListItem{}).Only("HeadingIcon", "Heading", "Text", "Link", "LinkText")

	eb.Field("Items").Nested(fb, &presets.DisplayFieldInSorter{Field: "Heading"})
}

func ListContentBody(data *ListContent, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("ListContent"))+"_%v", data.ID), data.AnchorID, "container-list_content container-lottie",
		data.BackgroundColor, "", "",
		"", data.AddTopSpace, data.AddBottomSpace, input.IsEditor, input.IsReadonly, "",
		Div(
			ListItemsBody(data.Items, input),
			If(data.LinkText != "" && data.Link != "",
				Div(
					LinkTextWithArrow(data.LinkText, data.Link),
				).Class("container-list_content-link").Attr("data-display", data.LinkDisplayOption),
			),
		).Class("container-wrapper"),
	)
	return
}

func ListItemsBody(items []*ListItem, input *pagebuilder.RenderInput) HTMLComponent {
	itemsDiv := Div().Class("container-list_content-grid")
	for _, i := range items {
		itemsDiv.AppendChildren(
			Div(
				Div(
					If(i.Link != "", A(
						If(i.HeadingIcon != "", Div(RawHTML(i.HeadingIcon)).Class("container-list_content-icon")),
						H3(i.Heading),
					).Class("container-list_content-heading").AttrIf("href", i.Link, i.Link != "")),
					If(i.Link == "", Div(
						If(i.HeadingIcon != "", Div(RawHTML(i.HeadingIcon)).Class("container-list_content-icon")),
						H3(i.Heading),
					).Class("container-list_content-heading")),
					Div(
						P(Text(i.Text)),
						LinkTextWithArrow(i.LinkText, i.Link),
					).Class("container-list_content-content"),
				).Class("container-list_content-inner"),
			).Class("container-list_content-item").AttrIf("data-has-icon", "true", i.HeadingIcon != ""),
		)
	}
	return itemsDiv
}
