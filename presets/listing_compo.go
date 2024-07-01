package presets

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
)

const (
	PerPageDefault = 50
	PerPageMax     = 1000
)

func init() {
	stateful.RegisterActionableCompoType((*ListingCompo)(nil))
}

type DisplayColumn struct {
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
}

type ListingCompo struct {
	lb *ListingBuilderX `inject:""`

	CompoID            string          `json:"compo_id"`
	Popup              bool            `json:"popup"`
	LongStyleSearchBox bool            `json:"long_style_search_box"`
	SelectedIds        []string        `json:"selected_ids" query:",omitempty"`
	Keyword            string          `json:"keyword" query:",omitempty"`
	OrderBys           []ColOrderBy    `json:"order_bys" query:",omitempty"`
	Page               int64           `json:"page" query:",omitempty"`
	PerPage            int64           `json:"per_page" query:",omitempty;cookie"`
	DisplayColumns     []DisplayColumn `json:"display_columns" query:",omitempty;cookie"`
	ActiveFilterTab    string          `json:"active_filter_tab" query:",omitempty"`
	FilterQuery        string          `json:"filter_query" query:";method:bare,f_"`
}

func (c *ListingCompo) CompoName() string {
	return fmt.Sprintf("ListingCompo_%s", c.CompoID)
}

// const notifListingLocalsUpdated = "NotifListingLocalsUpdated"

func (c *ListingCompo) locals() string {
	return stateful.LocalsActionable(c)
}

const (
	localsKeySelectedIds      = "selected_ids"
	localsKeyCurrentEditingId = "current_editing_id"
)

func (c *ListingCompo) wrapCompo(ctx context.Context, compo h.HTMLComponent) h.HTMLComponent {
	// for selected_ids front-end autonomy
	locals := c.locals()
	localsSelectedIds := locals + "." + localsKeySelectedIds
	localsNewAction := locals + "." + stateful.LocalsKeyNewAction
	return stateful.Actionable(ctx, c,
		// for sync locals between listing compo and actions compo
		web.DataSync(locals, locals),
		// onMounted
		web.RunScript(fmt.Sprintf(`function() {
	%s.%s = "";
	%s = %s || [];
	let orig = %s;
	%s = function() {
		let v = orig();
		v.compo.%s = this.%s;
		return v
	}
}`,
			locals, localsKeyCurrentEditingId,
			localsSelectedIds, h.JSONString(c.SelectedIds),
			localsNewAction,
			localsNewAction,
			localsKeySelectedIds, localsKeySelectedIds,
		)),
		compo,
	)
}

func (c *ListingCompo) MarshalHTML(ctx context.Context) (r []byte, err error) {
	localsSelectedIds := c.locals() + "." + localsKeySelectedIds
	return c.wrapCompo(ctx,
		h.Components(
			web.Observe(c.lb.mb.NotifModelsUpdated(), stateful.ReloadAction(ctx, c, nil).Go()),
			web.Observe(c.lb.mb.NotifModelsDeleted(), fmt.Sprintf(`
if (payload && payload.ids && payload.ids.length > 0) {
	%s = %s.filter(id => !payload.ids.includes(id));
}
%s`,
				localsSelectedIds, localsSelectedIds,
				stateful.ReloadAction(ctx, c, nil).Go(),
			)),
			VCard().Elevation(0).Children(
				c.tabsFilter(ctx),
				c.toolbarSearch(ctx),
				VCardText().Class("pa-2").Children(
					c.dataTable(ctx),
				),
				c.cardActionsFooter(ctx),
			),
		),
	).MarshalHTML(ctx)
}

func (c *ListingCompo) tabsFilter(ctx context.Context) (r h.HTMLComponent) {
	if c.lb.filterTabsFunc == nil {
		return
	}
	evCtx, _ := c.MustGetEventContext(ctx)

	activeIndex := -1
	fts := c.lb.filterTabsFunc(evCtx)
	tabs := VTabs().Class("mb-2").ShowArrows(true).Color(ColorPrimary).Density(DensityCompact)
	for i, ft := range fts {
		if ft.ID == "" {
			ft.ID = fmt.Sprintf("tab%d", i)
		}
		encodedQuery := ft.Query.Encode()
		if c.ActiveFilterTab == ft.ID && encodedQuery == c.FilterQuery {
			activeIndex = i
		}
		tabs.AppendChildren(
			VTab().
				Attr("@click", stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					target.ActiveFilterTab = ft.ID
					target.FilterQuery = encodedQuery
				}).Go()).
				Children(
					h.Iff(ft.AdvancedLabel != nil, func() h.HTMLComponent {
						return ft.AdvancedLabel
					}).Else(func() h.HTMLComponent {
						return h.Text(ft.Label)
					}),
				),
		)
	}
	return tabs.ModelValue(activeIndex)
}

