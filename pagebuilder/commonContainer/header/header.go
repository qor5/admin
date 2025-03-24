package header

import (
	"fmt"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type TailWindExampleHeader struct {
	ID uint

	Content headerContent
	Style   headerStyle
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

	ed := vb.Model(&TailWindExampleHeader{}).Editing("Tabs", "Content", "Style")

	ed.Field("Tabs").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		option := presets.TabsControllerOption{
			Tabs: []presets.TabControllerOption{
				{Tab: v.VTab().Text("Content"), Fields: []string{"Content"}},
				{Tab: v.VTab().Text("Style"), Fields: []string{"Style"}},
			},
		}
		return presets.TabsController(field, &option)
	})
	// vb.Model(&TailWindExampleHeader{})

	ed.Creating().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			p := obj.(*TailWindExampleHeader)
			p.Content.Href1 = "What we do"
			p.Content.Href2 = "Projects"
			p.Content.Href3 = "Why clients choose us"
			p.Content.Href4 = "Our company"
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

func headerBody(data *TailWindExampleHeader, input *pagebuilder.RenderInput) (body HTMLComponent) {
	logoUrl := data.Content.ImageUpload.URL()

	html := Div(
		Header(
			Div(
				// Span("Logo").Class("logo text-3xl cc-content font-bold tw-theme-text"),
				Img("").Src(logoUrl).Class("logo-image logo mr-3 shrink-0").Attr("width", "200"),
				Ul(
					Li(A(Text(data.Content.Href1)).Href("#").Class("cc-content")),
					Li(A(Text(data.Content.Href2)).Href("#").Class("cc-content")),
					Li(A(Text(data.Content.Href3)).Href("#").Class("cc-content")),
					Li(A(Text(data.Content.Href4)).Href("#").Class("cc-content")),
				).Class("list-none flex tw-theme-text text-[28px] gap-[62px] font-[500]").
					Attr("data-list-unset", "true"),
			).Class("w-[1152px] m-0 mx-auto leading-9 flex justify-between items-center"),
		).Class(" py-12 tw-theme-bg-base"),
	).Class("container-tailwind-inner")

	body = utils.TailwindContainerWrapper(
		"container-tailwind-example-header",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"container-tailwind-example-header", "id": %q}`, input.ContainerId)).Children(html),
	)
	return
}
