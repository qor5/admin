package containers

import (
	"fmt"

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
			p.Content.Text = `<h1
                class="text-3xl md:text-2xl text-black opacity-80 font-bold mb-1"
              >
                Testimonial introduction copy goes here, lorem ipsum dolor sit
                amet.
              </h1>
              <p class="text-black opacity-80 font-medium sm:text-sm">
                “Lorem ipsum dolor sit amet, consectetur adipiscing elit.
                suspendisse tincidunt sagitis eros. Quisque quis euismod lorem"
              </p>`

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
	imgUrl := data.Content.ImageUpload.URL()
	if data.Content.ImageUpload.URL() == "" {
		imgUrl = "https://via.placeholder.com/308x252"
	}

	heroBody := Div(
		Div(
			Div(
				Div(
					Div(
						Div(
						// Use a div with background image instead of Img
						).Style(fmt.Sprintf("background-image: url('%s');", imgUrl)).
							Class("w-full", "h-full", "bg-cover", "bg-center", "aspect-[3/4]"),
					).Class("lg:w-auto", "md:w-auto", "flex", "justify-start", "h-full"),
					Div(
						Div(
							If(data.Content.Text != "", Div(RawHTML(data.Content.Text)),
								If(data.Content.Text == "<p></p>", Div(
									H1("Testimonial introduction copy goes here, lorem ipsum dolor sit amet.").Class("text-3xl", "md:text-2xl", "text-black", "opacity-80", "font-bold", "mb-1"),
									P(Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. suspendisse tincidunt sagitis eros. Quisque quis euismod lorem")).Class("text-black", "opacity-80", "font-medium", "sm:text-sm"),
								)),
							),
						),
						Div(
							H2("Author Name").Class("text-2xl", "md:text-xl", "text-black", "opacity-80", "font-bold", "mb-2", "lg:mt-0", "md:mt-0", "mt-8"),
							P(Text("Co-Founder and CEO of Company")).Class("text-pretty", "text-black", "opacity-80", "font-medium", "sm:text-sm"),
						),
					).Class("lg:flex", "md:flex", "flex-col", "lg:h-[252px]", "md:h-[183px]", "justify-between", "lg:flex-1", "md:w-full", "text-left", "lg:pl-20", "md:pl-10", "lg:mt-0", "md:mt-0", "lg:text-left", "md:text-left", "text-center", "mt-8"),
				).Class("mx-auto", "max-w-5xl", "lg:flex", "md:flex", "items-center", "justify-center", "h-full", "lg:px-14", "lg:py-0", "md:py-0", "p-7"),
			).Class("flex-wrap", "lg:h-[400px]", "md:h-[300px]"),
		).Class("tailwind-scope"),
	).Class("container-hero-inner")

	body = tailwindContainerWrapper(
		"container-hero",
		Div(heroBody).Class("bg-gray-100"),
	)
	return
}