func (c *ListingCompo) textFieldSearch(ctx context.Context) h.HTMLComponent {
	if c.lb.keywordSearchOff {
		return nil
	}
	_, msgr := c.MustGetEventContext(ctx)
	return VTextField().
		Density(DensityCompact).
		Variant(FieldVariantOutlined).
		Label(msgr.Search).
		Flat(true).
		Clearable(true).
		HideDetails(true).
		SingleLine(true).
		ModelValue(c.Keyword).
		Attr("@keyup.enter", stateful.ReloadAction(ctx, c, nil,
			stateful.WithAppendFix(`v.compo.keyword = $event.target.value`),
		).Go()).
		Attr("@click:clear", stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
			target.Keyword = ""
		}).Go()).
		Children(
			web.Slot(VIcon("mdi-magnify")).Name("append-inner"),
		)
}

func (c *ListingCompo) filterSearch(ctx context.Context, fd vx.FilterData) h.HTMLComponent {
	if fd == nil {
		return nil
	}
	if !lo.ContainsBy(fd, func(d *vx.FilterItem) bool {
		return !d.Invisible
	}) {
		return nil
	}

	_, msgr := c.MustGetEventContext(ctx)

	for _, d := range fd {
		d.Translations = vx.FilterIndependentTranslations{
			FilterBy: msgr.FilterBy(d.Label),
		}
	}

	ft := vx.FilterTranslations{}
	ft.Clear = msgr.FiltersClear
	ft.Add = msgr.FiltersAdd
	ft.Apply = msgr.FilterApply
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

	return vx.VXFilter(fd).Translations(ft).UpdateModelValue(
		stateful.ReloadAction(ctx, c, nil,
			stateful.WithAppendFix(`
			if (v.compo.filter_query !== $event.encodedFilterData) {
				v.compo.filter_query = $event.encodedFilterData; 
				v.compo.active_filter_tab = "";
			}`),
		).Go(),
	)
}

func (c *ListingCompo) toolbarSearch(ctx context.Context) h.HTMLComponent {
	evCtx, _ := c.MustGetEventContext(ctx)

	var filterSearch h.HTMLComponent
	if c.lb.filterDataFunc != nil {
		fd := c.lb.filterDataFunc(evCtx)
		fd.SetByQueryString(c.FilterQuery)
		filterSearch = c.filterSearch(ctx, fd)
	}

	tfSearch := VResponsive().Children(
		c.textFieldSearch(ctx),
	)
	if filterSearch != nil || !c.LongStyleSearchBox {
		tfSearch.MaxWidth(200).MinWidth(200).Class("mr-4")
	} else {
		tfSearch.Width(100)
	}
	return VToolbar().Flat(true).Color("surface").AutoHeight(true).Class("pa-2").Children(
		tfSearch,
		filterSearch,
	)
}

func (c *ListingCompo) defaultCellWrapperFunc(cell h.MutableAttrHTMLComponent, id string, obj any, dataTableID string) h.HTMLComponent {
	if c.lb.mb.hasDetailing && !c.lb.mb.detailing.drawer {
		cell.SetAttr("@click", web.Plaid().PushStateURL(c.lb.mb.Info().DetailingHref(id)).Go())
		return cell
	}

	event := actions.Edit
	if c.lb.mb.hasDetailing {
		event = actions.DetailingDrawer
	}
	onClick := web.Plaid().EventFunc(event).Query(ParamID, id)
	// TODO: should use action ?
	// if c.Popup {
	// 	onClick.URL(evCtx.R.RequestURI).
	// 		Query(ParamOverlay, actions.Dialog).
	// 		Query(ParamInDialog, true).
	// 		Query(ParamListingQueries, ctx.Queries().Encode())
	// }
	// TODO: 需要更优雅的方式
	cell.SetAttr("@click", fmt.Sprintf(`
		%s; 
		%s.%s="%s-%s";`,
		onClick.Go(),
		c.locals(), localsKeyCurrentEditingId, dataTableID, id))
	return cell
}

