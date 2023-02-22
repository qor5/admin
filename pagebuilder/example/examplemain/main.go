package main

import (
	"log"
	"net/http"

	"github.com/qor5/admin/pagebuilder/example"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/web"
)

func main() {
	db := example.ConnectDB()

	p := presets.New().
		URIPrefix("/admin").
		DataOperator(gorm2op.DataOperator(db))
	pb := example.ConfigPageBuilder(db, "/page_builder", `<link rel="stylesheet" href="/frontstyle.css">`, p.I18n())

	pb.Configure(p, db, nil, nil)

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
	mux.Handle("/admin/", p)
	mux.Handle("/page_builder/", pb)
	log.Println("Listen on http://localhost:9600")
	log.Fatal(http.ListenAndServe(":9600", mux))
}
