package containers

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/pagebuilder"
	"github.com/qor/qor5/richeditor"
	. "github.com/theplant/htmlgo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	LINK_DISPLAY_OPTION_DESKTOP = "desktop"
	LINK_DISPLAY_OPTION_MOBILE  = "mobile"
	LINK_DISPLAY_OPTION_ALL     = "all"
)

var LinkDisplayOptions = []string{LINK_DISPLAY_OPTION_ALL, LINK_DISPLAY_OPTION_DESKTOP, LINK_DISPLAY_OPTION_MOBILE}

type Heading struct {
	ID                uint
	AddTopSpace       bool
	AddBottomSpace    bool
	AnchorID          string
	Heading           string
	FontColor         string
	BackgroundColor   string
	Link              string
	LinkText          string
	LinkDisplayOption string
	Text              string
}

func (*Heading) TableName() string {
	return "container_headings"
}

func RegisterHeadingContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("Heading").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*Heading)
			return HeadingBody(v)
		})
	ed := vb.Model(&Heading{}).Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Heading", "FontColor", "BackgroundColor", "Link", "LinkText", "LinkDisplayOption", "Text")
	ed.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Text").Plugins([]string{"alignment", "video", "imageinsert", "fontcolor"}).Value(obj.(*Heading).Text).Label(field.Label)
	})

	ed.Field("FontColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vuetify.VSelect().
			Items(FontColors).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})
	ed.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vuetify.VSelect().
			Items(BackgroundColors).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})
	ed.Field("LinkDisplayOption").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vuetify.VSelect().
			Items(LinkDisplayOptions).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})
}

func HeadingBody(data *Heading) (body HTMLComponent) {
	headingBody :=
		Div(
			Div(
				If(data.Heading != "",
					If(data.Link != "",
						A(H2(data.Heading).Class("container-heading-title")).Class("container-heading-title-link").Href(data.Link),
					),
					If(data.Link == "",
						H2(data.Heading).Class("container-heading-title"),
					),
				),
				If(data.Text != "", Div(RawHTML(data.Text)).Class("container-heading-content")),
			).Class("container-heading-wrap"),
			If(data.LinkText != "" && data.Link != "",
				Div(
					LinkTextWithArrow(data.LinkText, data.Link),
				).Class("container-heading-link").Attr("data-display", data.LinkDisplayOption),
			),
		).Class("container-heading-inner")

	body = ContainerWrapper(
		fmt.Sprintf("heading_%v", data.ID), data.AnchorID, "container-heading", data.BackgroundColor, data.FontColor, "",
		data.AddTopSpace, data.AddBottomSpace,
		Div(headingBody).Class("container-wrapper"),
	)
	return
}