func (c *ListingCompo) dataTable(ctx context.Context) h.HTMLComponent {
	if c.lb.Searcher == nil {
		panic(errors.New("function Searcher is not set"))
	}

	evCtx, msgr := c.MustGetEventContext(ctx)

	searchParams := &SearchParams{
		PageURL:       evCtx.R.URL,
		SQLConditions: c.lb.conditions,
	}

	if !c.lb.keywordSearchOff {
		searchParams.KeywordColumns = c.lb.searchColumns
		searchParams.Keyword = c.Keyword
	}

	orderBys := lo.Map(c.OrderBys, func(ob ColOrderBy, _ int) ColOrderBy {
		ob.OrderBy = strings.ToUpper(ob.OrderBy)
		if ob.OrderBy != OrderByASC && ob.OrderBy != OrderByDESC {
			ob.OrderBy = OrderByDESC
		}
		return ob
	})
	orderableFieldMap := make(map[string]string)
	for _, v := range c.lb.orderableFields {
		orderableFieldMap[v.FieldName] = v.DBColumn
	}
	dbOrderBys := []string{}
	for _, ob := range orderBys {
		dbCol, ok := orderableFieldMap[ob.FieldName]
		if !ok {
			continue
		}
		dbBy := ob.OrderBy
		dbOrderBys = append(dbOrderBys, fmt.Sprintf("%s %s", dbCol, dbBy))
	}
	var orderBySQL string
	if len(dbOrderBys) == 0 {
		if c.lb.orderBy != "" {
			orderBySQL = c.lb.orderBy
		} else {
			orderBySQL = fmt.Sprintf("%s %s", c.lb.mb.primaryField, OrderByDESC)
		}
	} else {
		orderBySQL = strings.Join(dbOrderBys, ", ")
	}
	searchParams.OrderBy = orderBySQL

	if !c.lb.disablePagination {
		perPage := c.PerPage
		if perPage > PerPageMax {
			perPage = PerPageMax
		}
		searchParams.PerPage = cmp.Or(perPage, PerPageDefault)
		searchParams.Page = cmp.Or(c.Page, 1)
	}

	var fd vx.FilterData
	if c.lb.filterDataFunc != nil {
		fd = c.lb.filterDataFunc(evCtx)
		cond, args := fd.SetByQueryString(c.FilterQuery)
		searchParams.SQLConditions = append(searchParams.SQLConditions, &SQLCondition{
			Query: cond,
			Args:  args,
		})
	}

	objs, totalCount, err := c.lb.Searcher(c.lb.mb.NewModelSlice(), searchParams, evCtx)
	if err != nil {
		panic(errors.Wrap(err, "searcher error"))
	}

	btnConfigColumns, columns := c.displayColumns(ctx)

	dataTable := vx.DataTableX(objs).
		HeadCellWrapperFunc(func(cell h.MutableAttrHTMLComponent, field string, title string) h.HTMLComponent {
			if _, exists := orderableFieldMap[field]; !exists {
				return cell
			}

			orderBy, orderByIdx, exists := lo.FindIndexOf(orderBys, func(ob ColOrderBy) bool {
				return ob.FieldName == field
			})
			if !exists {
				orderBy = ColOrderBy{
					FieldName: field,
					OrderBy:   OrderByDESC,
				}
			}

			icon := "mdi-arrow-down"
			if orderBy.OrderBy == OrderByASC {
				icon = "mdi-arrow-up"
			}
			return h.Th("").Style("cursor: pointer; white-space: nowrap;").
				Attr("@click.stop", stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					if orderBy.OrderBy == OrderByASC {
						orderBy.OrderBy = OrderByDESC
					} else {
						orderBy.OrderBy = OrderByASC
					}
					if exists {
						if orderBy.OrderBy == OrderByASC {
							target.OrderBys = append(target.OrderBys[:orderByIdx], target.OrderBys[orderByIdx+1:]...)
						} else {
							target.OrderBys[orderByIdx] = orderBy
						}
					} else {
						target.OrderBys = append(target.OrderBys, orderBy)
					}
				}).Go()).
				Children(
					h.Span(title).Style("text-decoration: underline;"),
					h.Span("").StyleIf("visibility: hidden;", !exists).Children(
						VIcon(icon).Size(SizeSmall),
						h.Span(fmt.Sprint(orderByIdx+1)),
					),
				)
		}).
		RowWrapperFunc(func(row h.MutableAttrHTMLComponent, id string, obj any, dataTableID string) h.HTMLComponent {
			// TODO: how to cancel active if not using vars.presetsRightDrawer
			row.SetAttr(":class", fmt.Sprintf(`{
					"vx-list-item--active primary--text": vars.presetsRightDrawer && %s.%s === "%s-%s",
				}`, c.locals(), localsKeyCurrentEditingId, dataTableID, id,
			))
			return row
		}).
		RowMenuHead(btnConfigColumns).
		RowMenuItemFuncs(c.lb.RowMenu().listingItemFuncs(evCtx)...).
		CellWrapperFunc(
			lo.If(c.lb.cellWrapperFunc != nil, c.lb.cellWrapperFunc).Else(c.defaultCellWrapperFunc),
		)

	if len(c.lb.bulkActions) > 0 {
		syncQuery := ""
		if stateful.IsSyncQuery(ctx) {
			syncQuery = web.Plaid().
				PushState(true).
				MergeQuery(true).
				Query("selected_ids", web.Var(`selected_ids`)).RunPushState()
		}
		dataTable.SelectedIds(c.SelectedIds).
			OnSelectionChanged(fmt.Sprintf(`function(selected_ids) {
					%s.%s = selected_ids;
					%s
				}`, c.locals(), localsKeySelectedIds, syncQuery)).
			SelectedCountLabel(msgr.ListingSelectedCountNotice).
			ClearSelectionLabel(msgr.ListingClearSelection)
	}

	for _, col := range columns {
		if !col.Visible {
			continue
		}
		// fill in empty compFunc and setter func with default
		f := c.lb.getFieldOrDefault(col.Name)
		dataTable.Column(col.Name).Title(col.Label).CellComponentFunc(c.lb.cellComponentFunc(f))
	}

	if c.lb.disablePagination {
		return dataTable
	}

	var dataTableAdditions h.HTMLComponent
	if totalCount <= 0 {
		dataTableAdditions = h.Div().Class("mt-10 text-center grey--text text--darken-2").Children(
			h.Text(msgr.ListingNoRecordToShow),
		)
	} else {
		dataTableAdditions = h.Div().Class("mt-2").Children(
			vx.VXTablePagination().
				Total(int64(totalCount)).
				CurrPage(searchParams.Page).
				PerPage(searchParams.PerPage).
				CustomPerPages([]int64{c.lb.perPage}).
				PerPageText(msgr.PaginationRowsPerPage).
				OnSelectPerPage(stateful.ReloadAction(ctx, c, nil,
					stateful.WithAppendFix(`v.compo.per_page = parseInt($event, 10)`),
				).Go()).
				OnPrevPage(stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					target.Page = searchParams.Page - 1
				}).Go()).
				OnNextPage(stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					target.Page = searchParams.Page + 1
				}).Go()),
		)
	}
	return h.Components(dataTable, dataTableAdditions)
}

