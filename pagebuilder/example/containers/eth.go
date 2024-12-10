package containers

import (
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

type Eth struct {
	ID uint
	B1
	B2
	B3
}
type B1 struct {
	AddTopSpace    bool
	AddBottomSpace bool
}

type B2 struct {
	AnchorID        string
	Heading         string
	FontColor       string
	BackgroundColor string
}

type B3 struct {
	Link              string
	LinkText          string
	LinkDisplayOption string
	Text              string
}

func RegisterEth(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("Eth").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*Eth)
			return EthBody(v, input)
		})
	cb := vb.Model(&Eth{})
	ed := cb.Editing()
	f1 := ed.Field("AddTopSpace").Label("tab1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return ed.Clone().Only("AddTopSpace", "AddBottomSpace").ToComponent(cb.GetModelBuilder().Info(), obj, ctx)

	})
	f2 := ed.Field("Heading").Label("Tab2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return ed.Clone().Only("AnchorID", "Heading").ToComponent(cb.GetModelBuilder().Info(), obj, ctx)
	})

	tb := presets.NewTabsFieldBuilder().
		TabsOrderFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) []string {
			return []string{"AddTopSpace", "Heading"}
		})
	ed.Viewing("AnchorID").Field("AnchorID").Tab(tb).AppendTabs(f1).AppendTabs(f2)
}

func EthBody(data *Eth, input *pagebuilder.RenderInput) (body HTMLComponent) {
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
