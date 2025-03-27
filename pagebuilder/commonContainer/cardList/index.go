package CardList

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/samber/lo"
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
			p.Style.ProductColumns = 3
			p.Content.Title = "This is a title"
			p.Content.Products = []*Product{
				{
					Title:       "Commerce",
					Description: "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable",
				},
				{
					Title:       "Commerce",
					Description: "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable",
				},
				{
					Title:       "Commerce",
					Description: "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable",
				},
			}
			p.Style.TopSpace = 40
			p.Style.BottomSpace = 40
			p.Style.LeftSpace = 20
			p.Style.RightSpace = 20
			p.Style.Visibility = []string{"title", "image", "productTitle", "description"}
			p.Style.ImageRatio = "1:1"

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

func imgList(data *CardList) HTMLComponent {
	image := "https://placehold.co/500x500"

	imgList := []HTMLComponent{}
	for _, product := range data.Content.Products {
		imgList = append(imgList, A(
			If(lo.Contains(data.Style.Visibility, "image"),
				Div(
					Img(func() string {
						if product.Image.URL() == "" {
							return image
						}
						return product.Image.URL()
					}()).Class("position-absolute w-full object-cover object-center h-full max-w-full left-0 top-0"),
				).Class(fmt.Sprintf("position-relative aspect-[%s] tw-theme-image-radius overflow-hidden mb-4 cc-image", data.Style.ImageRatio)),
			),
			If(lo.Contains(data.Style.Visibility, "productTitle"),
				H2("").Children(RawHTML(product.Title)).Class("tw-theme-text mb-2 cc-h2 text-bold xl:text-xl xl:leading-6 md:text-xl md:leading-7")),
			If(lo.Contains(data.Style.Visibility, "description"),
				Div(RawHTML(product.Description)).
					Class("tw-theme-text cc-content xl:text-base xl:leading-6 text-[14px] leading-[20px]")),
		).Href(product.Href),
		)
	}
	return Div(imgList...).Class(fmt.Sprintf("grid grid-cols-2 lg:grid-cols-%d gap-x-4 gap-y-8", data.Style.ProductColumns))
}

func CardListBody(data *CardList, input *pagebuilder.RenderInput) (body HTMLComponent) {
	heroBody := Div(
		Div(
			If(lo.Contains(data.Style.Visibility, "title"), H1("").Children(RawHTML(data.Content.Title)).
				Class("tw-theme-text cc-h1 text-center font-medium text-5xl leading-none mb-[60px]")),
			imgList(data),
		).Class(fmt.Sprintf("pl-[%dpx] pr-[%dpx] pt-[%dpx] pb-[%dpx] xl:w-[1280px] m-auto", data.Style.LeftSpace, data.Style.RightSpace, data.Style.TopSpace, data.Style.BottomSpace)),
	).Class("bg-no-repeat bg-cover bg-center cc-wrapper")

	body = utils.TailwindContainerWrapper(
		"container-hero",
		Tag("twind-scope").Attr("data-props", fmt.Sprintf(`{"type":"container-card-list", "id": %q}`, input.ContainerId)).Children(Div(heroBody)),
	)
	return
}
