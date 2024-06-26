package presets

import (
	"cmp"
	"context"
	"fmt"
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

func init() {
	stateful.RegisterActionableType((*ListingCompo)(nil))
}

type DisplayColumn struct {
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
}

type ListingCompo struct {
	lb *ListingBuilderX `inject:""`

	CompoID            string          `json:"compo_id"`
	LongStyleSearchBox bool            `json:"long_style_search_box"`
	SelectedIds        []string        `json:"selected_ids" query:"selected_ids"`
	Keyword            string          `json:"keyword" query:"keyword"`
	OrderBys           []ColOrderBy    `json:"order_bys" query:"order_bys"`
	Page               int64           `json:"page" query:"page"`
	PerPage            int64           `json:"per_page" query:"per_page"`
	DisplayColumns     []DisplayColumn `json:"display_columns" query:"display_columns"`
	ActiveFilterTab    string          `json:"active_filter_tab" query:"active_filter_tab"`
}

func (c *ListingCompo) CompoName() string {
	return fmt.Sprintf("ListingCompo:%s", c.CompoID)
}

const (
	listingLocals            = "listingLocals"
	listingLocalsSelectedIds = "listingLocals.selected_ids"
)

const notifListingLocalsUpdated = "NotifListingLocalsUpdated"

func (c *ListingCompo) MarshalHTML(ctx context.Context) (r []byte, err error) {
	observerScript := func(prevReload string) string {
		return fmt.Sprintf(`
				%s = locals
				%s
				%s`,
			listingLocals,
			prevReload,
			c.ReloadActionGo(ctx, nil),
		)
	}
	return stateful.Reloadable(c,
		web.Scope().VSlot(fmt.Sprintf("{ locals: %s }", listingLocals)).Init(fmt.Sprintf(`{
				currEditingListItemID: "",
				selected_ids: %s || [],
			}`, h.JSONString(c.SelectedIds))).
			Observe(c.lb.mb.NotifModelsUpdated(), observerScript("")).
			Observe(c.lb.mb.NotifModelsDeleted(), observerScript(fmt.Sprintf(`
				if (payload && payload.ids && payload.ids.length > 0) {
					%s = %s.filter(id => !payload.ids.includes(id));
				}`,
				listingLocalsSelectedIds, listingLocalsSelectedIds,
			))).
			// sync to actionsComponent
			// TODO: 可能还需要一个 OnMount 方法来触发这个事件，否则每次 reload 都要主动触发
			UseDebounce(1).OnChange(fmt.Sprintf(`vars.__sendNotification(%q, {
				compo_name: %q,
				locals: locals,
			})`, notifListingLocalsUpdated, c.CompoName())).
			Children(
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

	activeIndex := -1
	fts := c.lb.filterTabsFunc(web.MustGetEventContext(ctx))
	tabs := VTabs().Class("mb-2").ShowArrows(true).Color(ColorPrimary).Density(DensityCompact)
	for i, ft := range fts {
		if ft.ID == "" {
			ft.ID = fmt.Sprintf("tab%d", i)
		}
		if c.ActiveFilterTab == ft.ID {
			activeIndex = i
		}
		// TODO: ft.Query 需要和 filter 匹配才可
		tabs.AppendChildren(
			VTab().Attr("@click", c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
				cloned.ActiveFilterTab = ft.ID
			})).Children(
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
	msgr := c.MustGetMessages(ctx)
	onChanged := func(quoteKeyword string) string {
		return strings.Replace(
			c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
				cloned.Keyword = ""
			}),
			`"keyword": ""`,
			`"keyword": `+quoteKeyword,
			1,
		)
	}
	// TODO: isFocus 原来的目的是什么来着？
	return web.Scope().VSlot("{ locals }").Init(`{isFocus: false}`).Children(
		VTextField().
			Density(DensityCompact).
			Variant(FieldVariantOutlined).
			Label(msgr.Search).
			Flat(true).
			Clearable(true).
			HideDetails(true).
			SingleLine(true).
			ModelValue(c.Keyword).
			Attr("@keyup.enter", onChanged("$event.target.value")).
			Attr("@click:clear", onChanged(`""`)).
			Children(
				web.Slot(VIcon("mdi-magnify")).Name("append-inner"),
			),
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

	msgr := c.MustGetMessages(ctx)

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

	return vx.VXFilter(fd).Translations(ft)
	// TODO:
	// if inDialog {
	// 	filter.UpdateModelValue(Zone[*ListingZone](ctx).Plaid().
	// 		URL(ctx.R.RequestURI).
	// 		StringQuery(web.Var("$event.encodedFilterData")).
	// 		Query("page", 1).
	// 		ClearMergeQuery(web.Var("$event.filterKeys")).
	// 		EventFunc(actions.ReloadList).
	// 		Go())
	// }
	// return filter
}

func (c *ListingCompo) toolbarSearch(ctx context.Context) h.HTMLComponent {
	evCtx := web.MustGetEventContext(ctx)
	// msgr := c.MustGetMessages(ctx)

	var filterSearch h.HTMLComponent
	if c.lb.filterDataFunc != nil {
		fd := c.lb.filterDataFunc(evCtx)
		fd.SetByQueryString(evCtx.R.URL.RawQuery) // TODO:
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
		cell.SetAttr("@click.self", web.Plaid().PushStateURL(c.lb.mb.Info().DetailingHref(id)).Go())
		return cell
	}

	event := actions.Edit
	if c.lb.mb.hasDetailing {
		event = actions.DetailingDrawer
	}
	onClick := web.Plaid().EventFunc(event).Query(ParamID, id)
	// TODO: how to auto open in dialog ?
	// if inDialog {
	// 	onclick.URL(ctx.R.RequestURI).
	// 		Query(ParamOverlay, actions.Dialog).
	// 		Query(ParamInDialog, true).
	// 		Query(ParamListingQueries, ctx.Queries().Encode())
	// }
	// TODO: 需要更优雅的方式
	cell.SetAttr("@click.self", fmt.Sprintf(`%s; %s.currEditingListItemID="%s-%s"`, onClick.Go(), listingLocals, dataTableID, id))
	return cell
}

func (c *ListingCompo) dataTable(ctx context.Context) h.HTMLComponent {
	if c.lb.Searcher == nil {
		panic(errors.New("function Searcher is not set"))
	}

	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

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
		perPage := c.PerPage // TODO: sync cookie ?
		if perPage > 1000 {
			perPage = 1000
		}
		searchParams.PerPage = cmp.Or(perPage, 10) // TODO:
		searchParams.Page = cmp.Or(c.Page, 1)
	}

	var fd vx.FilterData
	if c.lb.filterDataFunc != nil {
		// TODO: how to stateful?
		fd = c.lb.filterDataFunc(evCtx)
		cond, args := fd.SetByQueryString(evCtx.R.URL.RawQuery)
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
				Attr("@click", c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
					if orderBy.OrderBy == OrderByASC {
						orderBy.OrderBy = OrderByDESC
					} else {
						orderBy.OrderBy = OrderByASC
					}
					if exists {
						if orderBy.OrderBy == OrderByASC {
							cloned.OrderBys = append(cloned.OrderBys[:orderByIdx], cloned.OrderBys[orderByIdx+1:]...)
						} else {
							cloned.OrderBys[orderByIdx] = orderBy
						}
					} else {
						cloned.OrderBys = append(cloned.OrderBys, orderBy)
					}
				})).
				Children(
					h.Span(title).Style("text-decoration: underline;"),
					h.Span("").StyleIf("visibility: hidden;", !exists).Children(
						VIcon(icon).Size(SizeSmall),
						h.Span(fmt.Sprint(orderByIdx+1)),
					),
				)
		}).
		RowWrapperFunc(func(row h.MutableAttrHTMLComponent, id string, obj any, dataTableID string) h.HTMLComponent {
			// TODO: how to cancel active ? 不可能都根据 vars.presetsRightDrawer 去 cancel
			row.SetAttr(":class", fmt.Sprintf(`{
					"vx-list-item--active primary--text": vars.presetsRightDrawer && %s.currEditingListItemID==="%s-%s",
				}`, listingLocals, dataTableID, id,
			))
			return row
		}).
		RowMenuHead(btnConfigColumns).
		RowMenuItemFuncs(c.lb.RowMenu().listingItemFuncs(evCtx)...).
		CellWrapperFunc(
			lo.If(c.lb.cellWrapperFunc != nil, c.lb.cellWrapperFunc).Else(c.defaultCellWrapperFunc),
		).
		VarSelectedIds(
			lo.If(len(c.lb.bulkActions) > 0, listingLocalsSelectedIds).Else(""),
		).
		SelectedCountLabel(msgr.ListingSelectedCountNotice).
		ClearSelectionLabel(msgr.ListingClearSelection)

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
				OnSelectPerPage(strings.Replace(
					c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
						cloned.PerPage = -1
					}),
					`"per_page": -1`,
					`"per_page": parseInt($event, 10)`,
					1,
				)).
				OnPrevPage(c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
					cloned.Page = searchParams.Page - 1
				})).
				OnNextPage(c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
					cloned.Page = searchParams.Page + 1
				})),
		)
	}
	return h.Components(dataTable, dataTableAdditions)
}

