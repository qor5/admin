package examples_presets

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/utils"
)

// @snippet_begin(PresetsDetailPageTopNotesSample)

type Note struct {
	ID         int
	SourceType string
	SourceID   int
	Content    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func addListener(_ *web.EventContext, v any) h.HTMLComponent {
	simpleReload := web.Plaid().MergeQuery(true).Go()
	return web.Listen(
		presets.NotifModelsCreated(v), simpleReload,
		presets.NotifModelsUpdated(v), simpleReload,
		presets.NotifModelsDeleted(v), simpleReload,
	)
}

func PresetsDetailPageTopNotes(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsEditingCustomizationValidation(b, db)
	err := db.AutoMigrate(&Note{})
	if err != nil {
		panic(err)
	}

	dp = cust.Detailing("TopNotes")

	dp.Field("TopNotes").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		cu := obj.(*Customer)

		title := cu.Name
		if len(title) == 0 {
			title = cu.Description
		}

		var notes []*Note
		err := db.Where("source_type = 'Customer' AND source_id = ?", cu.ID).
			Order("id DESC").
			Find(&notes).Error
		if err != nil {
			panic(err)
		}

		dt := vx.DataTable(notes).WithoutHeader(true).LoadMoreAt(2, "Show More")

		dt.Column("Content").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
			n := obj.(*Note)
			return h.Td(h.Div(
				h.Div(
					VIcon("mdi-message-reply-text").Color("blue").Size(SizeSmall).Class("pr-2"),
					h.Text(n.Content),
				).Class("body-1"),
				h.Div(
					h.Text(n.CreatedAt.Format("Jan 02,15:04 PM")),
					h.Text(" by Felix Sun"),
				).Class("grey--text pl-7 body-2"),
			).Class("my-3"))
		})

		cusID := fmt.Sprint(cu.ID)
		dt.RowMenuItemFuncs(presets.EditDeleteRowMenuItemFuncs(mi, mi.PresetsPrefix()+"/notes", url.Values{"model": []string{"Customer"}, "model_id": []string{cusID}})...)

		return vx.Card(dt).HeaderTitle(title).
			Actions(
				addListener(ctx, &Note{}),
				VBtn("Add Note").
					Attr("@click",
						web.POST().EventFunc(actions.New).
							Query("model", "Customer").
							Query("model_id", cusID).
							URL(mi.PresetsPrefix()+"/notes").
							Go(),
					),
			).Class("mb-4").Variant(VariantElevated)
	})

	b.Model(&Note{}).
		InMenu(false).
		Editing("Content").
		SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			note := obj.(*Note)
			note.SourceID = ctx.ParamAsInt("model_id")
			note.SourceType = ctx.R.FormValue("model")
		})
	return
}

// @snippet_end

// @snippet_begin(PresetsDetailPageDetailsSample)

