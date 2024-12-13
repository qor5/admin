package containers

import (
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

type Hero struct {
	ID uint

	Content heroContent
	Style   heroStyle
	// H2AreaStyle  H2AreaStyle
}

// type H1AreaStyle struct{}

// type H2AreaStyle struct{}

func (*Hero) TableName() string {
	return "container_hero"
}

func RegisterHeroContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("Hero").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			v := obj.(*Hero)

			return HeroBody(v, input)
		})
	ed := vb.Model(&Hero{}).Editing("Content", "Style")

	ed.Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			p := obj.(*Hero)
			p.Content.Title = "This is a title"
			p.Content.Body = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. suspendisse tincidunt sagitis eros. Quisque quis euismod lorem"

			if err = in(obj, id, ctx); err != nil {
				return
			}

			return
		}
	})
	// vb.Model(&Hero{}).Editing("heroStyle")

	// ed.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	extensions := tiptap.TiptapExtensions()
	// 	return tiptap.TiptapEditor(db, field.Name).
	// 		Extensions(extensions).
	// 		MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
	// 		Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
	// 		Label(field.Label).
	// 		Disabled(field.Disabled)
	// })
	// ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
	// 	p := obj.(*Hero)
	// 	if p.ID != 0 {
	// 		if p.LinkText == "" {
	// 			err.FieldError("LinkText", "LinkText 不能为空")
	// 		}
	// 	}
	// 	return
	// })
	// ed.Field("LinkText").LazyWrapSetterFunc(func(in presets.FieldSetterFunc) presets.FieldSetterFunc {
	// 	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	// 		if err = in(obj, field, ctx); err != nil {
	// 			return
	// 		}
	// 		p := obj.(*Hero)
	// 		p.LinkText = strings.Replace(p.LinkText, "{{Name}}", field.Name, -1)
	// 		return
	// 	}
	// 	}
	// })

	// ed.Field("FontColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	return presets.SelectField(obj, field, ctx).Items(FontColors)
	// })
	// ed.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	return presets.SelectField(obj, field, ctx).Items(BackgroundColors)
	// })
	// ed.Field("LinkDisplayOption").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	return presets.SelectField(obj, field, ctx).Items(LinkDisplayOptions)
	// })
	SetHeroContentComponent(pb, ed, db)
	SetHeroStyleComponent(pb, ed)
}

