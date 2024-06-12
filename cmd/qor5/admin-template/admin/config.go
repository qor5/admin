package admin

import (
	"net/http"

	"github.com/qor5/admin/v3/cmd/qor5/admin-template/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
)

func Initialize() *http.ServeMux {
	b := setupAdmin()
	mux := setupRouter(b)

	return mux
}

func setupAdmin() (b *presets.Builder) {
	db := ConnectDB()

	// Initialize the builder of QOR5
	b = presets.New()

	// Set up the project name, ORM and Homepage
	b.URIPrefix("/admin").
		BrandTitle("Admin").
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = VContainer(
				H1("Home"),
				P().Text("Change your home page here"))
			return
		})

	// Register Post into the builder
	// Use m to customize the model, Or config more models here.
	m := b.Model(&models.Post{})
	_ = m

	return
}
