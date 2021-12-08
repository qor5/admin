package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/qor/qor5/docsrc"
	"github.com/theplant/docgo"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8800"
	}
	fmt.Println("Starting docs at :" + port)
	http.Handle("/", docgo.New().
		Assets("/assets/", docsrc.Assets).
		Home(docsrc.Home).
		Build(),
	)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
