package examples_presets

import (
	"fmt"
	"log"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
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

func PresetsDetailSectionLabel(b *presets.Builder, db *gorm.DB) (
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
	dp = cust.Detailing("section1", "section2", "CreditCards", "Notes").Drawer(true)
	cust.Detailing().WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			c := obj.(*Customer)
			if c.CreditCards == nil {
				c.CreditCards = []*CreditCard{{Name: "Only is mock card, can't be save"}}
			}
			if c.Notes == nil {
				c.Notes = []*Note{{Content: "Only is mock note, can't be save"}}
			}
			return c, nil
		}
	})
	dp.Section("section1").Label("section_with_label").Editing("Name")
	dp.Section("section2").Label("section_without_label").DisableLabel().Editing("Email")
	dp.Section("CreditCards").Label("section_list_with_label").IsList(&CreditCard{}).
		Editing("Name").
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().VField(fmt.Sprintf("%s.Name", field.FormKey), card.Name)
		}).
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().Text(card.Name)
		})
	dp.Section("Notes").Label("section_list_without_label").IsList(&Note{}).DisableLabel().
		Editing("Content").
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			note := obj.(*Note)
			return vx.VXTextField().VField(fmt.Sprintf("%s.Name", field.FormKey), note.Content)
		}).
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			note := obj.(*Note)
			return vx.VXTextField().Text(note.Content)
		})

	return
}

func PresetsDetailInlineEditValidate(b *presets.Builder, db *gorm.DB) (
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
	dp = cust.Detailing("name_section").Drawer(true)
	dp.Section("name_section").Label("name must not be empty").Editing("Name").Viewing("Name").Validator(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		customer := obj.(*Customer)
		if customer.Name == "" {
			err.GlobalError("customer name must not be empty")
		}
		return
	})

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
	ccmb2.Listing().WrapDisplayColumns(func(in presets.DisplayColumnsProcessor) presets.DisplayColumnsProcessor {
		return func(evCtx *web.EventContext, displayColumns []*presets.DisplayColumn) ([]*presets.DisplayColumn, error) {
			displayColumns, err := in(evCtx, displayColumns)
			if err != nil {
				return nil, err
			}

			// You can get the current state of the listing compo this way, if you need.
			listCompo := presets.ListingCompoFromContext(evCtx.R.Context())
			log.Printf("ParentID: %v", listCompo.ParentID)

			for _, v := range displayColumns {
				if v.Name == "ExpireYearMonth" {
					v.Visible = false
				}
			}
			return displayColumns, nil
		}
	})

	dp.Field("CreditCards2").Use(ccmb2)
	return
}
