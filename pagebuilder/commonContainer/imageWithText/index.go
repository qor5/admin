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
			p.Content.ButtonStyle = "unset"
			p.Style.Layout = "left"
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
			p := obj.(*ImageWithText)

			if p.Content.ImageUpload.URL() != "" && !p.Content.ImgInitial {
				p.Content.ImgInitial = true
			}

			if p.Style.ImageBackground.URL() != "" && !p.Style.ImgInitial {
				p.Style.ImgInitial = true
			}

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
	backgroundImgUrl := data.Style.ImageBackground.URL()

	if heroImgUrl == "" && !data.Content.ImgInitial {
		heroImgUrl = "https://placehold.co/1024x585"
	}

	hasHeroImage := heroImgUrl != ""

	heroBody := Div(
		Div(Div(
			Div(
				Div(
					Div(
						H1("").Children(RawHTML(data.Content.Title)).Class("richEditor-content tw-theme-text tw-theme-h1 font-medium xl:text-[80px] md:text-[48px] text-[25.875px] xl:leading-[98px] md:leading-normal leading-[31.697px]"),
						Div().Children(RawHTML(data.Content.Content)).Class("richEditor-content tw-theme-text tw-theme-p xl:text-[24px] md:text-[22px] text-[12px] xl:py-10 md:py-6 py-3 font-medium xl:leading-[32px] leading-normal"),
						Div(
							Button(data.Content.Button).Class("tw-theme-bg-brand tw-theme-text-base tw-theme-p xl:text-[16px] md:text-[14px] xl:px-6 px-4 xl:py-3 md:py-2 py-[6px] rounded-[4px]", fmt.Sprintf("btn-%s", data.Content.ButtonStyle)),
						).Class("mt-auto").ClassIf("text-right", data.Style.Layout == "right"),
					).Attr("x-ref", "leftContent").Class("flex flex-col h-full"),
				).ClassIf("order-2 xl:ml-10 md:ml-[20px] ml-3", data.Style.Layout == "right"),
				Template(
					Div(
						Div().Attr("x-ref", "rightImageForCalc").
							Class("absolute -z-10 bg-cover bg-center bg-no-repeat flex-shrink-0 xl:w-[500px] md:w-[314px] w-[169px]").
							Attr(":style", fmt.Sprintf("`background-image: url(%s); aspect-ratio: ${imageAspectRatioForCalc};`", heroImgUrl)),

						Div().Attr("x-ref", "rightImage").
							Class("bg-cover absolute top-[50%] left-[50%] translate-x-[-50%] translate-y-[-50%] transition-all bg-center bg-no-repeat flex-shrink-0 xl:w-[500px] md:w-[314px] w-[169px]").
							Attr(":style", fmt.Sprintf("`background-image: url(%s); aspect-ratio: ${imageAspectRatio};`", heroImgUrl)),
					).
						Class("tw-theme-filter-container order-1 xl:w-[500px] xl:min-h-[400px] md:min-h-[251px] min-h-[135px] relative md:w-[314px] w-[169px] overflow-hidden flex-shrink-0 xl:ml-10 md:ml-[20px] ml-3").
						ClassIf("xl:ml-10 md:ml-[20px] ml-3", data.Style.Layout == "left"),
				).Attr("x-if", "hasHeroImage"),
			).Class("flex items-stretch justify-between xl:max-w-[1280px] mx-auto xl:p-[120px] md:p-[60px] p-8").Attr("x-data", fmt.Sprintf(`{
			imageAspectRatio: '5 / 4',
			hasHeroImage: %t,
			imageAspectRatioForCalc: '5 / 4',
            debounceTimer: null,
            debounceAdjust() {
              clearTimeout(this.debounceTimer);
              this.debounceTimer = setTimeout(() => {
                this.adjustAspectRatio();
              }, 0);
            },
            adjustAspectRatio() {
              this.$nextTick(async () => {
                const leftHeight = this.$refs.leftContent?.offsetHeight;
                const imageHeightForCalc = this.$refs.rightImageForCalc?.offsetHeight;
                this.imageAspectRatioForCalc = '5 / 4';

								if(!leftHeight || !imageHeightForCalc) return

                if (leftHeight > imageHeightForCalc) {
                  this.imageAspectRatio = '3 / 4';
                } else {
                  this.imageAspectRatio = '5 / 4';
                }
              });
            },
            init() {
              const resizeObserver = new ResizeObserver(() => {
                this.debounceAdjust();
              });
              resizeObserver.observe(this.$refs.leftContent);
            }
          }`, hasHeroImage)),
		).Class(fmt.Sprintf("tw-theme-bg-brand-20 text-[#212121] bg-no-repeat bg-cover bg-center pt-[%dpx] pb-[%dpx]", data.Style.TopSpace, data.Style.BottomSpace)).
			Style(fmt.Sprintf("background-image: url(%s)", backgroundImgUrl)),
		).Class(" tw-theme-bg-base"),
	).Class("container-hero-inner")

	body = utils.TailwindContainerWrapper(
		"container-hero",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"imageWithText", "id": %q}`, input.ContainerId)).Children(Div(heroBody).Class("bg-gray-100")),
	)
	return
}
