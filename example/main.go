package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/example/admin"
)

func main() {
	h := admin.Router()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9000"
	}
	fmt.Println("Served at http://localhost:" + port)

	mux := http.NewServeMux()
	mux.Handle("/",
		middleware.RequestID(
			middleware.Logger(
				middleware.Recoverer(h),
			),
		),
	)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
