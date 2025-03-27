package containers

import (
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	. "github.com/theplant/htmlgo"
)

type CommonStyle struct {
	MarginTop    int
	MarginBottom int
	MarginLeft   int
	MarginRight  int
}

func SetCommonStyleComponent(pb *pagebuilder.Builder, eb *presets.FieldBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&CommonStyle{}).Only("MarginTop", "MarginBottom", "MarginLeft", "MarginRight")

	fb.Field("MarginTop").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vx.VXField().
			Type("number").
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label)
	})
	fb.Field("MarginBottom").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vx.VXField().
			Type("number").
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label)
	})
	fb.Field("MarginLeft").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vx.VXField().
			Type("number").
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label)
	})
	fb.Field("MarginRight").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return vx.VXField().
			Type("number").
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label)
	})

	eb.Nested(fb)
}
