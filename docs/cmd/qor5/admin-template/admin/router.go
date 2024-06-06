package admin

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
)

func setupRouter(b *presets.Builder) (mux *http.ServeMux) {
	mux = http.NewServeMux()
	mux.Handle("/", b)
	return
}
