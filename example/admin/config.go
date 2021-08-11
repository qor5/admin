package admin

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gormop"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/media/media_library"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/media_library_view"
	h "github.com/theplant/htmlgo"
)

func NewConfig() (b *presets.Builder) {
	db := ConnectDB()

	b = presets.New()
	b.URIPrefix("/admin").
		BrandTitle("example").
		DataOperator(gormop.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = vuetify.VContainer(
				h.H1("Home"),
				h.P().Text("Change your home page here"))
			return
		})

	b.FieldDefaults(presets.WRITE).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(media_library_view.MediaBoxComponentFunc).
		SetterFunc(media_library_view.MediaBoxSetterFunc)

	m := b.Model(&models.Post{})
	ed := m.Editing("Title", "HeroImage", "Body", "BodyImage")
	ed.Field("HeroImage").
		WithContextValue(
			media_library_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{AllowType: "image"})
	ed.Field("BodyImage").
		WithContextValue(
			media_library_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{AllowType: "image"})
	_ = m
	// Use m to customize the model, Or config more models here.
	return
}
