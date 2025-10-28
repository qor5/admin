package containers

import (
	"fmt"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
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
	vb := pb.RegisterContainer("Heading").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*Heading)
			return HeadingBody(v, input)
		})
	ed := vb.Model(&Heading{}).Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Heading", "FontColor", "BackgroundColor", "Link", "LinkText", "LinkDisplayOption", "Text")
	ed.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		extensions := tiptap.TiptapExtensions()
		return tiptap.TiptapEditor(db, field.FormKey).
			Extensions(extensions).
			MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
			Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
			Label(field.Label).
			Disabled(field.Disabled)
	})
	ed.Field("Heading").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			return Components(presets.LinkageFieldsController(field, "AnchorID"), in(obj, field, ctx))
		}
	})
	ed.Field("AddTopSpace").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			return Components(presets.LinkageFieldsController(field, "AddBottomSpace"), in(obj, field, ctx))
		}
	})
	ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		p := obj.(*Heading)
		if p.ID != 0 {
			if p.LinkText == "" {
				err.FieldError("LinkText", "LinkText 不能为空")
			}
			if p.Heading == "" && p.AnchorID == "" {
				err.FieldError("Heading", "Heading or AnchorID 不能同时为空")
			}
			if !p.AddTopSpace && !p.AddBottomSpace {
				err.FieldError("AddTopSpace", "AddTopSpace or AddBottomSpace 不能同时为空")
			}
		}
		return
	})
	ed.Field("LinkText").LazyWrapSetterFunc(func(in presets.FieldSetterFunc) presets.FieldSetterFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			if err = in(obj, field, ctx); err != nil {
				return
			}
			p := obj.(*Heading)
			p.LinkText = strings.ReplaceAll(p.LinkText, "{{Name}}", field.Name)
			return
		}
	})

	ed.Field("FontColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(FontColors)
	})
	ed.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(BackgroundColors)
	})
	ed.Field("LinkDisplayOption").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(LinkDisplayOptions)
	})
}

func HeadingBody(data *Heading, input *pagebuilder.RenderInput) (body HTMLComponent) {
	headingBody := Div(
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
		data.AnchorID, "container-heading", data.BackgroundColor, "", data.FontColor,
		"", data.AddTopSpace, data.AddBottomSpace, "",
		Div(headingBody).Class("container-wrapper"),
	)
	return
}
