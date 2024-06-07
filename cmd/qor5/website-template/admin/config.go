package admin

import (
	"net/http"

	"github.com/qor/oss/filesystem"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const (
	PublishDir = "./publish"
)

type Config struct {
	pb          *presets.Builder
	pageBuilder *pagebuilder.Builder
}

func InitApp() *http.ServeMux {
	c := newPB()
	mux := SetupRouter(c)

	return mux
}

func newPB() Config {
	db := ConnectDB()

	b := presets.New()

	b.URIPrefix("/admin").DataOperator(gorm2op.DataOperator(db)).
		BrandFunc(func(ctx *web.EventContext) h.HTMLComponent {
			return vuetify.VContainer(
				h.Img(logo).Attr("width", "150"),
			).Class("ma-n4")
		}).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.Body = vuetify.VContainer(
				h.H1("Home"),
				h.P().Text("Change your home page here"))
			return
		})

	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*"),
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On("*:activity_logs:*"),
		),
	)

	utils.Install(b)
	mediab := media.New(db)
	ab := activity.New(db).CreatorContextKey(login.UserKey)

	pageBuilder := example.ConfigPageBuilder(db, "/admin/page_builder", ``, b.I18n())
	storage := filesystem.New(PublishDir)
	publisher := publish.New(db, storage)

	seoBuilder := seo.New(db)
	l10nBuilder := l10n.New(db).Activity(ab)
	pageBuilder.SEO(seoBuilder).Publisher(publisher).L10n(l10nBuilder).Activity(ab)

	l10nBuilder.
		RegisterLocales("International", "International", "International").
		RegisterLocales("China", "China", "China").
		SupportLocalesFunc(func(R *http.Request) []string {
			return l10nBuilder.GetSupportLocaleCodes()[:]
		})

	b.Use(
		mediab,
		ab,
		pageBuilder,
		seoBuilder,
		l10nBuilder,
		publisher,
	)

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.English, I18nExampleKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nExampleKey, Messages_zh_CN).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN_ModelsI18nModuleKey).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			return b.I18n().GetSupportLanguages()
		})

	b.MenuOrder(
		b.MenuGroup("Page Builder").SubItems("pages", "page_templates", "page_categories").Icon("mdi-web"),
		"shared_containers",
		"demo_containers",
		"media-library",
	)

	initMediaLibraryData(db)
	initWebsiteData(db)

	return Config{
		pb:          b,
		pageBuilder: pageBuilder,
	}
}

var logo = `data:image/png;base64,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`
