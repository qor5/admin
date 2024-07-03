package presets

import (
	"fmt"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type OrderableField struct {
	FieldName string
	DBColumn  string
}

type ListingBuilder struct {
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

func (b *ListingBuilder) CellWrapperFunc(cwf vx.CellWrapperFunc) (r *ListingBuilder) {
	b.cellWrapperFunc = cwf
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

func (b *ListingBuilder) Title(title string) (r *ListingBuilder) {
	b.title = title
	return b
}

func (b *ListingBuilder) KeywordSearchOff(v bool) (r *ListingBuilder) {
	b.keywordSearchOff = v
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

func (b *ListingBuilder) OrderBy(v string) (r *ListingBuilder) {
	b.orderBy = v
	return b
}

func (b *ListingBuilder) NewButtonFunc(v ComponentFunc) (r *ListingBuilder) {
	b.newBtnFunc = v
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

func (b *ListingBuilder) GetPageFunc() web.PageFunc {
	if b.pageFunc != nil {
		return b.pageFunc
	}
	return b.defaultPageFunc
}

func (b *ListingBuilder) cellComponentFunc(f *FieldBuilder) vx.CellComponentFunc {
	return func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		return f.compFunc(obj, b.mb.getComponentFuncField(f), ctx)
	}
}

func (b *ListingBuilder) injectorName() string {
	return strcase.ToSnake(strcase.ToCamel(b.mb.Info().ListingHref()))
}

func (b *ListingBuilder) setup() {
	b.once.Do(func() {
		injectorName := b.injectorName()
		stateful.RegisterInjector(injectorName)
		stateful.MustProvide(injectorName, func() *ListingBuilder {
			return b
		})
	})
}

func (b *ListingBuilder) defaultPageFunc(evCtx *web.EventContext) (r web.PageResponse, err error) {
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

func (b *ListingBuilder) openListingDialog(evCtx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *ListingBuilder) deleteConfirmation(evCtx *web.EventContext) (r web.EventResponse, err error) {
	msgr := MustGetMessages(evCtx.R)
	id := evCtx.R.FormValue(ParamID)
	promptID := id
	if v := evCtx.R.FormValue("prompt_id"); v != "" {
		promptID = v
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DeleteConfirmPortalName,
		Body: web.Scope().VSlot("{ locals }").Init(`{deleteConfirmation:true}`).Children(
			VDialog().MaxWidth("600px").Attr("v-model", "locals.deleteConfirmation").Children(
				VCard(
					VCardTitle(
						h.Text(msgr.DeleteConfirmationText(promptID)),
					),
					VCardActions(
						VSpacer(),
						VBtn(msgr.Cancel).
							Variant(VariantFlat).
							Class("ml-2").
							On("click", "locals.deleteConfirmation = false"),

						VBtn(msgr.Delete).
							Color("primary").
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", web.Plaid().
								EventFunc(actions.DoDelete).
								Queries(evCtx.Queries()).
								URL(b.mb.Info().ListingHref()).
								Go()),
					),
				),
			),
		),
	})
	return
}

// TODO: should remove ReloadList event func
func (b *ListingBuilder) notifReloadList() string {
	return "PresetsReloadList_" + b.injectorName()
}

func (b *ListingBuilder) reloadList(evCtx *web.EventContext) (r web.EventResponse, err error) {
	r.Emit(b.notifReloadList())
	return
}
