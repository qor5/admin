package admin

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/admin/example/models"
	"github.com/qor5/web"
	"github.com/qor5/x/sitemap"
)

//go:embed assets/favicon.ico
var favicon []byte

const (
	logoutURL                  = "/auth/logout"
	oauthCompleteInfoPageURL   = "/auth/complete-info"
	oauthCompleteInfoActionURL = "/auth/do-complete-info"

	exportOrdersURL = "/export-orders"
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
	// div {
	//	background-color:orange;
	// }
	// `)))

	mux.Handle("/page_builder/", c.pageBuilder)
	// example of seo
	mux.Handle("/posts/first", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		db.First(&post)
		seodata, _ := SeoCollection.Render(post, r).MarshalHTML(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>%s</head><body>%s</body></html>`, seodata, post.Body)
	}))

	mux.Handle("/", c.pb)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write(favicon)
		return
	})

	mux.Handle(oauthCompleteInfoActionURL, doOAuthCompleteInfo(db))
	mux.Handle(oauthCompleteInfoPageURL, c.pb.I18n().EnsureLanguage(web.New().Page(oauthCompleteInfoPage(vh, c.pb))))

	mux.Handle(exportOrdersURL, exportOrders(db))

	// example of sitemap and robot
	sitemap.SiteMap("product").RegisterRawString("https://dev.qor5.com/admin", "/product").MountTo(mux)
	robot := sitemap.Robots()
	robot.Agent(sitemap.AlexaAgent).Allow("/product1", "/product2").Disallow("/admin")
	robot.Agent(sitemap.GoogleAgent).Disallow("/admin")
	robot.MountTo(mux)

	cr := chi.NewRouter()
	cr.Use(
		loginBuilder.Middleware(),
		validateSessionToken(),
		isOAuthInfoCompleted(),
		withRoles(db),
		withNoteContext(),
		securityMiddleware(),
	)
	cr.Mount("/", mux)
	return cr
}
