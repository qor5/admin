// @snippet_begin(PresetHelloWorldSample)
package examples_presets

import (
	"fmt"
	"net/url"
	"time"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay/cursor"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/statusx"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

type Customer struct {
	ID                             int
	Name                           string
	Email                          string
	Description                    string
	CompanyID                      int
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
	ApprovedAt                     *time.Time
	TermAgreedAt                   *time.Time
	ApprovalComment                string
	Avatar                         media_library.MediaBox
	HTMLSanitizerPolicyTiptapInput string
	HTMLSanitizerPolicyUGCInput    string
	HTMLSanitizerPolicyStrictInput string
	HTMLSanitizerPolicyCustomInput string
	CreditCards                    []*CreditCard `gorm:"-"`
	Notes                          []*Note       `gorm:"-"`
}

type Address struct {
	ID       int
	Province string
	City     string
	District string
}

func PresetsHelloWorld(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	err := db.AutoMigrate(
		&Customer{},
		&Company{},
		&Address{},
	)
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&Customer{})
	cl = mb.Listing()
	ce = mb.Editing()
	return
}

// @snippet_end

func PresetsKeywordSearchOff(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsHelloWorld(b, db)
	cl.KeywordSearchOff(true)
	return
}

// @snippet_begin(PresetsListingCustomizationFieldsSample)

func PresetsRowMenuAction(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsHelloWorld(b, db)
	cl.KeywordSearchOff(true)
	rmb := cl.RowMenu()

	rmb.RowMenuItem("with-icon").Icon("mdi-close")
	rmb.RowMenuItem("Delete").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		return nil
	})
	return
}

type Company struct {
	ID   int
	Name string
}

func PresetsListingCustomizationFields(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.GetI18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN).
		RegisterForModule(language.English, presets.ModelsI18nModuleKey, Messages_en_US)

	mb, cl, ce, dp = PresetsHelloWorld(b, db)

	cl = mb.Listing("ID", "Name", "Company", "Email").
		SearchColumns("name", "email").SelectableColumns(true).
		OrderableFields([]*presets.OrderableField{
			{
				FieldName: "ID",
				DBColumn:  "id",
			},
			{
				FieldName: "Name",
				DBColumn:  "name",
			},
		})

	// if you want to customize the table head, you can use WrapColumns
	cl.WrapColumns(presets.CustomizeColumnHeader(func(evCtx *web.EventContext, col *presets.Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error) {
		th.SetAttr("style", "min-width: 123px; color: red;")
		return th, nil
	}, "ID"))

	cl.Field("Company").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c := obj.(*Customer)
		var comp Company
		if c.CompanyID == 0 {
			return h.Td()
		}

		db.First(&comp, "id = ?", c.CompanyID)
		return h.Td(
			h.A().Text(comp.Name).
				Class("text-decoration-none", "text-blue").
				Href("javascript:void(0)").
				Attr("@click.stop",
					web.POST().EventFunc(actions.Edit).
						Query(presets.ParamID, fmt.Sprint(comp.ID)).
						URL("companies").
						Go()),
			h.Text("-"),
			h.A().Text("(Open in Dialog)").
				Class("text-decoration-none", "text-blue").
				Href("javascript:void(0)").
				Attr("@click.stop",
					web.POST().EventFunc(actions.Edit).
						Query(presets.ParamID, fmt.Sprint(comp.ID)).
						Query(presets.ParamOverlay, actions.Dialog).
						URL("companies").
						Go(),
				),
		)
	})

	cl.Field("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c := obj.(*Customer)
		return h.Td(
			h.Div(h.Text(c.Name + "_" + "customizable")),
		)
	})

	ce = mb.Editing("Name", "CompanyID")

	mb.RegisterEventFunc("updateCompanyList", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		companyID := ctx.ParamAsInt(presets.ParamOverlayUpdateID)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "companyListPortal",
			Body: companyList(ctx, db, companyID),
		})
		return
	})

	ce.Field("CompanyID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c := obj.(*Customer)
		return web.Portal(companyList(ctx, db, c.CompanyID)).Name("companyListPortal")
	})

	comp := b.Model(&Company{})
	comp.Editing().ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Company)
		if len(c.Name) < 5 {
			err.GlobalError("name must longer than 5")
		}
		return
	})

	gcm, err := cursor.NewGCM([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		panic(err)
	}
	gcmMiddleware := cursor.GCM[any](gcm)
	cl.RelayPagination(
		gorm2op.KeysetBasedPagination(true, gcmMiddleware),
	)
	comp.Listing().RelayPagination(
		gorm2op.KeysetBasedPagination(true, gcmMiddleware),
	)

	return
}

