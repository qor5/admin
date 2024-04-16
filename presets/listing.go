package presets

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/presets/actions"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	h "github.com/theplant/htmlgo"
)

type ListingBuilder struct {
	mb              *ModelBuilder
	bulkActions     []*BulkActionBuilder
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
	FieldsBuilder
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

func (b *ListingBuilder) Title(title string) (r *ListingBuilder) {
	b.title = title
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

type OrderableField struct {
	FieldName string
	DBColumn  string
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

const bulkPanelOpenParamName = "bulkOpen"
const actionPanelOpenParamName = "actionOpen"
const DeleteConfirmPortalName = "deleteConfirm"
const dataTablePortalName = "dataTable"
const dataTableAdditionsPortalName = "dataTableAdditions"
const listingDialogContentPortalName = "listingDialogContentPortal"

func (b *ListingBuilder) defaultPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	if b.mb.Info().Verifier().Do(PermList).WithReq(ctx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	msgr := MustGetMessages(ctx.R)
	title := b.title
	if title == "" {
		title = msgr.ListingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, b.mb.label))
	}
	r.PageTitle = title

	r.Body = b.listingComponent(ctx, false)

	return
}

func (b *ListingBuilder) listingComponent(
	ctx *web.EventContext,
	inDialog bool,
) h.HTMLComponent {
	ctx.R = ctx.R.WithContext(context.WithValue(ctx.R.Context(), ctxInDialog, inDialog))

	msgr := MustGetMessages(ctx.R)

	var tabsAndActionsBar h.HTMLComponent
	{
		filterTabs := b.filterTabs(ctx, inDialog)

		actionsComponent := b.actionsComponent(msgr, ctx, inDialog)
		// if v := ; v != nil {
		//	actionsComponent = append(actionsComponent, v)
		// }
		// || len(actionsComponent) > 0
		if filterTabs != nil {
			tabsAndActionsBar = filterTabs
		}

		ctx.R = ctx.R.WithContext(context.WithValue(ctx.R.Context(), ctxActionsComponent, actionsComponent))
	}

	var filterBar h.HTMLComponent
	if b.filterDataFunc != nil {
		fd := b.filterDataFunc(ctx)
		fd.SetByQueryString(ctx.R.URL.RawQuery)
		filterBar = b.filterBar(ctx, msgr, fd, inDialog)
	}
	searchBoxDefault := VResponsive(
		web.Scope(
			VTextField().
				AppendInnerIcon("mdi-magnify").
				Variant(FieldVariantOutlined).
				// PrependIcon("mdi-magnify").
				Label(msgr.Search).
				Density(DensityCompact).
				// Flat(true).
				Clearable(true).
				HideDetails(true).
				SingleLine(true).
				ModelValue(ctx.R.URL.Query().Get("keyword")).
				// Color("grey-lighten-2").
				//Attr(":prepend-icon", `locals.isFocus?null:"mdi-magnify"`).
				//Attr(":bg-color", `locals.isFocus?"white":"blue-darken-1"`).
				//Attr("@update:focused", "locals.isFocus=!locals.isFocus").
				Attr("@keyup.enter", web.Plaid().
					ClearMergeQuery("page").
					Query("keyword", web.Var("[$event.target.value]")).
					MergeQuery(true).
					PushState(true).
					Go()).
				Attr("@click:clear", web.Plaid().
					Query("keyword", "").
					PushState(true).
					Go()).
				Class("mr-4"),
			// Attr("style", "width: 200px"), // ).Method("GET"),
		).VSlot("{ locals }").Init(`{isFocus: false}`),
	).MaxWidth(200).MinWidth(200)
	dataTable, dataTableAdditions := b.getTableComponents(ctx, inDialog)

	var dialogHeaderBar h.HTMLComponent
	if inDialog {
		title := msgr.ListingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, b.mb.label))
		var searchBox h.HTMLComponent
		if b.mb.layoutConfig == nil || !b.mb.layoutConfig.SearchBoxInvisible {
			searchBox = VTextField().
				PrependInnerIcon("search").
				Placeholder(msgr.Search).
				HideDetails(true).
				ModelValue(ctx.R.URL.Query().Get("keyword")).
				Attr("@keyup.enter", web.Plaid().
					URL(ctx.R.RequestURI).
					Query("keyword", web.Var("[$event.target.value]")).
					MergeQuery(true).
					EventFunc(actions.UpdateListingDialog).
					Go()).
				Attr("@click:clear", web.Plaid().
					URL(ctx.R.RequestURI).
					Query("keyword", "").
					MergeQuery(true).
					EventFunc(actions.UpdateListingDialog).
					Go()).
				Class("ma-0 pa-0 mr-6")
		}
		dialogHeaderBar = VAppBar(
			VToolbarTitle("").
				Children(h.Text(title)),
			VSpacer(),
			searchBox,
			VBtn("").Icon(true).
				Children(VIcon("mdi-close")).
				Size(SizeLarge).
				Attr("@click.stop", CloseListingDialogVarScript),
		).Color("white").Elevation(0).Density(DensityCompact)
	}
	return web.Scope(VContainer(
		dialogHeaderBar,
		tabsAndActionsBar,
		h.Div(
			VToolbar(
				searchBoxDefault,
				filterBar,
			).Flat(true).Color("white").Class("pb-2"),

			VCard(
				// VDivider(),
				VCardText(
					web.Portal(dataTable).Name(dataTablePortalName),
				).Class("pa-0"),
			).Variant("outlined").Color("blue-grey-lighten-4"),
			web.Portal(dataTableAdditions).Name(dataTableAdditionsPortalName),
		),
	).Fluid(true).
		Class("white py-0"),
	).VSlot("{ locals }").Init(`{currEditingListItemID: ""}`)
}

