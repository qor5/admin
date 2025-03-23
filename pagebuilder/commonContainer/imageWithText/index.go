package imageWithText

import (
	"fmt"
	"strconv"

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
			p.Style.ImageHeight = []string{"auto"}
			p.Style.ImageWidth = []string{"500px"}
			p.Style.TopSpace = 120
			p.Style.BottomSpace = 120
			p.Style.LeftSpace = 120
			p.Style.RightSpace = 120
			p.Style.Visibility = []string{"title", "content", "button", "image"}

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})

	ed.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			// p := obj.(*ImageWithText)

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

	SetHeroContentComponent(pb, ed, db)
	SetHeroStyleComponent(pb, ed)
}

func ImageWithTextBody(data *ImageWithText, input *pagebuilder.RenderInput) (body HTMLComponent) {
	heroImgUrl := data.Content.ImageUpload.URL()
	// backgroundImgUrl := data.Style.ImageBackground.URL()

	if heroImgUrl == "" {
		heroImgUrl = "https://placehold.co/500x400"
	}

	// Set default values for ImageHeight and ImageWidth if they're empty
	if len(data.Style.ImageHeight) == 0 {
		data.Style.ImageHeight = []string{"auto"}
	}
	if len(data.Style.ImageWidth) == 0 {
		data.Style.ImageWidth = []string{"500px"}
	}

	// For easier access and to avoid repeated indexing
	imageHeight := data.Style.ImageHeight[0]
	imageWidth := data.Style.ImageWidth[0]

	hasHeroImage := lo.Contains(data.Style.Visibility, "image")

	heroBody := Div(
		Div(Div(
			Div(
				Div(
					Div(
						If(lo.Contains(data.Style.Visibility, "title"), H1("").Children(RawHTML(data.Content.Title)).
							Class("richEditor-content tw-theme-text tw-theme-h1 font-medium xl:text-[80px] md:text-[48px] text-[25.875px] xl:leading-[98px] md:leading-normal leading-[31.697px]")),
						If(lo.Contains(data.Style.Visibility, "content"), Div().Children(RawHTML(data.Content.Content)).Class("richEditor-content tw-theme-text tw-theme-p xl:text-[24px] md:text-[22px] text-[12px] font-medium xl:leading-[32px] leading-normal").
							ClassIf("text-left", data.Style.HorizontalAlign == "item-start").
							ClassIf("text-center", data.Style.HorizontalAlign == "items-center").
							ClassIf("text-right", data.Style.HorizontalAlign == "items-end")),
						If(lo.Contains(data.Style.Visibility, "button"), Div(
							A().Attr("href", data.Content.ButtonHref).Attr("target", "_blank").Attr("rel", "noopener noreferrer").Children(Button(data.Content.Button).
								Class("tw-theme-bg-brand tw-theme-text-base tw-theme-p xl:text-[16px] md:text-[14px] text-[12px] xl:px-6 xl:py-3 md:px-4 md:py-2 px-3 py-[6px] rounded-[4px]")),
						)),
					).Attr("x-ref", "leftContent").Class(fmt.Sprintf("flex flex-col xl:gap-10 md:gap-6 gap-3 h-full %s %s", data.Style.VerticalAlign, data.Style.HorizontalAlign)),
				).ClassIf("order-2 xl:ml-10 md:ml-[20px] ml-6", data.Style.Layout == "right"),
				If(hasHeroImage, Div(
					Div(
						Img(heroImgUrl).Class("position-absolute w-full object-cover object-center h-full max-w-full left-0 top-0 flex-shrink-0"),
					).
						Class(fmt.Sprintf("tw-theme-filter-container flex-shrink-0 overflow-hidden xl:h-[%s] xl:w-[%s] md:h-[%s] md:w-[%s] h-[%s] w-[%s]",
							imageHeight,
							imageWidth,
							getScaledImageDimension(imageHeight, 314.0/500.0),
							getScaledImageDimension(imageWidth, 314.0/500.0),
							getScaledImageDimension(imageHeight, 169.0/500.0),
							getScaledImageDimension(imageWidth, 169.0/500.0),
						)),
				).Class(fmt.Sprintf("order-1 flex flex-col items-center %s", data.Style.VerticalAlign)).
					ClassIf("xl:ml-10 md:ml-[20px] ml-6", data.Style.Layout == "left").
					ClassIf("justify-center", data.Style.VerticalAlign == "justify-between")),
			).Class(fmt.Sprintf("flex justify-between xl:max-w-[1280px] md:max-w-[768px] max-w-[414px] mx-auto xl:pl-[%dpx] xl:pr-[%dpx]  md:pl-[%dpx] md:pr-[%dpx] pl-[%dpx] pr-[%dpx] xl:pt-[%dpx] xl:pb-[%dpx] md:pt-[%dpx] md:pb-[%dpx]  pt-[%dpx] pb-[%dpx]",
				data.Style.LeftSpace,
				data.Style.RightSpace,
				int(float64(data.Style.LeftSpace)*0.5),
				int(float64(data.Style.RightSpace)*0.5),
				int(float64(data.Style.LeftSpace)*0.26),
				int(float64(data.Style.RightSpace)*0.26),
				data.Style.TopSpace,
				data.Style.BottomSpace,
				int(float64(data.Style.TopSpace)*0.5),
				int(float64(data.Style.BottomSpace)*0.5),
				int(float64(data.Style.TopSpace)*0.26),
				int(float64(data.Style.BottomSpace)*0.26),
			)).Attr("x-data", fmt.Sprintf(`{
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

func getScaledImageDimension(value string, scale float64) string {
	if value == "auto" {
		return "auto"
	}

	// More robust parsing of CSS dimension values
	var size float64
	// Using strconv.ParseFloat with string manipulation
	numPart := ""
	for _, c := range value {
		if (c >= '0' && c <= '9') || c == '.' {
			numPart += string(c)
		} else {
			// We found the first non-numeric character
			// Try to parse what we have so far
			if size, err := strconv.ParseFloat(numPart, 64); err == nil {
				// Successfully parsed the number, apply scaling
				return fmt.Sprintf("%dpx", int(size*scale))
			}
			break
		}
	}

	// Fallback to original Sscanf approach
	n, err := fmt.Sscanf(value, "%fpx", &size)
	if err != nil || n != 1 {
		return value // Return original if parsing fails
	}

	// Scale and return the result with px units
	return fmt.Sprintf("%dpx", int(size*scale))
}
