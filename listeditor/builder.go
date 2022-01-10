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

	var haveSorter = b.value != nil

	var i = 0
	var sorter HTMLComponent
	if haveSorter {
		sorter1 := VList().Attr("v-if", "locals.isSorting")
		fmt.Println("b.value", b.value)

		funk.ForEach(b.value, func(obj interface{}) {
			i++

			var label = ""
			if b.displayFieldInSorter != "" {
				label = fmt.Sprint(reflectutils.MustGet(obj, b.displayFieldInSorter))
			} else {
				label = fmt.Sprintf("Item %d", i)
			}

			sorter1.AppendChildren(
				VListItem(
					VListItemContent(
						VListItemTitle(Text(label)),
					),
				),
			)
		})
		sorter = VCard(sorter1)

		if i < 2 {
			sorter = nil
			haveSorter = false
		}

	}

	return Div(
		web.Scope(
			Div(
				Label(b.fieldContext.Label).Class("v-label v-label--active").Style("font-size: 12px"),
				If(haveSorter,
					VBtn("Sort").Class("float-right").Icon(true).Children(
						VIcon("sort"),
					).Attr("@click", "locals.isSorting = !locals.isSorting"),
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
			).Attr("v-if", "!locals.isSorting"),
		).Init("{ isSorting: false }").VSlot("{ locals }"),
	).MarshalHTML(c)
}
