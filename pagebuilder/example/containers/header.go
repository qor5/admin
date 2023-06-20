package containers

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"

	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	. "github.com/theplant/htmlgo"
)

type WebHeader struct {
	ID    uint
	Color string
}

func (*WebHeader) TableName() string {
	return "container_headers"
}

func RegisterHeader(pb *pagebuilder.Builder) {
	header := pb.RegisterContainer("Header").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			header := obj.(*WebHeader)
			return HeaderTemplate(header, input)
		})

	ed := header.Model(&WebHeader{}).URIName(inflection.Plural(strcase.ToKebab("Header"))).Editing("Color")
	ed.Field("Color").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vuetify.VSelect().
			Items([]string{"black", "white"}).
			Value(field.Value(obj)).
			Label(field.Label).
			FieldName(field.FormKey)
	})
}

func HeaderTemplate(data *WebHeader, input *pagebuilder.RenderInput) (body HTMLComponent) {
	//bc := data.Color
	style := "color: #fff;background: #000;"
	//if input.IsEditor && (input.Device == "phone" || input.Device == "tablet") {
	//	bc = "black"
	//}
	if data.Color == "white" {
		style = "color: #000;background: #fff;"
	}

	body = ContainerWrapper(
		fmt.Sprintf(inflection.Plural(strcase.ToKebab("Header"))+"_%v", data.ID), "", "container-header", "", "", "",
		"", false, false, input.IsEditor, input.IsReadonly, style,
		Div(RawHTML(`
<a href="/" class="container-header-logo"><svg viewBox="0 0 29 30" fill="none" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M14.399 10.054V0L0 10.054V29.73h28.792V0L14.4 10.054z" fill="currentColor"><title>The Plant</title></path></svg></a>
<ul data-list-unset="true" class="container-header-links">
<li>
<a href="/what-we-do/">What we do</a>
</li>
<li>
<a href="/projects/">Projects</a>
</li>
<li>
<a href="/why-clients-choose-us/">Why clients choose us</a>
</li>
<li>
<a href="/our-company/">Our company</a>
</li>
</ul>
<button class="container-header-menu">
<span class="container-header-menu-icon"></span>
</button>`)).Class("container-wrapper"),
	)
	return
}
