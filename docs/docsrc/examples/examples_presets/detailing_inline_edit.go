package examples_presets

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
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

func PresetsDetailSectionView(b *presets.Builder, db *gorm.DB) (
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
		Editing("Name").ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VSwitch().Label("Prevent components covering the Edit button and making it unclickable").Color("primary").
			Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "Invalid"), true)...).
			Density(v.DensityCompact).
			Readonly(true)
	})

	return
}

func PresetsDetailTabsSection(b *presets.Builder, db *gorm.DB) (
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
	dp = cust.Detailing("tabs").Drawer(true)

	tb := presets.NewTabsFieldBuilder()
	dp.Field("tabs").Tab(tb)

	dp.Section("name").
		Editing("Name").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			custom := obj.(*Customer)
			return h.Div(
				v.VTextField().Attr(web.VField("name.Name", custom.Name)...).Label("Name"),
			)
		}).ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		custom := obj.(*Customer)
		return h.Div(
			h.Text(custom.Name),
		)
	})

	dp.Section("email").
		Editing("Email").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			custom := obj.(*Customer)
			return h.Div(
				v.VTextField().Attr(web.VField("email.Email", custom.Email)...).Label("Email"),
			)
		}).ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		custom := obj.(*Customer)
		return h.Div(
			h.Text(custom.Email),
		)
	})

	dp.Section("name").Tabs("tabs")
	dp.Section("email").Tabs("tabs")

	return
}

func PresetsDetailTabsSectionOrder(b *presets.Builder, db *gorm.DB) (
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
	dp = cust.Detailing("tabs").Drawer(true)

	const (
		nameSection  = "name"
		emailSection = "email"
	)
	tb := presets.NewTabsFieldBuilder().
		TabsOrderFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) []string {
			return []string{emailSection, nameSection}
		})
	dp.Field("tabs").Tab(tb)

	dp.Section(nameSection).Label("name_label").
		Editing("Name").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			custom := obj.(*Customer)
			key := fmt.Sprintf("%s.Name", nameSection)
			return h.Div(
				v.VTextField().Attr(web.VField(key, custom.Name)...).Label("Name"),
			)
		}).ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		custom := obj.(*Customer)
		return h.Div(
			h.Text(custom.Name),
		)
	}).Tabs("tabs")

	dp.Section(emailSection).
		Editing("Email").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			custom := obj.(*Customer)
			return h.Div(
				v.VTextField().Attr(web.VField("email.Email", custom.Email)...).Label("Email"),
			)
		}).ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		custom := obj.(*Customer)
		return h.Div(
			h.Text(custom.Email),
		)
	}).Tabs("tabs")

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

	type i18nMessage struct {
		CustomersFieldSectionTitle string
		CustomersSectionTitle      string
		CustomersSectionEN         string
	}
	b.GetI18n().SupportLanguages(language.English, language.Japanese).
		RegisterForModule(language.English, presets.ModelsI18nModuleKey, i18nMessage{
			CustomersFieldSectionTitle: "Field_sectionEN",
			CustomersSectionTitle:      "SectionEN",
			CustomersSectionEN:         "Wrong",
		}).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, i18nMessage{
			CustomersFieldSectionTitle: "Field_sectionJP",
			CustomersSectionTitle:      "SectionJP",
			CustomersSectionEN:         "Wrong",
		})

	cust = b.Model(&Customer{})
	dp = cust.Detailing("Details", "section").Drawer(true)
	sb := dp.Section("Details").
		Editing(&presets.FieldsSection{
			Title: "FieldSectionTitle",
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

	dp.Section("section").Label("SectionTitle").
		Editing([]string{"Name", "Email"})
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
	// section will use Editing().ValidateFunc() as validateFunc default
	cust.Editing().ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		customer := obj.(*Customer)
		if len(customer.Name) > 6 {
			err.FieldError("name_section.Name", "customer name must no longer than 6")
		}
		if len(customer.Name) > 20 {
			err.GlobalError("customer name must no longer than 20")
		}
		return
	})
	// This should inspect Notes attributes, When it is a list, It should show a standard table in detail page
	dp = cust.Detailing("name_section", "email_section", "CreditCards").Drawer(true)
	dp.Section("name_section").Label("name must not be empty, no longer than 6").
		Editing("Name").EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		customer := obj.(*Customer)
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}
		return v.VTextField().
			Variant(v.VariantOutlined).
			Density(v.DensityCompact).
			Attr(web.VField("name_section.Name", customer.Name)...).
			ErrorMessages(vErr.GetFieldErrors("name_section.Name")...)
	}).WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return in
	}).WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			err = in(obj, ctx)
			customer := obj.(*Customer)
			if customer.Name == "" {
				err.GlobalError("customer name must not be empty")
			}
			return err
		}
	})

	dp.Section("email_section").
		Label("email must not be empty, must longer than 6").
		Editing("Email").
		ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			customer := obj.(*Customer)
			if customer.Email == "" {
				err.GlobalError("customer email must not be empty")
			}
			if len(customer.Email) < 6 {
				err.FieldError("email_section.Email", "customer email must longer than 6")
			}
			return
		}).EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		customer := obj.(*Customer)
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}
		return v.VTextField().
			Variant(v.VariantOutlined).
			Density(v.DensityCompact).
			Attr(web.VField("email_section.Email", customer.Name)...).
			ErrorMessages(vErr.GetFieldErrors("email_section.Email")...)
	}).SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		return cust.Editing().Saver(obj, id, ctx)
	})

	dp.Section("CreditCards").IsList(&CreditCard{}).Editing("Name").
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			var errText []string
			if vErr, ok := ctx.Flash.(*web.ValidationErrors); ok {
				errText = vErr.GetFieldErrors("card")
			}
			return vx.VXField().ErrorMessages(errText...).
				Attr(web.VField(fmt.Sprintf("%s.Name", field.FormKey), card.Name)...)
		}).
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().Text(card.Name)
		}).ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		customer := obj.(*Customer)
		for index, card := range customer.CreditCards {
			if card.Name == "" {
				err.GlobalError("credit card name must not be empty")
			}
			if card.Name == "" {
				err.FieldError("card", fmt.Sprintf("credit card %d name must not be empty", index))
			}
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
	ccmb.Detailing("Detail").Drawer(true).Section("Detail").Editing("Name", "Phone")

	ccmb2 := mb.NestedMany(&CreditCard{}, "CustomerID")
	// force ignore ExpireYearMonth column if you need
	ccmb2.Listing().WrapColumns(func(in presets.ColumnsProcessor) presets.ColumnsProcessor {
		return func(evCtx *web.EventContext, columns []*presets.Column) ([]*presets.Column, error) {
			columns, err := in(evCtx, columns)
			if err != nil {
				return nil, err
			}

			// You can get the current state of the listing compo this way, if you need.
			listCompo := presets.ListingCompoFromContext(evCtx.R.Context())
			log.Printf("ParentID: %v", listCompo.ParentID)

			for _, v := range columns {
				if v.Name == "ExpireYearMonth" {
					v.Visible = false
				}
			}
			return columns, nil
		}
	})
	// You can also wrap the table if you need
	ccmb2.Listing().WrapTable(func(in presets.TableProcessor) presets.TableProcessor {
		return func(evCtx *web.EventContext, table *vx.DataTableBuilder) (*vx.DataTableBuilder, error) {
			table.Hover(false)
			return in(evCtx, table)
		}
	})

	dp.Field("CreditCards2").Use(ccmb2)
	return
}