func HeroBody(data *Hero, input *pagebuilder.RenderInput) (body HTMLComponent) {
	// if there is no image, use default image "https://via.placeholder.com/308x252"
	// imgUrl := data.Content.ImageUpload.URL()
	// if data.Content.ImageUpload.URL() == "" {
	// 	imgUrl = "https://via.placeholder.com/308x252"
	// }

	// 	Div(
	// 		Div(
	// 			Div(
	// 			// Use a div with background image instead of Img
	// 			).Style(fmt.Sprintf("background-image: url('%s');", imgUrl)).
	// 				Class("w-full", "h-full", "bg-cover", "bg-center", "aspect-[3/4]"),
	// 		).Class("lg:w-auto", "md:w-auto", "flex", "justify-start", "h-full"),
	// 		Div(
	// 			Div(
	// 				If(data.Content.Text != "", Div(RawHTML(data.Content.Text)),
	// 					If(data.Content.Text == "<p></p>", Div(
	// 						H1("Testimonial introduction copy goes here, lorem ipsum dolor sit amet.").Class("text-3xl", "md:text-2xl", "text-black", "opacity-80", "font-bold", "mb-1"),
	// 						P(Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. suspendisse tincidunt sagitis eros. Quisque quis euismod lorem")).Class("text-black", "opacity-80", "font-medium", "sm:text-sm"),
	// 					)),
	// 				),
	// 			),
	// 			Div(
	// 				H2("Author Name").Class("text-2xl", "md:text-xl", "text-black", "opacity-80", "font-bold", "mb-2", "lg:mt-0", "md:mt-0", "mt-8"),
	// 				P(Text("Co-Founder and CEO of Company")).Class("text-pretty", "text-black", "opacity-80", "font-medium", "sm:text-sm"),
	// 			),
	// 		).Class("lg:flex", "md:flex", "flex-col", "lg:h-[252px]", "md:h-[183px]", "justify-between", "lg:flex-1", "md:w-full", "text-left", "lg:pl-20", "md:pl-10", "lg:mt-0", "md:mt-0", "lg:text-left", "md:text-left", "text-center", "mt-8"),
	// 	).Class("mx-auto", "max-w-5xl", "lg:flex", "md:flex", "items-center", "justify-center", "h-full", "lg:px-14", "lg:py-0", "md:py-0", "p-7"),
	// ).Class("flex-wrap", "lg:h-[400px]", "md:h-[300px]"),

	heroBody := Div(
		Div(Div(
			Div(
				Div(
					Div(
						H1(data.Content.Title).Class("font-medium xl:text-[80px] md:text-[48px] text-[25.875px] xl:leading-[98px] md:leading-normal leading-[31.697px]"),
						P(Text(data.Content.Body)).Class("xl:text-[24px] md:text-[22px] text-[12px] xl:py-10 md:py-6 py-3 font-medium xl:leading-[32px] leading-normal"),
						Div(
							Button("Get Start").Class("bg-[#3E63DD] xl:text-[16px] md:text-[14px] xl:px-6 px-4 xl:py-3 md:py-2 py-[6px] rounded-[4px] text-white"),
						).Class("mt-auto"),
					).Attr("x-ref", "leftContent").Class("flex flex-col h-full"),
				),
				Div(
					Div().Attr("x-ref", "rightImageForCalc").
						Class("absolute -z-10 bg-cover bg-center bg-no-repeat flex-shrink-0 xl:w-[500px] md:w-[314px] w-[169px]").
						Attr(":style", "`background-image: url(https://s3-alpha-sig.figma.com/img/90f2/44c4/b4f73ff2ff491f9dff7524de1755baf2?Expires=1734912000&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4&Signature=JxwQ6TtGoE48HBsxpABrZu4GimuBm0HGZK~23eZ~IHRpM27EYGSZpYoo26o6b6uSA3eoY5oyMkwzv0Z2VkdCYvXzRILYIolGcYo6DeUTbXeZRXAGZnAcVU-MqArnRR3cb4Hj7KPpN9X0UudO4k~QMIjt11FOuAstJjFuI~wx3fIisp~YCtilCjqM0zhyZun2gSPqhyeU4NTNAuKNYnx2En-4xj~5CMOI1Pv6yFwH91HWGFLgR-BaGt5XYhVw2OOpfyThK83dfoPJn3SrhJ0A~67f-JqsRr0SJmwaoHqkdCAx3hPIoatpKZNZc~AM25Wi0XvxLfiIDV73J6pBOGOiBg__); aspect-ratio: ${imageAspectRatioForCalc};`"),

					Div().Attr("x-ref", "rightImage").
						Class("bg-cover absolute top-[50%] left-[50%] translate-x-[-50%] translate-y-[-50%] transition-all bg-center bg-no-repeat flex-shrink-0 xl:w-[500px] md:w-[314px] w-[169px]").
						Attr(":style", "`background-image: url(https://s3-alpha-sig.figma.com/img/90f2/44c4/b4f73ff2ff491f9dff7524de1755baf2?Expires=1734912000&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4&Signature=JxwQ6TtGoE48HBsxpABrZu4GimuBm0HGZK~23eZ~IHRpM27EYGSZpYoo26o6b6uSA3eoY5oyMkwzv0Z2VkdCYvXzRILYIolGcYo6DeUTbXeZRXAGZnAcVU-MqArnRR3cb4Hj7KPpN9X0UudO4k~QMIjt11FOuAstJjFuI~wx3fIisp~YCtilCjqM0zhyZun2gSPqhyeU4NTNAuKNYnx2En-4xj~5CMOI1Pv6yFwH91HWGFLgR-BaGt5XYhVw2OOpfyThK83dfoPJn3SrhJ0A~67f-JqsRr0SJmwaoHqkdCAx3hPIoatpKZNZc~AM25Wi0XvxLfiIDV73J6pBOGOiBg__); aspect-ratio: ${imageAspectRatio};`"),
				).Class("xl:w-[500px] xl:min-h-[400px] md:min-h-[251px] min-h-[135px] relative md:w-[314px] w-[169px] overflow-hidden flex-shrink-0 xl:ml-10 md:ml-[20px] ml-3"),
			).Class("flex items-stretch justify-between xl:max-w-[1280px] mx-auto xl:p-[120px] md:p-[60px] p-8").Attr("x-data", `{
            imageAspectRatio: '5 / 4',
            imageAspectRatioForCalc: '5 / 4',
            debounceTimer: null,
            debounceAdjust() {
              clearTimeout(this.debounceTimer);
              this.debounceTimer = setTimeout(() => {
                this.adjustAspectRatio();
              }, 500);
            },
            adjustAspectRatio() {
              this.$nextTick(() => {
                const leftHeight = this.$refs.leftContent.offsetHeight;
                const imageHeightForCalc = this.$refs.rightImageForCalc.offsetHeight;
                
                this.imageAspectRatioForCalc = '5 / 4';

                // console.log(leftHeight, imageHeightForCalc);

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
          }`),
		).Class("xl:min-h-[530px] md:min-h-[339px] min-h-[190px] bg-white text-[#212121] bg-no-repeat bg-cover").Style("background-image: url(https://s3-alpha-sig.figma.com/img/2091/847f/23aed5f700c8474f29fededd296394ba?Expires=1734912000&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4&Signature=LmpcBano56oNL0jVh6IHDRhAMCfMVmM2N6yEMlo-TCcBlKIdB3eYBOLkOCZWY~arsOiotjZN~yByJbLBXEZOG3b5JTWyXMjcE6CuONVBTimfHftQEqLPssW4UjIhEc1Xyl~BMt~5BXD4vg8fNz5qb-pUOqQEcwXDMayVI7hf3gGIqjDkpqDB0Xb8~v6-sZsbP-AxOYZg65eKECq188sWftjRsqmt85WIJvcNjtiSQvdd0fCQz3fbft535yMNpzsX6c9M9L7~H3ySGlxFIR-qj-02Z-8Lcyu-GgKFf~p31u-qXzqTMr5tcRz6VVQ1bIraAvF0sCK~8k1nmWNQHUV7HQ__)")).Class("tailwind-scope"),
	).Class("container-hero-inner")

	body = tailwindContainerWrapper(
		"container-hero",
		Div(heroBody).Class("bg-gray-100"),
	)
	return
}
