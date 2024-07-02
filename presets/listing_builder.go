package presets

import (
	"fmt"
	"sync"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type ListingBuilderX struct {
	mb              *ModelBuilder
	bulkActions     []*BulkActionBuilder
	footerActions   []*FooterActionBuilder
	actions         []*ActionBuilder
	actionsAsMenu   bool
	rowMenu         *RowMenuBuilder
	filterDataFunc  FilterDataFunc
	filterTabsFunc  FilterTabsFunc
	newBtnFunc      ComponentFunc
	pageFunc        web.PageFunc
	cellWrapperFunc vx.CellWrapperFunc
	Searcher        SearchFunc
	searchColumns   []string

	// title is the title of the listing page.
	// its default value is "Listing ${modelName}".
	title string

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

	orderBy           string
	orderableFields   []*OrderableField
	selectableColumns bool
	conditions        []*SQLCondition
	dialogWidth       string
	dialogHeight      string
	keywordSearchOff  bool
	FieldsBuilder

	once sync.Once
}

func (mb *ModelBuilder) ListingX(vs ...string) (r *ListingBuilderX) {
	r = mb.listingx
	if len(vs) == 0 {
		return
	}

	r.Only(vs...)
	return r
}

func (b *ListingBuilderX) Only(vs ...string) (r *ListingBuilderX) {
	r = b
	ivs := make([]interface{}, 0, len(vs))
	for _, v := range vs {
		ivs = append(ivs, v)
	}
	r.FieldsBuilder = *r.FieldsBuilder.Only(ivs...)
	return
}

func (b *ListingBuilderX) Except(vs ...string) (r *ListingBuilderX) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Except(vs...)
	return
}

func (b *ListingBuilderX) PageFunc(pf web.PageFunc) (r *ListingBuilderX) {
	b.pageFunc = pf
	return b
}

func (b *ListingBuilderX) CellWrapperFunc(cwf vx.CellWrapperFunc) (r *ListingBuilderX) {
	b.cellWrapperFunc = cwf
	return b
}

func (b *ListingBuilderX) DisablePagination(v bool) (r *ListingBuilderX) {
	b.disablePagination = v
	return b
}

func (b *ListingBuilderX) SearchFunc(v SearchFunc) (r *ListingBuilderX) {
	b.Searcher = v
	return b
}

func (b *ListingBuilderX) WrapSearchFunc(w func(in SearchFunc) SearchFunc) (r *ListingBuilderX) {
	b.Searcher = w(b.Searcher)
	return b
}

func (b *ListingBuilderX) Title(title string) (r *ListingBuilderX) {
	b.title = title
	return b
}

func (b *ListingBuilderX) KeywordSearchOff(v bool) (r *ListingBuilderX) {
	b.keywordSearchOff = v
	return b
}

func (b *ListingBuilderX) SearchColumns(vs ...string) (r *ListingBuilderX) {
	b.searchColumns = vs
	return b
}

func (b *ListingBuilderX) PerPage(v int64) (r *ListingBuilderX) {
	b.perPage = v
	return b
}

func (b *ListingBuilderX) OrderBy(v string) (r *ListingBuilderX) {
	b.orderBy = v
	return b
}

func (b *ListingBuilderX) NewButtonFunc(v ComponentFunc) (r *ListingBuilderX) {
	b.newBtnFunc = v
	return b
}

func (b *ListingBuilderX) ActionsAsMenu(v bool) (r *ListingBuilderX) {
	b.actionsAsMenu = v
	return b
}

func (b *ListingBuilderX) OrderableFields(v []*OrderableField) (r *ListingBuilderX) {
	b.orderableFields = v
	return b
}

func (b *ListingBuilderX) SelectableColumns(v bool) (r *ListingBuilderX) {
	b.selectableColumns = v
	return b
}

func (b *ListingBuilderX) Conditions(v []*SQLCondition) (r *ListingBuilderX) {
	b.conditions = v
	return b
}

func (b *ListingBuilderX) DialogWidth(v string) (r *ListingBuilderX) {
	b.dialogWidth = v
	return b
}

func (b *ListingBuilderX) DialogHeight(v string) (r *ListingBuilderX) {
	b.dialogHeight = v
	return b
}

func (b *ListingBuilderX) GetPageFunc() web.PageFunc {
	if b.pageFunc != nil {
		return b.pageFunc
	}
	return b.defaultPageFunc
}

func (b *ListingBuilderX) cellComponentFunc(f *FieldBuilder) vx.CellComponentFunc {
	return func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		return f.compFunc(obj, b.mb.getComponentFuncField(f), ctx)
	}
}

func (b *ListingBuilderX) injectorName() string {
	// TODO: 这个没准需要再考虑命名
	return b.mb.Info().URIName()
}

func (b *ListingBuilderX) setup() {
	b.once.Do(func() {
		injectorName := b.injectorName()
		stateful.Install(b.mb.p.builder)        // TODO: 为什么还需要这个呢？
		stateful.RegisterInjector(injectorName) // TODO: 全局的话貌似有点问题，例如 example 里存在一堆同类型的不同 presets
		stateful.MustProvide(injectorName, func() *ListingBuilderX {
			return b
		})
	})
}

func (b *ListingBuilderX) defaultPageFunc(evCtx *web.EventContext) (r web.PageResponse, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(evCtx.R).IsAllowed() != nil {
		return r, perm.PermissionDenied
	}

	msgr := MustGetMessages(evCtx.R)
	title := b.title
	if title == "" {
		title = msgr.ListingObjectTitle(i18n.T(evCtx.R, ModelsI18nModuleKey, b.mb.label))
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

	r.Body = VLayout(
		VMain(
			stateful.MustInject(injectorName, stateful.SyncQuery(compo)),
		),
	)
	return
}

func (b *ListingBuilderX) openListingDialog(evCtx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(evCtx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	msgr := MustGetMessages(evCtx.R)
	title := b.title
	if title == "" {
		title = msgr.ListingObjectTitle(i18n.T(evCtx.R, ModelsI18nModuleKey, b.mb.label))
	}

	evCtx.WithContextValue(ctxInDialog, true)

	injectorName := b.injectorName()
	compo := &ListingCompo{
		ID:                 injectorName + "_dialog",
		Popup:              true,
		LongStyleSearchBox: true,
	}

	compo.OnMounted = fmt.Sprintf(`
	var listingDialogElem = el.ownerDocument.getElementById(%q); 
	if (listingDialogElem && listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {
		listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';
	};`, compo.CompoID())

	content := VCard().Attr("id", compo.CompoID()).Children(
		VCardTitle().Class("d-flex align-center").Children(
			h.Text(title),
			VSpacer(),
			h.Div().Id(compo.ActionsComponentTeleportToID()),
			VBtn("").Elevation(0).Icon("mdi-close").Class("ml-2").Attr("@click", CloseListingDialogVarScript),
		),
		VCardText().Class("pa-0").Children(
			stateful.MustInject(injectorName, stateful.ParseQuery(compo)),
		),
	)
	dialog := VDialog(content).Attr("v-model", "vars.presetsListingDialog").Scrollable(true)
	if b.dialogWidth != "" {
		dialog.Width(b.dialogWidth)
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
