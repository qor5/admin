package admin

import (
	"net/http"
)

func Router() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	c := NewConfig()
	c.lb.Mount(mux)
	mux.Handle("/", autheticate(c.lb)(c.pb))
	return
}