func (b *ListingBuilder) cellComponentFunc(f *FieldBuilder) vx.CellComponentFunc {
	return func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		return f.compFunc(obj, b.mb.getComponentFuncField(f), ctx)
	}
}

func getSelectedIds(ctx *web.EventContext) (selected []string) {
	selectedValue := ctx.R.URL.Query().Get(ParamSelectedIds)
	if len(selectedValue) > 0 {
		selected = strings.Split(selectedValue, ",")
	}
	return selected
}

func (b *ListingBuilder) bulkPanel(
	bulk *BulkActionBuilder,
	selectedIds []string,
	processedSelectedIds []string,
	ctx *web.EventContext,
) (r h.HTMLComponent) {
	msgr := MustGetMessages(ctx.R)

	var errComp h.HTMLComponent
	if vErr, ok := ctx.Flash.(*web.ValidationErrors); ok {
		if gErr := vErr.GetGlobalError(); gErr != "" {
			errComp = VAlert(h.Text(gErr)).
				Border("left").
				Type("error").
				Elevation(2)
		}
	}
	var processSelectedIdsNotice h.HTMLComponent
	if len(processedSelectedIds) < len(selectedIds) {
		unactionables := make([]string, 0, len(selectedIds))
		{
			processedSelectedIdsM := make(map[string]struct{})
			for _, v := range processedSelectedIds {
				processedSelectedIdsM[v] = struct{}{}
			}
			for _, v := range selectedIds {
				if _, ok := processedSelectedIdsM[v]; !ok {
					unactionables = append(unactionables, v)
				}
			}
		}

		if len(unactionables) > 0 {
			var noticeText string
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				noticeText = bulk.selectedIdsProcessorNoticeFunc(selectedIds, processedSelectedIds, unactionables)
			} else {
				var idsText string
				if len(unactionables) <= 10 {
					idsText = strings.Join(unactionables, ", ")
				} else {
					idsText = fmt.Sprintf("%s...(+%d)", strings.Join(unactionables[:10], ", "), len(unactionables)-10)
				}
				noticeText = msgr.BulkActionSelectedIdsProcessNotice(idsText)
			}
			processSelectedIdsNotice = VAlert(h.Text(noticeText)).
				Type("warning")
		}
	}

	onOK := web.Plaid().EventFunc(actions.DoBulkAction).
		Query(ParamBulkActionName, bulk.name).
		MergeQuery(true)
	if isInDialogFromQuery(ctx) {
		onOK.URL(ctx.R.RequestURI)
	}
	return VCard(
		VCardTitle(
			h.Text(bulk.NameLabel.label),
		),
		VCardText(
			errComp,
			processSelectedIdsNotice,
			bulk.compFunc(selectedIds, ctx),
		),
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).
				Variant(VariantFlat).
				Class("ml-2").
				Attr("@click", closeDialogVarScript),

			VBtn(msgr.OK).
				Color("primary").
				Variant(VariantFlat).
				Theme(ThemeDark).
				Attr("@click", onOK.Go()),
		),
	)
}

func (b *ListingBuilder) actionPanel(action *ActionBuilder, ctx *web.EventContext) (r h.HTMLComponent) {
	msgr := MustGetMessages(ctx.R)

	var errComp h.HTMLComponent
	if vErr, ok := ctx.Flash.(*web.ValidationErrors); ok {
		if gErr := vErr.GetGlobalError(); gErr != "" {
			errComp = VAlert(h.Text(gErr)).
				Border("left").
				Type("error").
				Elevation(2)
		}
	}

	onOK := web.Plaid().EventFunc(actions.DoListingAction).
		Query(ParamListingActionName, action.name).
		MergeQuery(true)
	if isInDialogFromQuery(ctx) {
		onOK.URL(ctx.R.RequestURI)
	}

	var comp h.HTMLComponent
	if action.compFunc != nil {
		comp = action.compFunc("", ctx)
	}

	return VCard(
		VCardTitle(
			h.Text(action.NameLabel.label),
		),
		VCardText(
			errComp,
			comp,
		),
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).
				Variant(VariantFlat).
				Class("ml-2").
				Attr("@click", closeDialogVarScript),

			VBtn(msgr.OK).
				Color("primary").
				Variant(VariantFlat).
				Theme(ThemeDark).
				Attr("@click", onOK.Go()),
		),
	)
}

func (b *ListingBuilder) deleteConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	msgr := MustGetMessages(ctx.R)
	id := ctx.R.FormValue(ParamID)
	promptID := id
	if v := ctx.R.FormValue("prompt_id"); v != "" {
		promptID = v
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DeleteConfirmPortalName,
		Body: web.Scope(VDialog(
			VCard(
				VCardTitle(h.Text(msgr.DeleteConfirmationText(promptID))),
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
							Queries(ctx.Queries()).
							URL(ctx.R.URL.Path).
							Go()),
				),
			),
		).MaxWidth("600px").
			Attr("v-model", "locals.deleteConfirmation"),
		).VSlot("{ locals }").Init(`{deleteConfirmation:true}`),
	})
	return
}

