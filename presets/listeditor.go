package presets

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets/actions"
)

type ListEditorBuilder struct {
	fieldContext           *FieldContext
	value                  interface{}
	displayFieldInSorter   string
	addListItemRowEvent    string
	removeListItemRowEvent string
	sortListItemsEvent     string
	maxItems               int
	addRowBtnLabelFunc     func(*Messages) string
}

type ListSorter struct {
	Items []ListSorterItem `json:"items"`
}

type ListSorterItem struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}

func NewListEditor(v *FieldContext) *ListEditorBuilder {
	return &ListEditorBuilder{
		fieldContext:           v,
		addListItemRowEvent:    actions.AddRowEvent,
		removeListItemRowEvent: actions.RemoveRowEvent,
		sortListItemsEvent:     actions.SortEvent,
		addRowBtnLabelFunc: func(msgr *Messages) string {
			return msgr.AddRow
		},
	}
}

func (b *ListEditorBuilder) Value(v interface{}) (r *ListEditorBuilder) {
	if v == nil {
		return b
	}
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("value must be slice")
	}
	b.value = v
	return b
}

func (b *ListEditorBuilder) DisplayFieldInSorter(v string) (r *ListEditorBuilder) {
	b.displayFieldInSorter = v
	return b
}

func (b *ListEditorBuilder) AddListItemRowEvnet(v string) (r *ListEditorBuilder) {
	if v == "" {
		return b
	}
	b.addListItemRowEvent = v
	return b
}

func (b *ListEditorBuilder) RemoveListItemRowEvent(v string) (r *ListEditorBuilder) {
	if v == "" {
		return b
	}
	b.removeListItemRowEvent = v
	return b
}

func (b *ListEditorBuilder) SortListItemsEvent(v string) (r *ListEditorBuilder) {
	if v == "" {
		return b
	}
	b.sortListItemsEvent = v
	return b
}

func (b *ListEditorBuilder) MaxItems(v int) (r *ListEditorBuilder) {
	b.maxItems = v
	return b
}

func (b *ListEditorBuilder) AddRowBtnLabel(labelFunc interface{}) (r *ListEditorBuilder) {
	if labelFunc == nil {
		b.addRowBtnLabelFunc = func(msgr *Messages) string {
			return msgr.AddRow
		}
	} else {
		b.addRowBtnLabelFunc = labelFunc.(func(*Messages) string)
	}
	return b
}

// getAddRowBtnLabel returns the resolved label text based on the context
func (b *ListEditorBuilder) getAddRowBtnLabel(ctx *web.EventContext) string {
	msgr := MustGetMessages(ctx.R)
	if b.addRowBtnLabelFunc == nil {
		return msgr.AddRow
	}
	return b.addRowBtnLabelFunc(msgr)
}

