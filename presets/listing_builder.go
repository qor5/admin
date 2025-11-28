package presets

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"

	"github.com/qor5/admin/v3/presets/actions"
)

type ListingStyle string

const (
	ListingStylePage   ListingStyle = "Page"
	ListingStyleDialog ListingStyle = "Dialog"
	ListingStyleNested ListingStyle = "Nested"
)

type (
	ColumnsProcessor func(evCtx *web.EventContext, columns []*Column) ([]*Column, error)
	CellProcessor    func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error)
	RowProcessor     func(evCtx *web.EventContext, row h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error)
	TableProcessor   func(evCtx *web.EventContext, table *vx.DataTableBuilder) (*vx.DataTableBuilder, error)
)

type OrderableField struct {
	FieldName string
	DBColumn  string // Deprecated
}

type ListingBuilder struct {
	mb                     *ModelBuilder
	bulkActions            []*BulkActionBuilder
	footerActions          []*FooterActionBuilder
	actions                []*ActionBuilder
	actionsAsMenu          bool
	rowMenu                *RowMenuBuilder
	filterDataFunc         FilterDataFunc
	filterTabsFunc         FilterTabsFunc
	filterNotificationFunc FilterNotificationFunc
	newBtnFunc             ComponentFunc
	pageFunc               web.PageFunc
	cellWrapperFunc        vx.CellWrapperFunc
	cellProcessor          CellProcessor
	rowProcessor           RowProcessor
	tableProcessor         TableProcessor
	Searcher               SearchFunc
	searchColumns          []string
	titleFunc              func(evCtx *web.EventContext, style ListingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error)

	// perPage is the number of records per page.
	// if request query param "per_page" is set, it will be set to that value.
	// if the final value is less than 0, it will be set to 50.
	// if the final value is greater than 1000, it will be set to 1000.
	perPage int64

	// disablePagination is used to disable pagination, its default value is false.
	// if it is true, the following will happen:
	// 1. the pagination component will not display on listing page.
	// 2. the perPage will actually be ignored.
	// 3. all data will be returned in one page.
	disablePagination bool

	// If empty, regular pagination will be used
	relayPagination RelayPagination

	defaultOrderBy    []relay.Order
	defaultOrderBys   []relay.OrderBy
	orderableFields   []*OrderableField
	selectableColumns bool
	conditions        []*SQLCondition
	dialogWidth       string
	dialogHeight      string
	keywordSearchOff  bool
	columnsProcessor  ColumnsProcessor

	FieldsBuilder

	once                  sync.Once
	disableModelListeners bool

	dataTableFunc func(ctx *web.EventContext, searchParams *SearchParams, result *SearchResult, pagination h.HTMLComponent) h.HTMLComponent
}

func (mb *ModelBuilder) Listing(vs ...string) (r *ListingBuilder) {
	r = mb.listing
	if len(vs) == 0 {
		return
	}

	r.Only(vs...)
	return r
}

func (b *ListingBuilder) Only(vs ...string) (r *ListingBuilder) {
	r = b
	ivs := make([]interface{}, 0, len(vs))
	for _, v := range vs {
		ivs = append(ivs, v)
	}
	r.FieldsBuilder = *r.FieldsBuilder.Only(ivs...)
	return
}

func (b *ListingBuilder) Except(vs ...string) (r *ListingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Except(vs...)
	return
}

func (b *ListingBuilder) PageFunc(pf web.PageFunc) (r *ListingBuilder) {
	b.pageFunc = pf
	return b
}

// CellWrapperFunc Deprecated: Use WrapCell instead.
func (b *ListingBuilder) CellWrapperFunc(cwf vx.CellWrapperFunc) (r *ListingBuilder) {
	b.cellWrapperFunc = cwf
	return b
}

func (b *ListingBuilder) WrapCell(w func(in CellProcessor) CellProcessor) (r *ListingBuilder) {
	if b.cellProcessor == nil {
		b.cellProcessor = w(func(_ *web.EventContext, cell h.MutableAttrHTMLComponent, _ string, _ any) (h.MutableAttrHTMLComponent, error) {
			return cell, nil
		})
	} else {
		b.cellProcessor = w(b.cellProcessor)
	}
	return b
}