type DisplayColumnWrapper struct {
	DisplayColumn
	Label string `json:"label"`
}

func (c *ListingCompo) displayColumns(ctx context.Context) (btnConfigure h.HTMLComponent, wrappers []DisplayColumnWrapper) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	displayColumn := slices.Clone(c.DisplayColumns)
	var availableColumns []DisplayColumn
	for _, f := range c.lb.fields {
		if c.lb.mb.Info().Verifier().Do(PermList).SnakeOn("f_"+f.name).WithReq(evCtx.R).IsAllowed() != nil {
			continue
		}
		availableColumns = append(availableColumns, DisplayColumn{
			Name:    f.name,
			Visible: true,
		})
	}

	// if there is abnormal data, restore the default
	if len(displayColumn) != len(availableColumns) ||
		// names not match
		!lo.EveryBy(displayColumn, func(dc DisplayColumn) bool {
			return lo.ContainsBy(availableColumns, func(ac DisplayColumn) bool {
				return ac.Name == dc.Name
			})
		}) {
		displayColumn = availableColumns
	}

	allInvisible := lo.EveryBy(displayColumn, func(dc DisplayColumn) bool {
		return !dc.Visible
	})
	for _, col := range displayColumn {
		if allInvisible {
			col.Visible = true
		}
		wrappers = append(wrappers, DisplayColumnWrapper{
			DisplayColumn: col,
			Label:         i18n.PT(evCtx.R, ModelsI18nModuleKey, c.lb.mb.label, c.lb.mb.getLabel(c.lb.Field(col.Name).NameLabel)),
		})
	}

	if !c.lb.selectableColumns {
		return nil, wrappers
	}

	return web.Scope().
			VSlot("{ locals }").
			Init(fmt.Sprintf(`{selectColumnsMenu: false, displayColumns: %s}`, h.JSONString(wrappers))).
			Children(
				VMenu().CloseOnContentClick(false).Width(240).Attr("v-model", "locals.selectColumnsMenu").Children(
					web.Slot().Name("activator").Scope("{ props }").Children(
						VBtn("").Icon("mdi-cog").Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
					),
					VList().Density(DensityCompact).Children(
						h.Tag("vx-draggable").Attr("item-key", "name").Attr("v-model", "locals.displayColumns", "handle", ".handle", "animation", "300").Children(
							h.Template().Attr("#item", " { element } ").Children(
								VListItem(
									VListItemTitle(
										VSwitch().Density(DensityCompact).Color("primary").Class(" mt-2 ").Attr(
											"v-model", "element.visible",
											":label", "element.label",
										),
										VIcon("mdi-reorder-vertical").Class("handle cursor-grab mt-4"),
									).Class("d-flex justify-space-between "),
									VDivider(),
								),
							),
						),
						VListItem().Class("d-flex justify-space-between").Children(
							VBtn(msgr.Cancel).Elevation(0).Attr("@click", `locals.selectColumnsMenu = false`),
							VBtn(msgr.OK).Elevation(0).Color("primary").Attr("@click", fmt.Sprintf(`
								locals.selectColumnsMenu = false; 
								%s`,
								stateful.ReloadAction(ctx, c, nil,
									stateful.WithAppendFix(`v.compo.display_columns = locals.displayColumns.map(({ label, ...rest }) => rest)`),
								).Go(),
							)),
						),
					),
				),
			),
		wrappers
}

