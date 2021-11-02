package pagebuilder

import (
	"fmt"
	"github.com/qor/qor5/publish"
	"net/http"
	"reflect"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	. "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type RenderFunc func(obj interface{}, ctx *web.EventContext) h.HTMLComponent

type Builder struct {
	prefix            string
	wb                *web.Builder
	db                *gorm.DB
	containerBuilders []*ContainerBuilder
	ps                *presets.Builder
	pageStyle         h.HTMLComponent
	preview           http.Handler
}

func New(db *gorm.DB) *Builder {
	err := db.AutoMigrate(
		&Page{},
		&Container{},
	)

	if err != nil {
		panic(err)
	}

	r := &Builder{
		db:     db,
		wb:     web.New(),
		prefix: "/page_builder",
	}

	r.ps = presets.New().
		BrandTitle("Page Builder").
		DataOperator(gorm2op.DataOperator(db)).
		URIPrefix(r.prefix).
		LayoutFunc(r.pageEditorLayout).
		ExtraAsset("/vue-shadow-dom.js", "text/javascript", ShadowDomComponentsPack())

	type Editor struct {
	}
	r.ps.Model(&Editor{}).
		Detailing().
		PageFunc(r.Editor)
	r.ps.GetWebBuilder().RegisterEventFunc(AddContainerEvent, r.AddContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(DeleteContainerEvent, r.DeleteContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(MoveContainerEvent, r.MoveContainer)
	r.preview = r.ps.GetWebBuilder().Page(r.Preview)
	return r
}

func (b *Builder) Prefix(v string) (r *Builder) {
	b.ps.URIPrefix(v)
	b.prefix = v
	return b
}

func (b *Builder) PageStyle(v h.HTMLComponent) (r *Builder) {
	b.pageStyle = v
	return b
}

func (b *Builder) GetPresetsBuilder() (r *presets.Builder) {
	return b.ps
}

func (b *Builder) Configure(pb *presets.Builder, pm *presets.ModelBuilder) {
	list := pm.Listing("ID", "Title", "Slug", "Draft Count", "Status")
	list.Field("ID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		return h.Td(
			h.A().Children(
				h.Text(fmt.Sprintf("Editor for %d", p.ID)),
			).Href(fmt.Sprintf("%s/editors/%d", b.prefix, p.ID)).
				Target("_blank"),
			VIcon("open_in_new").Size(16).Class("ml-1"),
		)
	})
	list.Field("Draft Count").ComponentFunc(func(obj interface{}, _ *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		var count int64
		b.db.Model(&Page{}).Where("id = ? AND status = ?", obj.(*Page).ID, publish.StatusDraft).Count(&count)

		return h.Td(h.Text(fmt.Sprint(count)))
	})

	pm.Editing("Status", "Schedule", "Title", "Slug")
}

func (b *Builder) ContainerByName(name string) (r *ContainerBuilder) {
	for _, cb := range b.containerBuilders {
		if cb.name == name {
			return cb
		}
	}
	panic(fmt.Sprintf("No container: %s", name))
}

type ContainerBuilder struct {
	builder       *Builder
	name          string
	mb            *presets.ModelBuilder
	model         interface{}
	modelType     reflect.Type
	containerFunc RenderFunc
}

func (b *Builder) RegisterContainer(name string) (r *ContainerBuilder) {
	r = &ContainerBuilder{
		name:    name,
		builder: b,
	}
	b.containerBuilders = append(b.containerBuilders, r)
	return
}

func (b *ContainerBuilder) Model(m interface{}) *ContainerBuilder {
	b.model = m
	b.mb = b.builder.ps.Model(m)

	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Ptr {
		panic("model pointer type required")
	}

	b.modelType = val.Elem().Type()
	return b
}

func (b *ContainerBuilder) RenderFunc(v RenderFunc) *ContainerBuilder {
	b.containerFunc = v
	return b
}

func (b *ContainerBuilder) NewModel() interface{} {
	return reflect.New(b.modelType).Interface()
}

func (b *ContainerBuilder) ModelTypeName() string {
	return b.modelType.String()
}

func (b *ContainerBuilder) Editing(vs ...string) *presets.EditingBuilder {
	return b.mb.Editing(vs...)
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Index(r.RequestURI, b.prefix+"/preview") >= 0 {
		b.preview.ServeHTTP(w, r)
		return
	}
	b.ps.ServeHTTP(w, r)
}