func (b *ListEditorBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)

	id := ctx.Param(ParamID)

	formKey := b.fieldContext.FormKey
	var form h.HTMLComponent

	deletedIndexes := ContextModifiedIndexesBuilder(ctx)

	// Check if user has update permission for nested operations
	hasUpdatePermission := false
	if b.fieldContext.ModelInfo != nil {
		hasUpdatePermission = b.fieldContext.ModelInfo.Verifier().Do(PermUpdate).WithReq(ctx.R).IsAllowed() == nil
	}

	actualItemCount := 0
	if b.value != nil {
		totalItems := reflect.ValueOf(b.value).Len()
		deletedCount := 0
		for i := 0; i < totalItems; i++ {
			if deletedIndexes.DeletedContains(b.fieldContext.FormKey, i) {
				deletedCount++
			}
		}
		actualItemCount = totalItems - deletedCount
	}

	if b.value != nil {
		form = b.fieldContext.NestedFieldsBuilder.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, formKey string, content h.HTMLComponent, ctx *web.EventContext) h.HTMLComponent {
			return VCard(
				h.If(!b.fieldContext.Disabled && hasUpdatePermission,
					VBtn("").Icon("mdi-delete").Class("float-right ma-2").
						Attr("@click", web.Plaid().
							URL(b.fieldContext.ModelInfo.ListingHref()).
							EventFunc(b.removeListItemRowEvent).
							Queries(ctx.Queries()).
							Query(AddRowBtnKey(b.fieldContext.FormKey), "").
							Query(ParamID, id).
							Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
							Query(ParamRemoveRowFormKey, formKey).
							Go()),
				),
				content,
			).Class("mx-0 mb-2 px-4 pb-0 pt-4").Variant(VariantOutlined)
		})
	}

	isSortStart := ctx.R.FormValue(ParamIsStartSort) == "1" && ctx.R.FormValue(ParamSortSectionFormKey) == formKey
	haveSorterIcon := true
	var sorter h.HTMLComponent
	var sorterData ListSorter
	if b.value != nil {
		// deletedIndexes := ContextModifiedIndexesBuilder(ctx)  // 已在前面声明

		deletedIndexes.SortedForEach(b.value, formKey, func(obj interface{}, i int) {
			if deletedIndexes.DeletedContains(b.fieldContext.FormKey, i) {
				return
			}
			label := ""
			if b.displayFieldInSorter != "" {
				label = fmt.Sprint(reflectutils.MustGet(obj, b.displayFieldInSorter))
			} else {
				label = fmt.Sprintf("Item %d", i)
			}
			sorterData.Items = append(sorterData.Items, ListSorterItem{Label: label, Index: i})
		})
	}
	if len(sorterData.Items) < 2 {
		haveSorterIcon = false
	}
	if haveSorterIcon && isSortStart {
		sorter = VCard(
			VList(
				h.Tag("vx-draggable").Attr("v-model", "locals.items", "handle", ".handle", "animation", "300", "item-key", "index").Children(
					h.Template().Attr("#item", " { element } ").Children(
						VListItem(
							web.Slot(
								VIcon("mdi-drag").Class("handle mx-2 cursor-grab"),
							).Name("prepend"),
							VListItemTitle(h.Text("{{element.label}}")),
							VDivider(),
						),
					),
				),
			).Class("pa-0")).Variant(VariantOutlined).Class("mx-0 mt-1 mb-4")
	}
	addRowBtnId := fmt.Sprintf("%s_%s", b.fieldContext.FormKey, id)

	return h.Div(
		web.Scope(
			h.If(!b.fieldContext.Disabled && hasUpdatePermission,
				h.Div(
					h.Label(b.fieldContext.Label).Class("v-label theme--light text-caption"),
					VSpacer(),
					h.If(haveSorterIcon,
						h.If(!isSortStart,
							VBtn("").Icon("mdi-sort-variant").
								Class("mt-n4").
								Attr("@click",
									web.Plaid().
										URL(b.fieldContext.ModelInfo.ListingHref()).
										EventFunc(b.sortListItemsEvent).
										Queries(ctx.Queries()).
										Query(AddRowBtnKey(b.fieldContext.FormKey), "").
										Query(ParamID, id).
										Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
										Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
										Query(ParamIsStartSort, "1").
										Go(),
								),
						).Else(
							VBtn("").Icon("mdi-check").
								Class("mt-n4").
								Attr("@click",
									web.Plaid().
										URL(b.fieldContext.ModelInfo.ListingHref()).
										EventFunc(b.sortListItemsEvent).
										Queries(ctx.Queries()).
										Query(AddRowBtnKey(b.fieldContext.FormKey), "").
										Query(ParamID, id).
										Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
										Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
										FieldValue(ParamSortResultFormKey, web.Var("JSON.stringify(locals.items)")).
										Query(ParamIsStartSort, "0").
										Go(),
								),
						),
					),
				).Class("d-flex align-end"),
			),
			sorter,
			h.If(!b.fieldContext.Disabled && hasUpdatePermission,
				h.Div(
					form,
					VBtn(b.getAddRowBtnLabel(ctx)).
						Variant(VariantText).
						Color("primary").
						Attr("id", addRowBtnId).
						Disabled(b.maxItems > 0 && actualItemCount >= b.maxItems).
						Attr("@click", web.Plaid().
							URL(b.fieldContext.ModelInfo.ListingHref()).
							EventFunc(b.addListItemRowEvent).
							Queries(ctx.Queries()).
							Query(AddRowBtnKey(b.fieldContext.FormKey), addRowBtnId).
							Query(ParamID, id).
							Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
							Query(ParamAddRowFormKey, b.fieldContext.FormKey).
							Go()),
				).Attr("v-show", h.JSONString(!isSortStart)).
					Class("mt-1 mb-4"),
			),
		).Init(h.JSONString(sorterData)).VSlot("{ locals }"),

		// Read-only view when user doesn't have update permission or field is disabled
		h.If(b.fieldContext.Disabled || !hasUpdatePermission,
			h.Div(
				h.Label(b.fieldContext.Label).Class("v-label theme--light text-caption"),
				form,
			),
		),
	).MarshalHTML(c)
}

