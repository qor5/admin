package header

import (
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
)

type TailWindExampleHeader struct {
	ID uint
}

func (b *TailWindExampleHeader) TableName() string {
	return "container_tailwind_header"
}

func RegisterContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("TailWindExampleHeader").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*TailWindExampleHeader)
			return headerBody(v, input)
		})

	vb.Model(&TailWindExampleHeader{})
}

func headerBody(data *TailWindExampleHeader, input *pagebuilder.RenderInput) (body HTMLComponent) {
	html := Div(
		Header(
			Div(
				Span("Logo").Class("logo text-3xl tw-theme-p font-bold tw-theme-text"),
				Ul(
					Li(A(Text("What we do")).Href("#").Class("tw-theme-p")),
					Li(A(Text("Projects")).Href("#").Class("tw-theme-p")),
					Li(A(Text("Why clients choose us")).Href("#").Class("tw-theme-p")),
					Li(A(Text("Our company")).Href("#").Class("tw-theme-p")),
				).Class("list-none flex tw-theme-text text-[28px] gap-[72px] font-[500]").
					Attr("data-list-unset", "true"),
			).Class("w-[1152px] m-0 mx-auto leading-9 flex justify-between"),
		).Class(" py-12 tw-theme-bg-base"),
	).Class("container-tailwind-inner")

	body = utils.TailwindContainerWrapper(
		"container-tailwind-example-header",
		Tag("twind-scope").Children(html),
	)
	return
}
