package presets

import (
	"cmp"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/stateful"
	"github.com/qor5/x/v3/i18n"
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

const (
	OrderByASC  = "ASC"
	OrderByDESC = "DESC"
)

type ColOrderBy struct {
	FieldName string
	OrderBy   string
}

type ListingCompo struct {
	lb *ListingBuilder `inject:""`

	activeFilterTabQuery string

	ID                 string           `json:"id"`
	Popup              bool             `json:"popup"`
	LongStyleSearchBox bool             `json:"long_style_search_box"`
	SelectedIds        []string         `json:"selected_ids" query:",omitempty"`
	Keyword            string           `json:"keyword" query:",omitempty"`
	OrderBys           []ColOrderBy     `json:"order_bys" query:",omitempty"`
	Page               int64            `json:"page" query:",omitempty"`
	PerPage            int64            `json:"per_page" query:",omitempty;cookie"`
	DisplayColumns     []*DisplayColumn `json:"display_columns" query:",omitempty;cookie"`
	ActiveFilterTab    string           `json:"active_filter_tab" query:",omitempty"`
	FilterQuery        string           `json:"filter_query" query:";method:bare,f_"`

	OnMounted string `json:"on_mounted"`
	ParentID  string `json:"parent_id,omitempty"`
}

func (c *ListingCompo) CompoID() string {
	return fmt.Sprintf("ListingCompo_%s", c.ID)
}

type ctxKeyListingCompo struct{}

func ListingCompoFromContext(ctx context.Context) *ListingCompo {
	v, _ := ctx.Value(ctxKeyListingCompo{}).(*ListingCompo)
	return v
}

func ListingCompoFromEventContext(evCtx *web.EventContext) *ListingCompo {
	return ListingCompoFromContext(evCtx.R.Context())
}

const ListingCompo_JsPreFixWhenNotifModelsDeleted = `
if (payload && payload.ids && payload.ids.length > 0) {
	locals.selected_ids = locals.selected_ids.filter(id => !payload.ids.includes(id));
}
`

const ListingCompo_JsScrollToTop = "locals.document.querySelector(`#vt-app > div.v-layout > main`).scrollTop = 0"

func (c *ListingCompo) VarCurrentActive() string {
	return fmt.Sprintf("__current_active_of_%s__", stateful.MurmurHash3(c.CompoID()))
}

const ListingCompo_CurrentActiveClass = "vx-list-item--active primary--text"

const (
	ParamVarCurrentActive = "var_current_active"
)

func ListingCompo_GetVarCurrentActive(evCtx *web.EventContext) string {
	var varCurrentActive string
	c := ListingCompoFromEventContext(evCtx)
	if c != nil {
		varCurrentActive = c.VarCurrentActive()
	}
	if varCurrentActive == "" {
		varCurrentActive = evCtx.R.FormValue(ParamVarCurrentActive)
	}
	if varCurrentActive == "" {
		panic(errors.New("var_current_active is not set"))
	}
	return varCurrentActive
}

func (c *ListingCompo) MarshalHTML(ctx context.Context) (r []byte, err error) {
	ctx = context.WithValue(ctx, ctxKeyListingCompo{}, c)

	evCtx, _ := c.MustGetEventContext(ctx)
	evCtx.WithContextValue(ctxKeyListingCompo{}, c)

	return stateful.Actionable(ctx, c,
		// onMounted for selected_ids front-end autonomy
		web.RunScript(fmt.Sprintf(`({el}) => {
			locals.dialog = false;
			locals.document = el.ownerDocument;
			locals.selected_ids = %s || [];
			let orig = locals.%s;
			locals.%s = function() {
				let v = orig();
				v.compo.selected_ids = this.selected_ids;
				return v
			}
		}`,
			h.JSONString(c.SelectedIds),
			stateful.LocalsKeyNewAction,
			stateful.LocalsKeyNewAction,
		)),
		h.Iff(c.OnMounted != "", func() h.HTMLComponent {
			return h.Div().Attr("v-on-mounted", c.OnMounted)
		}),
		h.Iff(!c.lb.disableModelListeners, func() h.HTMLComponent {
			return web.Listen(
				c.lb.mb.NotifModelsCreated(), stateful.ReloadAction(ctx, c, nil).Go(),
				c.lb.mb.NotifModelsUpdated(), stateful.ReloadAction(ctx, c, nil).Go(),
				c.lb.mb.NotifModelsDeleted(), fmt.Sprintf(`%s%s`, ListingCompo_JsPreFixWhenNotifModelsDeleted, stateful.ReloadAction(ctx, c, nil).Go()),
			)
		}),
		// the dialog is handled internally so that it can make good use of locals
		web.Portal().Name(c.actionDialogPortalName()),
		// user should locate it self
		// https://vuejs.org/guide/built-ins/teleport.html
		h.Tag("Teleport").Attr("to", "#"+c.ActionsComponentTeleportToID()).Children(
			c.actionsComponent(ctx),
		),
		VCard().Elevation(0).Children(
			h.Tag("div").Children(
				c.tabsFilter(ctx),
			).Class("px-2"),
			h.Div(
				c.toolbarSearch(ctx),
			),
			VCardText().Class("list-table-wrap").Children(
				c.dataTable(ctx),
			),
			h.Div(
				c.cardActionsFooter(ctx),
			).Class(("pb-2")),
		).Class("listing-compo-wrap"),
	).MarshalHTML(ctx)
}

func (c *ListingCompo) tabsFilter(ctx context.Context) h.HTMLComponent {
	if c.lb.filterTabsFunc == nil {
		return nil
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
		if c.ActiveFilterTab == ft.ID && stateful.IsRawQuerySubset(c.FilterQuery, encodedQuery) {
			activeIndex = i
			c.activeFilterTabQuery = encodedQuery
		}
		tabs.AppendChildren(
			VTab().
				Attr("@click", stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					target.Page = 0
					target.ActiveFilterTab = ft.ID
					target.FilterQuery = encodedQuery
				}).ThenScript(ListingCompo_JsScrollToTop).Go()).
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

func (c *ListingCompo) textFieldSearchID() string {
	return c.CompoID() + "_textFieldSearch"
}

func (c *ListingCompo) textFieldSearch(ctx context.Context) h.HTMLComponent {
	if c.lb.keywordSearchOff {
		return nil
	}
	_, msgr := c.MustGetEventContext(ctx)
	newReloadAction := func() string {
		return fmt.Sprintf(`
			const targetKeyword = xlocals.keyword || "";
			if (targetKeyword === %q) {
				return;
			}
			%s
			`,
			c.Keyword,
			stateful.ReloadAction(ctx, c,
				func(target *ListingCompo) {
					target.Page = 0
				},
				stateful.WithAppendFix(`v.compo.keyword = targetKeyword;`),
			).ThenScript(ListingCompo_JsScrollToTop).Go(),
		)
	}
	return web.Scope().VSlot("{ locals: xlocals }").Init(fmt.Sprintf("{ keyword: %q }", c.Keyword)).Children(
		VTextField().
			Id(c.textFieldSearchID()).
			Density(DensityCompact).
			Variant(FieldVariantOutlined).
			Label(msgr.Search).
			Flat(true).
			Clearable(true).
			HideDetails(true).
			SingleLine(true).
			Attr("v-model", "xlocals.keyword").
			Attr("@blur", fmt.Sprintf("xlocals.keyword = %q", c.Keyword)).
			Attr("@keyup.enter", newReloadAction()).
			Attr("@click:clear", newReloadAction()).
			Children(
				web.Slot(VIcon("mdi-magnify").Attr("@click", newReloadAction())).Name("append-inner"),
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

	existsInvisibleQuery := ""

	invisibleKeys := map[string]bool{}
	for _, item := range fd {
		if item.Invisible {
			invisibleKeys[item.Key] = true
		}
	}
	if len(invisibleKeys) > 0 {
		if c.FilterQuery != "" {
			qs, err := url.ParseQuery(c.FilterQuery)
			if err == nil {
				for k := range qs {
					if !invisibleKeys[k] {
						delete(qs, k)
					}
				}
				existsInvisibleQuery = qs.Encode()
			}
		}
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

	opts := []stateful.PostActionOption{
		stateful.WithAppendFix(`v.compo.filter_query = $event.encodedFilterData + "";`),
	}
	if existsInvisibleQuery != "" {
		opts = append(opts, stateful.WithAppendFix(fmt.Sprintf(`
			if (v.compo.filter_query !== "" && !v.compo.filter_query.endsWith("&")) {
				v.compo.filter_query += "&";
			}
			v.compo.filter_query += %q;`, existsInvisibleQuery),
		))
	}
	// method tabsFilter need to be called first, it will set activeFilterTabQuery
	if c.activeFilterTabQuery != "" {
		opts = append(opts, stateful.WithAppendFix(fmt.Sprintf(`
			if (!plaid().isRawQuerySubset(v.compo.filter_query, %q)) {
				v.compo.active_filter_tab = "";
			}`, c.activeFilterTabQuery),
		))
	}
	opts = append(opts, stateful.WithAppendFix(`
		if (xlocals.textFieldSearchElem) {
			v.compo.keyword = xlocals.textFieldSearchElem.value;
		}`))
	return web.Scope().VSlot("{locals:xlocals}").Init("{textFieldSearchElem: null}").Children(
		vx.VXFilter(fd).Translations(ft).
			UpdateModelValue(stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
				target.Page = 0
			}, opts...).ThenScript(ListingCompo_JsScrollToTop).Go()).
			Attr("v-on-mounted", fmt.Sprintf(`({el}) => { 
				xlocals.textFieldSearchElem = el.ownerDocument.getElementById(%q); 
			}`, c.textFieldSearchID())),
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

	textFieldSearch := c.textFieldSearch(ctx)
	if textFieldSearch != nil {
		wrapper := VResponsive(textFieldSearch)
		if filterSearch != nil || !c.LongStyleSearchBox {
			wrapper.MaxWidth(200).MinWidth(200).Class("mr-4")
		} else {
			wrapper.Width(100)
		}
		textFieldSearch = wrapper
	}
	return VToolbar().Flat(true).Color("surface").AutoHeight(true).Class("pa-2").Children(
		textFieldSearch,
		filterSearch,
	)
}

func (c *ListingCompo) defaultCellWrapperFunc(_ context.Context) func(cell h.MutableAttrHTMLComponent, id string, obj any, dataTableID string) h.HTMLComponent {
	return func(cell h.MutableAttrHTMLComponent, id string, obj any, _ string) h.HTMLComponent {
		if c.lb.mb.hasDetailing && !c.lb.mb.detailing.drawer {
			cell.SetAttr("@click", web.Plaid().PushStateURL(c.lb.mb.Info().DetailingHref(id)).Go())
			return cell
		}

		event := actions.Edit
		if c.lb.mb.hasDetailing {
			event = actions.DetailingDrawer
		}
		onClick := web.Plaid().EventFunc(event).Query(ParamID, id)
		if c.Popup {
			onClick.URL(c.lb.mb.Info().ListingHref()).Query(ParamOverlay, actions.Dialog)
		}
		if c.ParentID != "" {
			onClick.Query(ParamParentID, c.ParentID)
		}
		onClick.Query(ParamVarCurrentActive, c.VarCurrentActive())
		cell.SetAttr("@click", onClick.Go())
		return cell
	}
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
	var dbOrderBys []string
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
		perPage := cmp.Or(c.PerPage, c.lb.perPage, PerPageDefault)
		if perPage > PerPageMax {
			perPage = PerPageMax
		}
		searchParams.PerPage = perPage
	}
	searchParams.Page = cmp.Or(c.Page, 1)

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

	btnConfigColumns, columns, err := c.getColumns(ctx)
	if err != nil {
		panic(errors.Wrap(err, "get columns error"))
	}

	fieldColumn := lo.SliceToMap(columns, func(col *Column) (string, *Column) {
		return col.Name, col
	})
	dataTable := vx.DataTable(objs).Hover(true).HoverClass("cursor-pointer").
		HeadCellWrapperFunc(func(_ h.MutableAttrHTMLComponent, field string, title string) (compo h.HTMLComponent) {
			defer func() {
				th, ok := compo.(h.MutableAttrHTMLComponent)
				if !ok {
					return
				}
				col, ok := fieldColumn[field]
				if ok && col.WrapHeader != nil {
					wrapper, err := col.WrapHeader(evCtx, col, th)
					if err != nil {
						panic(err)
					}
					compo = wrapper
				} else {
					th.SetAttr("style", "min-width: 100px;")
				}
			}()

			if _, exists := orderableFieldMap[field]; !exists {
				return h.Th(title)
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
			return h.Th("").
				Attr("@click.stop", stateful.ReloadAction(ctx, c, func(target *ListingCompo) {
					target.Page = 0
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
				}).ThenScript(ListingCompo_JsScrollToTop).Go()).
				Children(
					h.Div().Style("cursor: pointer; white-space: nowrap;").Children(
						h.Span(title).Style("text-decoration: underline;"),
						h.Span("").StyleIf("visibility: hidden;", !exists).Children(
							VIcon(icon).Size(SizeSmall),
							h.Span(fmt.Sprint(orderByIdx+1)),
						),
					),
				)
		}).
		RowWrapperFunc(func(row h.MutableAttrHTMLComponent, id string, obj any, _ string) h.HTMLComponent {
			if c.lb.rowProcessor != nil {
				compo, err := c.lb.rowProcessor(evCtx, row, id, obj)
				if err != nil {
					panic(err)
				}
				return compo
			}
			row.SetAttr(":class", fmt.Sprintf(`{ %q: vars.%s === %q }`, ListingCompo_CurrentActiveClass, c.VarCurrentActive(), id))
			return row
		}).
		RowMenuHead(btnConfigColumns).
		RowMenuItemFuncs(c.lb.RowMenu().listingItemFuncs(evCtx)...).
		CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
			if c.lb.cellProcessor != nil {
				compo, err := c.lb.cellProcessor(evCtx, cell, id, obj)
				if err != nil {
					panic(err)
				}
				return compo
			}
			if c.lb.cellWrapperFunc != nil {
				return c.lb.cellWrapperFunc(cell, id, obj, dataTableID)
			}
			return c.defaultCellWrapperFunc(ctx)(cell, id, obj, dataTableID)
		})

	if len(c.lb.bulkActions) > 0 {
		syncQuery := ""
		if stateful.IsSyncQuery(ctx) {
			syncQuery = web.Plaid().PushState(true).MergeQuery(true).Query("selected_ids", web.Var(`selected_ids`)).RunPushState()
		}
		dataTable.SelectedIds(c.SelectedIds).
			OnSelectionChanged(fmt.Sprintf(`function(selected_ids) { locals.selected_ids = selected_ids; %s }`, syncQuery)).
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

	if c.lb.tableProcessor != nil {
		dataTable, err = c.lb.tableProcessor(evCtx, dataTable)
		if err != nil {
			panic(err)
		}
	}

	var dataTableAdditions h.HTMLComponent
	if totalCount <= 0 {
		dataTableAdditions = h.Div().Class("mt-10 text-center grey--text text--darken-2").Children(
			h.Text(msgr.ListingNoRecordToShow),
		)
	} else {
		if c.lb.disablePagination {
			return dataTable
		}
		dataTableAdditions = h.Div().Style("margin-top:26px").Children(
			vx.VXTablePagination().
				Total(int64(totalCount)).
				CurrPage(searchParams.Page).
				PerPage(searchParams.PerPage).
				CustomPerPages([]int64{c.lb.perPage}).
				PerPageText(msgr.PaginationRowsPerPage).
				NoOffsetPart(true).
				TotalVisible(cmp.Or(c.lb.paginationTotalVisible, int64(5))).
				OnSelectPerPage(stateful.ReloadAction(ctx, c,
					func(target *ListingCompo) {
						target.Page = 0
					},
					stateful.WithAppendFix(`v.compo.per_page = parseInt($event, 10)`),
				).ThenScript(ListingCompo_JsScrollToTop).Go()).
				OnSelectPage(stateful.ReloadAction(ctx, c, nil,
					stateful.WithAppendFix(`v.compo.page = parseInt(value,10);`),
				).ThenScript(ListingCompo_JsScrollToTop).Go()),
		)
	}
	return h.Components(dataTable, dataTableAdditions)
}

type Column struct {
	*DisplayColumn
	Label      string                                                                                                        `json:"label"`
	WrapHeader func(evCtx *web.EventContext, col *Column, th h.MutableAttrHTMLComponent) (h.MutableAttrHTMLComponent, error) `json:"-"`
}

func (c *ListingCompo) getColumns(ctx context.Context) (btnConfigure h.HTMLComponent, columns []*Column, err error) {
	evCtx, msgr := c.MustGetEventContext(ctx)

	var availableColumns []*DisplayColumn
	for _, f := range c.lb.fields {
		if c.lb.mb.Info().Verifier().Do(PermList).SnakeOn("f_"+f.name).WithReq(evCtx.R).IsAllowed() != nil {
			continue
		}
		availableColumns = append(availableColumns, &DisplayColumn{
			Name:    f.name,
			Visible: true,
		})
	}

	var displayColumns []*DisplayColumn
	if err := JsonCopy(&displayColumns, c.DisplayColumns); err != nil {
		return nil, nil, err
	}
	// if there is abnormal data, restore the default
	if len(displayColumns) != len(availableColumns) ||
		// names not match
		!lo.EveryBy(displayColumns, func(dc *DisplayColumn) bool {
			return lo.ContainsBy(availableColumns, func(ac *DisplayColumn) bool {
				return ac.Name == dc.Name
			})
		}) {
		displayColumns = availableColumns
	}

	allInvisible := lo.EveryBy(displayColumns, func(dc *DisplayColumn) bool {
		return !dc.Visible
	})
	for _, col := range displayColumns {
		if allInvisible {
			col.Visible = true
		}
		columns = append(columns, &Column{
			DisplayColumn: col,
			Label:         i18n.PT(evCtx.R, ModelsI18nModuleKey, c.lb.mb.label, c.lb.mb.getLabel(c.lb.Field(col.Name).NameLabel)),
		})
	}

	if c.lb.columnsProcessor != nil {
		var err error
		columns, err = c.lb.columnsProcessor(evCtx, columns)
		if err != nil {
			return nil, nil, err
		}
	}

	if !c.lb.selectableColumns {
		return nil, columns, nil
	}

	return web.Scope().
			VSlot("{ locals: xlocals }").
			Init(fmt.Sprintf(`{selectColumnsMenu: false, columns: %s}`, h.JSONString(columns))).
			Children(
				VMenu().CloseOnContentClick(false).Width(240).Attr("v-model", "xlocals.selectColumnsMenu").Children(
					web.Slot().Name("activator").Scope("{ props }").Children(
						VBtn("").Icon("mdi-cog").Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
					),
					VList().Density(DensityCompact).Children(
						h.Tag("vx-draggable").Attr("item-key", "name").Attr("v-model", "xlocals.columns", "handle", ".handle", "animation", "300").Children(
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
							VBtn(msgr.Cancel).Elevation(0).Attr("@click", `xlocals.selectColumnsMenu = false`),
							VBtn(msgr.OK).Elevation(0).Color("primary").Attr("@click", fmt.Sprintf(`
								xlocals.selectColumnsMenu = false; 
								%s`,
								stateful.ReloadAction(ctx, c, nil,
									stateful.WithAppendFix(`v.compo.display_columns = xlocals.columns.map(({ label, ...rest }) => rest)`),
								).Go(),
							)),
						),
					),
				),
			),
		columns, nil
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

func (c *ListingCompo) actionDialogPortalName() string {
	return fmt.Sprintf("%s_action_dialog", c.CompoID())
}

func (c *ListingCompo) actionDialogContentPortalName() string {
	return fmt.Sprintf("%s_action_dialog_content", c.CompoID())
}

func (c *ListingCompo) closeActionDialog() string {
	return fmt.Sprintf("locals.dialog = false;")
}

func (c *ListingCompo) dialog(r *web.EventResponse, comp h.HTMLComponent, width string) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: c.actionDialogPortalName(),
		Body: web.Scope().VSlot("{ form }").Children(
			VDialog().Attr("v-model", "locals.dialog").Width(cmp.Or(width, c.lb.mb.rightDrawerWidth)).Children(
				web.Portal(comp).Name(c.actionDialogContentPortalName()),
			),
		),
	})
	web.AppendRunScripts(r, "setTimeout(function(){ locals.dialog = true }, 100);")
}

func (c *ListingCompo) ActionsComponentTeleportToID() string {
	return fmt.Sprintf("%s_actions", c.CompoID())
}

func (c *ListingCompo) actionsComponent(ctx context.Context) (r h.HTMLComponent) {
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

		label := i18n.PT(evCtx.R, ModelsI18nModuleKey, c.lb.mb.label, c.lb.mb.getLabel(ba.NameLabel))
		buttons = append(buttons, VBtn(label).
			Color(cmp.Or(ba.buttonColor, ColorSecondary)).Variant(VariantFlat).Class("ml-2").
			Attr("@click", stateful.PostAction(ctx, c, c.OpenBulkActionDialog, OpenBulkActionDialogRequest{
				Name: ba.name,
			}).Go()),
		)
	}

	for _, ba := range c.lb.actions {
		if c.lb.mb.Info().Verifier().SnakeDo(permDoListingAction, ba.name).WithReq(evCtx.R).IsAllowed() != nil {
			continue
		}

		if ba.buttonCompFunc != nil {
			buttons = append(buttons, ba.buttonCompFunc(evCtx))
			continue
		}

		label := i18n.PT(evCtx.R, ModelsI18nModuleKey, c.lb.mb.label, c.lb.mb.getLabel(ba.NameLabel))
		buttons = append(buttons, VBtn(label).
			Color(cmp.Or(ba.buttonColor, ColorPrimary)).Variant(VariantElevated).Class("ml-2").
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
		if c.Popup {
			onClick.URL(c.lb.mb.Info().ListingHref()).Query(ParamOverlay, actions.Dialog)
		}
		if c.ParentID != "" {
			onClick.Query(ParamParentID, c.ParentID)
		}
		return VBtn(msgr.New).
			Color(ColorPrimary).
			Variant(VariantElevated).
			Theme("light").Class("ml-2").
			Attr("@click", onClick.Go())
	}()

	if c.lb.actionsAsMenu {
		buttons = append([]h.HTMLComponent{buttonNew}, buttons...)
		return VMenu().OpenOnHover(true).Children(
			web.Slot().Name("activator").Scope("{ props }").Children(
				VBtn(msgr.ButtonLabelActionsMenu).Size(SizeSmall).Attr("v-bind", "props"),
			),
			VList(lo.Map(buttons, func(item h.HTMLComponent, _ int) h.HTMLComponent {
				return VListItem(item)
			})...),
		)
	}

	buttons = append(buttons, buttonNew)
	return h.Div(buttons...)
}

func (c *ListingCompo) bulkPanel(ctx context.Context, bulk *BulkActionBuilder, selectedIds []string, actionableIds []string) h.HTMLComponent {
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
			alertCompo = VAlert(h.Text(notice)).Type(ColorWarning)
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
			VBtn(msgr.Cancel).Variant(VariantFlat).Class("ml-2").Attr("@click", c.closeActionDialog()),
			VBtn(msgr.OK).Color("primary").Variant(VariantFlat).Theme(ThemeDark).Attr("@click",
				stateful.PostAction(ctx, c, c.DoBulkAction, DoBulkActionRequest{
					Name: bulk.name,
				}).Go(),
			),
		),
	)
}

