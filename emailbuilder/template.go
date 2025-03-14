package emailbuilder

import (
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets"
)

var (
	cardHeight        = 146
	cardContentHeight = 56
)

func DefaultMailTemplate(pb *presets.Builder) *presets.ModelBuilder {
	mb := pb.Model(&EmailTemplate{}).Label("Email Templates")
	editing := mb.Editing("Name", "JSONBody", "HTMLBody")
	editing.Field("JSONBody").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Div(in(obj, field, ctx)).Style("display:none")
		}
	})
	editing.Field("HTMLBody").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Div(in(obj, field, ctx)).Style("display:none")
		}
	})
	editing.Creating("Name")
	mb.Detailing(EmailEditorField)
	return mb
}
