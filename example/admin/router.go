package admin

import (
	"net/http"
)

func Router() (mux *http.ServeMux) {
	mux = http.NewServeMux()
	b := NewConfig()
	mux.Handle("/", b)
	return
}
