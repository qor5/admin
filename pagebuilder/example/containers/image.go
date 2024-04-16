package containers

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ImageContainer struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string

	Image                     media_library.MediaBox `sql:"type:text;"`
	BackgroundColor           string
	TransitionBackgroundColor string
}

func (*ImageContainer) TableName() string {
	return "container_images"
}

func RegisterImageContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("Image").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ImageContainer)
			return ImageContainerBody(v, input)
		})
	mb := vb.Model(&ImageContainer{}).URIName(inflection.Plural(strcase.ToKebab("Image")))
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "BackgroundColor", "TransitionBackgroundColor", "Image")
	eb.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VSelect().
			Items([]string{"white", "blue", "grey"}).
			Variant(vuetify.FieldVariantUnderlined).
			Label(field.Label).
			Attr(web.VField(field.FormKey, field.Value(obj))...)
	})
	eb.Field("TransitionBackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VSelect().
			Items([]string{"white", "blue", "grey"}).
			Label(field.Label).
			Variant(vuetify.FieldVariantUnderlined).
			Attr(web.VField(field.FormKey, field.Value(obj))...)
	})

}

func ImageContainerBody(data *ImageContainer, input *pagebuilder.RenderInput) (body HTMLComponent) {
	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("Image"))+"_%v", data.ID), data.AnchorID, "container-image",
		data.BackgroundColor, data.TransitionBackgroundColor, "",
		"", data.AddTopSpace, data.AddBottomSpace, input.IsEditor, input.IsReadonly, "", input,
		Div(
			ImageHtml(data.Image),
			Div().Class("container-image-corner"),
		).Class("container-wrapper"),
	)
	return
}