func (c *ListingCompo) cardActionsFooter(ctx context.Context) h.HTMLComponent {
	if len(c.lb.footerActions) <= 0 {
		return nil
	}
	evCtx, _ := c.MustGetEventContext(ctx)
	compos := []h.HTMLComponent{VSpacer()}
	for _, action := range c.lb.footerActions {
		compos = append(compos, action.buttonCompFunc(evCtx))
	}
	return VCardActions(compos...)
}

func (c *ListingCompo) actionsComponent(ctx context.Context) (r h.HTMLComponent) {
	defer func() {
		if r != nil {
			rr := r
			r = h.ComponentFunc(func(ctx context.Context) (r []byte, err error) {
				ctx = stateful.WithPortalName(ctx, c.CompoName()+"_actions")
				return c.wrapCompo(ctx, rr).MarshalHTML(ctx)
			})
		}
	}()
	evCtx, msgr := c.MustGetEventContext(ctx)

	var buttons []h.HTMLComponent

	for _, ba := range c.lb.bulkActions {
		if c.lb.mb.Info().Verifier().SnakeDo(permBulkActions, ba.name).WithReq(evCtx.R).IsAllowed() != nil {
			continue
		}

		if ba.buttonCompFunc != nil {
			buttons = append(buttons, ba.buttonCompFunc(evCtx))
			continue
		}

		buttons = append(buttons, VBtn(c.lb.mb.getLabel(ba.NameLabel)).
			Color(cmp.Or(ba.buttonColor, ColorSecondary)).Variant(VariantFlat).Class("ml-2").
			Attr("@click", stateful.PostAction(ctx, c, c.OpenBulkActionDialog, OpenBulkActionDialogRequest{
				Name: ba.name,
			}).Go()),
		)
	}

	for _, ba := range c.lb.actions {
		if c.lb.mb.Info().Verifier().SnakeDo(permActions, ba.name).WithReq(evCtx.R).IsAllowed() != nil {
			continue
		}

		if ba.buttonCompFunc != nil {
			buttons = append(buttons, ba.buttonCompFunc(evCtx))
			continue
		}

		buttons = append(buttons, VBtn(c.lb.mb.getLabel(ba.NameLabel)).
			Color(cmp.Or(ba.buttonColor, ColorPrimary)).Variant(VariantFlat).Class("ml-2").
			Attr("@click", stateful.PostAction(ctx, c, c.OpenActionDialog, OpenActionDialogRequest{
				Name: ba.name,
			}).Go()),
		)
	}

	buttonNew := func() h.HTMLComponent {
		if c.lb.mb.Info().Verifier().Do(PermCreate).WithReq(evCtx.R).IsAllowed() != nil {
			return nil
		}
		if c.lb.newBtnFunc != nil {
			return c.lb.newBtnFunc(evCtx)
		}
		onClick := web.Plaid().EventFunc(actions.New)
		// TODO:
		// if inDialog {
		// 	onclick.URL(ctx.R.RequestURI).
		// 		Query(ParamOverlay, actions.Dialog).
		// 		Query(ParamInDialog, true).
		// 		Query(ParamListingQueries, ctx.Queries().Encode())
		// }
		return VBtn(msgr.New).
			Color(ColorPrimary).
			Variant(VariantFlat).
			Theme("dark").Class("ml-2").
			Attr("@click", onClick.Go())
	}()

	if c.lb.actionsAsMenu {
		buttons = append([]h.HTMLComponent{buttonNew}, buttons...)
		return VMenu().OpenOnHover(true).Children(
			web.Slot().Name("activator").Scope("{ on, props }").Children(
				VBtn(msgr.ButtonLabelActionsMenu).Size(SizeSmall).Attr("v-bind", "props").Attr("v-on", "on"),
			),
			VList(lo.Map(buttons, func(item h.HTMLComponent, _ int) h.HTMLComponent {
				return VListItem(item)
			})...),
		)
	}

	buttons = append(buttons, buttonNew)
	return h.Div(buttons...)
}

