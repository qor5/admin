package admin

import (
	"net/http"
)

func SetupRouter(c Config) (mux *http.ServeMux) {
	mux = http.NewServeMux()
	mux.Handle("/", c.pb)
	mux.Handle("/admin/page_builder/", c.pageBuilder)

	return
}
