package presets

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
)

func TestInputWithDefaults(t *testing.T) {
	type Person struct {
		Name string
	}
	ctx := context.Background()
	obj := Person{Name: "Tom"}
	field := &FieldContext{
		Name:    "Name",
		FormKey: "NameFormKey",
		Label:   "NameLabel",
	}

	// HTML component
	t.Run("Input", func(t *testing.T) {
		expect, _ := h.Input("").Attr(web.VFieldName(field.FormKey)...).Value(field.StringValue(obj)).MarshalHTML(ctx)
		result, _ := InputWithDefaults(h.Input(""), obj, field).MarshalHTML(ctx)
		if diff := cmp.Diff(expect, result); diff != "" {
			t.Fatal(diff)
		}
	})

	// Vuetify component
	t.Run("VSelect", func(t *testing.T) {
		expect, _ := v.VSelect().FieldName(field.FormKey).Value(field.Value(obj)).Label(field.Label).Outlined(true).MarshalHTML(ctx)
		result, _ := InputWithDefaults(v.VSelect(), obj, field).Outlined(true).MarshalHTML(ctx)
		if diff := cmp.Diff(expect, result); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("VTextField", func(t *testing.T) {
		expect, _ := v.VTextField().FieldName(field.FormKey).Value(field.Value(obj)).Label(field.Label).MarshalHTML(ctx)
		result, _ := InputWithDefaults(v.VTextField(), obj, field).MarshalHTML(ctx)
		if diff := cmp.Diff(expect, result); diff != "" {
			t.Fatal(diff)
		}
	})
}