func (b *ListingBuilder) openActionDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	actionName := ctx.R.URL.Query().Get(actionPanelOpenParamName)
	action := getAction(b.actions, actionName)
	if action == nil {
		err = errors.New("cannot find requested action")
		return
	}

	b.mb.p.dialog(
		&r,
		b.actionPanel(action, ctx),
		action.dialogWidth,
	)
	return
}

func (b *ListingBuilder) openBulkActionDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	msgr := MustGetMessages(ctx.R)
	selected := getSelectedIds(ctx)
	bulkName := ctx.R.URL.Query().Get(bulkPanelOpenParamName)
	bulk := getBulkAction(b.bulkActions, bulkName)

	if bulk == nil {
		err = errors.New("cannot find requested action")
		return
	}

	if len(selected) == 0 {
		ShowMessage(&r, "Please select record", "warning")
		return
	}

	// If selectedIdsProcessorFunc is not nil, process the request in it and skip the confirmation dialog
	var processedSelectedIds []string
	if bulk.selectedIdsProcessorFunc != nil {
		processedSelectedIds, err = bulk.selectedIdsProcessorFunc(selected, ctx)
		if err != nil {
			return
		}
		if len(processedSelectedIds) == 0 {
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				ShowMessage(&r, bulk.selectedIdsProcessorNoticeFunc(selected, processedSelectedIds, selected), "warning")
			} else {
				ShowMessage(&r, msgr.BulkActionNoAvailableRecords, "warning")
			}
			return
		}
	} else {
		processedSelectedIds = selected
	}

	b.mb.p.dialog(
		&r,
		b.bulkPanel(bulk, selected, processedSelectedIds, ctx),
		bulk.dialogWidth,
	)
	return
}

func (b *ListingBuilder) doBulkAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	bulk := getBulkAction(b.bulkActions, ctx.R.FormValue(ParamBulkActionName))
	if bulk == nil {
		panic("bulk required")
	}

	if b.mb.Info().Verifier().SnakeDo(PermBulkActions, bulk.name).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	selectedIds := getSelectedIds(ctx)

	var err1 error
	var processedSelectedIds []string
	if bulk.selectedIdsProcessorFunc != nil {
		processedSelectedIds, err1 = bulk.selectedIdsProcessorFunc(selectedIds, ctx)
	} else {
		processedSelectedIds = selectedIds
	}

	if err1 == nil {
		err1 = bulk.updateFunc(processedSelectedIds, ctx)
	}

	if err1 != nil {
		if _, ok := err1.(*web.ValidationErrors); !ok {
			vErr := &web.ValidationErrors{}
			vErr.GlobalError(err1.Error())
			ctx.Flash = vErr
		}
	}

	if ctx.Flash != nil {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: b.bulkPanel(bulk, selectedIds, processedSelectedIds, ctx),
		})
		return
	}

	msgr := MustGetMessages(ctx.R)
	ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	if isInDialogFromQuery(ctx) {
		qs := ctx.Queries()
		qs.Del(bulkPanelOpenParamName)
		qs.Del(ParamBulkActionName)
		web.AppendRunScripts(&r,
			closeDialogVarScript,
			web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				Queries(qs).
				Go(),
		)
	} else {
		r.PushState = web.Location(url.Values{bulkPanelOpenParamName: []string{}}).MergeQuery(true)
	}

	return
}

func (b *ListingBuilder) doListingAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamListingActionName))
	if action == nil {
		panic("action required")
	}

	if b.mb.Info().Verifier().SnakeDo(PermDoListingAction, action.name).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	if err := action.updateFunc("", ctx); err != nil {
		if _, ok := err.(*web.ValidationErrors); !ok {
			vErr := &web.ValidationErrors{}
			vErr.GlobalError(err.Error())
			ctx.Flash = vErr
		}
	}

	if ctx.Flash != nil {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: b.actionPanel(action, ctx),
		})
		return
	}

	msgr := MustGetMessages(ctx.R)
	ShowMessage(&r, msgr.SuccessfullyUpdated, "")

	if isInDialogFromQuery(ctx) {
		qs := ctx.Queries()
		qs.Del(actionPanelOpenParamName)
		qs.Del(ParamListingActionName)
		web.AppendRunScripts(&r,
			closeDialogVarScript,
			web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				Queries(qs).
				Go(),
		)
	} else {
		r.PushState = web.Location(url.Values{actionPanelOpenParamName: []string{}}).MergeQuery(true)
	}

	return
}

const ActiveFilterTabQueryKey = "active_filter_tab"