func (c *ListingCompo) fetchBulkAction(ctx context.Context, name string) (*BulkActionBuilder, error) {
	bulk, exists := lo.Find(c.lb.bulkActions, func(ba *BulkActionBuilder) bool {
		return ba.name == name
	})
	if !exists {
		return nil, errors.New("cannot find requested bulk action")
	}

	if bulk.updateFunc == nil {
		return nil, errors.New("bulk.updateFunc not set")
	}

	if bulk.compFunc == nil {
		return nil, errors.New("bulk.compFunc not set")
	}

	evCtx, msgr := c.MustGetEventContext(ctx)
	err := c.lb.mb.Info().Verifier().SnakeDo(permBulkActions, name).WithReq(evCtx.R).IsAllowed()
	if err != nil {
		return nil, err
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
		ShowMessage(&r, err.Error(), ColorError)
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
				ShowMessage(&r, bulk.selectedIdsProcessorNoticeFunc(c.SelectedIds, actionableIds, c.SelectedIds), ColorError)
			} else {
				ShowMessage(&r, msgr.BulkActionNoAvailableRecords, ColorError)
			}
			return r, nil
		}
	}

	c.dialog(&r, c.bulkPanel(ctx, bulk, c.SelectedIds, actionableIds), bulk.dialogWidth)
	return r, nil
}

type DoBulkActionRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) DoBulkAction(ctx context.Context, req DoBulkActionRequest) (r web.EventResponse, err error) {
	evCtx, _ := c.MustGetEventContext(ctx)

	bulk, err := c.fetchBulkAction(ctx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), ColorError)
		return r, nil
	}

	actionableIds := c.SelectedIds
	if bulk.selectedIdsProcessorFunc != nil {
		actionableIds, err = bulk.selectedIdsProcessorFunc(c.SelectedIds, evCtx)
	}

	if err == nil {
		err = bulk.updateFunc(actionableIds, evCtx, &r)
	}

	if err != nil || evCtx.Flash != nil {
		if evCtx.Flash == nil {
			evCtx.Flash = toValidationErrors(err)
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: c.actionDialogContentPortalName(),
			Body: c.bulkPanel(ctx, bulk, c.SelectedIds, actionableIds),
		})
		return r, nil
	}

	web.AppendRunScripts(&r, c.closeActionDialog())
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
			VBtn(msgr.Cancel).Variant(VariantFlat).Class("ml-2").Attr("@click", c.closeActionDialog()),
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

	if action.updateFunc == nil {
		return nil, errors.New("action.updateFunc not set")
	}

	if action.compFunc == nil {
		return nil, errors.New("action.compFunc not set")
	}

	err := c.lb.mb.Info().Verifier().SnakeDo(permDoListingAction, action.name).WithReq(evCtx.R).IsAllowed()
	if err != nil {
		return nil, err
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
		ShowMessage(&r, err.Error(), ColorError)
		return r, nil
	}

	c.dialog(&r, c.actionPanel(ctx, action), action.dialogWidth)
	return r, nil
}

type DoActionRequest struct {
	Name string `json:"name"`
}

func (c *ListingCompo) DoAction(ctx context.Context, req DoActionRequest) (r web.EventResponse, err error) {
	evCtx, _ := c.MustGetEventContext(ctx)

	action, err := c.fetchAction(evCtx, req.Name)
	if err != nil {
		ShowMessage(&r, err.Error(), ColorError)
		return r, nil
	}

	if err := action.updateFunc("", evCtx, &r); err != nil || evCtx.Flash != nil {
		if evCtx.Flash == nil {
			evCtx.Flash = toValidationErrors(err)
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: c.actionDialogContentPortalName(),
			Body: c.actionPanel(ctx, action),
		})
		return r, nil
	}

	web.AppendRunScripts(&r, c.closeActionDialog())
	return r, nil
}

func (c *ListingCompo) MustGetEventContext(ctx context.Context) (*web.EventContext, *Messages) {
	evCtx := web.MustGetEventContext(ctx)
	return evCtx, MustGetMessages(evCtx.R)
}