type DisplayColumnWrapper struct {
	DisplayColumn
	Label string `json:"label"`
}

func (c *ListingCompo) displayColumns(ctx context.Context) (btnConfigure h.HTMLComponent, wrappers []DisplayColumnWrapper) {
	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

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
	if len(c.DisplayColumns) != len(availableColumns) ||
		// names not match
		!lo.EveryBy(c.DisplayColumns, func(dc DisplayColumn) bool {
			return lo.ContainsBy(availableColumns, func(ac DisplayColumn) bool {
				return ac.Name == dc.Name
			})
		}) {
		// TODO: 对于状态的修正是否应该在 MarshalHTML 的头部提前统一处理？
		c.DisplayColumns = availableColumns
	}

	allInvisible := lo.EveryBy(c.DisplayColumns, func(dc DisplayColumn) bool {
		return !dc.Visible
	})
	for _, col := range c.DisplayColumns {
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
								strings.Replace(
									c.ReloadActionGo(ctx, func(cloned *ListingCompo) {
										cloned.DisplayColumns = []DisplayColumn{}
									}),
									`"display_columns": []`,
									`"display_columns": locals.displayColumns.map(({ label, ...rest }) => rest)`,
									1,
								),
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
	evCtx := web.MustGetEventContext(ctx)
	compos := []h.HTMLComponent{VSpacer()}
	for _, action := range c.lb.footerActions {
		compos = append(compos, action.buttonCompFunc(evCtx))
	}
	return VCardActions(compos...)
}

func (c *ListingCompo) actionsComponent(ctx context.Context) (r h.HTMLComponent) {
	defer func() {
		if r != nil {
			r = web.Scope(r).VSlot(fmt.Sprintf("{ locals: %s }", listingLocals)).
				Init(
					fmt.Sprintf(`{
						currEditingListItemID: "",
						selected_ids: %s || [],
					}`,
						h.JSONString(c.SelectedIds),
					),
				).
				Observe(notifListingLocalsUpdated, fmt.Sprintf(`
					if (!payload.compo_name || payload.compo_name != %q) {
						return
					}
					Object.keys(locals).forEach(key => {
						if (payload.locals.hasOwnProperty(key)) {
							locals[key] = payload.locals[key];
						}
					});
				`, c.CompoName()))
		}
	}()
	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

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
			Attr("@click", c.PostActionGo(ctx, c.OpenBulkActionDialog, OpenBulkActionDialogRequest{
				Name: ba.name,
			})),
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
			Attr("@click", c.PostActionGo(ctx, c.OpenActionDialog, OpenActionDialogRequest{
				Name: ba.name,
			})),
		)
	}

	if c.lb.actionsAsMenu {
		// TODO: no new ? 为什么以前写在这呢？
		return VMenu().OpenOnHover(true).Children(
			web.Slot().Name("activator").Scope("{ on, props }").Children(
				// TODO: i18n ?
				VBtn("Actions").Size(SizeSmall).Attr("v-bind", "props").Attr("v-on", "on"),
			),
			VList(lo.Map(buttons, func(item h.HTMLComponent, _ int) h.HTMLComponent {
				return VListItem(item)
			})...),
		)
	}

	if c.lb.newBtnFunc != nil { // TODO: perm ? 即使自己提供了也应该做前置权限判断吧？
		if button := c.lb.newBtnFunc(evCtx); button != nil {
			buttons = append(buttons, button)
		}
	} else if c.lb.mb.Info().Verifier().Do(PermCreate).WithReq(evCtx.R).IsAllowed() == nil {
		// TODO:
		// onclick := web.Plaid().EventFunc(actions.New)
		// if inDialog {
		// 	onclick.URL(ctx.R.RequestURI).
		// 		Query(ParamOverlay, actions.Dialog).
		// 		Query(ParamInDialog, true).
		// 		Query(ParamListingQueries, ctx.Queries().Encode())
		// }
		buttons = append(buttons, VBtn(msgr.New).
			Color(ColorPrimary).
			Variant(VariantFlat).
			Theme("dark").Class("ml-2"), // TODO: why theme dark ?
		// .Attr("@click", onclick.Go())
		)
	}

	return h.Div(buttons...)
}

