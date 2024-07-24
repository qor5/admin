package examples_presets

import (
	"log"

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
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &Note{})
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
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &Note{})
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
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &Note{})
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
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &Note{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	cust = b.Model(&Customer{})
	// This should inspect Notes attributes, When it is a list, It should show a standard table in detail page
	dp = cust.Detailing("CreditCards").Drawer(true)

	return
}

func PresetsDetailNestedMany(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Customer{}, &CreditCard{}, &Note{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&Customer{}).RightDrawerWidth("1000")
	dp = mb.Detailing("Name", "CreditCards", "CreditCards2").Drawer(true)

	ccmb := mb.NestedMany(&CreditCard{}, "CustomerID")
	dp.Field("CreditCards").Use(ccmb)

	ccmb2 := mb.NestedMany(&CreditCard{}, "CustomerID")
	// force ignore ExpireYearMonth column if you need
	ccmb2.Listing().DisplayColumnsProcessor(func(evCtx *web.EventContext, displayColumns []*presets.DisplayColumn) ([]*presets.DisplayColumn, error) {
		// You can get the current state of the listing compo this way, if you need.
		listCompo := presets.ListingCompoFromContext(evCtx.R.Context())
		log.Printf("ParentID: %v", listCompo.ParentID)

		for _, v := range displayColumns {
			if v.Name == "ExpireYearMonth" {
				v.Visible = false
			}
		}
		return displayColumns, nil
	})

	dp.Field("CreditCards2").Use(ccmb2)
	return
}
