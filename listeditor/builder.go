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
	"github.com/thoas/go-funk"
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

	var form HTMLComponent
	if b.value != nil {
		form = b.fieldContext.ListItemBuilder.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, formKey string, content HTMLComponent, ctx *web.EventContext) HTMLComponent {
			return Tr(
				Td(
					VCard(
						VCardText(
							content,
						),
					).Class("mb-2").Outlined(true),
				),
				Td(
					VBtn("Delete").Icon(true).
						Color("error").Children(
						VIcon("clear"),
					).Attr("@click", web.Plaid().
						EventFunc(removeRowEvent).
						Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
						Query(ParamRemoveRowFormKey, formKey).
						Go()),
				).Style("width: 1%"),
			)
		})
	}

	isSortStart := ctx.R.FormValue(ParamIsStartSort) == "1" && ctx.R.FormValue(ParamSortSectionFormKey) == b.fieldContext.FormKey
	haveSorterIcon := true
	var i = 0
	var j = 0
	var sorter HTMLComponent
	var sorterData Sorter
	if b.value != nil {
		deletedIndexes := presets.ContextModifiedIndexesBuilder(ctx)

		funk.ForEach(b.value, func(obj interface{}) {
			defer func() { i++ }()
			if deletedIndexes.DeletedContains(b.fieldContext.FormKey, i) {
				return
			}
			var label = ""
			if b.displayFieldInSorter != "" {
				label = fmt.Sprint(reflectutils.MustGet(obj, b.displayFieldInSorter))
			} else {
				j++
				label = fmt.Sprintf("Item %d", j)
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
						VListItemIcon(VIcon("reorder")),
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
			Div(
				Label(b.fieldContext.Label).Class("v-label v-label--active").Style("font-size: 12px"),
				If(haveSorterIcon,
					If(!isSortStart,
						VBtn("SortStart").Class("float-right").Icon(true).Children(
							VIcon("sort"),
						).Attr("@click",
							web.Plaid().
								EventFunc(sortEvent).
								Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
								Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
								Query(ParamIsStartSort, "1").
								Go(),
						),
					).Else(
						VBtn("SortDone").Class("float-right").Icon(true).Children(
							VIcon("done"),
						).Attr("@click",
							web.Plaid().
								EventFunc(sortEvent).
								Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
								Query(ParamSortSectionFormKey, b.fieldContext.FormKey).
								FieldValue(ParamSortResultFormKey, web.Var("JSON.stringify(locals.items)")).
								Query(ParamIsStartSort, "0").
								Go(),
						),
					),
				),
			),
			sorter,
			Table(
				Tbody(
					form,
					Tr(
						Td(
							VBtn("Add row").
								Text(true).
								Color("primary").
								Attr("@click", web.Plaid().
									EventFunc(addRowEvent).
									Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
									Query(ParamAddRowFormKey, b.fieldContext.FormKey).
									Go()),
						),
						Td(),
					),
				),
			).Attr("v-show", JSONString(!isSortStart)),
		).Init(JSONString(sorterData)).VSlot("{ locals }"),
	).MarshalHTML(c)
}
