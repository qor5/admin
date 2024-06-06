package examples_admin

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
)

func Mux(mux *http.ServeMux, prefix string) http.Handler {
	examples_vuetify.Mux(mux, prefix)

	im := &examples_vuetify.IndexMux{Mux: http.NewServeMux()}
	SamplesHandler(im, prefix)

	mux.Handle("/samples/",
		middleware.Logger(
			middleware.RequestID(
				im.Mux,
			),
		),
	)

	return mux
}

func SamplesHandler(mux examples.Muxer, prefix string) {
	examples_vuetify.SamplesHandler(mux, prefix)
	examples_presets.SamplesHandler(mux, prefix)

	examples.AddPresetExample(mux, ListingExample)
	examples.AddPresetExample(mux, WorkerExample)
	examples.AddPresetExample(mux, ActionWorkerExample)
	examples.AddPresetExample(mux, InternationalizationExample)
	examples.AddPresetExample(mux, LocalizationExample)
	examples.AddPresetExample(mux, PublishExample)
	examples.AddPresetExample(mux, SEOExampleBasic)
	examples.AddPresetExample(mux, ActivityExample)
	examples.AddPresetExample(mux, PageBuilderExample)
}
