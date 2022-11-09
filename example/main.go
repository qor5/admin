package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/qor5/admin/example/admin"
)

func main() {
	mux := admin.Router()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9000"
	}
	fmt.Println("Served at http://localhost:" + port + "/admin")
	http.Handle("/",
		middleware.RequestID(
			middleware.Logger(
				middleware.Recoverer(mux),
			),
		),
	)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