func companyList(ctx *web.EventContext, db *gorm.DB, companyID int) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, presets.ModelsI18nModuleKey, Messages_en_US).(*Messages)
	var comps []Company
	db.Find(&comps)
	return h.Div(
		v.VSelect().
			Label(msgr.CustomersCompanyID).
			Variant("underlined").
			Items(comps).
			Attr(web.VField("CompanyID", companyID)...).
			ItemTitle("Name").ItemValue("ID"),
		h.A().Text("Add Company").
			Class("text-decoration-none", "text-blue").
			Href("javascript:void(0)").Attr("@click",
			web.POST().
				URL("companies").
				EventFunc(actions.New).
				Query(presets.ParamOverlay, actions.Dialog).
				Query(presets.ParamOverlayAfterUpdateScript,
					web.POST().EventFunc("updateCompanyList").Go()).
				Go(),
		),
	)
}

// @snippet_end

// @snippet_begin(PresetsListingCustomizationFiltersSample)

func PresetsListingCustomizationFilters(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsListingCustomizationFields(b, db)

	cl.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, presets.ModelsI18nModuleKey, Messages_en_US).(*Messages)
		var companyOptions []*vuetifyx.SelectItem
		err := db.Model(&Company{}).Select("name as text, id as value").Scan(&companyOptions).Error
		if err != nil {
			panic(err)
		}

		return []*vuetifyx.FilterItem{
			{
				Key:      "created",
				Label:    msgr.CustomersFilterCreated,
				ItemType: vuetifyx.ItemTypeDatetimeRange,
				// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
				SQLCondition: `created_at %s ?`,
				DateOptions:  &[]vuetifyx.DateOption{{Label: "StartAt"}, {Label: "EndAt"}},

				ValidateFunc: func(ctx *web.EventContext, vErr *web.ValidationErrors, it *vuetifyx.FilterItem) {
					if it.ValueFrom >= it.ValueTo {
						vErr.GlobalError("CreatedAt Error")
					}
				},
			},
			{
				Key:      "approved",
				Label:    msgr.CustomersFilterApproved,
				ItemType: vuetifyx.ItemTypeDatetimeRange,
				// SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
				SQLCondition: `created_at %s ?`,
				DateOptions:  &[]vuetifyx.DateOption{{Label: "Approved_Start_At", ClearText: "Cancel1"}, {Label: "Approved_End_At"}},
				ValidateFunc: func(ctx *web.EventContext, vErr *web.ValidationErrors, it *vuetifyx.FilterItem) {
					if it.ValueFrom >= it.ValueTo {
						vErr.GlobalError("ApprovedAt Error")
					}
				},
			},
			{
				Key:          "name",
				Label:        msgr.CustomersFilterName,
				ItemType:     vuetifyx.ItemTypeString,
				SQLCondition: `name %s ?`,
			},
			{
				Key:          "company",
				Label:        msgr.CustomersFilterCompany,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `company_id %s ?`,
				Options:      companyOptions,
			},
		}
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsListingCustomizationTabsSample)

func PresetsListingCustomizationTabs(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsListingCustomizationFilters(b, db)

	cl.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		var c Company
		db.First(&c)
		return []*presets.FilterTab{
			{
				Label: "All",
				Query: url.Values{},
			},
			{
				Label: "Felix",
				Query: url.Values{"name.ilike": []string{"felix"}},
			},
			{
				Label: "The Plant",
				Query: url.Values{"company": []string{fmt.Sprint(c.ID)}},
			},
			{
				Label: "Approved",
				Query: url.Values{"approved.gte": []string{time.Time{}.Format("2006-01-02 15:04")}},
			},
		}
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsListingCustomizationBulkActionsSample)

func PresetsListingCustomizationBulkActions(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, _ = PresetsListingCustomizationTabs(b, db)

	cl.BulkAction("Approve").Label("Approve").
		UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
			comment := ctx.R.FormValue("ApprovalComment")
			if len(comment) < 10 {
				ctx.Flash = "comment should larger than 10"
				return
			}
			err = db.Model(&Customer{}).
				Where("id IN (?)", selectedIds).
				Updates(map[string]interface{}{"approved_at": time.Now(), "approval_comment": comment}).Error
			if err != nil {
				ctx.Flash = err.Error()
			} else {
				r.Emit(
					presets.NotifModelsUpdated(&Customer{}),
					presets.PayloadModelsUpdated{Ids: selectedIds},
				)
			}
			return
		}).
		ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
			comment := ctx.R.FormValue("ApprovalComment")
			errorMessage := ""
			if ctx.Flash != nil {
				errorMessage = ctx.Flash.(string)
			}
			return v.VTextField().
				Variant("underlined").
				Attr(web.VField("ApprovalComment", comment)...).
				Label("Content").
				ErrorMessages(errorMessage)
		})

	cl.BulkAction("Delete").Label("Delete").
		UpdateFunc(func(selectedIds []string, ctx *web.EventContext, r *web.EventResponse) (err error) {
			err = db.Where("id IN (?)", selectedIds).Delete(&Customer{}).Error
			if err == nil {
				r.Emit(
					presets.NotifModelsDeleted(&Customer{}),
					presets.PayloadModelsDeleted{
						Ids: selectedIds,
					},
				)
			}
			return
		}).
		ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
			return h.Div().Text(fmt.Sprintf("Are you sure you want to delete %s ?", selectedIds)).Class("title deep-orange--text")
		})

	return
}