func (c *ListingCompo) bulkPanel(ctx context.Context, bulk *BulkActionBuilder, selectedIds []string, processedSelectedIds []string) (r h.HTMLComponent) {
	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

	var errCompo h.HTMLComponent
	if vErr, ok := evCtx.Flash.(*web.ValidationErrors); ok {
		if gErr := vErr.GetGlobalError(); gErr != "" {
			errCompo = VAlert(h.Text(gErr)).Border("left").Type("error").Elevation(2)
		}
	}

	var alertCompo h.HTMLComponent
	if len(processedSelectedIds) < len(selectedIds) {
		unactionables := lo.Without(selectedIds, processedSelectedIds...)
		if len(unactionables) > 0 {
			var notice string
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				notice = bulk.selectedIdsProcessorNoticeFunc(selectedIds, processedSelectedIds, unactionables)
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

func (c *ListingCompo) fetchBulkAction(evCtx *web.EventContext, name string) (*BulkActionBuilder, error) {
	bulk, exists := lo.Find(c.lb.bulkActions, func(ba *BulkActionBuilder) bool {
		return ba.name == name
	})
	if !exists {
		return nil, errors.New("cannot find requested bulk action")
	}

	if len(c.SelectedIds) == 0 {
		return nil, errors.New("Please select record") // TODO: i18n ?
	}

	if c.lb.mb.Info().Verifier().SnakeDo(permBulkActions, bulk.name).WithReq(evCtx.R).IsAllowed() != nil {
		return nil, perm.PermissionDenied
	}

	return bulk, nil
}

type OpenBulkActionDialogRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) OpenBulkActionDialog(ctx context.Context, req OpenBulkActionDialogRequest) (r web.EventResponse, err error) {
	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

	bulk, err := c.fetchBulkAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	// If selectedIdsProcessorFunc is not nil, process the request in it and skip the confirmation dialog
	processedSelectedIds := c.SelectedIds
	if bulk.selectedIdsProcessorFunc != nil {
		processedSelectedIds, err = bulk.selectedIdsProcessorFunc(c.SelectedIds, evCtx)
		if err != nil {
			return r, err
		}
		if len(processedSelectedIds) == 0 {
			if bulk.selectedIdsProcessorNoticeFunc != nil {
				ShowMessage(&r, bulk.selectedIdsProcessorNoticeFunc(c.SelectedIds, processedSelectedIds, c.SelectedIds), "warning")
			} else {
				ShowMessage(&r, msgr.BulkActionNoAvailableRecords, "warning")
			}
			return r, nil
		}
	}

	c.lb.mb.p.dialog(&r, c.bulkPanel(ctx, bulk, c.SelectedIds, processedSelectedIds), bulk.dialogWidth)
	return r, nil
}

type DoBulkActionRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) DoBulkAction(ctx context.Context, req DoBulkActionRequest) (r web.EventResponse, err error) {
	evCtx := web.MustGetEventContext(ctx)

	bulk, err := c.fetchBulkAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	processedSelectedIds := c.SelectedIds
	if bulk.selectedIdsProcessorFunc != nil { // TODO: 这个地方的逻辑和 OpenBulkActionDialog 里的逻辑不重复吗？
		processedSelectedIds, err = bulk.selectedIdsProcessorFunc(c.SelectedIds, evCtx)
	}

	if err == nil {
		// TODO: 这个方法貌似应该反馈哪些 ids 已经不存在了，便于下面的 reload 之前剔除
		err = bulk.updateFunc(processedSelectedIds, evCtx)
	}

	if err != nil {
		evCtx.Flash = toValidationErrors(err)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: c.bulkPanel(ctx, bulk, c.SelectedIds, processedSelectedIds),
		})
		return r, nil
	}

	msgr := c.MustGetMessages(ctx)
	ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	web.AppendRunScripts(&r, closeDialogVarScript)
	stateful.AppendReloadToResponse(&r, c)
	return r, nil
}

