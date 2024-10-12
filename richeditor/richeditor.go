package richeditor

import (
	"context"
	"fmt"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/redactor"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// how to add more plugins from https://imperavi.com/redactor/plugins/
// 1. add {{plugin}}.min.js to redactor dir
// 2. add plugin name in Plugins array

// how to add own plugins
// 1. load plugin jss,css to PluginsJS,PluginsCSS
// 2. add plugin names in Plugins array
var (
	Plugins    = []string{"alignment", "table", "video", "imageinsert"}
	PluginsJS  [][]byte
	PluginsCSS [][]byte
)

type RichEditorBuilder struct {
	db            *gorm.DB
	name          string
	value         string
	label         string
	placeholder   string
	plugins       []string
	setPlugins    bool
	rawConfig     interface{}
	errorMessages []string
	disabled      bool
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

// Note: RawConfig overwrites Plugins
func (b *RichEditorBuilder) RawConfig(v interface{}) (r *RichEditorBuilder) {
	b.rawConfig = v
	return b
}

func (b *RichEditorBuilder) ErrorMessages(errMsgs ...string) (r *RichEditorBuilder) {
	b.errorMessages = errMsgs
	return b
}

func (b *RichEditorBuilder) Disabled(disabled bool) (r *RichEditorBuilder) {
	b.disabled = disabled
	return b
}

func (b *RichEditorBuilder) MarshalHTML(ctx context.Context) ([]byte, error) {
	p := Plugins
	if b.setPlugins {
		p = b.plugins
	}
	redactorB := redactor.New().Placeholder(b.placeholder).Attr(web.VField(b.name, b.value)...)
	if b.rawConfig != nil {
		redactorB.RawConfig(b.rawConfig)
	} else {
		redactorB.Config(redactor.Config{Plugins: p})
	}

	r := h.Components(
		v.VSheet(
			h.Label(b.label).Class("v-label theme--light"),
			h.Div().Style("position:relative").
				ClassIf("redactor-error", len(b.errorMessages) > 0 && !b.disabled).
				ClassIf("redactor-disable", b.disabled).
				Children(
					redactorB,
					h.If(b.disabled, h.Div().Class("redactor-disable-overlay")),
				),
			h.Div(
				media.QMediaBox(b.db).FieldName(fmt.Sprintf("%s_richeditor_medialibrary", b.name)).
					Value(&media_library.MediaBox{}).Config(&media_library.MediaBoxConfig{
					AllowType: "image",
				}),
			).Class("hidden-screen-only"),
			h.Iff(len(b.errorMessages) > 0, func() h.HTMLComponent {
				var compos []h.HTMLComponent
				for _, errMsg := range b.errorMessages {
					compos = append(compos, h.Div().Attr("v-pre", true).Text(errMsg))
				}
				return h.Div().Class("d-flex flex-column ps-4 py-1 ga-1 text-caption").
					ClassIf("text-error", len(b.errorMessages) > 0 && !b.disabled).
					ClassIf("text-grey", b.disabled).Children(compos...)
			}),
		).Class("pb-4").Rounded(true).Attr("data-type", "redactor").Attr("style", "position: relative; z-index:1;"),
	)
	return r.MarshalHTML(ctx)
}
