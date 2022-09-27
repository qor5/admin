package admin

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/sitemap"
)

func Router() http.Handler {
	db := ConnectDB()
	c := NewConfig()

	mux := http.NewServeMux()
	loginBuilder.Mount(mux)
	//	mux.Handle("/frontstyle.css", c.pb.GetWebBuilder().PacksHandler("text/css", web.ComponentsPack(`
	// :host {
	//	all: initial;
	//	display: block;
	//div {
	//	background-color:orange;
	//}
	//`)))

	mux.Handle("/admin/page_builder/", c.pageBuilder)
	// example of seo
	mux.Handle("/posts/first", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		db.First(&post)
		seodata, _ := SeoCollection.Render(post, r).MarshalHTML(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>%s</head><body>%s</body></html>`, seodata, post.Body)
	}))
	mux.Handle("/", c.pb)

	// example of sitemap and robot
	sitemap.SiteMap("product").RegisterRawString("https://dev.qor5.com/admin", "/product").MountTo(mux)
	robot := sitemap.Robots()
	robot.Agent(sitemap.AlexaAgent).Allow("/product1", "/product2").Disallow("/admin")
	robot.Agent(sitemap.GoogleAgent).Disallow("/admin")
	robot.MountTo(mux)

	cr := chi.NewRouter()
	cr.Use(
		login.Authenticate(loginBuilder),
		withRoles(db),
		withNoteContext(),
		withTokenAuth(),
	)
	cr.Mount("/", mux)
	return cr
}
