package listeditor

import (
	"context"
	"fmt"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	. "github.com/theplant/htmlgo"
)

type EditorBuilder struct {
	fieldName string
	value     interface{}
	cf        presets.ComponentFunc
}

func Editor() *EditorBuilder {
	return &EditorBuilder{}
}

func (b *EditorBuilder) FieldName(v string) (r *EditorBuilder) {
	b.fieldName = v
	return b
}

func (b *EditorBuilder) Value(v interface{}) (r *EditorBuilder) {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("value must be slice")
	}
	b.value = v
	return b
}

func (b *EditorBuilder) newElementValue() interface{} {
	return reflect.New(reflect.TypeOf(b.value).Elem()).Interface()
}

func (b *EditorBuilder) ComponentFunc(v presets.ComponentFunc) (r *EditorBuilder) {
	b.cf = v
	return b
}

func (b *EditorBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	ctx := web.MustGetEventContext(c)
	var localName = fmt.Sprintf("listeditor_%s", b.fieldName)
	var newElementValueJSON = JSONString(b.newElementValue())
	return Div(
		Div(
			b.cf(ctx),
		).Attr("v-for", fmt.Sprintf("(obj, index) in locals.%s", localName)),
		VBtn("Add row").
			Text(true).
			Color("primary").
			Attr("@click", fmt.Sprintf("locals.%s.push(%s)", localName, newElementValueJSON)),
	).Attr(web.InitContextLocals, fmt.Sprintf(`{%s: %s}`, localName, JSONString(b.value))).MarshalHTML(c)
}
