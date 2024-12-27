package admin

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/sitemap"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/admin/v3/utils"
)

//go:embed assets/favicon.ico
var favicon []byte

const (
	exportOrdersURL = "/export-orders"
)

func TestHandlerComplex(db *gorm.DB, u *models.User, enableWork bool, opts ...ConfigOption) (http.Handler, Config) {
	mux := http.NewServeMux()
	c := NewConfig(db, enableWork, opts...)
	if u == nil {
		u = &models.User{
			Model: gorm.Model{ID: 888},
			Roles: []role.Role{
				{
					Name: models.RoleAdmin,
				},
			},
		}
	}
	m := login.MockCurrentUser(u)
	mux.Handle("/page_builder/", m(c.pageBuilder))
	mux.Handle("/", m(c.pb))
	return mux, c
}

func TestHandler(db *gorm.DB, u *models.User) http.Handler {
	mux, _ := TestHandlerComplex(db, u, false)
	return mux
}

func TestHandlerWorker(db *gorm.DB, u *models.User) http.Handler {
	mux, _ := TestHandlerComplex(db, u, true)
	return mux
}

func TestL18nHandler(db *gorm.DB) (http.Handler, Config) {
	mux := http.NewServeMux()
	c := NewConfig(db, false)
	c.loginSessionBuilder.Secret("test")
	c.loginSessionBuilder.Mount(mux)
	mux.Handle("/", c.pb)
	return mux, c
}

// a reverse proxy for development
func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}

func Router(db *gorm.DB) http.Handler {
	c := NewConfig(db, true)

	mux := http.NewServeMux()
	c.loginSessionBuilder.Mount(mux)
	//	mux.Handle("/frontstyle.css", c.pb.GetWebBuilder().PacksHandler("text/css", web.ComponentsPack(`
	// :host {
	//	all: initial;
	//	display: block;
	// div {
	//	background-color:orange;
	// }
	// `)))
	// example of seo
	mux.Handle("/posts/first", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		db.First(&post)
		seodata, _ := seoBuilder.Render(post, r).MarshalHTML(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>%s</head><body>%s</body></html>`, seodata, post.Body)
	}))

	mux.Handle("/page_builder/", c.pageBuilder)

	// email_builder  start
	if utils.IsDevelopment() {

		// development mode：reverse proxy to Vite http server
		proxy, err := NewReverseProxy("http://localhost:3000/email_builder/")
		if err != nil {
			fmt.Printf("%s", fmt.Sprintf("Failed to create reverse proxy: %v", err))
		}

		mux.Handle("/email_builder/", http.StripPrefix("/email_builder", proxy))
		mux.Handle("/email_builder", http.RedirectHandler("/email_builder/", http.StatusMovedPermanently))
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		distPath := filepath.Join(currentDir, "../emailbuilder/dist")
		absDistPath, err := filepath.Abs(distPath)
		if err != nil {
			panic(err)
		}
		if _, err := os.Stat(absDistPath); os.IsNotExist(err) {
			panic("Dist directory not found: " + absDistPath)
		}

		// for /email_builder/* fallback is index.html
		mux.HandleFunc("/email_builder/", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/email_builder/")
			if path == "" {
				path = "index.html"
			}
			if strings.Contains(path, "..") || strings.Contains(path, "\\") {
				http.Error(w, "Invalid file name", http.StatusBadRequest)
				return
			}
			filePath := filepath.Join(absDistPath, path)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(absDistPath, "index.html"))
				return
			}
			http.ServeFile(w, r, filePath)
		})

		// for exactly /email_builder
		mux.Handle("/email_builder", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(absDistPath, "index.html"))
		}))
	}
	// email_builder register end

	mux.Handle("/", c.pb)
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Write(favicon)
		return
	})

	mux.Handle(exportOrdersURL, exportOrders(db))

	// example of sitemap and robot
	sitemap.SiteMap("product").RegisterRawString("https://dev.qor5.com/admin", "/product").MountTo(mux)
	robot := sitemap.Robots()
	robot.Agent(sitemap.AlexaAgent).Allow("/product1", "/product2").Disallow("/admin")
	robot.Agent(sitemap.GoogleAgent).Disallow("/admin")
	robot.MountTo(mux)

	cr := chi.NewRouter()
	cr.Use(
		c.loginSessionBuilder.Middleware(),
		withRoles(db),
		securityMiddleware(),
	)
	cr.Mount("/", mux)
	return cr
}
