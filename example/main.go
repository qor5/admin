package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/v3/example/admin"
	"github.com/theplant/osenv"
)

func main() {
	h := admin.Router(admin.ConnectDB())

	host := osenv.Get("HOST", "The host to serve the admin on", "")
	port := osenv.Get("PORT", "The port to serve the admin on", "9000")

	fmt.Println("Served at http://localhost:" + port)

	mux := http.NewServeMux()
	mux.Handle("/",
		middleware.RequestID(
			middleware.Logger(
				middleware.Recoverer(h),
			),
		),
	)
	err := http.ListenAndServe(host+":"+port, mux)
	if err != nil {
		panic(err)
	}
}
