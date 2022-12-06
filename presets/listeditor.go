package presets

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/qor5/admin/presets/actions"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type ListEditorBuilder struct {
	fieldContext         *FieldContext
	value                interface{}
	displayFieldInSorter string
}

type ListSorter struct {
	Items []ListSorterItem `json:"items"`
}

type ListSorterItem struct {
	Index int    `json:"index"`
	Label string `json:"label"`
}

func NewListEditor(v *FieldContext) *ListEditorBuilder {
	return &ListEditorBuilder{fieldContext: v}
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

func (b *ListEditorBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)
	formKey := b.fieldContext.FormKey
	var form h.HTMLComponent
	if b.value != nil {
		form = b.fieldContext.NestedFieldsBuilder.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, formKey string, content h.HTMLComponent, ctx *web.EventContext) h.HTMLComponent {
			return VCard(
				VBtn("Delete").Icon(true).Class("float-right ma-2").
					Children(
						VIcon("delete"),
					).Attr("@click", web.Plaid().
					URL(b.fieldContext.ModelInfo.ListingHref()).
					EventFunc(actions.RemoveRowEvent).
					Query(ParamID, ctx.R.FormValue(ParamID)).
					Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
					Query(ParamRemoveRowFormKey, formKey).
					Go()),
				VCardText(
					content,
				),
			).Class("mb-2").Outlined(true)
		})
	}

	isSortStart := ctx.R.FormValue(ParamIsStartSort) == "1" && ctx.R.FormValue(ParamSortSectionFormKey) == formKey
	haveSorterIcon := true
	var sorter h.HTMLComponent
	var sorterData ListSorter
	if b.value != nil {
		deletedIndexes := ContextModifiedIndexesBuilder(ctx)

		deletedIndexes.SortedForEach(b.value, formKey, func(obj interface{}, i int) {
			if deletedIndexes.DeletedContains(b.fieldContext.FormKey, i) {
				return
			}
			var label = ""
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
		sorter = VCard(VList(
			h.Tag("vx-draggable").Attr("v-model", "locals.items", "draggable", ".item", "animation", "300").Children(
				h.Div(
					VListItem(
						VListItemIcon(VIcon("drag_handle")),
						VListItemContent(
							VListItemTitle(h.Text("{{item.label}}")),
						),
					),
					VDivider().Attr("v-if", "index < locals.items.length - 1", ":key", "index"),
				).Attr("v-for", "(item, index) in locals.items", ":key", "item.index", "class", "item"),
			),
		))
	}

	return h.Div(
		web.Scope(
			VToolbar(
				h.Label(b.fieldContext.Label).Class("v-label v-label--active").Style("font-size: 12px"),
				VSpacer(),
				h.If(haveSorterIcon,
					h.If(!isSortStart,
						VBtn("SortStart").Icon(true).Children(
							VIcon("sort"),
						).Attr("@click",
							web.Plaid().
								URL(b.fieldContext.ModelInfo.ListingHref()).
								EventFunc(actions.SortEvent).
								Query(ParamID, ctx.R.FormValue(ParamID)).
								Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
								Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
								Query(ParamIsStartSort, "1").
								Go(),
						),
					).Else(
						VBtn("SortDone").Icon(true).Children(
							VIcon("done"),
						).Attr("@click",
							web.Plaid().
								URL(b.fieldContext.ModelInfo.ListingHref()).
								EventFunc(actions.SortEvent).
								Query(ParamID, ctx.R.FormValue(ParamID)).
								Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
								Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
								FieldValue(ParamSortResultFormKey, web.Var("JSON.stringify(locals.items)")).
								Query(ParamIsStartSort, "0").
								Go(),
						),
					),
				),
			).Flat(true).Dense(true),
			sorter,
			h.Div(
				form,
				VBtn("Add row").
					Text(true).
					Color("primary").
					Attr("@click", web.Plaid().
						URL(b.fieldContext.ModelInfo.ListingHref()).
						EventFunc(actions.AddRowEvent).
						Query(ParamID, ctx.R.FormValue(ParamID)).
						Query(ParamOverlay, ctx.R.FormValue(ParamOverlay)).
						Query(ParamAddRowFormKey, b.fieldContext.FormKey).
						Go()),
			).Attr("v-show", h.JSONString(!isSortStart)),
		).Init(h.JSONString(sorterData)).VSlot("{ locals }"),
	).MarshalHTML(c)
}

func addListItemRow(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(ParamID), false, ctx)
		formKey := ctx.R.FormValue(ParamAddRowFormKey)
		t := reflectutils.GetType(obj, formKey+"[0]")
		newVal := reflect.New(t.Elem()).Interface()
		err = reflectutils.Set(obj, formKey+"[]", newVal)
		if err != nil {
			panic(err)
		}
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}

func removeListItemRow(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(ParamID), false, ctx)

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
		return
	}
}

func sortListItems(mb *ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(ParamID), false, ctx)
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
			ContextModifiedIndexesBuilder(ctx).SetSorted(sortSectionFormKey, indexes)
		}

		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}
