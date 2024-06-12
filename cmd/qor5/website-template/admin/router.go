package admin

import (
	"net/http"

	"gorm.io/gorm"
)

func Router(db *gorm.DB) (mux *http.ServeMux) {
	c := newConfig(db)
	mux = http.NewServeMux()
	mux.Handle("/", c.pb)
	mux.Handle("/admin/page_builder/", c.pageBuilder)

	return
}