// @snippet_end

// @snippet_begin(PresetsListingCustomizationSearcherSample)

func PresetsListingCustomizationSearcher(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&Customer{})
	mb.Listing().SearchFunc(func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
		// only display approved customers
		qdb := db.Where("approved_at IS NOT NULL")
		return gorm2op.DataOperator(qdb).Search(ctx, params)
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsListingDatatableFuncSample)

func PresetsListingDatatableFunc(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	mb = b.Model(&Customer{})
	mb.Listing().DataTableFunc(func(context *web.EventContext, params *presets.SearchParams, result *presets.SearchResult, pagination h.HTMLComponent) h.HTMLComponent {
		rows := v.VRow()
		reflectutils.ForEach(result.Nodes, func(obj interface{}) {
			p := obj.(*Customer)
			rows.AppendChildren(v.VCol(
				v.VCard(
					v.VCardItem(
						v.VCard(
							v.VCardText(
								h.Text(p.Name),
							).Class("pa-0", v.H100, "bg-"+v.ColorGreyLighten4),
						).Elevation(0).Height(80),
					).Class("pa-0", v.W100),
					v.VCardItem(
						v.VCard(
							v.VCardItem(
								h.Div(
									h.Text(p.Email),
								).Class(v.W100, "d-flex", "justify-space-between", "align-center"),
							).Class("pa-2"),
						).Color(v.ColorGreyLighten5).Height(80),
					).Class("pa-0"),
				).Elevation(0),
			).Cols(3))
		})
		return h.Components(
			v.VContainer(
				rows,
			).Fluid(true).Class("pa-0"),
			pagination,
		)
	})
	return
}

// @snippet_end

// @snippet_begin(PresetsListingFilterNotificationFuncSample)

func PresetsListingFilterNotificationFunc(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(gorm2op.DataOperator(db))
	err := db.AutoMigrate(&Customer{})
	if err != nil {
		panic(err)
	}
	mb = b.Model(&Customer{})
	mb.Listing().FilterNotificationFunc(func(_ *web.EventContext) h.HTMLComponent {
		return h.Div().Text("Filter Notification").Class(fmt.Sprintf("text-%s", v.ColorWarning))
	})
	return
}

// @snippet_end
type mockGRPCDataOperatorWrapper struct {
	next presets.DataOperator
}

func mockGRPCDataOperator(next presets.DataOperator) presets.DataOperator {
	return &mockGRPCDataOperatorWrapper{next: next}
}

func (w *mockGRPCDataOperatorWrapper) Search(eventCtx *web.EventContext, params *presets.SearchParams) (*presets.SearchResult, error) {
	return w.next.Search(eventCtx, params)
}

func (w *mockGRPCDataOperatorWrapper) Fetch(obj any, id string, eventCtx *web.EventContext) (any, error) {
	return w.next.Fetch(obj, id, eventCtx)
}

func (w *mockGRPCDataOperatorWrapper) Save(obj any, id string, eventCtx *web.EventContext) error {
	var fvs statusx.FieldViolations
	p := obj.(*Customer)
	if p.Name == "system" {
		fvs = append(fvs, statusx.NewFieldViolation("Name", "name can`t set system", "name can`t set system"))
		return statusx.BadRequest(fvs).Err()
	}
	return nil
}

func (w *mockGRPCDataOperatorWrapper) Delete(obj any, id string, eventCtx *web.EventContext) error {
	return w.next.Delete(obj, id, eventCtx)
}

func PresetsDataOperatorWithGRPC(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	b.DataOperator(presets.DataOperatorWithGRPC(mockGRPCDataOperator(gorm2op.DataOperator(db))))
	err := db.AutoMigrate(&Customer{})
	if err != nil {
		panic(err)
	}
	mb = b.Model(&Customer{})
	cl = mb.Listing()
	ce = mb.Editing()
	return
}