type UserCreditCard struct {
	gorm.Model
	Name         string
	CreditCards  creditCards `gorm:"type:text"`
	CreditCards2 creditCards `gorm:"type:text"`
}
type creditCards []*CreditCard

func (creditCard creditCards) Value() (driver.Value, error) {
	json, err := json.Marshal(creditCard)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (creditCard *creditCards) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		return json.Unmarshal(values, &creditCard)
	case string:
		return creditCard.Scan([]byte(values))
	}
	return nil
}

func PresetsDetailListSection(b *presets.Builder, db *gorm.DB) (cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&UserCreditCard{})
	if err != nil {
		panic(err)
	}
	b.DataOperator(gorm2op.DataOperator(db))
	cust = b.Model(&UserCreditCard{})
	dp = cust.Detailing("CreditCards", "CreditCards2").Drawer(true)
	dp.WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (r interface{}, err error) {
			o, _ := in(obj, id, ctx)
			us := o.(*UserCreditCard)
			if len(us.CreditCards2) == 0 {
				us.CreditCards2 = nil
			}
			return us, nil
		}
	})
	dp.Section("CreditCards").Label("cards").IsList(&CreditCard{}).AlwaysShowListLabel().
		Editing("Name", "Phone").Viewing("Name", "Phone")

	dp.Section("CreditCards2").Label("cards2").IsList(&CreditCard{}).AlwaysShowListLabel().
		Editing("Name", "Phone").Viewing("Name", "Phone")
	return
}
