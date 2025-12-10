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
	section := presets.NewSectionBuilder(cust, "Details").
		Editing("Name", "Email", "Description", "Avatar")
	dp.Section(section)
	cust.Editing("Details").Section(section.Clone())

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
	section := presets.NewSectionBuilder(cust, "Details").
		Editing("Name").ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VSwitch().Label("Prevent components covering the Edit button and making it unclickable").Color("primary").
			Attr(web.VField(fmt.Sprintf("%s.%s", field.FormKey, "Invalid"), true)...).
			Density(v.DensityCompact).
			Readonly(true)
	})
	dp.Section(section)

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

	nameSection := presets.NewSectionBuilder(cust, "name").
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
	dp.Section(nameSection)

	emailSection := presets.NewSectionBuilder(cust, "email").
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
	dp.Section(emailSection)

	tb := presets.NewTabsFieldBuilder()
	dp.Field("tabs").Tab(tb).AppendTabs(dp.Field("name")).AppendTabs(dp.Field("email"))

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

	sectionName := presets.NewSectionBuilder(cust, nameSection).Label("name_label").
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
	})
	sectionEmail := presets.NewSectionBuilder(cust, emailSection).
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
	dp.Section(sectionEmail).Section(sectionName)
	tb := presets.NewTabsFieldBuilder().
		TabsOrderFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) []string {
			return []string{emailSection, nameSection}
		})
	dp.Field("tabs").Tab(tb).AppendTabs(dp.Field(nameSection)).AppendTabs(dp.Field(emailSection))
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
	sectionDetail := presets.NewSectionBuilder(cust, "Details").
		Editing(&presets.FieldsSection{
			Title: "FieldSectionTitle",
			Rows: [][]string{
				{"Name", "Email"},
				{"Description"},
			},
		}, "Avatar")

	sectionDetail.EditingField("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Input("").Attr(web.VField("Details."+field.Name, field.Value(obj))...)
	})

	sectionDetail.ViewingField("Email").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Strong(obj.(*Customer).Email)
	})

	section2 := presets.NewSectionBuilder(cust, "section").Label("SectionTitle").
		Editing([]string{"Name", "Email"}).
		HiddenFuncs(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			return h.Div(
				h.Input("").Type("hidden").
					Attr(web.VField("TestHiddenFunc", "This is hidden input")...),
			)
		})

	dp.Section(sectionDetail, section2)
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
	cust.Editing().WrapFetchFunc(func(in presets.FetchFunc) presets.FetchFunc {
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
	section1 := presets.NewSectionBuilder(cust, "section1").Label("section_with_label").Editing("Name")
	section2 := presets.NewSectionBuilder(cust, "section2").Label("section_without_label").DisableLabel().Editing("Email").Viewing("Email")
	creditCardssection := presets.NewSectionBuilder(cust, "CreditCards").Label("section_list_with_label").IsList(&CreditCard{}).
		Editing("Name").
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().VField(fmt.Sprintf("%s.Name", field.FormKey), card.Name)
		}).
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().Text(card.Name)
		})

	notesSection := presets.NewSectionBuilder(cust, "Notes").Label("section_list_without_label").IsList(&Note{}).DisableLabel().
		Editing("Content").
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			note := obj.(*Note)
			return vx.VXTextField().VField(fmt.Sprintf("%s.Name", field.FormKey), note.Content)
		}).
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			note := obj.(*Note)
			return vx.VXTextField().Text(note.Content)
		})
	dp.Section(section1, section2, creditCardssection, notesSection)
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
	cust.Editing("name_section", "email_section").ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		customer := obj.(*Customer)
		if len(customer.Name) > 6 {
			err.FieldError("Name", "customer name must no longer than 6")
		}
		if len(customer.Name) > 20 {
			err.GlobalError("customer name must no longer than 20")
		}
		if customer.Name == "" {
			err.GlobalError("customer name must not be empty")
		}
		if customer.Email == "" {
			err.GlobalError("customer email must not be empty")
		}
		if len(customer.Email) < 6 {
			err.FieldError("Email", "customer email must longer than 6")
		}
		return
	})
	// This should inspect Notes attributes, When it is a list, It should show a standard table in detail page
	dp = cust.Detailing("name_section", "email_section").Drawer(true)
	nameSection := presets.NewSectionBuilder(cust, "name_section").Label("name must not be empty, no longer than 6").
		Editing("Name").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			customer := obj.(*Customer)
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return v.VTextField().
				Variant(v.VariantOutlined).
				Density(v.DensityCompact).
				Attr(presets.VFieldError("Name", customer.Name, vErr.GetFieldErrors("Name"))...)
		}).
		WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				customer := obj.(*Customer)
				if len(customer.Name) > 6 {
					err.FieldError("Name", "customer name must no longer than 6")
				}
				return err
			}
		})

	emailSection := presets.NewSectionBuilder(cust, "email_section").
		Label("email must not be empty, must longer than 6").
		Editing("Email").
		EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			customer := obj.(*Customer)
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			return v.VTextField().
				Variant(v.VariantOutlined).
				Density(v.DensityCompact).
				Attr(presets.VFieldError("Email", customer.Email, vErr.GetFieldErrors("Email"))...)
		})

	dp.Section(nameSection, emailSection)
	cust.Editing("name_section", "email_section")
	cust.Editing().Section(nameSection.Clone(), emailSection.Clone())
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

	detailSection := presets.NewSectionBuilder(mb, "Detail").Editing("Name", "Phone")
	ccmb.Detailing("Detail").Drawer(true).Section(detailSection)

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
	cardsSection1 := presets.NewSectionBuilder(cust, "CreditCards").Label("cards").IsList(&CreditCard{}).AlwaysShowListLabel().
		Editing("Name", "Phone").Viewing("Name", "Phone")
	cardsSection1.HiddenFuncs(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(
			h.Input("").Type("hidden").
				Attr(web.VField("TestHiddenFunc", "This is hidden input")...),
		)
	})
	cardsSection1.WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			if in != nil {
				err = in(obj, ctx)
			}
			user := obj.(*UserCreditCard)
			for i, card := range user.CreditCards {
				if card.Name == "" {
					err.FieldError(fmt.Sprintf("CreditCards[%d].Name", i), "card name must not be empty")
				}
				if len(card.Name) > 10 {
					err.FieldError(fmt.Sprintf("CreditCards[%d].Name", i), "card name must not exceed 10 characters")
				}
			}
			return
		}
	})

	cardsSection2 := presets.NewSectionBuilder(cust, "CreditCards2").Label("cards2").IsList(&CreditCard{}).AlwaysShowListLabel().
		Editing("Name", "Phone").Viewing("Name", "Phone")
	cardsSection2.WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			if in != nil {
				err = in(obj, ctx)
			}
			user := obj.(*UserCreditCard)
			for i, card := range user.CreditCards2 {
				if card.Phone == "" {
					err.FieldError(fmt.Sprintf("CreditCards2[%d].Phone", i), "card phone must not be empty")
				}
			}
			return
		}
	})
	dp.Section(cardsSection1, cardsSection2)
	return
}
