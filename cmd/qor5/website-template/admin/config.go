package admin

import (
	"context"
	"embed"
	"encoding/base64"
	"net/http"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/oss/filesystem"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/admin/v3/utils"
)

var PublishDir = osenv.Get(
	"PUBLISH_DIR",
	"The dir that static files published to",
	"/tmp/qor5_publish",
)

type config struct {
	pb          *presets.Builder
	pageBuilder *pagebuilder.Builder
}

func newConfig(db *gorm.DB) config {
	err := db.AutoMigrate(&MyHeader{})
	if err != nil {
		panic(err)
	}

	b := presets.New()

	b.URIPrefix("/admin").DataOperator(gorm2op.DataOperator(db)).
		BrandFunc(func(ctx *web.EventContext) HTMLComponent {
			return vuetify.VContainer(
				Img(logo).Attr("width", "150"),
			).Class("ma-n4")
		}).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = vuetify.VContainer(
				H1("Home"),
				P().Text("Change your home page here"))
			return
		})

	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*"),
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On("*:activity_logs:*"),
		),
	)

	utils.Install(b)

	storage := filesystem.New(PublishDir)

	mediaBuilder := media.New(db)
	ab := activity.New(db, func(ctx context.Context) (*activity.User, error) {
		return &activity.User{
			ID:   "1",
			Name: "John",
		}, nil
	}).AutoMigrate()
	publisher := publish.New(db, storage)
	seoBuilder := seo.New(db).AutoMigrate()

	pageBuilder := pagebuilder.New(b.GetURIPrefix()+"/page_builder", db, b).
		AutoMigrate().
		Publisher(publisher)

	header := pageBuilder.RegisterContainer("MyHeader").Group("Navigation").
		RenderFunc(func(obj interface{}, input *pagebuilder.RenderInput, ctx *web.EventContext) HTMLComponent {
			c := obj.(*MyHeader)
			u := Ul()
			for _, item := range c.MenuItems {
				u.AppendChildren(Li(
					A().Href(item.Link).Text(item.Text),
				))
			}
			return u
		})

	ed := header.Model(&MyHeader{}).Editing("MenuItems")

	menuItemFb := b.NewFieldsBuilder(presets.WRITE).Model(&MenuItem{}).Only("Text", "Link")

	ed.Field("MenuItems").Nested(menuItemFb)

	pageBuilder.SEO(seoBuilder).Publisher(publisher).Activity(ab)

	b.Use(
		mediaBuilder,
		ab,
		pageBuilder,
		seoBuilder,
		publisher,
	)

	b.GetI18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.English, I18nExampleKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nExampleKey, Messages_zh_CN).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, MessagesModels_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			return b.GetI18n().GetSupportLanguages()
		})

	b.MenuOrder(
		b.MenuGroup("Page Builder").
			SubItems(
				"pages",
				"page_templates",
				"page_categories",
				"shared_containers",
				"demo_containers",
			).
			Icon("mdi-web"),
		"media-library",
	)

	// initMediaLibraryData(db)
	// initWebsiteData(db)

	return config{
		pb:          b,
		pageBuilder: pageBuilder,
	}
}

//go:embed qor-logo.png
var logoFs embed.FS
var logo = readLogo()

func readLogo() string {
	b, err := logoFs.ReadFile("qor-logo.png")
	if err != nil {
		panic(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
}