func PresetsDetailPageDetails(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsDetailPageTopNotes(b, db)
	err := db.AutoMigrate(&CreditCard{})
	if err != nil {
		panic(err)
	}
	dp = cust.Detailing("TopNotes", "Details")
	dp.Field("Details").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		cu := obj.(*Customer)
		cusID := fmt.Sprint(cu.ID)

		var termAgreed string
		if cu.TermAgreedAt != nil {
			termAgreed = cu.TermAgreedAt.Format("Jan 02,15:04 PM")
		}

		detail := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(cu.Name).ZeroLabel("No Name")).Label("Name"),
				vx.DetailField(vx.OptionalText(cu.Email).ZeroLabel("No Email")).Label("Email"),
				vx.DetailField(vx.OptionalText(cusID).ZeroLabel("No ID")).Label("ID"),
				vx.DetailField(vx.OptionalText(cu.CreatedAt.Format("Jan 02,15:04 PM")).ZeroLabel("")).Label("Created"),
				vx.DetailField(vx.OptionalText(termAgreed).ZeroLabel("Not Agreed Yet")).Label("Terms Agreed"),
			).Header("ACCOUNT INFORMATION"),
			vx.DetailColumn(
				vx.DetailField(h.RawHTML(cu.Description)).Label("Description"),
			).Header("DETAILS"),
		)

		return vx.Card(detail).HeaderTitle("Details").Variant(VariantElevated).
			Actions(
				web.Listen(
					cust.NotifModelsUpdated(), web.Plaid().MergeQuery(true).Go(),
				),
				VBtn("Agree Terms").
					Class("mr-2").
					Attr("@click", web.POST().
						EventFunc(actions.Action).
						Query(presets.ParamAction, "AgreeTerms").
						Query(presets.ParamID, cusID).
						Go(),
					),

				VBtn("Update details").
					Attr("@click", web.POST().
						EventFunc(actions.Edit).
						Query(presets.ParamOverlay, actions.Dialog).
						Query(presets.ParamID, cusID).
						URL(mi.PresetsPrefix()+"/customers").
						Go(),
					),
			).Class("mb-4")
	})

	dp.Action("AgreeTerms").UpdateFunc(func(id string, ctx *web.EventContext, r *web.EventResponse) (err error) {
		if ctx.R.FormValue("Agree") != "true" {
			ve := &web.ValidationErrors{}
			ve.GlobalError("You must agree the terms")
			err = ve
			return
		}

		err = db.Model(&Customer{}).Where("id = ?", id).
			Updates(map[string]interface{}{"term_agreed_at": time.Now()}).Error
		if err == nil {
			r.Emit(presets.NotifModelsUpdated(&Customer{}), presets.PayloadModelsUpdated{Ids: []string{id}})
			presets.ShowMessage(r, "Terms agreed", ColorSuccess)
		}
		return
	}).ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
		var alert h.HTMLComponent

		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			alert = VAlert(h.Text(ve.GetGlobalError())).Border("start").
				Type("error").
				Elevation(2)
		}

		var agreedAt *time.Time
		db.Model(&Customer{}).Select("term_agreed_at").Where("id = ?", id).Scan(&agreedAt)

		return h.Components(
			alert,
			VCheckbox().Attr(web.VField("Agree", agreedAt != nil && !agreedAt.IsZero())...).Label("Agree the terms"),
		)
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsDetailPageCardsSample)

type CreditCard struct {
	ID              int
	CustomerID      int
	Number          string
	ExpireYearMonth string
	Name            string
	Type            string
	Phone           string
	Email           string
}

func PresetsDetailPageCards(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsDetailPageDetails(b, db)
	err := db.AutoMigrate(&CreditCard{})
	if err != nil {
		panic(err)
	}

	dp = cust.RightDrawerWidth("800").Detailing("TopNotes", "Details", "Cards").Drawer(true)

	dp.Field("Cards").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		cu := obj.(*Customer)
		cusID := fmt.Sprint(cu.ID)

		var cards []*CreditCard
		err := db.Where("customer_id = ?", cu.ID).Order("id ASC").Find(&cards).Error
		if err != nil {
			panic(err)
		}

		dt := vx.DataTable(cards).
			WithoutHeader(true).
			RowExpandFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
				card := obj.(*CreditCard)
				return vx.DetailInfo(
					vx.DetailColumn(
						vx.DetailField(vx.OptionalText(card.Name).ZeroLabel("No Name")).Label("Name"),
						vx.DetailField(vx.OptionalText(card.Number).ZeroLabel("No Number")).Label("Number"),
						vx.DetailField(vx.OptionalText(card.ExpireYearMonth).ZeroLabel("No Expires")).Label("Expires"),
						vx.DetailField(vx.OptionalText(card.Type).ZeroLabel("No Type")).Label("Type"),
						vx.DetailField(vx.OptionalText(card.Phone).ZeroLabel("No phone provided")).Label("Phone"),
						vx.DetailField(vx.OptionalText(card.Email).ZeroLabel("No email provided")).Label("Email"),
					),
				)
			}).RowMenuItemFuncs(presets.EditDeleteRowMenuItemFuncs(mi, mi.PresetsPrefix()+"/credit-cards", url.Values{"customerID": []string{cusID}})...)

		dt.Column("Type")
		dt.Column("Number")
		dt.Column("ExpireYearMonth")

		return vx.Card(dt).HeaderTitle("Cards").
			Actions(

				addListener(ctx, &CreditCard{}),
				VBtn("Add Card").
					Attr("@click",
						web.POST().
							EventFunc(actions.New).
							Query("customerID", cusID).
							Query(presets.ParamOverlay, actions.Dialog).
							URL(mi.PresetsPrefix()+"/credit-cards").
							Go(),
					).Class("mb-4"),
			)
	})

	cc := b.Model(&CreditCard{}).
		InMenu(false)

	ccedit := cc.Editing("ExpireYearMonth", "Phone", "Email").
		SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			card := obj.(*CreditCard)
			card.CustomerID = ctx.ParamAsInt("customerID")
		})

	ccedit.Creating("Number")
	return
}