func (c *ListingCompo) actionPanel(ctx context.Context, action *ActionBuilder) (r h.HTMLComponent) {
	evCtx := web.MustGetEventContext(ctx)
	msgr := c.MustGetMessages(ctx)

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

	// TODO: 到底应该是 permActions 还是 permListingActions
	if c.lb.mb.Info().Verifier().SnakeDo(permActions, action.name).WithReq(evCtx.R).IsAllowed() != nil {
		return nil, perm.PermissionDenied
	}

	return action, nil
}

type OpenActionDialogRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) OpenActionDialog(ctx context.Context, req OpenBulkActionDialogRequest) (r web.EventResponse, err error) {
	evCtx := web.MustGetEventContext(ctx)

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
	evCtx := web.MustGetEventContext(ctx)

	action, err := c.fetchAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	if err := action.updateFunc("", evCtx); err != nil {
		evCtx.Flash = toValidationErrors(err)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: c.actionPanel(ctx, action),
		})
		return r, nil
	}

	msgr := c.MustGetMessages(ctx)
	ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	web.AppendRunScripts(&r, closeDialogVarScript)
	stateful.AppendReloadToResponse(&r, c)
	return r, nil
}

func (c *ListingCompo) ReloadActionGo(ctx context.Context, f func(cloned *ListingCompo)) string {
	return strings.Replace( // TODO: 这种 replace 的方式会影响到 sync_query 机制，那块需要改进
		stateful.ReloadAction(ctx, c, func(cloned *ListingCompo) {
			cloned.SelectedIds = []string{}
			if f != nil {
				f(cloned)
			}
		}).Go(),
		`"selected_ids": []`,
		`"selected_ids": `+listingLocalsSelectedIds,
		1,
	)
}

func (c *ListingCompo) PostActionGo(ctx context.Context, method any, request any) string {
	cloned := stateful.MustClone(c)
	cloned.SelectedIds = []string{}
	return strings.Replace(
		stateful.PostAction(ctx, cloned, method, request).Go(),
		`"selected_ids": []`,
		`"selected_ids": `+listingLocalsSelectedIds,
		1,
	)
}

func (c *ListingCompo) MustGetMessages(ctx context.Context) *Messages {
	return MustGetMessages(web.MustGetEventContext(ctx).R)
}
