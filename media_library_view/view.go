package media_library_view

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/media/media_library"
	h "github.com/theplant/htmlgo"
)

type MediaBoxConfigKey int

const MediaBoxConfig MediaBoxConfigKey = iota

func MediaBoxComponentFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	cfg := field.ContextValue(MediaBoxConfig).(*media_library.MediaBoxConfig)
	return VContainer(h.Text("123" + cfg.AllowType))
}

func MediaBoxSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	return
}
