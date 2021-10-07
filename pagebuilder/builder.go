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

	r.Body = VContainer(comps...).Attr(web.InitContextLocals, `{width: "width: 600px;"}`).
		Class("mt-6").
		Fluid(true)

	return
}

func (b *Builder) containerEditor(obj interface{}, ec *editorContainer, c h.HTMLComponent, ctx *web.EventContext) (r h.HTMLComponent) {

	return VRow(
		VCol(
			h.Div(
				h.Tag("shadow-root").Children(
					h.Div(
						b.pageStyle,
						c,
					),
				),
			).Class("page-builder-container elevation-10 mx-auto").Attr(":style", "locals.width"),
		).Cols(10).Class("pa-0"),

		VCol(
			VBtn("").Attr("@click",
				web.Plaid().
					URL(ec.builder.mb.Info().ListingHref()).
					EventFunc(actions.DrawerEdit, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
					Go(),
			).Color("primary").Children(
				VIcon("settings"),
				h.Text(ec.builder.name),
			).Class("ma-2 float-right"),
		).Cols(2).Class("pa-0"),
	).Attr("style", "border-top: 0.5px dashed gray")

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
<script >

(function (global, factory) {
    typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('vue')) :
    typeof define === 'function' && define.amd ? define(['exports', 'vue'], factory) :
    (global = typeof globalThis !== 'undefined' ? globalThis : global || self, factory(global.shadow = {}, global.Vue));
}(this, (function (exports, Vue) { 'use strict';

    function _interopDefaultLegacy (e) { return e && typeof e === 'object' && 'default' in e ? e : { 'default': e }; }

    var Vue__default = /*#__PURE__*/_interopDefaultLegacy(Vue);

    function makeShadow(el) {
        makeAbstractShadow(el, el.childNodes);
    }
    function makeAbstractShadow(rootEl, childNodes) {
        const fragment = document.createDocumentFragment();
        for (const node of childNodes) {
            fragment.appendChild(node);
        }
        const shadowroot = rootEl.attachShadow({ mode: 'closed' });
        shadowroot.appendChild(fragment);
    }
    function data() {
        return {
            pabstract: false,
            pstatic: false
        };
    }
    const ShadowRoot = Vue__default['default'].extend({
        render(h) {
            return h(this.tag, {}, [
                this.pstatic ? this.$slots.default : h(this.slotTag, { attrs: { id: this.slotId }, 'class': this.slotClass }, [
                    this.$slots.default
                ])
            ]);
        },
        props: {
            abstract: {
                type: Boolean,
                default: false
            },
            static: {
                type: Boolean,
                default: false,
            },
            tag: {
                type: String,
                default: 'div',
            },
            slotTag: {
                type: String,
                default: 'div',
            },
            slotClass: {
                type: String,
            },
            slotId: {
                type: String
            }
        },
        data,
        beforeMount() {
            this.pabstract = this.abstract;
            this.pstatic = this.static;
        },
        mounted() {
            if (this.pabstract) {
                makeAbstractShadow(this.$el.parentElement, this.$el.childNodes);
            }
            else {
                makeShadow(this.$el);
            }
        },
    });
    function install(vue) {
        vue.component('shadow-root', ShadowRoot);
        vue.directive('shadow', {
            bind(el) {
                makeShadow(el);
            }
        });
    }
    if (typeof window != null && window.Vue) {
        install(window.Vue);
    }
    var shadow = { ShadowRoot, shadow_root: ShadowRoot, install };

    exports.ShadowRoot = ShadowRoot;
    exports.default = shadow;
    exports.install = install;
    exports.shadow_root = ShadowRoot;

    Object.defineProperty(exports, '__esModule', { value: true });

})));

</script>
<style>
	.page-builder-container {
		overflow: hidden;
	}
</style>

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
				h.Div(
					VBtn("").Icon(true).Children(
						VIcon("phone_iphone"),
					).Attr("@click", `locals.width="width: 600px"`).
						Class("mr-10"),
					VBtn("").Icon(true).Children(
						VIcon("tablet_mac"),
					).Attr("@click", `locals.width="width: 960px"`).
						Class("mr-10"),

					VBtn("").Icon(true).Children(
						VIcon("laptop_mac"),
					).Attr("@click", `locals.width="width: 1264px"`),
				).Class("mx-auto"),
			).Dark(true).
				Color("primary").
				App(true),

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
