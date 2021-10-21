package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/qor/qor5/example/admin"
	"github.com/qor/qor5/sitemap"
)

func main() {
	mux := admin.Router()
	sitemap.SiteMap("product").RegisterRawString("https://dev.qor5.com/admin", "/product").MountTo(mux)
	robot := sitemap.Robots()
	robot.Agent(sitemap.AlexaAgent).Allow("/product1", "/product2").Disallow("/admin")
	robot.Agent(sitemap.GoogleAgent).Disallow("/admin")
	robot.MountTo(mux)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "9000"
	}
	fmt.Println("Served at http://localhost:" + port + "/admin")
	http.Handle("/",
		middleware.Logger(
			middleware.RequestID(mux),
		),
	)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
