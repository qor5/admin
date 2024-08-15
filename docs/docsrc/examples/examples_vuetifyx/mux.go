package examples_vuetifyx

import (
	"net/http"

	"github.com/qor5/admin/v3/docs/docsrc/assets"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

func Mux(mux *http.ServeMux, prefix string) http.Handler {
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

	vuetifyx.HandleMaterialDesignIcons(prefix, mux)

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write(assets.Favicon)
		return
	})

	return mux
}

func SamplesHandler(mux examples.Muxer) {
	mux.Handle(
		VuetifyComponentsLinkageSelectPath,
		VuetifyComponentsLinkageSelectPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	mux.Handle(
		ExpansionPanelDemoPath,
		ExpansionPanelDemoPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	mux.Handle(
		KeyInfoDemoPath,
		KeyInfoDemoPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	mux.Handle(
		FilterDemoPath,
		FilterDemoPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	mux.Handle(
		DatePickersPath,
		DatePickersPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	mux.Handle(
		AutoCompleteDemoPath,
		AutoCompleteDemoPB.Wrap(examples_vuetify.DemoVuetifyLayout),
	)
	return
}
