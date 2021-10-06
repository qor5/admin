package pagebuilder

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	"github.com/goplaid/x/presets/gorm2op"
	. "github.com/goplaid/x/vuetify"
	media_view "github.com/qor/qor5/media/views"
	"github.com/sunfmin/reflectutils"
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
	return r
}

func (b *Builder) Prefix(v string) (r *Builder) {
	b.prefix = v
	return b
}

func (b *Builder) Configure(pb *presets.Builder) {
	pb.Model(&Page{})
	pb.Model(&Container{})
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

func (b *Builder) Editor(ctx *web.EventContext) (r web.PageResponse, err error) {

	segs := strings.Split(ctx.R.URL.Path, "/")
	slug := segs[len(segs)-1]
	if len(slug) == 0 && len(segs) > 1 {
		slug = segs[len(segs)-2]
	}

	var page Page
	err = b.db.First(&page, "slug = ?", slug).Error
	if err != nil {
		return
	}

	var cons []*Container
	err = b.db.Find(&cons, "page_id = ?", page.ID).Order("order ASC").Error
	if err != nil {
		return
	}

	var comps []h.HTMLComponent
	cbs := b.getContainerBuilders(cons)
	for _, ec := range cbs {
		obj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(obj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}

		pure := ec.builder.containerFunc(obj, ctx)

		comps = append(comps, b.containerEditor(obj, ec, pure, ctx))
	}

	r.Body = h.Components(comps...)

	return
}

func (b *Builder) containerEditor(obj interface{}, ec *editorContainer, c h.HTMLComponent, ctx *web.EventContext) (r h.HTMLComponent) {

	return VCard(
		c,
		VCardActions(
			VSpacer(),
			VBtn("Edit").Attr("@click",
				web.Plaid().
					URL(ec.builder.mb.Info().ListingHref()).
					EventFunc(actions.DrawerEdit, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
					Go(),
			).Text(true).Color("primary"),
		),
	).Class("mb-2")
}

type editorContainer struct {
	builder   *ContainerBuilder
	container *Container
}

func (b *Builder) getContainerBuilders(cs []*Container) (r []*editorContainer) {
	for _, c := range cs {
		for _, cb := range b.containerBuilders {
			if cb.name == c.Name {
				r = append(r, &editorContainer{
					builder:   cb,
					container: c,
				})
			}
		}
	}
	return
}

func (b *Builder) pageEditorLayout(in web.PageFunc) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {

		ctx.Injector.HeadHTML(strings.Replace(`
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto+Mono">
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500">
			<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
			<link rel="stylesheet" href="{{prefix}}/assets/main.css">
			<script src='{{prefix}}/assets/vue.js'></script>
			<style>
				[v-cloak] {
					display: none;
				}
			</style>
		`, "{{prefix}}", b.prefix, -1))

		if len(os.Getenv("DEV_PRESETS")) > 0 {
			ctx.Injector.TailHTML(`
<script src='http://localhost:3080/js/chunk-vendors.js'></script>
<script src='http://localhost:3080/js/app.js'></script>
<script src='http://localhost:3100/js/chunk-vendors.js'></script>
<script src='http://localhost:3100/js/app.js'></script>
			`)

		} else {
			ctx.Injector.TailHTML(strings.Replace(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix, -1))
		}

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}

		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, "Page Editor")
		pr.Body = VApp(

			web.Portal().Name(presets.RightDrawerPortalName),

			VAppBar(
				h.Text("Hello"),
			).Dark(true).
				Color("primary").
				App(true).
				ClippedLeft(true),

			VMain(
				innerPr.Body.(h.HTMLComponent),
			),
		).Id("vt-app")

		return
	}
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.ps.ServeHTTP(w, r)
}
