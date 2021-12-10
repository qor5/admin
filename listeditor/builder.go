package listeditor

import (
	"context"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	. "github.com/theplant/htmlgo"
	"github.com/thoas/go-funk"
)

type Builder struct {
	fieldContext *presets.FieldContext
	value        interface{}
	cf           presets.ComponentFunc
	presets.FieldBuilders
}

func New() *Builder {
	return &Builder{}
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

func (b *Builder) ComponentFunc(v presets.ComponentFunc) (r *Builder) {
	b.cf = v
	return b
}

func (b *Builder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)

	var rows []HTMLComponent
	var mi *presets.ModelInfo
	if b.fieldContext != nil {
		mi = b.fieldContext.ModelInfo
	}

	funk.ForEach(b.value, func(obj interface{}) {
		rows = append(rows,
			Tr(
				Td(
					VCard(
						VCardText(
							b.FieldBuilders.ToComponent(mi, obj, ctx),
						),
					).Class("mb-2"),
				),
				Td(
					VBtn("Delete").Icon(true).
						Color("error").Children(
						VIcon("clear"),
					),
				).Style("width: 1%"),
			),
		)
	})

	return Div(

		Table(
			Tbody(
				HTMLComponents(rows),
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
