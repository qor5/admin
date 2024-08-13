package main

import (
	"fmt"
	"net/http"

	"github.com/qor5/admin/v3/autocomplete"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	"github.com/theplant/osenv"
)

var (
	host = osenv.Get("HOST", "The host to serve the admin on", "")
	port = osenv.Get("PORT", "The port to serve on", "7800")
)

func main() {
	addr := host + ":" + port
	fmt.Println("Starting docs at http://" + addr)

	mux := http.NewServeMux()
	examples_vuetify.Mux(mux)

	im := &examples.IndexMux{Mux: http.NewServeMux()}
	ab := autocomplete.New().Prefix("/complete")
	examples_admin.SamplesHandler(im, ab)
	mux.Handle("/complete/", ab)
	mux.Handle("/examples/",
		middleware.Logger(
			middleware.RequestID(
				im.Mux,
			),
		),
	)
	mux.Handle("/", web.New().Page(im.Page))

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