func addListItemRow(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		id := ctx.Param(ParamID)
		obj, _ := me.FetchAndUnmarshal(id, false, ctx)

		if mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			ShowMessage(&r, perm.PermissionDenied.Error(), ColorError)
			return r, nil
		}

		formKey := ctx.Param(ParamAddRowFormKey)
		t := reflectutils.GetType(obj, formKey+"[0]")
		newVal := reflect.New(t.Elem()).Interface()
		err = reflectutils.Set(obj, formKey+"[]", newVal)
		if err != nil {
			panic(err)
		}
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		web.AppendRunScripts(&r, fmt.Sprintf(`setTimeout(function(){%s},200)`, web.Emit(mb.NotifRowUpdated(), PayloadRowUpdated{Id: id})))
		return
	}
}

func removeListItemRow(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		id := ctx.Param(ParamID)
		obj, _ := me.FetchAndUnmarshal(id, false, ctx)

		if mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			ShowMessage(&r, perm.PermissionDenied.Error(), ColorError)
			return r, nil
		}

		formKey := ctx.R.FormValue(ParamRemoveRowFormKey)
		lb := strings.LastIndex(formKey, "[")
		sliceField := formKey[0:lb]
		strIndex := formKey[lb+1 : strings.LastIndex(formKey, "]")]

		var index int
		index, err = strconv.Atoi(strIndex)
		if err != nil {
			return
		}
		ContextModifiedIndexesBuilder(ctx).AppendDeleted(sliceField, index)
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		web.AppendRunScripts(&r, fmt.Sprintf(`setTimeout(function(){%s},200)`, web.Emit(mb.NotifRowUpdated(), PayloadRowUpdated{Id: id})))
		return
	}
}

func sortListItems(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		id := ctx.Param(ParamID)
		obj, _ := me.FetchAndUnmarshal(id, false, ctx)

		if mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			ShowMessage(&r, perm.PermissionDenied.Error(), ColorError)
			return r, nil
		}

		sortSectionFormKey := ctx.R.FormValue(ParamSortSectionFormKey)

		isStartSort := ctx.R.FormValue(ParamIsStartSort)
		if isStartSort != "1" {
			sortResult := ctx.R.FormValue(ParamSortResultFormKey)

			var result []ListSorterItem
			err = json.Unmarshal([]byte(sortResult), &result)
			if err != nil {
				return
			}
			var indexes []string
			for _, i := range result {
				indexes = append(indexes, fmt.Sprint(i.Index))
			}
			ContextModifiedIndexesBuilder(ctx).Sorted(sortSectionFormKey, indexes)
			web.AppendRunScripts(&r, fmt.Sprintf(`setTimeout(function(){%s},200)`, web.Emit(mb.NotifRowUpdated(), PayloadRowUpdated{Id: id})))
		}

		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}

func AddRowBtnKey(fromKey string) string {
	return fmt.Sprintf("%sAddRowBtnID", fromKey)
}
