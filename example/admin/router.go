package admin

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/sitemap"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/role"
)

//go:embed assets/favicon.ico
var favicon []byte

const (
	exportOrdersURL = "/export-orders"
)

func TestHandlerComplex(db *gorm.DB, u *models.User, enableWork bool) (http.Handler, Config) {
	mux := http.NewServeMux()
	c := NewConfig(db, enableWork)
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

	mux.Handle("/system/media_libraries/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/system/media_libraries/", http.FileServer(http.Dir("public/system/media_libraries"))).ServeHTTP(w, r)
	}))

	mux.Handle("/posts/first", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var post models.Post
		db.First(&post)
		seodata, _ := seoBuilder.Render(post, r).MarshalHTML(r.Context())
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>%s</head><body>%s</body></html>`, seodata, post.Body)
	}))

	mux.Handle("/page_builder/", c.pageBuilder)

	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 构建 dist 目录的绝对路径
	distPath := filepath.Join(currentDir, "../../react-project/dist")

	// 获取绝对路径
	absDistPath, err := filepath.Abs(distPath)
	if err != nil {
		panic(err)
	}

	// 检查目录是否存在
	if _, err := os.Stat(absDistPath); os.IsNotExist(err) {
		panic("Dist directory not found: " + absDistPath)
	}

	// 创建文件服务器
	fs := http.FileServer(http.Dir(absDistPath))

	// 处理 /new_page_builder 路由
	mux.HandleFunc("/new_page_builder", func(w http.ResponseWriter, r *http.Request) {
		// 如果请求的是根路径，则提供 index.html
		if r.URL.Path == "/new_page_builder" {
			http.ServeFile(w, r, filepath.Join(absDistPath, "index.html"))
			return
		}
		// 否则，将请求传递给文件服务器
		http.StripPrefix("/new_page_builder", fs).ServeHTTP(w, r)
	})

	// 处理所有其他静态文件
	mux.Handle("/new_page_builder/", http.StripPrefix("/new_page_builder/", fs))

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
