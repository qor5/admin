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
	fields       *presets.FieldBuilders
}

func New(fields *presets.FieldBuilders) *Builder {
	return &Builder{
		fields: fields,
	}
}

func (b *Builder) FieldContext(v *presets.FieldContext) (r *Builder) {
	b.fieldContext = v
	return b
}

func (b *Builder) Value(v interface{}) (r *Builder) {
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

	form := b.fields.ToComponentForEach(b.fieldContext, b.value, ctx, func(obj interface{}, content HTMLComponent, ctx *web.EventContext) HTMLComponent {
		return Tr(
			Td(
				VCard(
					VCardText(
						content,
					),
				).Class("mb-2"),
			),
			Td(
				VBtn("Delete").Icon(true).
					Color("error").Children(
					VIcon("clear"),
				),
			).Style("width: 1%"),
		)
	})

	return Div(
		Table(
			Tbody(
				form,
				Tr(
					Td(
						VBtn("Add row").
							Text(true).
							Color("primary"),
					),
					Td(),
				),
			),
		),
	).MarshalHTML(c)
}
