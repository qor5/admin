package admin

import (
	"embed"
	"os"

	"github.com/qor/qor5/utils"

	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/example/pages"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/media/oss"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/pagebuilder"
	"github.com/qor/qor5/pagebuilder/example"
	"github.com/qor/qor5/richeditor"
	"github.com/qor/qor5/slug"
	"github.com/qor/qor5/worker"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

//go:embed assets
var assets embed.FS

type Config struct {
	pb          *presets.Builder
	lb          *login.Builder
	pageBuilder *pagebuilder.Builder
}

func NewConfig() Config {
	db := ConnectDB()

	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: sess,
	})

	media.RegisterCallbacks(db)

	b := presets.New().RightDrawerWidth(700).VuetifyOptions(`
{
  icons: {
	iconfont: 'md', // 'mdi' || 'mdiSvg' || 'md' || 'fa' || 'fa4'
  },
  theme: {
    themes: {
      light: {
		  primary: "#673ab7",
		  secondary: "#009688",
		  accent: "#ff5722",
		  error: "#f44336",
		  warning: "#ff9800",
		  info: "#8bc34a",
		  success: "#4caf50"
      },
    },
  },
}
`)
	js, _ := assets.ReadFile("assets/fontcolor.min.js")
	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert", "fontcolor"}
	richeditor.PluginsJS = [][]byte{js}
	b.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	b.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
	b.URIPrefix("/admin").
		BrandTitle("example").
		ProfileFunc(profile).
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.PageTitle = "Home"
			r.Body = vuetify.VContainer(
				h.H1("Home"),
				h.P().Text("Change your home page here"),
			)
			return
		})

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN)
	utils.Configure(b)

	media_view.Configure(b, db)
	//media_view.MediaLibraryPerPage = 3
	models.ConfigureSeo(b, db)

	b.MenuOrder(
		"InputHarness",
		"Post",
		"User",
		b.MenuGroup("Site Management").SubItems(
			"Setting",
			"QorSEOSetting",
		).Icon("settings"),
	)

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
	configUser(b, db)

	_ = m
	// Use m to customize the model, Or config more models here.

	type Setting struct{}
	b.Model(&Setting{}).Listing().PageFunc(pages.Settings(db))

	pageBuilder := example.ConfigPageBuilder(db)
	publisher := publish.New(db, oss.Storage).WithValue("pagebuilder", pageBuilder)
	publish_view.Configure(b, db, publisher)
	publish_view.RegisterPublishModels(&pagebuilder.Page{})
	pageBuilder.
		PageStyle(h.RawHTML(`<link rel="stylesheet" href="/frontstyle.css">`)).
		Prefix("/admin/page_builder")
	pageBuilder.Configure(b)

	w := worker.New(db)
	addJobs(w)
	w.Configure(b)

	return Config{
		pb:          b,
		lb:          newLoginBuilder(db),
		pageBuilder: pageBuilder,
	}
}
