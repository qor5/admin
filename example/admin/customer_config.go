package admin

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/listeditor"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configCustomer(b *presets.Builder, db *gorm.DB) {
	cust := b.Model(&models.Customer{}).RightDrawerWidth("50%")
	listeditor.Configure(cust)

	addFb := b.NewFieldsBuilder(presets.WRITE).Model(&models.Address{}).Only("Street", "HomeImage", "Phones")

	addFb.Field("HomeImage").WithContextValue(media_view.MediaBoxConfig, &media_library.MediaBoxConfig{
		AllowType: "image",
		Sizes: map[string]*media.Size{
			"thumb": {
				Width:  400,
				Height: 300,
			},
			"main": {
				Width:  800,
				Height: 500,
			},
		},
	})

	var phoneFb = b.NewFieldsBuilder(presets.WRITE).Model(&models.Phone{}).Only("Number")
	addFb.ListField("Phones", phoneFb).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return listeditor.New(field).Value(field.Value(obj))
	})

	ed := cust.Editing("Name", "Addresses")
	ed.ListField("Addresses", addFb).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return listeditor.New(field).Value(field.Value(obj))
	})

	ed.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		return gorm2op.DataOperator(db.Preload("Addresses.Phones")).Fetch(obj, id, ctx)
	})

	ed.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		return gorm2op.DataOperator(db.Session(&gorm.Session{FullSaveAssociations: true})).Save(obj, id, ctx)
	})
}
