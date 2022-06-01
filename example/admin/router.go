package admin

import (
	"fmt"
	"net/http"

	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/sitemap"
)

func Router() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	c := NewConfig()
	c.lb.Mount(mux)
	//	mux.Handle("/frontstyle.css", c.pb.GetWebBuilder().PacksHandler("text/css", web.ComponentsPack(`
	//:host {
	//	all: initial;
	//	display: block;
	//}
	//div {
	//	background-color:orange;
	//}
	//`)))

	mux.Handle("/admin/page_builder/", authenticate(c.lb)(c.pageBuilder))
	// example of seo
	mux.Handle("/posts/first", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		db.First(&post)
		seodata, _ := SeoCollection.Render(post, r).MarshalHTML(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>%s</head><body>%s</body></html>`, seodata, post.Body)
	}))
	mux.Handle("/", authenticate(c.lb)(c.pb))

	// example of sitemap and robot
	sitemap.SiteMap("product").RegisterRawString("https://dev.qor5.com/admin", "/product").MountTo(mux)
	robot := sitemap.Robots()
	robot.Agent(sitemap.AlexaAgent).Allow("/product1", "/product2").Disallow("/admin")
	robot.Agent(sitemap.GoogleAgent).Disallow("/admin")
	robot.MountTo(mux)

	return
}