func (b *ListingBuilder) filterTabs(
	ctx *web.EventContext,
	inDialog bool,
) (r h.HTMLComponent) {
	if b.filterTabsFunc == nil {
		return
	}

	qs := ctx.R.URL.Query()

	tabs := VTabs().
		Class("mb-2").
		ShowArrows(true).
		Color("primary").
		Density(DensityCompact)

	tabsData := b.filterTabsFunc(ctx)
	for i, tab := range tabsData {
		if tab.ID == "" {
			tab.ID = fmt.Sprintf("tab%d", i)
		}
	}
	value := -1
	activeTabValue := qs.Get(ActiveFilterTabQueryKey)

	for i, td := range tabsData {
		// Find selected tab by active_filter_tab=xx in the url query
		if activeTabValue == td.ID {
			value = i
		}

		tabContent := h.Text(td.Label)
		if td.AdvancedLabel != nil {
			tabContent = td.AdvancedLabel
		}

		totalQuery := url.Values{}
		totalQuery.Set(ActiveFilterTabQueryKey, td.ID)
		for k, v := range td.Query {
			totalQuery[k] = v
		}

		onclick := web.Plaid().Queries(totalQuery)
		if inDialog {
			onclick.URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog)
		} else {
			onclick.PushState(true)
		}
		tabs.AppendChildren(
			VTab(tabContent).
				Attr("@click", onclick.Go()),
		)
	}
	return tabs.ModelValue(value)
}

type selectColumns struct {
	DisplayColumns []string       `json:"displayColumns,omitempty"`
	SortedColumns  []sortedColumn `json:"sortedColumns,omitempty"`
}
type sortedColumn struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

func (b *ListingBuilder) selectColumnsBtn(
	pageURL *url.URL,
	ctx *web.EventContext,
	inDialog bool,
) (btn h.HTMLComponent, displaySortedFields []*FieldBuilder) {
	var (
		_, respath         = path.Split(pageURL.Path)
		displayColumnsName = fmt.Sprintf("%s_display_columns", respath)
		sortedColumnsName  = fmt.Sprintf("%s_sorted_columns", respath)
		originalColumns    []string
		displayColumns     []string
		sortedColumns      []string
	)

	for _, f := range b.fields {
		if b.mb.Info().Verifier().Do(PermList).SnakeOn("f_"+f.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}
		originalColumns = append(originalColumns, f.name)
	}

	// get the columns setting from url params or cookie data
	if urldata := pageURL.Query().Get(displayColumnsName); urldata != "" {
		if urlColumns := strings.Split(urldata, ","); len(urlColumns) > 0 {
			displayColumns = urlColumns
		}
	}

	if urldata := pageURL.Query().Get(sortedColumnsName); urldata != "" {
		if urlColumns := strings.Split(urldata, ","); len(urlColumns) > 0 {
			sortedColumns = urlColumns
		}
	}

	// get the columns setting from  cookie data
	if len(displayColumns) == 0 {
		cookiedata, err := ctx.R.Cookie(displayColumnsName)
		if err == nil {
			if cookieColumns := strings.Split(cookiedata.Value, ","); len(cookieColumns) > 0 {
				displayColumns = cookieColumns
			}
		}
	}

	if len(sortedColumns) == 0 {
		cookiedata, err := ctx.R.Cookie(sortedColumnsName)
		if err == nil {
			if cookieColumns := strings.Split(cookiedata.Value, ","); len(cookieColumns) > 0 {
				sortedColumns = cookieColumns
			}
		}
	}

	// check if listing fileds is changed. if yes, use the original columns
	var originalFiledsChanged bool

	if len(sortedColumns) > 0 && len(originalColumns) != len(sortedColumns) {
		originalFiledsChanged = true
	}

	if len(sortedColumns) > 0 && !originalFiledsChanged {
		for _, sortedColumn := range sortedColumns {
			var find bool
			for _, originalColumn := range originalColumns {
				if sortedColumn == originalColumn {
					find = true
					break
				}
			}
			if !find {
				originalFiledsChanged = true
				break
			}
		}
	}

	if len(displayColumns) > 0 && !originalFiledsChanged {
		for _, displayColumn := range displayColumns {
			var find bool
			for _, originalColumn := range originalColumns {
				if displayColumn == originalColumn {
					find = true
					break
				}
			}
			if !find {
				originalFiledsChanged = true
				break
			}
		}
	}

	// save display columns setting on cookie
	if !originalFiledsChanged && len(displayColumns) > 0 {
		http.SetCookie(ctx.W, &http.Cookie{
			Name:  displayColumnsName,
			Value: strings.Join(displayColumns, ","),
		})
	}

	// save sorted columns setting on cookie
	if !originalFiledsChanged && len(sortedColumns) > 0 {
		http.SetCookie(ctx.W, &http.Cookie{
			Name:  sortedColumnsName,
			Value: strings.Join(sortedColumns, ","),
		})
	}

	// set the data for displaySortedFields on data table
	if originalFiledsChanged || (len(sortedColumns) == 0 && len(displayColumns) == 0) {
		displaySortedFields = b.fields
	}

	if originalFiledsChanged || len(displayColumns) == 0 {
		displayColumns = originalColumns
	}

	if originalFiledsChanged || len(sortedColumns) == 0 {
		sortedColumns = originalColumns
	}

	if len(displaySortedFields) == 0 {
		for _, sortedColumn := range sortedColumns {
			for _, displayColumn := range displayColumns {
				if sortedColumn == displayColumn {
					displaySortedFields = append(displaySortedFields, b.Field(sortedColumn))
					break
				}
			}
		}
	}

	// set the data for selected columns on toolbar
	selectColumns := selectColumns{
		DisplayColumns: displayColumns,
	}
	for _, sc := range sortedColumns {
		selectColumns.SortedColumns = append(selectColumns.SortedColumns, sortedColumn{
			Name:  sc,
			Label: i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.mb.getLabel(b.Field(sc).NameLabel)),
		})
	}

	msgr := MustGetMessages(ctx.R)
	onOK := web.Plaid().
		Query(displayColumnsName, web.Var("locals.displayColumns")).
		Query(sortedColumnsName, web.Var("locals.sortedColumns.map(column => column.name )")).
		MergeQuery(true)
	if inDialog {
		onOK.URL(ctx.R.RequestURI).
			EventFunc(actions.UpdateListingDialog)
	}
	// add the HTML component of columns setting into toolbar
	btn = web.Scope(VMenu(
		web.Slot(
			VBtn("").Icon("mdi-cog").Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
		).Name("activator").Scope("{ props }"),
		VList(
			h.Tag("vx-draggable").Attr("item-key", "name").Attr("v-model", "locals.sortedColumns", "handle", ".handle", "animation", "300").Children(
				h.Template(
					VListItem(
						VListItemTitle(
							VSwitch().Density(DensityCompact).Attr("v-model", "locals.displayColumns", ":value",
								"element.name",
								":label", "element.label").Color("primary").Class(" mt-2 "),
							VIcon("mdi-reorder-vertical").Class("handle cursor-grab mt-4"),
						).Class("d-flex justify-space-between "),
						VDivider(),
					),
				).Attr("#item", " { element } "),
			),
			VListItem(
				VBtn(msgr.Cancel).Elevation(0).Attr("@click", `locals.selectColumnsMenu = false`),
				VBtn(msgr.OK).Elevation(0).Color("primary").Attr("@click", `locals.selectColumnsMenu = false ; `+onOK.Go()),
			).Class("d-flex justify-space-between"),
		).Density(DensityCompact),
	).CloseOnContentClick(false).Width(240).
		Attr("v-model", "locals.selectColumnsMenu")).
		VSlot("{ locals }").Init(fmt.Sprintf(`{selectColumnsMenu: false,...%s}`, h.JSONString(selectColumns)))
	return
}

