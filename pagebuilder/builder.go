package pagebuilder

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	media_view "github.com/qor/qor5/media/views"
	h "github.com/theplant/htmlgo"
	goji "goji.io"
	"gorm.io/gorm"
)

type RenderFunc func(obj interface{}, ctx *web.EventContext) h.HTMLComponent

type Builder struct {
	prefix            string
	mux               *goji.Mux
	wb                *web.Builder
	db                *gorm.DB
	containerBuilders []*ContainerBuilder
	ps                *presets.Builder
	pageStyle         h.HTMLComponent
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
		DataOperator(gorm2op.DataOperator(db)).
		URIPrefix(r.prefix).
		LayoutFunc(r.pageEditorLayout)

	media_view.Configure(r.ps, db)

	type Editor struct {
	}
	r.ps.Model(&Editor{}).
		Detailing().
		PageFunc(r.Editor)
	r.ps.GetWebBuilder().RegisterEventFunc(AddContainerEvent, r.AddContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(DeleteContainerEvent, r.DeleteContainer)
	return r
}

func (b *Builder) Prefix(v string) (r *Builder) {
	b.ps.URIPrefix(v)
	return b
}

func (b *Builder) PageStyle(v h.HTMLComponent) (r *Builder) {
	b.pageStyle = v
	return b
}

func (b *Builder) Configure(pb *presets.Builder) {
	pb.Model(&Page{})
	pb.Model(&Container{})
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
	b.ps.ServeHTTP(w, r)
}
