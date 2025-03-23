package heroImageVertical

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

type TailWindHeroVertical struct {
	ID uint

	Content heroContent
	Style   heroStyle
}

func (*TailWindHeroVertical) TableName() string {
	return "container_tailwind_hero_vertical"
}

func RegisterContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("TailWindHeroVertical").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*TailWindHeroVertical)

			return HeroBody(v, input)
		})

	ed := vb.Model(&TailWindHeroVertical{}).Editing("Tabs", "Content", "Style")

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
			p := obj.(*TailWindHeroVertical)
			p.Content.Title = "This is a title"
			p.Content.Body = "From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation."
			p.Content.Button = "Get Start"
			p.Content.ButtonStyle = "unset"
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
			p := obj.(*TailWindHeroVertical)

			if p.Content.ImageUpload.URL() != "" && !p.Content.ImgInitial {
				p.Content.ImgInitial = true
			}

			// if p.Style.ImageBackground.URL() != "" && !p.Style.ImgInitial {
			// 	p.Style.ImgInitial = true
			// }

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})

	SetHeroContentComponent(pb, ed, db)
	SetHeroStyleComponent(pb, ed)
}

func HeroBody(data *TailWindHeroVertical, input *pagebuilder.RenderInput) (body HTMLComponent) {
	heroImgUrl := data.Content.ImageUpload.URL()
	// backgroundImgUrl := data.Style.ImageBackground.URL()

	if heroImgUrl == "" && !data.Content.ImgInitial {
		heroImgUrl = "https://placehold.co/1024x585"
	}

	heroBody := Div(
		Div(
			Div(
				H1(data.Content.Title).Class("tw-theme-text cc-h1 text-center font-medium xl:text-[80px] md:text-[48px] text-[25.875px] xl:leading-[98px] md:leading-normal leading-[31.697px]"),

				P(Text(data.Content.Body)).
					Class("tw-theme-text cc-content text-2xl text-center break-all"),

				Button(data.Content.Button).Class("tw-theme-bg-brand  cc-content tw-theme-text-base mx-auto py-3 px-6 rounded-[4px] text-[16px] leading-6", fmt.Sprintf("btn-%s", data.Content.ButtonStyle)),

				Div().Class("tw-theme-filter-container w-[1024px] m-auto h-[585px] bg-cover bg-no-repeat bg-center").
					Style(fmt.Sprintf("background-image: url(%s)", heroImgUrl)).
					Attr("alt", ""),
			).Class("p-[120px] gap-10 flex flex-col w-[1280px] m-auto"),
		).Class(" tw-theme-bg-base"),
	).Class("container-hero-inner")

	body = utils.TailwindContainerWrapper(
		"container-hero",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"heroImageVertical", "id": %q}`, input.ContainerId)).Children(Div(heroBody).Class("bg-gray-100")),
	)
	return
}
