package CardList

import (
	"fmt"

	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	v "github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
	"github.com/qor5/admin/v3/presets"
)

type CardList struct {
	ID uint

	Content cardListContent
	Style   cardListStyle
}

func (*CardList) TableName() string {
	return "container_card_list"
}

func RegisterContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("CardList").Group("Content").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*CardList)

			return CardListBody(v, input)
		})

	ed := vb.Model(&CardList{}).Editing("Tabs", "Content", "Style")

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
			p := obj.(*CardList)
			p.Content.Title = "This is a title"
			p.Content.ProductTitle1 = "Commerce"
			p.Content.ProductDescription1 = "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable"
			p.Content.ProductTitle2 = "Commerce"
			p.Content.ProductDescription2 = "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable"
			p.Content.ProductTitle3 = "Commerce"
			p.Content.ProductDescription3 = "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable"
			// p.Content.Body = "From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation."
			// p.Content.Button = "Get Start"
			// p.Content.ButtonStyle = "unset"
			// p.Style.Layout = "left"
			p.Style.TopSpace = 0
			p.Style.BottomSpace = 0

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})

	ed.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			// p := obj.(*TailWindHeroList)

			// if p.Content.ImageUpload.URL() != "" && !p.Content.ImgInitial {
			// 	p.Content.ImgInitial = true
			// }

			// if p.Style.ImageBackground.URL() != "" && !p.Style.ImgInitial {
			// 	p.Style.ImgInitial = true
			// }

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})

	SetContentComponent(pb, ed, db)
	SetStyleComponent(pb, ed)
}

func CardListBody(data *CardList, input *pagebuilder.RenderInput) (body HTMLComponent) {
	image1 := "https://placehold.co/500x500"
	image2 := "https://placehold.co/500x500"
	image3 := "https://placehold.co/500x500"

	// heroImgUrl := data.Content.ImageUpload.URL()
	// backgroundImgUrl := data.Style.ImageBackground.URL()

	// if heroImgUrl == "" && !data.Content.ImgInitial {
	// 	heroImgUrl = "https://placehold.co/1024x585"
	// }

	heroBody := Div(
		Div(
			H1(data.Content.Title).Class("tw-theme-text cc-h1 text-center font-medium text-5xl leading-none"),

			Ul(
				Li(
					Div().Class("aspect-square bg-center bg-cover").
						Style(fmt.Sprintf("background-image: url(%s)", image1)),
					H2(data.Content.ProductTitle1).Class("tw-theme-text cc-h2 text-bold xl:text-xl xl:leading-6 md:text-xl md:leading-7 mt-4"),
					P(Text(data.Content.ProductDescription1)).
						Class("mt-2 tw-theme-text cc-content xl:text-base xl:leading-6 text-[14px] leading-[20px]"),
				),
				Li(
					Div().Class("aspect-square bg-center bg-cover").
						Style(fmt.Sprintf("background-image: url(%s)", image2)),
					H2(data.Content.ProductTitle2).Class("tw-theme-text cc-h2 text-bold xl:text-xl xl:leading-6 md:text-xl md:leading-7 mt-4"),
					P(Text(data.Content.ProductDescription2)).
						Class("mt-2 tw-theme-text cc-content xl:text-base xl:leading-6 text-[14px] leading-[20px]"),
				),
				Li(
					Div().Class("aspect-square bg-center bg-cover").
						Style(fmt.Sprintf("background-image: url(%s)", image3)),
					H2(data.Content.ProductTitle3).Class("tw-theme-text cc-h2 text-bold xl:text-xl xl:leading-6 md:text-xl md:leading-7 mt-4"),
					P(Text(data.Content.ProductDescription3)).
						Class("mt-2 tw-theme-text cc-content xl:text-base xl:leading-6 text-[14px] leading-[20px]"),
				),
			).Class("grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 mt-[60px]"),
		).Class("px-5 py-10 xl:w-[1280px] m-auto"),
	).Class("bg-no-repeat bg-cover bg-center cc-wrapper")

	body = utils.TailwindContainerWrapper(
		"container-hero",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"card-list", "id": %q}`, input.ContainerId)).Children(Div(heroBody)),
	)
	return
}
