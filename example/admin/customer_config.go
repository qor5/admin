package admin

import (
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

func configNestedFieldDemo(b *presets.Builder, db *gorm.DB) {
	cust := b.Model(&models.Customer{}).RightDrawerWidth("50%").
		Label("NestedFieldDemos").URIName("nested-field-demos")

	addFb := b.NewFieldsBuilder(presets.WRITE).Model(&models.Address{}).Only("Street", "HomeImage", "Phones")

	addFb.Field("HomeImage").WithContextValue(media.MediaBoxConfig, &media_library.MediaBoxConfig{
		AllowType: "image",
		Sizes: map[string]*base.Size{
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

	phoneFb := b.NewFieldsBuilder(presets.WRITE).Model(&models.Phone{}).Only("Number")
	addFb.Field("Phones").Nested(phoneFb, &presets.DisplayFieldInSorter{Field: "Number"})
	ed := cust.Editing("Name", "Addresses", "MembershipCard")
	ed.Field("Addresses").Nested(addFb, &presets.DisplayFieldInSorter{Field: "Street"})

	cardFb := b.NewFieldsBuilder(presets.WRITE).Model(&models.MembershipCard{}).Only("Number", "ValidBefore")
	ed.Field("MembershipCard").Nested(cardFb)

	ed.FetchFunc(func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
		return gorm2op.DataOperator(db.Preload("Addresses.Phones").Preload("MembershipCard")).Fetch(obj, id, ctx)
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
