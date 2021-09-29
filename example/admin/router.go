package admin

import (
	"net/http"
)

func Router() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	c := NewConfig()
	c.lb.Mount(mux)
	mux.Handle("/", c.lb.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		c.pb.ServeHTTP(w, r)
	}))
	return
}