func (b *ListingBuilder) filterBar(
	ctx *web.EventContext,
	msgr *Messages,
	fd vx.FilterData,
	inDialog bool,
) (filterBar h.HTMLComponent) {
	if fd == nil {
		return nil
	}
	noVisiableItem := true
	for _, d := range fd {
		if !d.Invisible {
			noVisiableItem = false
			break
		}
	}
	if noVisiableItem {
		return nil
	}

	ft := vx.FilterTranslations{}
	ft.Clear = msgr.FiltersClear
	ft.Add = msgr.FiltersAdd
	ft.Apply = msgr.FilterApply
	for _, d := range fd {
		d.Translations = vx.FilterIndependentTranslations{
			FilterBy: msgr.FilterBy(d.Label),
		}
	}

	ft.Date.To = msgr.FiltersDateTo

	ft.Number.And = msgr.FiltersNumberAnd
	ft.Number.Equals = msgr.FiltersNumberEquals
	ft.Number.Between = msgr.FiltersNumberBetween
	ft.Number.GreaterThan = msgr.FiltersNumberGreaterThan
	ft.Number.LessThan = msgr.FiltersNumberLessThan

	ft.String.Equals = msgr.FiltersStringEquals
	ft.String.Contains = msgr.FiltersStringContains

	ft.MultipleSelect.In = msgr.FiltersMultipleSelectIn
	ft.MultipleSelect.NotIn = msgr.FiltersMultipleSelectNotIn

	filter := vx.VXFilter(fd).Translations(ft)
	if inDialog {
		filter.UpdateModelValue(web.Plaid().
			URL(ctx.R.RequestURI).
			StringQuery(web.Var("$event.encodedFilterData")).
			Query("page", 1).
			ClearMergeQuery(web.Var("$event.filterKeys")).
			EventFunc(actions.UpdateListingDialog).
			Go())
	}
	return filter
}

func getLocalPerPage(
	ctx *web.EventContext,
	mb *ModelBuilder,
) int64 {
	// c is the cookie value of a serials of per page value, split by "$".
	// each value is split by "#".
	// the first part is the uri name of the model builder.
	// the second part is the per page value.
	c, err := ctx.R.Cookie("_perPage")
	if err != nil {
		return 0
	}
	vals := strings.Split(c.Value, "$")
	for _, v := range vals {
		vvs := strings.Split(v, "#")
		if len(vvs) != 2 {
			continue
		}
		if vvs[0] == mb.uriName {
			r, _ := strconv.ParseInt(vvs[1], 10, 64)
			return r
		}
	}

	return 0
}

// setLocalPerPage set the per page value to cookie.
// v is the per page value to set.
func setLocalPerPage(
	ctx *web.EventContext,
	mb *ModelBuilder,
	v int64,
) {
	var oldVals []string
	{
		c, err := ctx.R.Cookie("_perPage")
		if err == nil {
			oldVals = strings.Split(c.Value, "$")
		}
	}
	newVals := []string{fmt.Sprintf("%s#%d", mb.uriName, v)}
	for _, v := range oldVals {
		vvs := strings.Split(v, "#")
		if len(vvs) != 2 {
			continue
		}
		if vvs[0] == mb.uriName {
			continue
		}
		newVals = append(newVals, v)
	}
	http.SetCookie(ctx.W, &http.Cookie{
		Name:  "_perPage",
		Value: strings.Join(newVals, "$"),
	})
}

