package picker

import (
	"context"
	"fmt"

	"github.com/goplaid/web"
	. "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
)

type PickerBuilder struct {
	value     interface{}
	label     string
	fieldName string
	comp      h.MutableAttrHTMLComponent
}

func Picker(c h.MutableAttrHTMLComponent) (r *PickerBuilder) {
	r = &PickerBuilder{comp: c}
	return
}

func (b *PickerBuilder) Value(v interface{}) (r *PickerBuilder) {
	b.value = v
	return b
}

func (b *PickerBuilder) Label(v string) (r *PickerBuilder) {
	b.label = v
	return b
}

func (b *PickerBuilder) FieldName(v string) (r *PickerBuilder) {
	b.fieldName = v
	return b
}

func (b *PickerBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
	menuLocal := fmt.Sprintf("picker_%s_menu", b.fieldName)
	valueLocal := fmt.Sprintf("picker_%s_value", b.fieldName)

	b.comp.SetAttr("v-field-name", h.JSONString(b.fieldName))
	b.comp.SetAttr("@change", fmt.Sprintf("locals.%s = false; locals.%s = $event", menuLocal, valueLocal))

	return h.Div(
		VMenu(
			web.Slot(
				VTextField().
					Label(b.label).
					Value(b.value).
					Readonly(true).
					PrependIcon("edit_calendar").
					Attr("v-model", fmt.Sprintf("locals.%s", valueLocal)).
					Attr("v-bind", "attrs").
					Attr("v-on", "on"),
			).Name("activator").Scope("{ on, attrs }"),

			b.comp,
		).Attr("v-model", fmt.Sprintf("locals.%s", menuLocal)).
			CloseOnContentClick(false).
			MaxWidth(290),
	).Attr(web.InitContextLocals, fmt.Sprintf(`{%s: %s, %s: false}`, valueLocal, h.JSONString(b.value), menuLocal)).
		MarshalHTML(ctx)

}
