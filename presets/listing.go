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

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/perm"
	"github.com/qor/qor5/presets/actions"
	"github.com/goplaid/ui/stripeui"
	s "github.com/goplaid/ui/stripeui"
	. "github.com/goplaid/ui/vuetify"
	"github.com/goplaid/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type ListingBuilder struct {
	mb                *ModelBuilder
	bulkActions       []*ActionBuilder
	actions           []*ActionBuilder
	actionsAsMenu     bool
	rowMenu           *RowMenuBuilder
	filterDataFunc    FilterDataFunc
	filterTabsFunc    FilterTabsFunc
	newBtnFunc        ComponentFunc
	pageFunc          web.PageFunc
	cellWrapperFunc   stripeui.CellWrapperFunc
	Searcher          SearchFunc
	searchColumns     []string
	perPage           int64
	totalVisible      int64
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

func (b *ListingBuilder) PageFunc(pf web.PageFunc) (r *ListingBuilder) {
	b.pageFunc = pf
	return b
}

func (b *ListingBuilder) CellWrapperFunc(cwf stripeui.CellWrapperFunc) (r *ListingBuilder) {
	b.cellWrapperFunc = cwf
	return b
}

func (b *ListingBuilder) SearchFunc(v SearchFunc) (r *ListingBuilder) {
	b.Searcher = v
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

func (b *ListingBuilder) TotalVisible(v int64) (r *ListingBuilder) {
	b.totalVisible = v
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
	title := msgr.ListingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, b.mb.label))
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

		var actionsComponent h.HTMLComponents
		if v := b.actionsComponent(msgr, ctx, inDialog); v != nil {
			actionsComponent = append(actionsComponent, v)
		}
		if b.newBtnFunc != nil {
			if btn := b.newBtnFunc(ctx); btn != nil {
				actionsComponent = append(actionsComponent, b.newBtnFunc(ctx))
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
				actionsComponent = append(actionsComponent, VBtn(msgr.New).
					Color("primary").
					Depressed(true).
					Dark(true).Class("ml-2").
					Disabled(disableNewBtn).
					Attr("@click", onclick.Go()))
			}
		}

		if filterTabs != nil || len(actionsComponent) > 0 {
			tabsAndActionsBar = VToolbar(
				filterTabs,
				VSpacer(),
				actionsComponent,
			).Flat(true)
		}
	}

	var filterBar h.HTMLComponent
	if b.filterDataFunc != nil {
		fd := b.filterDataFunc(ctx)
		fd.SetByQueryString(ctx.R.URL.RawQuery)
		filterBar = b.filterBar(ctx, msgr, fd, inDialog)
	}

	dataTable, dataTableAdditions := b.getTableComponents(ctx, inDialog)

	var dialogHeadbar h.HTMLComponent
	if inDialog {
		title := msgr.ListingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, b.mb.label))
		var searchBox h.HTMLComponent
		if b.mb.layoutConfig == nil || !b.mb.layoutConfig.SearchBoxInvisible {
			searchBox = VTextField().
				PrependInnerIcon("search").
				Placeholder(msgr.Search).
				HideDetails(true).
				Value(ctx.R.URL.Query().Get("keyword")).
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
		dialogHeadbar = VAppBar(
			VToolbarTitle("").
				Children(h.Text(title)),
			VSpacer(),
			searchBox,
			VBtn("").Icon(true).
				Children(VIcon("close")).
				Large(true).
				Attr("@click.stop", CloseListingDialogVarScript),
		).Color("white").Elevation(0).Dense(true)
	}

	return VContainer(
		dialogHeadbar,
		tabsAndActionsBar,
		h.Div(
			VCard(
				filterBar,
				VDivider(),
				VCardText(
					web.Portal(dataTable).Name(dataTablePortalName),
				).Class("pa-0"),
			),
			web.Portal(dataTableAdditions).Name(dataTableAdditionsPortalName),
		).Class("mt-2"),
	).Fluid(true).
		Class("white").
		Attr(web.InitContextVars, `{currEditingListItemID: ''}`)
}

func (b *ListingBuilder) cellComponentFunc(f *FieldBuilder) s.CellComponentFunc {
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
	bulk *ActionBuilder,
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
				Elevation(2).
				ColoredBorder(true)
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
				Depressed(true).
				Class("ml-2").
				Attr("@click", closeDialogVarScript),

			VBtn(msgr.OK).
				Color("primary").
				Depressed(true).
				Dark(true).
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
				Elevation(2).
				ColoredBorder(true)
		}
	}

	onOK := web.Plaid().EventFunc(actions.DoListingAction).
		Query(ParamListingActionName, action.name).
		MergeQuery(true)
	if isInDialogFromQuery(ctx) {
		onOK.URL(ctx.R.RequestURI)
	}

	return VCard(
		VCardTitle(
			h.Text(action.NameLabel.label),
		),
		VCardText(
			errComp,
			action.compFunc([]string{}, ctx), // because action and bulk action shared the same func, so pass blank slice here
		),
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).
				Depressed(true).
				Class("ml-2").
				Attr("@click", closeDialogVarScript),

			VBtn(msgr.OK).
				Color("primary").
				Depressed(true).
				Dark(true).
				Attr("@click", onOK.Go()),
		),
	)
}