const PresetsDetailPageCardsPath = "/samples/presets-detail-page-cards"

// @snippet_end

func PresetsDetailTabsField(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsHelloWorld(b, db)

	dp = cust.RightDrawerWidth("800").Detailing("tab").Drawer(true)

	dp.Only("Name", "Email", "tab")
	emailField := dp.NewFieldWithName("Email").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		customer := obj.(*Customer)
		text := fmt.Sprintf("This is Email Tabs, Email: %s", customer.Email)
		return h.Text(text)
	})
	nameField := dp.NewFieldWithName("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		customer := obj.(*Customer)
		text := fmt.Sprintf("This is Name Tabs, Email: %s", customer.Name)
		return h.Text(text)
	})
	dp.Field("tab").
		AppendTabs(nameField).
		AppendTabs(emailField)
	return
}

func PresetsDetailAfterTitle(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsHelloWorld(b, db)

	dp = cust.Detailing("Name")

	dp.AfterTitleCompFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return h.Text("After Title")
	})
	return
}

func PresetsDetailSidePanel(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsHelloWorld(b, db)

	dp = cust.Detailing().Drawer(true)

	dp.WrapSidePanel(func(in presets.ObjectComponentFunc) presets.ObjectComponentFunc {
		return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			return h.Div(
				in(obj, ctx),
				h.Text("Side Panel 1"),
			)
		}
	})

	dp.WrapSidePanel(func(in presets.ObjectComponentFunc) presets.ObjectComponentFunc {
		return func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			return h.Div(
				in(obj, ctx),
				h.Text("Side Panel 2"),
			)
		}
	})

	return
}

// Test helper: configure a detailing page with list section and constant field error for index 0
func PresetsDetailListSectionStatusxFieldViolations(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	// Ensure listing/search is available (prevents "function Searcher is not set")
	b.DataOperator(gorm2op.DataOperator(db))
	if err := db.AutoMigrate(&UserCreditCard{}); err != nil {
		panic(err)
	}
	mb = b.Model(&UserCreditCard{})
	dp = mb.Detailing("CreditCards").Drawer(true)

	section := presets.NewSectionBuilder(mb, "CreditCards").IsList(&CreditCard{}).
		Editing("Name").
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			var errs []string
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				errs = ve.GetFieldErrors(fmt.Sprintf("%s.Name", field.FormKey))
			}
			return vx.VXField().
				Attr(presets.VFieldError(fmt.Sprintf("%s.Name", field.FormKey), card.Name, errs)...)
		}).
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().Text(card.Name)
		})
	dp.Section(section)

	section.WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (vErr web.ValidationErrors) {
			if in != nil {
				ve := in(obj, ctx)
				_ = vErr.Merge(&ve)
			}
			card := obj.(*UserCreditCard)
			for i, c := range card.CreditCards {
				if c.Name == "" {
					vErr.FieldError(fmt.Sprintf("CreditCards[%d].Name", i), "name is required")
				}
			}
			return
		}
	})
	return
}

