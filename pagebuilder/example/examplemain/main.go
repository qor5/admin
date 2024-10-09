package main

import (
	"log"
	"net/http"

	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func main() {
	db := example.ConnectDB()

	p := presets.New().
		URIPrefix("/admin").
		DataOperator(gorm2op.DataOperator(db))
	pb := example.ConfigPageBuilder(db, "/page_builder", `<link rel="stylesheet" href="/frontstyle.css">`, p)

	pb.Install(p)

	mux := http.NewServeMux()

	mux.Handle("/frontstyle.css", p.GetWebBuilder().PacksHandler("text/css", web.ComponentsPack(`
:host {
	all: initial;
	display: block;
}
div {
	background-color:orange;
}
`)))
	mux.Handle("/admin", p)
	mux.Handle("/admin/", p)
	mux.Handle("/page_builder", pb)
	mux.Handle("/page_builder/", pb)
	log.Println("Listen on http://localhost:9600")
	log.Fatal(http.ListenAndServe(":9600", mux))
}
