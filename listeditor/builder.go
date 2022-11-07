package listeditor

import (
	"context"
	"fmt"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
)

type Builder struct {
	fieldContext         *presets.FieldContext
	value                interface{}
	displayFieldInSorter string
}

func New(v *presets.FieldContext) *Builder {
	return &Builder{fieldContext: v}
}

func (b *Builder) Value(v interface{}) (r *Builder) {
	if v == nil {
		return b
	}
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("value must be slice")
	}
	b.value = v
	return b
}

func (b *Builder) DisplayFieldInSorter(v string) (r *Builder) {
	b.displayFieldInSorter = v
	return b
}

func (b *Builder) newElementValue() interface{} {
	return reflect.New(reflect.TypeOf(b.value).Elem()).Interface()
}

func (b *Builder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)
	formKey := b.fieldContext.FormKey
	var form HTMLComponent
	if b.value != nil {
		form = b.fieldContext.ListItemBuilder.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, formKey string, content HTMLComponent, ctx *web.EventContext) HTMLComponent {
			return VCard(
				VBtn("Delete").Icon(true).Class("float-right ma-2").
					Children(
						VIcon("delete"),
					).Attr("@click", web.Plaid().
					URL(b.fieldContext.ModelInfo.ListingHref()).
					EventFunc(removeRowEvent).
					Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
					Query(presets.ParamOverlay, ctx.R.FormValue(presets.ParamOverlay)).
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
	var sorter HTMLComponent
	var sorterData Sorter
	if b.value != nil {
		deletedIndexes := presets.ContextModifiedIndexesBuilder(ctx)

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
			sorterData.Items = append(sorterData.Items, SorterItem{Label: label, Index: i})
		})
	}
	if len(sorterData.Items) < 2 {
		haveSorterIcon = false
	}
	if haveSorterIcon && isSortStart {
		sorter = VCard(VList(
			Tag("vx-draggable").Attr("v-model", "locals.items", "draggable", ".item", "animation", "300").Children(
				Div(
					VListItem(
						VListItemIcon(VIcon("drag_handle")),
						VListItemContent(
							VListItemTitle(Text("{{item.label}}")),
						),
					),
					VDivider().Attr("v-if", "index < locals.items.length - 1", ":key", "index"),
				).Attr("v-for", "(item, index) in locals.items", ":key", "item.index", "class", "item"),
			),
		))
	}

	return Div(
		web.Scope(
			VToolbar(
				Label(b.fieldContext.Label).Class("v-label v-label--active").Style("font-size: 12px"),
				VSpacer(),
				If(haveSorterIcon,
					If(!isSortStart,
						VBtn("SortStart").Icon(true).Children(
							VIcon("sort"),
						).Attr("@click",
							web.Plaid().
								URL(b.fieldContext.ModelInfo.ListingHref()).
								EventFunc(sortEvent).
								Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
								Query(presets.ParamOverlay, ctx.R.FormValue(presets.ParamOverlay)).
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
								EventFunc(sortEvent).
								Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
								Query(presets.ParamOverlay, ctx.R.FormValue(presets.ParamOverlay)).
								Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
								FieldValue(ParamSortResultFormKey, web.Var("JSON.stringify(locals.items)")).
								Query(ParamIsStartSort, "0").
								Go(),
						),
					),
				),
			).Flat(true).Dense(true),
			sorter,
			Div(
				form,
				VBtn("Add row").
					Text(true).
					Color("primary").
					Attr("@click", web.Plaid().
						URL(b.fieldContext.ModelInfo.ListingHref()).
						EventFunc(addRowEvent).
						Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
						Query(presets.ParamOverlay, ctx.R.FormValue(presets.ParamOverlay)).
						Query(ParamAddRowFormKey, b.fieldContext.FormKey).
						Go()),
			).Attr("v-show", JSONString(!isSortStart)),
		).Init(JSONString(sorterData)).VSlot("{ locals }"),
	).MarshalHTML(c)
}
