package main

import (
	"fmt"
	"net/http"

	"github.com/qor5/admin/v3/cmd/qor5/admin-template/admin"
	"github.com/theplant/osenv"
)

func main() {
	// Setup project
	mux := admin.Initialize()

	port := osenv.Get("PORT", "The port to serve the admin on", "9000")

	fmt.Println("Served at http://localhost:" + port + "/admin")

	http.Handle("/", mux)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