func (b *ListingBuilder) deleteConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	msgr := MustGetMessages(ctx.R)
	id := ctx.R.FormValue(ParamID)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: DeleteConfirmPortalName,
		Body: VDialog(
			VCard(
				VCardTitle(h.Text(msgr.DeleteConfirmationText(id))),
				VCardActions(
					VSpacer(),
					VBtn(msgr.Cancel).
						Depressed(true).
						Class("ml-2").
						On("click", "vars.deleteConfirmation = false"),

					VBtn(msgr.Delete).
						Color("primary").
						Depressed(true).
						Dark(true).
						Attr("@click", web.Plaid().
							EventFunc(actions.DoDelete).
							Queries(ctx.Queries()).
							URL(ctx.R.URL.Path).
							Go()),
				),
			),
		).MaxWidth("600px").
			Attr("v-model", "vars.deleteConfirmation").
			Attr(web.InitContextVars, `{deleteConfirmation: false}`),
	})

	r.VarsScript = "setTimeout(function(){ vars.deleteConfirmation = true }, 100)"
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
	bulk := getAction(b.bulkActions, bulkName)

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
	bulk := getAction(b.bulkActions, ctx.R.FormValue(ParamBulkActionName))
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
		web.AppendVarsScripts(&r,
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

func (b ListingBuilder) doListingAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamListingActionName))
	if action == nil {
		panic("action required")
	}

	if b.mb.Info().Verifier().SnakeDo(PermListingActions, action.name).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	err1 := action.updateFunc([]string{}, ctx)

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
		web.AppendVarsScripts(&r,
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

	tabs := VTabs().ShowArrows(true)
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
	return tabs.Value(value)
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
		if b.mb.Info().Verifier().Do(PermList).SnakeOn(f.name).WithReq(ctx.R).IsAllowed() != nil {
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
	btn = VMenu(
		web.Slot(
			VBtn("").Children(VIcon("settings")).Attr("v-on", "on").Text(true).Fab(true).Small(true),
		).Name("activator").Scope("{ on }"),

		web.Scope(VList(
			h.Tag("vx-draggable").Attr("v-model", "locals.sortedColumns", "draggable", ".vx_column_item", "animation", "300").Children(
				h.Div(
					VListItem(
						VListItemContent(
							VListItemTitle(
								VSwitch().Dense(true).Attr("v-model", "locals.displayColumns", ":value", "column.name", ":label", "column.label", "@click", "event.preventDefault()"),
							),
						),
						VListItemIcon(
							VIcon("reorder"),
						).Attr("style", "margin-top: 28px"),
					),
					VDivider(),
				).Attr("v-for", "(column, index) in locals.sortedColumns", ":key", "column.name", "class", "vx_column_item"),
			),
			VListItem(
				VListItemAction(VBtn(msgr.Cancel).Elevation(0).Attr("@click", `vars.selectColumnsMenu = false`)),
				VListItemAction(VBtn(msgr.OK).Elevation(0).Color("primary").Attr("@click", `vars.selectColumnsMenu = false;`+onOK.Go()))),
		).Dense(true)).
			Init(h.JSONString(selectColumns)).
			VSlot("{ locals }"),
	).OffsetY(true).CloseOnClick(false).CloseOnContentClick(false).
		Attr(web.InitContextVars, `{selectColumnsMenu: false}`).
		Attr("v-model", "vars.selectColumnsMenu")
	return
}

func (b *ListingBuilder) filterBar(
	ctx *web.EventContext,
	msgr *Messages,
	fd vuetifyx.FilterData,
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

	ft := vuetifyx.FilterTranslations{}
	ft.Clear = msgr.FiltersClear
	ft.Add = msgr.FiltersAdd
	ft.Apply = msgr.FilterApply
	for _, d := range fd {
		d.Translations = vuetifyx.FilterIndependentTranslations{
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

	filter := vuetifyx.VXFilter(fd).Translations(ft)
	if inDialog {
		filter.OnChange(web.Plaid().
			URL(ctx.R.RequestURI).
			StringQuery(web.Var("$event.encodedFilterData")).
			ClearMergeQuery(web.Var("$event.filterKeys")).
			EventFunc(actions.UpdateListingDialog).
			Go())
	}
	return VToolbar(
		filter,
	).Flat(true).AutoHeight(true).Class("py-2")
}

func getLocalPerPage(
	ctx *web.EventContext,
	mb *ModelBuilder,
) int64 {
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

func GetOrderBysFromQuery(query url.Values) []*ColOrderBy {
	r := make([]*ColOrderBy, 0)
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
	newOrderBysQueryValue := []string{}
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

	var requestPerPage int64
	qPerPageStr := qs.Get("per_page")
	qPerPage, _ := strconv.ParseInt(qPerPageStr, 10, 64)
	if qPerPage != 0 {
		setLocalPerPage(ctx, b.mb, qPerPage)
		requestPerPage = qPerPage
	} else if cPerPage := getLocalPerPage(ctx, b.mb); cPerPage != 0 {
		requestPerPage = cPerPage
	}
	perPage := b.perPage
	if requestPerPage != 0 {
		perPage = requestPerPage
	}
	if perPage == 0 {
		perPage = 50
	}
	if perPage > 1000 {
		perPage = 1000
	}

	totalVisible := b.totalVisible
	if totalVisible == 0 {
		totalVisible = 10
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

	var fd vuetifyx.FilterData
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

	pagesCount := int(int64(totalCount)/searchParams.PerPage + 1)
	if int64(totalCount)%searchParams.PerPage == 0 {
		pagesCount--
	}

	var cellWraperFunc = func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		tdbind := cell
		if b.mb.hasDetailing && !b.mb.detailing.drawer {
			tdbind.SetAttr("@click.self", web.Plaid().
				PushStateURL(
					b.mb.Info().
						DetailingHref(id)).
				Go())
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
				onclick.Go()+fmt.Sprintf(`; vars.currEditingListItemID="%s-%s"`, dataTableID, id))
		}
		return tdbind
	}
	if b.cellWrapperFunc != nil {
		cellWraperFunc = b.cellWrapperFunc
	}

	var displayFields = b.fields
	var selectColumnsBtn h.HTMLComponent
	if b.selectableColumns {
		selectColumnsBtn, displayFields = b.selectColumnsBtn(ctx.R.URL, ctx, inDialog)
	}

	sDataTable := s.DataTable(objs).
		CellWrapperFunc(cellWraperFunc).
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
							VIcon("arrow_drop_up").Small(true),
							h.Span(fmt.Sprint(orderByIdx)),
						).ElseIf(orderBy == "DESC",
							VIcon("arrow_drop_down").Small(true),
							h.Span(fmt.Sprint(orderByIdx)),
						).Else(
							// take up place
							h.Span("").Style("visibility: hidden;").Children(
								VIcon("arrow_drop_down").Small(true),
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
			row.SetAttr(":class", fmt.Sprintf(`{"blue lighten-5": vars.presetsRightDrawer && vars.currEditingListItemID==="%s-%s"}`, dataTableID, id))
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
		if b.mb.Info().Verifier().Do(PermList).SnakeOn(f.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}
		f = b.getFieldOrDefault(f.name) // fill in empty compFunc and setter func with default
		dataTable.(*stripeui.DataTableBuilder).Column(f.name).
			Title(i18n.PT(ctx.R, ModelsI18nModuleKey, b.mb.label, b.mb.getLabel(f.NameLabel))).
			CellComponentFunc(b.cellComponentFunc(f))
	}

	if totalCount > 0 {
		tpb := vuetifyx.VXTablePagination().
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

		datatableAdditions = tpb
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
				Depressed(true).
				Dark(true).
				Class("ml-2").
				Attr("@click", onclick.Go())
		}

		actionBtns = append(actionBtns, btn)
	}

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
				Depressed(true).
				Dark(true).
				Class("ml-2").
				Attr("@click", onclick.Go())
		}

		actionBtns = append(actionBtns, btn)
	}

	if len(actionBtns) == 0 {
		return nil
	}

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
		).OpenOnHover(true).
			OffsetY(true).
			AllowOverflow(true))
	}
	return h.Components(actionBtns...)
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
		Name: listingDialogPortalName,
		Body: web.Scope(dialog).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){ vars.presetsListingDialog = true }, 100)"
	return
}

func (b *ListingBuilder) updateListingDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: listingDialogContentPortalName,
		Body: b.listingComponent(ctx, true),
	})

	web.AppendVarsScripts(&r, `
var listingDialogElem = document.getElementById('listingDialog'); 
if (listingDialogElem.offsetHeight > parseInt(listingDialogElem.style.minHeight || '0', 10)) {
    listingDialogElem.style.minHeight = listingDialogElem.offsetHeight+'px';
};`)

	return
}
