package containers

import (
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

type WebHeader struct {
	ID    uint
	Color string
}

func (*WebHeader) TableName() string {
	return "container_headers"
}

func RegisterHeader(pb *pagebuilder.Builder) {
	header := pb.RegisterContainer("Header").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			header := obj.(*WebHeader)
			return HeaderTemplate(header, input)
		})

	ed := header.Model(&WebHeader{}).Editing("Color")
	ed.Field("Color").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items([]string{"black", "white"})
	})
}

func HeaderTemplate(data *WebHeader, input *pagebuilder.RenderInput) (body HTMLComponent) {
	// bc := data.Color
	style := "color: #fff;background: #000;"
	// if input.IsEditor && (input.Device == "phone" || input.Device == "tablet") {
	//	bc = "black"
	// }
	if data.Color == "white" {
		style = "color: #000;background: #fff;"
	}

	body = ContainerWrapper(
		"", "container-header", "", "", "",
		"", false, false, style,
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
