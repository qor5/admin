package admin

import (
	"net/http"

	"github.com/goplaid/web"
)

func Router() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	c := NewConfig()
	c.lb.Mount(mux)
	mux.Handle("/frontstyle.css", c.pb.GetWebBuilder().PacksHandler("text/css", web.ComponentsPack(`
:host {
	all: initial;
	display: block;
}
div {
	background-color:orange;
}
`)))

	mux.Handle("/admin/page_builder/", autheticate(c.lb)(c.pageBuilder))
	mux.Handle("/", autheticate(c.lb)(c.pb))

	return
}
