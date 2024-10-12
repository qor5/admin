package examples_vuetify

import (
	"net/http"

	"github.com/qor5/admin/v3/docs/docsrc/assets"
	"github.com/qor5/web/v3"
	wexamples "github.com/qor5/web/v3/examples"
	"github.com/qor5/x/v3/ui/tiptap"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

// @snippet_begin(TipTapLayoutSample)
func tiptapLayout(in web.PageFunc) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		ctx.Injector.HeadHTML(`
			<link rel="stylesheet" href="/assets/tiptap.css">
			<script src='/assets/vue.js'></script>
		`)

		ctx.Injector.TailHTML(`
<script src='/assets/tiptap.js'></script>
<script src='/assets/main.js'></script>
`)
		ctx.Injector.HeadHTML(`
		<style>
			[v-cloak] {
				display: none;
			}
		</style>
		`)

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}

		pr.Body = innerPr.Body

		return
	}
}

// @snippet_end

// @snippet_begin(DemoVuetifyLayoutSample)
func DemoVuetifyLayout(in web.PageFunc) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		ctx.Injector.HeadHTML(`
			<link rel="stylesheet" href="/vuetify/assets/index.css">
			<script src='/assets/vue.js'></script>
		`)

		ctx.Injector.TailHTML(`
			<script src='/assets/main.js'></script>
		`)
		ctx.Injector.HeadHTML(`
		<style>
			[v-cloak] {
				display: none;
			}
		</style>
		`)

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}

		pr.Body = VApp(
			VMain(
				innerPr.Body,
			),
		)

		return
	}
}

// @snippet_end
func Mux(mux *http.ServeMux) http.Handler {
	// @snippet_begin(ComponentsPackSample)
	mux.Handle("/assets/main.js",
		web.PacksHandler("text/javascript",
			vuetifyx.JSComponentsPack(),
			web.JSComponentsPack(),
		),
	)

	mux.Handle("/assets/vue.js",
		web.PacksHandler("text/javascript",
			web.JSVueComponentsPack(),
		),
	)

	// @snippet_end

	// @snippet_begin(TipTapComponentsPackSample)
	mux.Handle("/assets/tiptap.js",
		web.PacksHandler("text/javascript",
			tiptap.JSComponentsPack(),
		),
	)

	mux.Handle("/assets/tiptap.css",
		web.PacksHandler("text/css",
			tiptap.CSSComponentsPack(),
		),
	)
	// @snippet_end

	// @snippet_begin(VuetifyComponentsPackSample)
	vuetifyx.HandleMaterialDesignIcons("", mux)
	// @snippet_end

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write(assets.Favicon)
		return
	})

	return mux
}

func SamplesHandler(mux wexamples.Muxer) {
	mux.Handle(
		HelloVuetifyListPath,
		HelloVuetifyListPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		HelloVuetifyMenuPath,
		HelloVuetifyMenuPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifyGridPath,
		VuetifyGridPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifyBasicInputsPath,
		VuetifyBasicInputsPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifyVariantSubFormPath,
		VuetifyVariantSubFormPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifyComponentsKitchenPath,
		VuetifyComponentsKitchenPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifyNavigationDrawerPath,
		VuetifyNavigationDrawerPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		LazyPortalsAndReloadPath,
		LazyPortalsAndReloadPB.Wrap(DemoVuetifyLayout),
	)

	mux.Handle(
		VuetifySnackBarsPath,
		VuetifySnackBarsPB.Wrap(DemoVuetifyLayout),
	)

	return
}