// Test helper: detailing page with list section; error only when name is empty
func PresetsDetailListSection_ItemStateIsolation(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	// Ensure listing/search is available (prevents "function Searcher is not set")
	b.DataOperator(gorm2op.DataOperator(db))
	if err := db.AutoMigrate(&UserCreditCard{}); err != nil {
		panic(err)
	}
	mb = b.Model(&UserCreditCard{})
	dp = mb.Detailing("CreditCards").Drawer(true)

	section := presets.NewSectionBuilder(mb, "CreditCards").IsList(&CreditCard{}).
		Editing("Name").
		ElementEditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			formkey := fmt.Sprintf("%s.Name", field.FormKey)
			return vx.VXField().
				Attr(presets.VFieldError(formkey, card.Name, vErr.GetFieldErrors(formkey))...)
		}).
		ElementShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			card := obj.(*CreditCard)
			return vx.VXTextField().Text(card.Name)
		})
	dp.Section(section)

	section.WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (vErr web.ValidationErrors) {
			if in != nil {
				ve := in(obj, ctx)
				_ = vErr.Merge(&ve)
			}
			if uc, ok := obj.(*UserCreditCard); ok {
				for i, c := range uc.CreditCards {
					if strings.TrimSpace(c.Name) == "" {
						vErr.FieldError(fmt.Sprintf("CreditCards[%d].Name", i), "name is required")
					}
				}
			}
			return
		}
	})
	return
}

func PresetsUtilsDialog(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsHelloWorld(b, db)

	dp = cust.RightDrawerWidth("800").Detailing("tab").Drawer(true)

	dp.Only("Delected Dialog")
	dp.Field("Delected Dialog").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		dialogPayload := utils.UtilDialogPayloadType{
			Title:     "Confirm",
			TypeField: "error",
			ContentEl: h.Div(h.Text("are you sure?")),
			Msgr:      utils.MustGetMessages(ctx.R),
		}

		return utils.CustomDialog(dialogPayload)
	})

	return
}

func PresetsDetailDisableSave(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(&Company{})
	if err != nil {
		panic(err)
	}
	cust, cl, ce, dp = PresetsHelloWorld(b, db)
	dp = cust.Detailing("DisabledSection", "Normal")
	section := presets.NewSectionBuilder(cust, "DisabledSection").Editing("Name", "Email")
	section.EditingField("Btn").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(
			vx.VXBtn("Disabled").Attr("@click", fmt.Sprintf("dash.disabled.%s=true", presets.DisabledKeyButtonSave)).Color(ColorSecondary),
			vx.VXBtn("Savable").Attr("@click", fmt.Sprintf("dash.disabled.%s=false", presets.DisabledKeyButtonSave)).Color(ColorPrimary),
		).Class("d-flex d-flex-inline ga-4")
	})
	normalSec := presets.NewSectionBuilder(cust, "Normal").Editing("Email")

	dp.Section(section, normalSec)
	mb := b.Model(&Company{})
	dpc := mb.Detailing("DisabledSectionCompany").Drawer(true)
	sec := presets.NewSectionBuilder(mb, "DisabledSectionCompany").Editing("Name")
	sec.EditingField("Btn").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(
			vx.VXBtn("Disabled").Attr("@click", fmt.Sprintf("dash.disabled.%s=true", presets.DisabledKeyButtonSave)).Color(ColorSecondary),
			vx.VXBtn("Savable").Attr("@click", fmt.Sprintf("dash.disabled.%s=false", presets.DisabledKeyButtonSave)).Color(ColorPrimary),
		).Class("d-flex d-flex-inline ga-4")
	})
	dpc.Section(sec)
	return
}

func PresetsDetailSaverValidation(b *presets.Builder, db *gorm.DB) (
	cust *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	cust, cl, ce, dp = PresetsHelloWorld(b, db)
	dp = cust.Detailing("Customer")
	section := presets.NewSectionBuilder(cust, "Customer").Editing("Name", "Email")
	section.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			ve := web.ValidationErrors{}
			if obj.(*Customer).Name == "system" {
				ve.GlobalError("You can not use system as name")
			}
			if ve.HaveErrors() {
				return &ve
			}
			return in(obj, id, ctx)
		}
	})
	dp.Section(section)
	return
}