type ColOrderBy struct {
	FieldName string
	// ASC, DESC
	OrderBy string
}

// GetOrderBysFromQuery gets order bys from query string.
func GetOrderBysFromQuery(query url.Values) []*ColOrderBy {
	r := make([]*ColOrderBy, 0)
	// qs is like "field1_ASC,field2_DESC"
	qs := strings.Split(query.Get("order_by"), ",")
	for _, q := range qs {
		ss := strings.Split(q, "_")
		ssl := len(ss)
		if ssl == 1 {
			continue
		}
		if ss[ssl-1] != "ASC" && ss[ssl-1] != "DESC" {
			continue
		}
		r = append(r, &ColOrderBy{
			FieldName: strings.Join(ss[:ssl-1], "_"),
			OrderBy:   ss[ssl-1],
		})
	}

	return r
}

func newQueryWithFieldToggleOrderBy(query url.Values, fieldName string) url.Values {
	oldOrderBys := GetOrderBysFromQuery(query)
	var newOrderBysQueryValue []string
	existed := false
	for _, oob := range oldOrderBys {
		if oob.FieldName == fieldName {
			existed = true
			if oob.OrderBy == "ASC" {
				newOrderBysQueryValue = append(newOrderBysQueryValue, oob.FieldName+"_DESC")
			}
			continue
		}
		newOrderBysQueryValue = append(newOrderBysQueryValue, oob.FieldName+"_"+oob.OrderBy)
	}
	if !existed {
		newOrderBysQueryValue = append(newOrderBysQueryValue, fieldName+"_ASC")
	}

	newQuery := make(url.Values)
	for k, v := range query {
		newQuery[k] = v
	}
	newQuery.Set("order_by", strings.Join(newOrderBysQueryValue, ","))
	return newQuery
}

