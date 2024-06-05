package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ListContentWithImage struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Items ImageListItems
}

type ImageListItems []*ImageListItem

type ImageListItem struct {
	Image      media_library.MediaBox `sql:"type:text;"`
	Link       string
	Heading    string
	Subheading string
	Text       string
}

func (*ListContentWithImage) TableName() string {
	return "container_list_content_with_image"
}

func (this ImageListItems) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *ImageListItems) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func RegisterListContentWithImageContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("ListContentWithImage").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ListContentWithImage)
			return ListContentWithImageBody(v, input)
		})
	mb := vb.Model(&ListContentWithImage{})
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Items")

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&ImageListItem{}).
		Only("Image", "Link", "Heading", "Subheading", "Text")

	eb.Field("Items").Nested(fb)
}

func ListContentWithImageBody(data *ListContentWithImage, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		data.AnchorID, "container-list_content_with_image",
		"", "", "",
		"", data.AddTopSpace, data.AddBottomSpace, "",
		Div(
			ImageListItemsBody(data.Items),
		).Class("container-wrapper"),
	)
	return
}

func ImageListItemsBody(items []*ImageListItem) HTMLComponent {
	listItemsDiv := Div().Class("container-list_content_with_image-inner")
	for _, i := range items {
		listItemsDiv.AppendChildren(
			Div(
				If(i.Link != "", A().Class("container-list_content_with_image-link").Href(i.Link)),
				Div().Class("container-list_content_with_image-image").Style(fmt.Sprintf("background-image: url(%s)", i.Image.URL())),
				If(i.Heading != "" || i.Subheading != "" || i.Text != "",
					Div(
						If(i.Heading != "", H3(i.Heading).Class("container-list_content_with_image-heading")),
						If(i.Subheading != "", Div(Text(i.Subheading)).Class("container-list_content_with_image-subheading h5")),
						If(i.Text != "", P(Text(i.Text)).Class("container-list_content_with_image-text")),
					).Class("container-list_content_with_image-content"),
				),
			).Class("container-list_content_with_image-item"),
		)
	}
	return listItemsDiv
}
