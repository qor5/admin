package main

import (
	"fmt"
	"net/http"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/emailbuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func main() {
	withCORS := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	db := emailbuilder.ConnectDB()
	err := db.AutoMigrate(&emailbuilder.EmailTemplate{})
	if err != nil {
		panic(err)
	}
	b := presets.New()
	b.URIPrefix("/").
		BrandTitle("Admin").
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(_ *web.EventContext) (r web.PageResponse, err error) {
			r.Body = vuetify.VContainer(
				h.H1("Home"),
				h.P().Text("Change your home page here"))
			return
		})

	eb := emailbuilder.New(b, db, emailbuilder.DefaultMailTemplate(b)).AutoMigrate()
	emailbuilder.DefaultMailCampaign(b, db).Use(eb)
	b.Use(eb)
	mux := http.NewServeMux()
	mux.Handle("/", b)
	mux.Handle("/email_template/", http.StripPrefix("/email_template", eb))
	fmt.Println("Listen on http://localhost:9800")
	if err := http.ListenAndServe(":9800", withCORS(mux)); err != nil {
		panic(err)
	}
}
