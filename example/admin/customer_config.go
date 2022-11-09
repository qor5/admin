package admin

import (
	"github.com/qor5/web"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/listeditor"
	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	media_view "github.com/qor5/admin/media/views"
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
		return listeditor.New(field).Value(field.Value(obj)).DisplayFieldInSorter("Number")
	})

	ed := cust.Editing("Name", "Addresses")
	ed.ListField("Addresses", addFb).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return listeditor.New(field).Value(field.Value(obj)).DisplayFieldInSorter("Street")
	})

	ed.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		return gorm2op.DataOperator(db.Preload("Addresses.Phones")).Fetch(obj, id, ctx)
	})

	ed.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		c := obj.(*models.Customer)
		err = db.Delete(&models.Phone{}, "address_id IN (select id from addresses where customer_id = ?)", c.ID).Error
		if err != nil {
			panic(err)
		}
		err = db.Delete(&models.Address{}, "customer_id = ?", c.ID).Error
		if err != nil {
			panic(err)
		}
		return gorm2op.DataOperator(db.Session(&gorm.Session{FullSaveAssociations: true})).Save(obj, id, ctx)
	})
}
