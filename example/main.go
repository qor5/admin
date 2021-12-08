package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/qor/qor5/example/admin"
)

func main() {
	mux := admin.Router()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9000"
	}
	fmt.Println("Served at http://localhost:" + port + "/admin")
	http.Handle("/",
		middleware.Logger(
			middleware.RequestID(mux),
		),
	)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