func (b *ListingBuilder) DataTableFunc(v func(*web.EventContext, *SearchParams, *SearchResult, h.HTMLComponent) h.HTMLComponent) *ListingBuilder {
	b.dataTableFunc = v
	return b
}

func (b *ListingBuilder) WrapRow(w func(in RowProcessor) RowProcessor) (r *ListingBuilder) {
	if b.rowProcessor == nil {
		b.rowProcessor = w(func(_ *web.EventContext, row h.MutableAttrHTMLComponent, _ string, _ any) (h.MutableAttrHTMLComponent, error) {
			return row, nil
		})
	} else {
		b.rowProcessor = w(b.rowProcessor)
	}
	return b
}

func (b *ListingBuilder) WrapTable(w func(in TableProcessor) TableProcessor) (r *ListingBuilder) {
	if b.tableProcessor == nil {
		b.tableProcessor = w(func(_ *web.EventContext, table *vx.DataTableBuilder) (*vx.DataTableBuilder, error) {
			return table, nil
		})
	} else {
		b.tableProcessor = w(b.tableProcessor)
	}
	return b
}

func (b *ListingBuilder) DisablePagination(v bool) (r *ListingBuilder) {
	b.disablePagination = v
	return b
}

func (b *ListingBuilder) SearchFunc(v SearchFunc) (r *ListingBuilder) {
	b.Searcher = v
	return b
}

func (b *ListingBuilder) WrapSearchFunc(w func(in SearchFunc) SearchFunc) (r *ListingBuilder) {
	b.Searcher = w(b.Searcher)
	return b
}

// Title The title must not return empty, and titleCompo can return nil
func (b *ListingBuilder) Title(f func(evCtx *web.EventContext, style ListingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error)) (r *ListingBuilder) {
	b.titleFunc = f
	return b
}

func (b *ListingBuilder) KeywordSearchOff(v bool) (r *ListingBuilder) {
	b.keywordSearchOff = v
	return b
}

func (b *ListingBuilder) WrapColumns(w func(in ColumnsProcessor) ColumnsProcessor) (r *ListingBuilder) {
	if b.columnsProcessor == nil {
		b.columnsProcessor = w(func(_ *web.EventContext, columns []*Column) ([]*Column, error) {
			return columns, nil
		})
	} else {
		b.columnsProcessor = w(b.columnsProcessor)
	}
	return b
}

func (b *ListingBuilder) SearchColumns(vs ...string) (r *ListingBuilder) {
	b.searchColumns = vs
	return b
}

func (b *ListingBuilder) PerPage(v int64) (r *ListingBuilder) {
	b.perPage = v
	return b
}

func (b *ListingBuilder) DefaultOrderBy(v ...relay.Order) (r *ListingBuilder) {
	b.defaultOrderBy = v
	return b
}

// Deprecated: Use DefaultOrderBy instead.
func (b *ListingBuilder) DefaultOrderBys(v ...relay.OrderBy) (r *ListingBuilder) {
	b.defaultOrderBys = v
	return b
}

func (b *ListingBuilder) RelayPagination(v RelayPagination) (r *ListingBuilder) {
	b.relayPagination = v
	return b
}

func (b *ListingBuilder) FilterNotificationFunc(v FilterNotificationFunc) (r *ListingBuilder) {
	b.filterNotificationFunc = v
	return b
}

func (b *ListingBuilder) NewButtonFunc(v ComponentFunc) (r *ListingBuilder) {
	b.newBtnFunc = v
	return b
}

func (b *ListingBuilder) WarpNewButtonFunc(w func(ComponentFunc) ComponentFunc) (r *ListingBuilder) {
	b.newBtnFunc = w(b.newBtnFunc)
	return b
}

func (b *ListingBuilder) ActionsAsMenu(v bool) (r *ListingBuilder) {
	b.actionsAsMenu = v
	return b
}

func (b *ListingBuilder) OrderableFields(v []*OrderableField) (r *ListingBuilder) {
	b.orderableFields = v
	return b
}

func (b *ListingBuilder) SelectableColumns(v bool) (r *ListingBuilder) {
	b.selectableColumns = v
	return b
}

