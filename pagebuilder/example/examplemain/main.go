package main

import (
	"log"
	"net/http"

	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	"github.com/qor/qor5/pagebuilder/example"
)

func main() {
	db := example.ConnectDB()

	p := presets.New().
		URIPrefix("/admin").
		DataOperator(gorm2op.DataOperator(db))
	pb := example.ConfigPageBuilder(db)
	pb.Configure(p)

	mux := http.NewServeMux()
	mux.Handle("/admin/", p)
	mux.Handle("/page_builder/", pb)
	log.Println("Listen on http://localhost:9600")
	log.Fatal(http.ListenAndServe(":9600", mux))
}
