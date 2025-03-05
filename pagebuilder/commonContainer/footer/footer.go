package footer

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
)

type TailWindExampleFooter struct {
	ID uint

	Content footerContent
	Style   footerStyle
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

	ed := vb.Model(&TailWindExampleFooter{}).Editing("Tabs", "Content", "Style")

	// vb.Model(&TailWindExampleFooter{})
	ed.Field("Tabs").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		option := presets.TabsControllerOption{
			Tabs: []presets.TabControllerOption{
				{Tab: v.VTab().Text("Content"), Fields: []string{"Content"}},
				{Tab: v.VTab().Text("Style"), Fields: []string{"Style"}},
			},
		}
		return presets.TabsController(field, &option)
	})

	ed.Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			p := obj.(*TailWindExampleFooter)
			p.Content.Group1 = "What We Do"
			p.Content.Group1Item1 = "Commerce"
			p.Content.Group1Item2 = "Content"
			p.Content.Group1Item3 = "Consulting"
			p.Content.Group1Item4 = "Personalization"

			p.Content.Group2 = "Projects"
			p.Content.Group2Item1 = "Commerce"
			p.Content.Group2Item2 = "Content"
			p.Content.Group2Item3 = "Consulting"
			p.Content.Group2Item4 = "Personalization"

			p.Content.Group3 = "Why clients choose us"
			p.Content.Group3Item1 = "Commerce"
			p.Content.Group3Item2 = "Content"
			p.Content.Group3Item3 = "Consulting"
			p.Content.Group3Item4 = "Personalization"

			p.Content.Group4 = "Follow Us"

			p.Content.Group5 = "Our Company"
			p.Content.Group5Item1 = "Commerce"
			p.Content.Group5Item2 = "Content"
			p.Content.Group5Item3 = "Consulting"
			p.Content.Group5Item4 = "Personalization"

			// p.Content.Href1 = "What we do"
			// p.Content.Href2 = "Projects"
			// p.Content.Href3 = "Why clients choose us"
			// p.Content.Href4 = "Our company"
			p.Style.Layout = "left"
			p.Style.TopSpace = 0
			p.Style.BottomSpace = 0

			if err = in(obj, id, ctx); err != nil {
				return
			}
			return err
		}
	})

	SetHeroContentComponent(pb, ed, db)
	SetHeroStyleComponent(pb, ed)
}

func footerBody(data *TailWindExampleFooter, input *pagebuilder.RenderInput) (body HTMLComponent) {
	html := Div(

		Div(
			Div(
				Div(
					Ul(
						Li(
							H2(data.Content.Group1).Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text(data.Content.Group1Item1)),
								Li(Text(data.Content.Group1Item2)),
								Li(Text(data.Content.Group1Item3)),
								Li(Text(data.Content.Group1Item4)),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2(data.Content.Group2).Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text(data.Content.Group2Item1)),
								Li(Text(data.Content.Group2Item2)),
								Li(Text(data.Content.Group2Item3)),
								Li(Text(data.Content.Group2Item4)),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2(data.Content.Group3).Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text(data.Content.Group3Item1)),
								Li(Text(data.Content.Group3Item2)),
								Li(Text(data.Content.Group3Item3)),
								Li(Text(data.Content.Group3Item4)),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
						Li(
							H2(data.Content.Group5).Class("tw-theme-text tw-theme-h2 text-2xl mb-6"),
							Ul(
								Li(Text(data.Content.Group5Item1)),
								Li(Text(data.Content.Group5Item2)),
								Li(Text(data.Content.Group5Item3)),
								Li(Text(data.Content.Group5Item4)),
							).Class("tw-theme-text tw-theme-p flex flex-col gap-4").Attr("data-list-unset", "true"),
						),
					).Class("grid grid-cols-[repeat(auto-fill,_minmax(235px,_1fr))] gap-8").Attr("data-list-unset", "true"),
				).Class("flex-1"),
				Div(
					H2(data.Content.Group4).Class("tw-theme-text tw-theme-h2 text-2xl"),
				),
			).Class("py-[120px] px-[82px] flex justify-between gap-8 w-[1280px] m-auto"),
		).Class(" tw-theme-bg-base"),
	).Class("container-tailwind-inner").Attr("data-twind-scope", true)

	body = utils.TailwindContainerWrapper(
		"container-tailwind-example-header",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"container-tailwind-example-header", "id": %q}`, input.ContainerId)).Children(html),
	)
	return
}
