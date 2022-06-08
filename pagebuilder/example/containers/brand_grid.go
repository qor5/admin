package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/listeditor"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/pagebuilder"
	. "github.com/theplant/htmlgo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type BrandGrid struct {
	ID             uint
	AddTopSpace    bool
	AddBottomSpace bool
	AnchorID       string
	Brands         Brands `sql:"type:text;"`
}

type Brand struct {
	Image media_library.MediaBox `sql:"type:text;"`
	Name  string
}

type Brands []*Brand

func (this Brands) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *Brands) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func RegisterBrandGridContainer(pb *pagebuilder.Builder, db *gorm.DB) {
	vb := pb.RegisterContainer("BrandGrid").
		RenderFunc(func(obj interface{}, ctx *web.EventContext) HTMLComponent {
			v := obj.(*BrandGrid)
			return BrandGridBody(v)
		})
	mb := vb.Model(&BrandGrid{})
	listeditor.Configure(mb.GetModelBuilder())
	eb := mb.Editing("AddTopSpace", "AddBottomSpace", "AnchorID", "Brands")

	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&Brand{}).Only("Image", "Name")
	fb.Field("Image").WithContextValue(media_view.MediaBoxConfig, &media_library.MediaBoxConfig{
		AllowType: "image",
	})

	eb.ListField("Brands", fb).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return listeditor.New(field).Value(field.Value(obj)).DisplayFieldInSorter("Name")
	})
}

func BrandGridBody(data *BrandGrid) (body HTMLComponent) {
	body = ContainerWrapper(
		fmt.Sprintf("brand_grid_%v", data.ID), data.AnchorID, "container-brand_grid", "", "", "",
		data.AddTopSpace, data.AddBottomSpace,
		Div(
			BrandsBody(data.Brands),
		).Class("container-wrapper"),
	)
	return
}

func BrandsBody(brands []*Brand) HTMLComponent {
	brandsDiv := Div().Class("container-brand_grid-wrap")
	for _, b := range brands {
		brandsDiv.AppendChildren(
			Div(
				LazyImageHtml(b.Image),
			).Class("container-brand_grid-item"),
		)
	}
	return brandsDiv
}
