package pagebuilder

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	. "github.com/goplaid/x/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
)

func (b *Builder) Editor(ctx *web.EventContext) (r web.PageResponse, err error) {

	id := pat.Param(ctx.R, "id")

	var page Page
	err = b.db.First(&page, "id = ?", id).Error
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

	r.Body = h.Components(
		VAppBar(
			VSpacer(),

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

			VSpacer(),
			b.addContainerMenu(id),
		).Dark(true).
			Color("primary").
			App(true),

		VMain(

			VContainer(comps...).Attr(web.InitContextLocals, `{width: "width: 600px;"}`).
				Class("mt-6").
				Fluid(true),
		),
	)

	return
}

const AddContainerEvent = "page_builder_AddContainerEvent"
const DeleteContainerEvent = "page_builder_DeleteContainerEvent"

func (b *Builder) AddContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.Event.ParamAsInt(0)
	containerName := ctx.Event.Params[1]

	err = b.AddContainerToPage(pageID, containerName)

	r.PushState = web.PushState(url.Values{})
	return
}

func (b *Builder) DeleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.Event.ParamAsInt(0)
	containerID := ctx.Event.ParamAsInt(1)

	err = b.db.Delete(&Container{}, "id = ? AND page_id = ?", containerID, pageID).Error
	if err != nil {
		return
	}
	r.PushState = web.PushState(url.Values{})
	return
}

func (b *Builder) AddContainerToPage(pageID int, containerName string) (err error) {
	model := b.ContainerByName(containerName).NewModel()
	err = b.db.Create(model).Error
	if err != nil {
		return
	}

	var maxOrder float64
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ?", pageID).Scan(&maxOrder).Error
	if err != nil {
		return
	}

	err = b.db.Create(&Container{
		PageID:       uint(pageID),
		Name:         containerName,
		ModelID:      reflectutils.MustGet(model, "ID").(uint),
		DisplayOrder: maxOrder + 8,
	}).Error
	if err != nil {
		return
	}
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
			).Class("page-builder-container mx-auto").Attr(":style", "locals.width"),
		).Cols(10).Class("pa-0"),

		VCol(
			VMenu(
				web.Slot(

					VBtn("").Color("primary").Children(
						VIcon("settings"),
						h.Text(ec.builder.name),
					).Class("ma-2 float-right").
						Attr("v-bind", "attrs", "v-on", "on"),
				).Name("activator").Scope("{ on, attrs }"),

				VList(
					VListItem(
						VListItemTitle(h.Text("Edit")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(actions.DrawerEdit, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
							Go(),
					),

					VDivider(),

					VListItem(
						VListItemTitle(h.Text("Delete")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(DeleteContainerEvent,
								fmt.Sprint(ec.container.PageID),
								fmt.Sprint(ec.container.ID)).
							Go(),
					),
				),
			).OffsetY(true),
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
		box-shadow: -10px 0px 13px -7px rgba(0,0,0,.3), 10px 0px 13px -7px rgba(0,0,0,.18), 5px 0px 15px 5px rgba(0,0,0,.12);	
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

			innerPr.Body.(h.HTMLComponent),
		).Id("vt-app")

		return
	}
}

func (b *Builder) addContainerMenu(id string) h.HTMLComponent {
	var items []h.HTMLComponent

	for _, builder := range b.containerBuilders {
		items = append(items,
			VCol(
				VCard(

					VCardTitle(h.Text(builder.name)),
					VCardActions(
						VSpacer(),
						VBtn("Select").
							Text(true).
							Color("primary").Attr("@click",
							web.Plaid().EventFunc(AddContainerEvent, id, builder.name).
								Go(),
						),
					),
				),
			).Cols(4),
		)
	}

	return VMenu(
		web.Slot(
			VBtn("Add Container").Text(true).
				Attr("v-bind", "attrs", "v-on", "on"),
		).Name("activator").Scope("{ on, attrs }"),
		VSheet(
			VContainer(
				VRow(
					items...,
				),
			),
		),
	).OffsetY(true).NudgeWidth(600).CloseOnContentClick(false)
}