func (b *ListingBuilder) getTableComponents(
	ctx *web.EventContext,
	inDialog bool,
) (
	dataTable h.HTMLComponent,
	// pagination, no-record message
	datatableAdditions h.HTMLComponent,
) {
	msgr := MustGetMessages(ctx.R)

	qs := ctx.R.URL.Query()

	var perPage int64 = 0
	if !b.disablePagination {
		var requestPerPage int64
		qPerPageStr := qs.Get("per_page")
		qPerPage, _ := strconv.ParseInt(qPerPageStr, 10, 64)
		if qPerPage != 0 {
			setLocalPerPage(ctx, b.mb, qPerPage)
			requestPerPage = qPerPage
		} else if cPerPage := getLocalPerPage(ctx, b.mb); cPerPage != 0 {
			requestPerPage = cPerPage
		}

		perPage = b.perPage
		if requestPerPage != 0 {
			perPage = requestPerPage
		}
		if perPage == 0 {
			perPage = 50
		}
		if perPage > 1000 {
			perPage = 1000
		}

	}

	var orderBySQL string
	orderBys := GetOrderBysFromQuery(qs)
	// map[FieldName]DBColumn
	orderableFieldMap := make(map[string]string)
	for _, v := range b.orderableFields {
		orderableFieldMap[v.FieldName] = v.DBColumn
	}
	for _, ob := range orderBys {
		dbCol, ok := orderableFieldMap[ob.FieldName]
		if !ok {
			continue
		}
		orderBySQL += fmt.Sprintf("%s %s,", dbCol, ob.OrderBy)
	}
	// remove the last ","
	if orderBySQL != "" {
		orderBySQL = orderBySQL[:len(orderBySQL)-1]
	}
	if orderBySQL == "" {
		if b.orderBy != "" {
			orderBySQL = b.orderBy
		} else {
			orderBySQL = fmt.Sprintf("%s DESC", b.mb.primaryField)
		}
	}
	searchParams := &SearchParams{
		KeywordColumns: b.searchColumns,
		Keyword:        qs.Get("keyword"),
		PerPage:        perPage,
		OrderBy:        orderBySQL,
		PageURL:        ctx.R.URL,
		SQLConditions:  b.conditions,
	}

	searchParams.Page, _ = strconv.ParseInt(qs.Get("page"), 10, 64)
	if searchParams.Page == 0 {
		searchParams.Page = 1
	}

	var fd vx.FilterData
	if b.filterDataFunc != nil {
		fd = b.filterDataFunc(ctx)
		cond, args := fd.SetByQueryString(ctx.R.URL.RawQuery)

		searchParams.SQLConditions = append(searchParams.SQLConditions, &SQLCondition{
			Query: cond,
			Args:  args,
		})
	}

	if b.Searcher == nil || b.mb.p.dataOperator == nil {
		panic("presets.New().DataOperator(...) required")
	}

	var objs interface{}
	var totalCount int
	var err error

	objs, totalCount, err = b.Searcher(b.mb.NewModelSlice(), searchParams, ctx)

	if err != nil {
		panic(err)
	}

	haveCheckboxes := len(b.bulkActions) > 0

	var cellWrapperFunc = func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		tdbind := cell
		if b.mb.hasDetailing && !b.mb.detailing.drawer {
			tdbind.SetAttr("@click.self", web.Plaid().
				PushStateURL(b.mb.Info().DetailingHref(id)).Go())
		} else {
			event := actions.Edit
			if b.mb.hasDetailing {
				event = actions.DetailingDrawer
			}
			onclick := web.Plaid().
				EventFunc(event).
				Query(ParamID, id)
			if inDialog {
				onclick.URL(ctx.R.RequestURI).
					Query(ParamOverlay, actions.Dialog).
					Query(ParamInDialog, true).
					Query(ParamListingQueries, ctx.Queries().Encode())
			}
			tdbind.SetAttr("@click.self",
				onclick.Go()+fmt.Sprintf(`; locals.currEditingListItemID="%s-%s"`, dataTableID, id))
		}
		return tdbind
	}
	if b.cellWrapperFunc != nil {
		cellWrapperFunc = b.cellWrapperFunc
	}

	var displayFields = b.fields
	var selectColumnsBtn h.HTMLComponent
	if b.selectableColumns {
		selectColumnsBtn, displayFields = b.selectColumnsBtn(ctx.R.URL, ctx, inDialog)
	}

	sDataTable := vx.DataTable(objs).
		CellWrapperFunc(cellWrapperFunc).
		HeadCellWrapperFunc(func(cell h.MutableAttrHTMLComponent, field string, title string) h.HTMLComponent {
			if _, ok := orderableFieldMap[field]; ok {
				var orderBy string
				var orderByIdx int
				for i, ob := range orderBys {
					if ob.FieldName == field {
						orderBy = ob.OrderBy
						orderByIdx = i + 1
						break
					}
				}
				th := h.Th("").Style("cursor: pointer; white-space: nowrap;").
					Children(
						h.Span(title).
							Style("text-decoration: underline;"),
						h.If(orderBy == "ASC",
							VIcon("arrow_drop_up").Size(SizeSmall),
							h.Span(fmt.Sprint(orderByIdx)),
						).ElseIf(orderBy == "DESC",
							VIcon("arrow_drop_down").Size(SizeSmall),
							h.Span(fmt.Sprint(orderByIdx)),
						).Else(
							// take up place
							h.Span("").Style("visibility: hidden;").Children(
								VIcon("arrow_drop_down").Size(SizeSmall),
								h.Span(fmt.Sprint(orderByIdx)),
							),
						),
					)
				qs.Del("__execute_event__")
				newQuery := newQueryWithFieldToggleOrderBy(qs, field)
				onclick := web.Plaid().
					Queries(newQuery)
				if inDialog {
					onclick.URL(ctx.R.RequestURI).
						EventFunc(actions.UpdateListingDialog)
				} else {
					onclick.PushState(true)
				}
				th.Attr("@click", onclick.Go())

				cell = th
			}

			return cell
		}).
		RowWrapperFunc(func(row h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
			row.SetAttr(":class", fmt.Sprintf(`{"vx-list-item--active primary--text": vars.presetsRightDrawer && locals.currEditingListItemID==="%s-%s"}`, dataTableID, id))
			return row
		}).
		RowMenuItemFuncs(b.RowMenu().listingItemFuncs(ctx)...).
		Selectable(haveCheckboxes).
		SelectionParamName(ParamSelectedIds).
		SelectedCountLabel(msgr.ListingSelectedCountNotice).
		SelectableColumnsBtn(selectColumnsBtn).
		ClearSelectionLabel(msgr.ListingClearSelection)
	if inDialog {
		sDataTable.OnSelectAllFunc(func(idsOfPage []string, ctx *web.EventContext) string {
			return web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				Query(ParamSelectedIds,
					web.Var(fmt.Sprintf(`{value: %s, add: $event, remove: !$event}`, h.JSONString(idsOfPage))),
				).
				MergeQuery(true).
				Go()
		})
		sDataTable.OnSelectFunc(func(id string, ctx *web.EventContext) string {
			return web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				Query(ParamSelectedIds,
					web.Var(fmt.Sprintf(`{value: %s, add: $event, remove: !$event}`, h.JSONString(id))),
				).
				MergeQuery(true).
				Go()
		})
		sDataTable.OnClearSelectionFunc(func(ctx *web.EventContext) string {
			return web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				Query(ParamSelectedIds, "").
				MergeQuery(true).
				Go()
		})
	}
	dataTable = sDataTable

	for _, f := range displayFields {
		if b.mb.Info().Verifier().Do(PermList).SnakeOn("f_"+f.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}
		f = b.getFieldOrDefault(f.name) // fill in empty compFunc and setter func with default
		dataTable.(*vx.DataTableBuilder).Column(f.name).
			Title(i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.mb.getLabel(f.NameLabel))).
			CellComponentFunc(b.cellComponentFunc(f))
	}

	if b.disablePagination {
		// if disable pagination, we don't need to add
		// the pagination component and the no-record message to page.
		return
	}
	if totalCount > 0 {
		tpb := vx.VXTablePagination().
			Total(int64(totalCount)).
			CurrPage(searchParams.Page).
			PerPage(searchParams.PerPage).
			CustomPerPages([]int64{b.perPage}).
			PerPageText(msgr.PaginationRowsPerPage)

		if inDialog {
			tpb.OnSelectPerPage(web.Plaid().
				URL(ctx.R.RequestURI).
				Query("per_page", web.Var("[$event]")).
				MergeQuery(true).
				EventFunc(actions.UpdateListingDialog).
				Go())
			tpb.OnPrevPage(web.Plaid().
				URL(ctx.R.RequestURI).
				Query("page", searchParams.Page-1).
				MergeQuery(true).
				EventFunc(actions.UpdateListingDialog).
				Go())
			tpb.OnNextPage(web.Plaid().
				URL(ctx.R.RequestURI).
				Query("page", searchParams.Page+1).
				MergeQuery(true).
				EventFunc(actions.UpdateListingDialog).
				Go())
		}

		datatableAdditions = h.Div(tpb).Class("mt-2")
	} else {
		datatableAdditions = h.Div(h.Text(msgr.ListingNoRecordToShow)).Class("mt-10 text-center grey--text text--darken-2")
	}

	return
}

