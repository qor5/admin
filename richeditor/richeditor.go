package richeditor

import (
	"context"
	"fmt"

	"github.com/qor5/ui/redactor"
	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/admin/media/media_library"
	media_view "github.com/qor5/admin/media/views"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
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

type RichEditorBuilder struct {
	db          *gorm.DB
	name        string
	value       string
	label       string
	placeholder string
	plugins     []string
	setPlugins  bool
}

func RichEditor(db *gorm.DB, name string) (r *RichEditorBuilder) {
	r = &RichEditorBuilder{db: db, name: name}
	return
}

func (b *RichEditorBuilder) Value(v string) (r *RichEditorBuilder) {
	b.value = v
	return b
}

func (b *RichEditorBuilder) Label(v string) (r *RichEditorBuilder) {
	b.label = v
	return b
}

func (b *RichEditorBuilder) Placeholder(v string) (r *RichEditorBuilder) {
	b.placeholder = v
	return b
}

func (b *RichEditorBuilder) Plugins(v []string) (r *RichEditorBuilder) {
	b.plugins = v
	b.setPlugins = true
	return b
}

func (b *RichEditorBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
	p := Plugins
	if b.setPlugins {
		p = b.plugins
	}
	r := h.Components(
		v.VSheet(
			h.Label(b.label).Class("v-label theme--light"),
			redactor.New().Value(b.value).Placeholder(b.placeholder).Config(redactor.Config{Plugins: p}).Attr(web.VFieldName(b.name)...),
			h.Div(
				media_view.QMediaBox(b.db).FieldName(fmt.Sprintf("%s_richeditor_medialibrary", b.name)).
					Value(&media_library.MediaBox{}).Config(&media_library.MediaBoxConfig{
					AllowType: "image",
				}),
			).Class("hidden-screen-only"),
		).Class("pb-4").Rounded(true).Attr("data-type", "redactor").Attr("style", "position: relative; z-index:1;"),
	)
	return r.MarshalHTML(ctx)
}
