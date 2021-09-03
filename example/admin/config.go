package admin

import (
	"embed"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gormop"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/media/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/cropper"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/example/pages"
	"github.com/qor/qor5/media_library_view"
	"github.com/qor/qor5/richeditor"
	h "github.com/theplant/htmlgo"
)

//go:embed assets
var assets embed.FS

func NewConfig() (b *presets.Builder) {
	db := ConnectDB()

	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  "test-juice",
		Region:  "ap-northeast-1",
		Session: sess,
	})

	media.RegisterCallbacks(db)

	b = presets.New().RightDrawerWidth(700)
	js, _ := assets.ReadFile("assets/fontcolor.min.js")
	richeditor.Plugins = []string{"alignment", "video", "fontcolor", "imageinsert"}
	richeditor.PluginsJS = [][]byte{js}
	b.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	b.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
	b.ExtraAsset("/cropper.js", "text/javascript", cropper.JSComponentsPack())
	b.ExtraAsset("/cropper.css", "text/css", cropper.CSSComponentsPack())
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
		ComponentFunc(media_library_view.MediaBoxComponentFunc(db)).
		SetterFunc(media_library_view.MediaBoxSetterFunc(db))

	b.FieldDefaults(presets.LIST).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(media_library_view.MediaBoxListFunc())
	//media_library_view.MediaLibraryPerPage = 3

	m := b.Model(&models.Post{})
	m.Listing("ID", "Title", "HeroImage", "Body").
		SearchColumns("title", "body").
		PerPage(10)
	ed := m.Editing("Title", "HeroImage", "Body", "BodyImage")
	ed.Field("HeroImage").
		WithContextValue(
			media_library_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*media.Size{
					"thumb": {
						Width:  400,
						Height: 300,
					},
					"main": {
						Width:  800,
						Height: 500,
					},
				},
			})
	ed.Field("BodyImage").
		WithContextValue(
			media_library_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{AllowType: "image"})

	ed.Field("Body").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Body", obj.(*models.Post).Body, field.Label, "")
	})
	_ = m
	// Use m to customize the model, Or config more models here.

	type Setting struct{}
	b.Model(&Setting{}).Listing().PageFunc(pages.Settings(db))

	return
}
