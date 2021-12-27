package listeditor

import (
	"context"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	. "github.com/theplant/htmlgo"
)

type Builder struct {
	fieldContext *presets.FieldContext
	value        interface{}
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

func (b *Builder) newElementValue() interface{} {
	return reflect.New(reflect.TypeOf(b.value).Elem()).Interface()
}

func (b *Builder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)

	var form HTMLComponent
	if b.value != nil {
		form = b.fieldContext.ListItemBuilder.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, content HTMLComponent, ctx *web.EventContext) HTMLComponent {
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
					),
				).Style("width: 1%"),
			)
		})
	}

	return Div(
		Table(
			Tbody(
				Label(b.fieldContext.Label).Class("v-label v-label--active").Style("font-size: 12px"),
				form,
				Tr(
					Td(
						VBtn("Add row").
							Text(true).
							Color("primary").
							Attr("@click", web.Plaid().
								EventFunc(addRowEvent).
								Query(presets.ParamID, ctx.R.FormValue(presets.ParamID)).
								Query(ParamOpFormKey, b.fieldContext.FormKey).
								Go()),
					),
					Td(),
				),
			),
		),
	).MarshalHTML(c)
}
