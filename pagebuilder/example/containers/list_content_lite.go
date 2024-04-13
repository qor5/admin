package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/richeditor"
	v "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ListContentLite struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Items           ListItemLites
	BackgroundColor string
}

type ListItemLites []*ListItemLite

type ListItemLite struct {
	Heading string
	Text    string
}

func (this ListItemLites) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *ListItemLites) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func (*ListContentLite) TableName() string {
	return "container_list_content_lite"
}

func RegisterListContentLiteContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("ListContentLite").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ListContentLite)
			return ListContentLiteBody(v, input)
		})
	mb := vb.Model(&ListContentLite{})

	eb := mb.Editing(
		"AddTopSpace", "AddBottomSpace", "AnchorID",
		"Items", "BackgroundColor",
	)

	eb.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return v.VAutocomplete().
			Attr(web.VField(field.Name, field.Value(obj))...).
			Variant(v.FieldVariantUnderlined).
			Label(field.Label).
			Items([]string{White, Grey})
	})

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&ListItemLite{}).Only("Heading", "Text")
	fb.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return richeditor.RichEditor(db, field.FormKey).
			Plugins([]string{"alignment", "video", "imageinsert", "fontcolor"}).
			Value(obj.(*ListItemLite).Text).Label(field.Label)
	})
	eb.Field("Items").Nested(fb, &presets.DisplayFieldInSorter{Field: "Heading"})
}

func ListContentLiteBody(data *ListContentLite, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("ListContentLite"))+"_%v", data.ID), data.AnchorID, "container-list_content_lite",
		data.BackgroundColor, "", "",
		"", data.AddTopSpace, data.AddBottomSpace, input.IsEditor, input.IsReadonly, "",
		Div(LiteItemsBody(data.Items)).Class("container-wrapper"),
	)
	return
}

func LiteItemsBody(items []*ListItemLite) HTMLComponent {
	itemsDiv := Div().Class("container-list_content_lite-grid")
	for _, i := range items {
		itemsDiv.AppendChildren(
			Div(
				H3(i.Heading).Class("container-list_content_lite-heading"),
				Div(
					RawHTML(i.Text),
				).Class("container-list_content_lite-text"),
			).Class("container-list_content_lite-item"),
		)
	}
	return itemsDiv
}
