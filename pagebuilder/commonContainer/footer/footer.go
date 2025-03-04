package footer

import (
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
)

type TailWindExampleFooter struct {
	ID uint
}

func (b *TailWindExampleFooter) TableName() string {
	return "container_tailwind_footer"
}

func RegisterContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("TailWindExampleFooter").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*TailWindExampleFooter)
			return footerBody(v, input)
		})

	vb.Model(&TailWindExampleFooter{})
}

func footerBody(data *TailWindExampleFooter, input *pagebuilder.RenderInput) (body HTMLComponent) {
	html := Div(

		Div(
			Div(
				Div(
					Ul(
						Li(
							H2("What We Do").Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text("Commerce")),
								Li(Text("Content")),
								Li(Text("Consulting")),
								Li(Text("Personalization")),
								Li(),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2("Projects").Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text("Commerce")),
								Li(Text("Content")),
								Li(Text("Consulting")),
								Li(Text("Personalization")),
								Li(),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2("Why clients choose us").Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text("Commerce")),
								Li(Text("Content")),
								Li(Text("Consulting")),
								Li(Text("Personalization")),
								Li(),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2("Our company").Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text("Commerce")),
								Li(Text("Content")),
								Li(Text("Consulting")),
								Li(Text("Personalization")),
								Li(),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
					).Class("grid grid-cols-[repeat(auto-fill,_minmax(235px,_1fr))] gap-8").Attr("data-list-unset", "true"),
				).Class("flex-1"),
				Div(
					H2("Follow us").Class("tw-theme-text tw-theme-h2 text-2xl"),
				),
			).Class("py-[120px] px-[82px] flex justify-between gap-8 w-[1280px] m-auto"),
		).Class(" tw-theme-bg-base"),
	).Class("container-tailwind-inner").Attr("data-twind-scope", true)

	body = utils.TailwindContainerWrapper(
		"container-tailwind-example-header",
		Tag("twind-scope").Children(html),
	)
	return
}
