package imageWithText

import (
	"fmt"

	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	v "github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/samber/lo"
)

type ImageWithText struct {
	ID uint

	Content imageWithTextContent
	Style   imageWithTextStyle
}

func (*ImageWithText) TableName() string {
	return "container_image_with_text"
}

func RegisterContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("ImageWithText").Group("Content").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*ImageWithText)

			return ImageWithTextBody(v, input)
		})

	ed := vb.Model(&ImageWithText{}).Editing("Tabs", "Content", "Style")

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
			p := obj.(*ImageWithText)
			p.Content.Title = "This is a title"
			p.Content.Content = "From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation."
			p.Content.Button = "Get Start"
			p.Content.ButtonHref = ""
			p.Style.Layout = "left"
			p.Style.VerticalAlign = "justify-between"
			p.Style.HorizontalAlign = "left"
			p.Style.TopSpace = 120
			p.Style.BottomSpace = 120
			p.Style.Visibility = []string{"title", "content", "button", "image"}

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})

	ed.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			p := obj.(*ImageWithText)

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

func ImageWithTextBody(data *ImageWithText, input *pagebuilder.RenderInput) (body HTMLComponent) {
	heroImgUrl := data.Content.ImageUpload.URL()
	// backgroundImgUrl := data.Style.ImageBackground.URL()

	if heroImgUrl == "" && !data.Content.ImgInitial {
		heroImgUrl = "https://placehold.co/500x400"
	}

	hasHeroImage := lo.Contains(data.Style.Visibility, "image")

	heroBody := Div(
		Div(Div(
			Div(
				Div(
					Div(
						If(lo.Contains(data.Style.Visibility, "title"), H1("").Children(RawHTML(data.Content.Title)).
							Class("richEditor-content tw-theme-text tw-theme-h1 font-medium xl:text-[80px] md:text-[48px] text-[25.875px] xl:leading-[98px] md:leading-normal leading-[31.697px]")),
						If(lo.Contains(data.Style.Visibility, "content"), Div().Children(RawHTML(data.Content.Content)).Class("richEditor-content tw-theme-text tw-theme-p xl:text-[24px] md:text-[22px] text-[12px] xl:mt-10 md:mt-6 mt-3 font-medium xl:leading-[32px] leading-normal").
							ClassIf("text-left", data.Style.HorizontalAlign == "item-start").
							ClassIf("text-center", data.Style.HorizontalAlign == "items-center").
							ClassIf("text-right", data.Style.HorizontalAlign == "items-end")),
						If(lo.Contains(data.Style.Visibility, "button"), Div(
							A().Attr("href", data.Content.ButtonHref).Attr("target", "_blank").Attr("rel", "noopener noreferrer").Children(Button(data.Content.Button).
								Class("tw-theme-bg-brand tw-theme-text-base tw-theme-p xl:text-[16px] md:text-[14px] text-[12px] xl:mt-10 md:mt-6 mt-3 xl:px-6 xl:py-3 md:px-4 md:py-2 px-3 py-[6px] rounded-[4px]")),
						).ClassIf("text-right", data.Style.Layout == "right")),
					).Attr("x-ref", "leftContent").Class(fmt.Sprintf("flex flex-col h-full %s %s", data.Style.VerticalAlign, data.Style.HorizontalAlign)),
				).ClassIf("order-2 xl:ml-10 md:ml-[20px] ml-6", data.Style.Layout == "right"),
				Template(
					Div(
						Div().
							Class("tw-theme-filter-container bg-cover  xl:w-[500px] md:w-[314px] w-[169px] bg-center bg-no-repeat flex-shrink-0 ").
							Attr(":style", fmt.Sprintf("`background-image: url(%s); aspect-ratio: ${imageAspectRatio};`", heroImgUrl)),
					).ClassIf("xl:ml-10 md:ml-[20px] ml-6", data.Style.Layout == "left").
						Class(fmt.Sprintf("order-1 flex flex-col items-center %s", data.Style.VerticalAlign)).
						ClassIf("justify-center", data.Style.VerticalAlign == "justify-between"),
				).Attr("x-if", "hasHeroImage"),
			).Class(fmt.Sprintf("flex justify-between xl:max-w-[1280px] mx-auto xl:px-[120px] xl:pt-[%dpx] xl:pb-[%dpx] md:px-[60px] md:pt-[%dpx] md:pb-[%dpx] px-8 pt-[%dpx] pb-[%dpx]",
				data.Style.TopSpace, data.Style.BottomSpace, int(float64(data.Style.TopSpace)*0.5), int(float64(data.Style.BottomSpace)*0.5), int(float64(data.Style.TopSpace)*0.26), int(float64(data.Style.BottomSpace)*0.26))).Attr("x-data", fmt.Sprintf(`{
			imageAspectRatio: '5 / 4',
			hasHeroImage: %t,
		}`, hasHeroImage)),
		).Class("tw-theme-bg-brand-20 text-[#212121] bg-no-repeat bg-cover bg-center"),
		// Style(fmt.Sprintf("background-image: url(%s)", backgroundImgUrl)),
		).Class(" tw-theme-bg-base"),
	).Class("container-hero-inner")

	body = utils.TailwindContainerWrapper(
		"container-hero",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"imageWithText", "id": %q}`, input.ContainerId)).Children(Div(heroBody).Class("bg-gray-100")),
	)
	return
}
