package admin

import (
	"embed"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gormop"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/cropper"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/example/pages"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/media/oss"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/richeditor"
	"github.com/qor/qor5/slug"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

//go:embed assets
var assets embed.FS

func NewConfig() (b *presets.Builder) {
	db := ConnectDB()

	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: sess,
	})

	media.RegisterCallbacks(db)

	b = presets.New().RightDrawerWidth(700)
	js, _ := assets.ReadFile("assets/fontcolor.min.js")
	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert", "fontcolor"}
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

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN)

	media_view.Configure(b, db)
	//media_view.MediaLibraryPerPage = 3
	models.ConfigureSeo(b, db)

	m := b.Model(&models.Post{})
	m.Listing("ID", "Title", "TitleWithSlug", "HeroImage", "Body").
		SearchColumns("title", "body").
		PerPage(10)
	ed := m.Editing("Title", "TitleWithSlug", "Seo", "HeroImage", "Body", "BodyImage")
	ed.Field("HeroImage").
		WithContextValue(
			media_view.MediaBoxConfig,
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
			media_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{})

	ed.Field("Body").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Body").Plugins([]string{"alignment", "video", "imageinsert", "fontcolor"}).Value(obj.(*models.Post).Body).Label(field.Label)
	})

	ed.Field("Title").ComponentFunc(slug.SlugEditingComponentFunc)
	ed.Field("TitleWithSlug").SetterFunc(slug.SlugEditingSetterFunc).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) { return })
	configInputHarness(b, db)

	_ = m
	// Use m to customize the model, Or config more models here.

	type Setting struct{}
	b.Model(&Setting{}).Listing().PageFunc(pages.Settings(db))
	return
}
