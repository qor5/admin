package richeditor

import (
	"github.com/goplaid/web"
	v "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
)

func RichEditor(name, value, label, placeholder string) h.HTMLComponent {
	return h.Components(
		v.VSheet(
			h.Label(label).Class("v-label theme--light"),
			Redactor().Value(value).Placeholder(placeholder).Attr(web.VFieldName(name)...),
		).Class("pb-4").Rounded(true),
	)
}
