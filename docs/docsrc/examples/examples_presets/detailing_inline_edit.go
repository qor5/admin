package examples_presets

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func PresetsDetailSimple(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &ActivityNote{})
	if err != nil {
		panic(err)
	}
	mediaBuilder := media.New(db)
	b.DataOperator(gorm2op.DataOperator(db)).Use(mediaBuilder)

	cust = b.Model(&Customer{})
	dp = cust.Detailing("Name", "Email", "Description", "Avatar").Drawer(true)

	return
}

func PresetsDetailInlineEditDetails(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &ActivityNote{})
	if err != nil {
		panic(err)
	}
	mediaBuilder := media.New(db)
	b.DataOperator(gorm2op.DataOperator(db)).Use(mediaBuilder)

	cust = b.Model(&Customer{})
	dp = cust.Detailing("Details").Drawer(true)
	dp.Section("Details").
		Editing("Name", "Email", "Description", "Avatar")

	return
}

func PresetsDetailInlineEditFieldSections(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &ActivityNote{})
	if err != nil {
		panic(err)
	}
	mediaBuilder := media.New(db)
	b.DataOperator(gorm2op.DataOperator(db)).Use(mediaBuilder)

	cust = b.Model(&Customer{})
	dp = cust.Detailing("Details").Drawer(true)
	sb := dp.Section("Details").
		Editing(&presets.FieldsSection{
			Title: "Hello",
			Rows: [][]string{
				{"Name", "Email"},
				{"Description"},
			},
		}, "Avatar")

	sb.EditingField("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Input("").Attr(web.VField("Details."+field.Name, field.Value(obj))...)
	})

	sb.ViewingField("Email").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Strong(obj.(*Customer).Email)
	})

	return
}

func PresetsDetailInlineEditInspectTables(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &ActivityNote{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	cust = b.Model(&Customer{})
	// This should inspect Notes attributes, When it is a list, It should show a standard table in detail page
	dp = cust.Detailing("CreditCards").Drawer(true)

	return
}

func PresetsDetailInlineEditDetailsInspectShowFields(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &ActivityNote{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	cust = b.Model(&Customer{})
	b.URIPrefix(examples.URLPathByFunc(PresetsDetailInlineEditDetailsInspectShowFields))
	dp = cust.Detailing("Details", "CreditCards").Drawer(true)
	dp.WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			var cus Customer
			db.Find(&cus)

			var cc []*CreditCard
			db.Find(&cc)
			cus.CreditCards = cc
			r = cus
			return
		}
	})
	dp.Section("Details").
		Editing("Name", "Email2", "Description")

	dp.Field("Email2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div().Text("abc")
	})

	ccm := b.Model(&CreditCard{}).InMenu(false)
	ccm.Editing("Number")
	l := ccm.Listing("Name")
	l.Field("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div()
	})
	return
}