func (b *ListingBuilder) Conditions(v []*SQLCondition) (r *ListingBuilder) {
	b.conditions = v
	return b
}

func (b *ListingBuilder) DialogWidth(v string) (r *ListingBuilder) {
	b.dialogWidth = v
	return b
}

func (b *ListingBuilder) DialogHeight(v string) (r *ListingBuilder) {
	b.dialogHeight = v
	return b
}

func (b *ListingBuilder) DisableModelListeners(v bool) (r *ListingBuilder) {
	b.disableModelListeners = v
	return b
}

func (b *ListingBuilder) GetPageFunc() web.PageFunc {
	if b.pageFunc != nil {
		return b.pageFunc
	}
	return b.defaultPageFunc
}

func (b *ListingBuilder) cellComponentFunc(f *FieldBuilder) vx.CellComponentFunc {
	return func(obj interface{}, _ string, ctx *web.EventContext) h.HTMLComponent {
		return f.lazyCompFunc().FieldComponent(obj, b.mb.getComponentFuncField(f), ctx)
	}
}

func (b *ListingBuilder) injectorName() string {
	return strcase.ToSnake(strcase.ToCamel(b.mb.Info().ListingHref()))
}

func (b *ListingBuilder) setup() {
	b.once.Do(func() {
		injectorName := b.injectorName()
		b.mb.p.dc.RegisterInjector(injectorName)
		b.mb.p.dc.MustProvide(injectorName, func() *ListingBuilder {
			return b
		})
	})
}

func (b *ListingBuilder) defaultPageFunc(evCtx *web.EventContext) (r web.PageResponse, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
	}

	title, titleCompo, err := b.getTitle(evCtx, ListingStylePage)
	if err != nil {
		return r, err
	}
	if titleCompo != nil {
		evCtx.WithContextValue(CtxPageTitleComponent, titleCompo)
	}
	r.PageTitle = title

	evCtx.WithContextValue(ctxInDialog, false)

	injectorName := b.injectorName()
	compo := &ListingCompo{
		ID:                 injectorName + "_page",
		Popup:              false,
		LongStyleSearchBox: false,
	}
	evCtx.WithContextValue(ctxActionsComponentTeleportToID, compo.ActionsComponentTeleportToID())

	r.Body = v.VLayout(
		v.VMain(
			b.mb.p.dc.MustInject(injectorName, stateful.SyncQuery(compo)),
		),
	)
	return
}

func (b *ListingBuilder) getTitle(evCtx *web.EventContext, style ListingStyle) (title string, titleCompo h.HTMLComponent, err error) {
	title = b.mb.mustGetMessages(evCtx.R).ListingObjectTitle(b.mb.Info().LabelName(evCtx, false))
	if b.titleFunc != nil {
		return b.titleFunc(evCtx, style, title)
	}
	return title, nil, nil
}

func (b *ListingBuilder) openListingDialog(evCtx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(evCtx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	title, titleCompo, err := b.getTitle(evCtx, ListingStyleDialog)
	if err != nil {
		return r, err
	}
	if titleCompo == nil {
		titleCompo = h.Div().Attr("v-pre", true).Text(title)
	}

	evCtx.WithContextValue(ctxInDialog, true)

	injectorName := b.injectorName()
	compo := &ListingCompo{
		ID:                 injectorName + "_dialog",
		Popup:              true,
		LongStyleSearchBox: true,
	}

	compo.OnMounted = fmt.Sprintf(`({el}) => {
		var listingDialogElem = el.ownerDocument.getElementById(%q); 
		if (listingDialogElem && listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {
			listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';
		};
	}`, compo.CompoID())

	content := v.VCard().Attr("id", compo.CompoID()).Children(
		v.VCardTitle().Class("d-flex align-center h-abs-26 py-6 px-6 content-box").Children(
			titleCompo,
			v.VSpacer(),
			h.Div().Id(compo.ActionsComponentTeleportToID()),
			v.VBtn("").Elevation(0).Size(v.SizeXSmall).Icon("mdi-close").Class("ml-2 dialog-close-btn").Attr("@click", CloseListingDialogVarScript),
		),
		v.VCardText().Class("pa-0").Children(
			b.mb.p.dc.MustInject(injectorName, stateful.ParseQuery(compo)),
		),
	)
	dialog := vuetifyx.VXDialog(content).Attr("v-model", "vars.presetsListingDialog").
		ContentOnlyMode(true).
		ContentPadding("0")

	if b.dialogWidth != "" {
		dialog.Attr("width", b.dialogWidth)
	}
	if b.dialogHeight != "" {
		content.Attr("height", b.dialogHeight)
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: ListingDialogPortalName,
		Body: web.Scope(dialog).VSlot("{ form }"),
	})
	r.RunScript = "setTimeout(function(){ vars.presetsListingDialog = true }, 100)"
	return
}