func (c *ListingCompo) bulkPanel(ctx context.Context, bulk *BulkActionBuilder, selectedIds []string, actionableIds []string) (r h.HTMLComponent) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	var errCompo h.HTMLComponent
	if vErr, ok := evCtx.Flash.(*web.ValidationErrors); ok {
		if gErr := vErr.GetGlobalError(); gErr != "" {
			errCompo = VAlert(h.Text(gErr)).Border("left").Type("error").Elevation(2)
		}
	}

	var alertCompo h.HTMLComponent
	if len(actionableIds) < len(selectedIds) {
		unactionables := lo.Without(selectedIds, actionableIds...)
		if len(unactionables) > 0 {
			var notice string
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				notice = bulk.selectedIdsProcessorNoticeFunc(selectedIds, actionableIds, unactionables)
			} else {
				var ids string
				if len(unactionables) <= 10 {
					ids = strings.Join(unactionables, ", ")
				} else {
					ids = fmt.Sprintf("%s...(+%d)", strings.Join(unactionables[:10], ", "), len(unactionables)-10)
				}
				notice = msgr.BulkActionSelectedIdsProcessNotice(ids)
			}
			alertCompo = VAlert(h.Text(notice)).Type("warning")
		}
	}

	return VCard(
		VCardTitle(
			h.Text(bulk.NameLabel.label),
		),
		VCardText(
			errCompo,
			alertCompo,
			bulk.compFunc(selectedIds, evCtx),
		),
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).Variant(VariantFlat).Class("ml-2").Attr("@click", closeDialogVarScript),
			VBtn(msgr.OK).Color("primary").Variant(VariantFlat).Theme(ThemeDark).Attr("@click",
				stateful.PostAction(ctx, c, c.DoBulkAction, DoBulkActionRequest{
					Name: bulk.name,
				}).Go(),
			),
		),
	)
}

func (c *ListingCompo) fetchBulkAction(ctx context.Context, name string) (*BulkActionBuilder, error) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	if c.lb.mb.Info().Verifier().SnakeDo(permBulkActions, name).WithReq(evCtx.R).IsAllowed() != nil {
		return nil, perm.PermissionDenied
	}

	bulk, exists := lo.Find(c.lb.bulkActions, func(ba *BulkActionBuilder) bool {
		return ba.name == name
	})
	if !exists {
		return nil, errors.New("cannot find requested bulk action")
	}

	if len(c.SelectedIds) == 0 {
		return nil, errors.New(msgr.BulkActionNoRecordsSelected)
	}

	return bulk, nil
}

type OpenBulkActionDialogRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) OpenBulkActionDialog(ctx context.Context, req OpenBulkActionDialogRequest) (r web.EventResponse, err error) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	bulk, err := c.fetchBulkAction(ctx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	actionableIds := c.SelectedIds
	if bulk.selectedIdsProcessorFunc != nil {
		actionableIds, err = bulk.selectedIdsProcessorFunc(c.SelectedIds, evCtx)
		if err != nil {
			return r, err
		}
		// if no actionable ids, skip dialog
		if len(actionableIds) == 0 {
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				ShowMessage(&r, bulk.selectedIdsProcessorNoticeFunc(c.SelectedIds, actionableIds, c.SelectedIds), "warning")
			} else {
				ShowMessage(&r, msgr.BulkActionNoAvailableRecords, "warning")
			}
			return r, nil
		}
	}

	c.lb.mb.p.dialog(&r, c.bulkPanel(ctx, bulk, c.SelectedIds, actionableIds), bulk.dialogWidth)
	return r, nil
}

type DoBulkActionRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) DoBulkAction(ctx context.Context, req DoBulkActionRequest) (r web.EventResponse, err error) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	bulk, err := c.fetchBulkAction(ctx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	actionableIds := c.SelectedIds
	if bulk.selectedIdsProcessorFunc != nil {
		actionableIds, err = bulk.selectedIdsProcessorFunc(c.SelectedIds, evCtx)
	}

	if err == nil {
		err = bulk.updateFunc(actionableIds, evCtx, &r)
	}

	if err != nil {
		evCtx.Flash = toValidationErrors(err)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: c.bulkPanel(ctx, bulk, c.SelectedIds, actionableIds),
		})
		return r, nil
	}

	ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	web.AppendRunScripts(&r, closeDialogVarScript)
	return r, nil
}

func (c *ListingCompo) actionPanel(ctx context.Context, action *ActionBuilder) (r h.HTMLComponent) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	var errCompo h.HTMLComponent
	if vErr, ok := evCtx.Flash.(*web.ValidationErrors); ok {
		if gErr := vErr.GetGlobalError(); gErr != "" {
			errCompo = VAlert(h.Text(gErr)).Border("left").Type("error").Elevation(2)
		}
	}

	return VCard(
		VCardTitle(
			h.Text(action.NameLabel.label),
		),
		VCardText(
			errCompo,
			h.Iff(action.compFunc != nil, func() h.HTMLComponent {
				return action.compFunc("", evCtx)
			}),
		),
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).Variant(VariantFlat).Class("ml-2").Attr("@click", closeDialogVarScript),
			VBtn(msgr.OK).Color("primary").Variant(VariantFlat).Theme(ThemeDark).Attr("@click",
				stateful.PostAction(ctx, c, c.DoAction, DoActionRequest{
					Name: action.name,
				}).Go(),
			),
		),
	)
}

func (c *ListingCompo) fetchAction(evCtx *web.EventContext, name string) (*ActionBuilder, error) {
	action, exists := lo.Find(c.lb.actions, func(a *ActionBuilder) bool {
		return a.name == name
	})
	if !exists {
		return nil, errors.New("cannot find requested action")
	}

	if c.lb.mb.Info().Verifier().SnakeDo(permActions, action.name).WithReq(evCtx.R).IsAllowed() != nil {
		return nil, perm.PermissionDenied
	}

	return action, nil
}

type OpenActionDialogRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) OpenActionDialog(ctx context.Context, req OpenBulkActionDialogRequest) (r web.EventResponse, err error) {
	evCtx, _ := c.MustGetEventContext(ctx)

	action, err := c.fetchAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	c.lb.mb.p.dialog(&r, c.actionPanel(ctx, action), action.dialogWidth)
	return r, nil
}

type DoActionRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) DoAction(ctx context.Context, req DoActionRequest) (r web.EventResponse, err error) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	action, err := c.fetchAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	if err := action.updateFunc("", evCtx, &r); err != nil {
		evCtx.Flash = toValidationErrors(err)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: c.actionPanel(ctx, action),
		})
		return r, nil
	}

	ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	web.AppendRunScripts(&r, closeDialogVarScript)
	return r, nil
}

func (c *ListingCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, MustGetMessages(evCtx.R)
}