func (b *ListingBuilder) reloadList(ctx *web.EventContext) (r web.EventResponse, err error) {
	dataTable, dataTableAdditions := b.getTableComponents(ctx, false)
	r.UpdatePortals = append(r.UpdatePortals,
		&web.PortalUpdate{
			Name: dataTablePortalName,
			Body: dataTable,
		},
		&web.PortalUpdate{
			Name: dataTableAdditionsPortalName,
			Body: dataTableAdditions,
		},
	)

	return
}

func (b *ListingBuilder) actionsComponent(
	msgr *Messages,
	ctx *web.EventContext,
	inDialog bool,
) h.HTMLComponent {
	var actionBtns []h.HTMLComponent

	// Render bulk actions
	for _, ba := range b.bulkActions {
		if b.mb.Info().Verifier().SnakeDo(PermBulkActions, ba.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		var btn h.HTMLComponent
		if ba.buttonCompFunc != nil {
			btn = ba.buttonCompFunc(ctx)
		} else {
			buttonColor := ba.buttonColor
			if buttonColor == "" {
				buttonColor = ColorSecondary
			}
			onclick := web.Plaid().EventFunc(actions.OpenBulkActionDialog).
				Queries(url.Values{bulkPanelOpenParamName: []string{ba.name}}).
				MergeQuery(true)
			if inDialog {
				onclick.URL(ctx.R.RequestURI).
					Query(ParamInDialog, inDialog)
			}
			btn = VBtn(b.mb.getLabel(ba.NameLabel)).
				Color(buttonColor).
				Variant(VariantFlat).
				Class("ml-2").
				Attr("@click", onclick.Go())
		}

		actionBtns = append(actionBtns, btn)
	}
	b.newBtnFunc = nil
	// Render actions
	for _, ba := range b.actions {
		if b.mb.Info().Verifier().SnakeDo(PermActions, ba.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		var btn h.HTMLComponent
		if ba.buttonCompFunc != nil {
			btn = ba.buttonCompFunc(ctx)
		} else {
			buttonColor := ba.buttonColor
			if buttonColor == "" {
				buttonColor = ColorPrimary
			}

			onclick := web.Plaid().EventFunc(actions.OpenActionDialog).
				Queries(url.Values{actionPanelOpenParamName: []string{ba.name}}).
				MergeQuery(true)
			if inDialog {
				onclick.URL(ctx.R.RequestURI).
					Query(ParamInDialog, inDialog)
			}
			btn = VBtn(b.mb.getLabel(ba.NameLabel)).
				Color(buttonColor).
				Variant(VariantFlat).
				Class("ml-2").
				Attr("@click", onclick.Go())
		}

		actionBtns = append(actionBtns, btn)
	}

	// if len(actionBtns) == 0 {
	//	return nil
	// }

	if b.actionsAsMenu {
		var listItems []h.HTMLComponent
		for _, btn := range actionBtns {
			listItems = append(listItems, VListItem(btn))
		}
		return h.Components(VMenu(
			web.Slot(
				VBtn("Actions").
					Attr("v-bind", "attrs").
					Attr("v-on", "on"),
			).Name("activator").Scope("{ on, attrs }"),
			VList(listItems...),
		).OpenOnHover(true))
	}

	if b.newBtnFunc != nil {
		if btn := b.newBtnFunc(ctx); btn != nil {
			actionBtns = append(actionBtns, b.newBtnFunc(ctx))
		}
	} else {
		disableNewBtn := b.mb.Info().Verifier().Do(PermCreate).WithReq(ctx.R).IsAllowed() != nil
		if !disableNewBtn {
			onclick := web.Plaid().EventFunc(actions.New)
			if inDialog {
				onclick.URL(ctx.R.RequestURI).
					Query(ParamOverlay, actions.Dialog).
					Query(ParamInDialog, true).
					Query(ParamListingQueries, ctx.Queries().Encode())
			}
			actionBtns = append(actionBtns, VBtn(msgr.New).
				Color("primary").
				Variant(VariantFlat).
				Theme("dark").Class("ml-2").
				Disabled(disableNewBtn).
				Attr("@click", onclick.Go()))
		}
	}
	return h.Div(actionBtns...)
}

func (b *ListingBuilder) openListingDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	content := VCard(
		web.Portal(b.listingComponent(ctx, true)).
			Name(listingDialogContentPortalName),
	).Attr("id", "listingDialog")
	dialog := VDialog(content).
		Attr("v-model", "vars.presetsListingDialog")
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

func (b *ListingBuilder) updateListingDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: listingDialogContentPortalName,
		Body: b.listingComponent(ctx, true),
	})

	web.AppendRunScripts(&r, `
var listingDialogElem = document.getElementById('listingDialog'); 
if (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {
    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';
};`)

	return
}