func (b *ListingBuilder) deleteConfirmation(evCtx *web.EventContext) (r web.EventResponse, err error) {
	var (
		msgr    = b.mb.mustGetMessages(evCtx.R)
		message = msgr.DeleteConfirmationText
		length  = len(strings.Split(evCtx.Param(ParamID), ","))
	)
	if length > 1 {
		message = msgr.DeleteObjectsConfirmationText(length)
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DeleteConfirmPortalName,
		Body: web.Scope().VSlot("{ locals }").Init(`{deleteConfirmation:true}`).Children(
			vx.VXDialog(
				h.Span(message),
			).Title(msgr.DialogTitleDefault).
				CancelText(msgr.Cancel).
				OkText(msgr.Delete).
				Attr("@click:ok", web.Plaid().
					EventFunc(actions.DoDelete).
					Queries(evCtx.Queries()).
					URL(b.mb.Info().ListingHref()).
					Go()).
				Attr("v-model", "locals.deleteConfirmation"),
		),
	})
	return
}

func CustomizeColumnHeader(f func(evCtx *web.EventContext, col *Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error), fields ...string) func(in ColumnsProcessor) ColumnsProcessor {
	m := lo.SliceToMap(fields, func(field string) (string, struct{}) { return field, struct{}{} })
	return func(in ColumnsProcessor) ColumnsProcessor {
		return func(evCtx *web.EventContext, columns []*Column) ([]*Column, error) {
			columns, err := in(evCtx, columns)
			if err != nil {
				return nil, err
			}

			for _, dc := range columns {
				if len(m) > 0 {
					if _, ok := m[dc.Name]; !ok {
						continue
					}
				}
				w := dc.WrapHeader
				dc.WrapHeader = func(evCtx *web.EventContext, col *Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error) {
					if w != nil {
						var err error
						th, err = w(evCtx, col, th)
						if err != nil {
							return nil, err
						}
					}
					return f(evCtx, col, th)
				}
			}
			return columns, nil
		}
	}
}

func CustomizeColumnLabel(mapper func(evCtx *web.EventContext) (map[string]string, error)) func(in ColumnsProcessor) ColumnsProcessor {
	return func(in ColumnsProcessor) ColumnsProcessor {
		return func(evCtx *web.EventContext, columns []*Column) ([]*Column, error) {
			columns, err := in(evCtx, columns)
			if err != nil {
				return nil, err
			}

			m, err := mapper(evCtx)
			if err != nil {
				return nil, err
			}
			for _, dc := range columns {
				v, ok := m[dc.Name]
				if ok {
					dc.Label = v
				}
			}
			return columns, nil
		}
	}
}

func (b *ListingBuilder) defaultNewBtnFunc(ctx *web.EventContext) h.HTMLComponent {
	if b.mb.Info().Verifier().Do(PermCreate).WithReq(ctx.R).IsAllowed() != nil {
		return nil
	}
	var (
		lc   = ListingCompoFromContext(ctx.R.Context())
		msgr = b.mb.mustGetMessages(ctx.R)
	)
	onClick := web.Plaid().EventFunc(actions.New)
	if lc.Popup {
		onClick.URL(b.mb.Info().ListingHref()).Query(ParamOverlay, actions.Dialog)
	}
	if lc.ParentID != "" {
		onClick.Query(ParamParentID, lc.ParentID)
	}
	return v.VBtn(msgr.New).
		Color(v.ColorPrimary).
		Variant(v.VariantElevated).
		Theme("light").Class("ml-2").
		Attr("@click", onClick.Go())
}
