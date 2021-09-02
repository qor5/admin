package richeditor

import (
	"github.com/goplaid/web"
	v "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/media/media_library"
	"github.com/qor/qor5/media_library_view"
	h "github.com/theplant/htmlgo"
)

// how to add more plugins from https://imperavi.com/redactor/plugins/
// 1. add {{plugin}}.min.js to redactor dir
// 2. add plugin name in Plugins array

// how to add own plugins
// 1. load plugin jss,css to PluginsJS,PluginsCSS
// 2. add plugin names in Plugins array
var Plugins = []string{"alignment", "table", "video", "imageinsert"}
var PluginsJS [][]byte
var PluginsCSS [][]byte

func RichEditor(db *gorm.DB, name, value, label, placeholder string) h.HTMLComponent {
	return h.Components(
		v.VSheet(
			h.Label(label).Class("v-label theme--light"),
			Redactor().Value(value).Placeholder(placeholder).Config(RedactorConfig{Plugins: Plugins}).Attr(web.VFieldName(name)...),
			h.Div(
				media_library_view.QMediaBox(db).FieldName("richeditor").
					Value(&media_library.MediaBox{}).Config(&media_library.MediaBoxConfig{
					AllowType: "image",
				}),
			).Class("hidden-screen-only"),
		).Class("pb-4").Rounded(true).Attr("data-type", "redactor"),
	)
}
