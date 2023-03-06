package admin

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/biter777/countries"
	"github.com/qor/oss"
	"github.com/qor/oss/filesystem"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/example/pages"
	"github.com/qor5/admin/l10n"
	l10n_view "github.com/qor5/admin/l10n/views"
	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	media_oss "github.com/qor5/admin/media/oss"
	media_view "github.com/qor5/admin/media/views"
	microsite_views "github.com/qor5/admin/microsite/views"
	"github.com/qor5/admin/note"
	"github.com/qor5/admin/pagebuilder"
	"github.com/qor5/admin/pagebuilder/example"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	publish_view "github.com/qor5/admin/publish/views"
	"github.com/qor5/admin/richeditor"
	"github.com/qor5/admin/role"
	"github.com/qor5/admin/slug"
	"github.com/qor5/admin/utils"
	"github.com/qor5/admin/worker"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/login"
	"github.com/qor5/x/perm"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

//go:embed assets
var assets embed.FS

var (
	// PublishStorage is used to storage static pages published by page builder.
	PublishStorage oss.StorageInterface = filesystem.New("publish")
)

type Config struct {
	pb          *presets.Builder
	pageBuilder *pagebuilder.Builder
}

func NewConfig() Config {
	db := ConnectDB()
	domain := os.Getenv("Site_Domain")
	sess := session.Must(session.NewSession())
	media_oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
		Session: sess,
	})
	PublishStorage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Publish_Bucket"),
		Region:  os.Getenv("S3_Publish_Region"),
		ACL:     s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Session: sess,
	})
	b := presets.New().RightDrawerWidth("700").VuetifyOptions(`
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
	b.BrandTitle("QOR5 Example").
		ProfileFunc(profile).
		NotificationFunc(notifierComponent(db), notifierCount(db)).
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.PageTitle = "Home"
			r.Body = vuetify.VContainer(
				h.H1("Home"),
				h.P().Text("Change your home page here"),
			)
			return
		}).
		NotFoundPageLayoutConfig(&presets.LayoutConfig{
			SearchBoxInvisible:          true,
			NotificationCenterInvisible: true,
		})
	perm.Verbose = true
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*"),
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:orders:*"),
			perm.PolicyFor("root").WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*"),
			perm.PolicyFor("viewer").WhoAre(perm.Denied).ToDo(presets.PermGet).On("*:products:*:price:"),
			perm.PolicyFor("viewer").WhoAre(perm.Denied).ToDo(presets.PermList).On("*:products:price:"),
			perm.PolicyFor("editor").WhoAre(perm.Denied).ToDo(presets.PermUpdate).On("*:products:*:price:"),
		).SubjectsFunc(func(r *http.Request) []string {
			u := getCurrentUser(r)
			if u == nil {
				return nil
			}
			return u.GetRoles()
		}).DBPolicy(perm.NewDBPolicy(db)),
	)

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN_ModelsI18nModuleKey).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, Messages_ja_JP_ModelsI18nModuleKey).
		RegisterForModule(language.English, I18nExampleKey, Messages_en_US).
		RegisterForModule(language.Japanese, I18nExampleKey, Messages_ja_JP).
		RegisterForModule(language.SimplifiedChinese, I18nExampleKey, Messages_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			//// Example:
			//user := getCurrentUser(r)
			//var supportedLanguages []language.Tag
			//for _, role := range user.GetRoles() {
			//	switch role {
			//	case "English Group":
			//		supportedLanguages = append(supportedLanguages, language.English)
			//	case "Chinese Group":
			//		supportedLanguages = append(supportedLanguages, language.SimplifiedChinese)
			//	}
			//}
			//return supportedLanguages
			return b.I18n().GetSupportLanguages()
		})

	l10nBuilder := l10n.New()
	l10nBuilder.
		RegisterLocales(countries.International, "International", "International").
		RegisterLocales(countries.China, "China", "China").
		RegisterLocales(countries.Japan, "Japan", "Japan").
		//RegisterLocales(countries.Russia, "Russia", "Russia").
		GetSupportLocalesFromRequestFunc(func(R *http.Request) []countries.CountryCode {
			return l10nBuilder.GetSupportLocales()[:]
		})

	utils.Configure(b)

	media_view.Configure(b, db)
	// media_view.MediaLibraryPerPage = 3
	// vips.UseVips(vips.Config{EnableGenerateWebp: true})
	ConfigureSeo(b, db)

	b.MenuOrder(
		b.MenuGroup("Page Builder").SubItems(
			"Page",
			"shared_containers",
			"demo_containers",
			"page_templates",
			"page_categories",
		).Icon("view_quilt"),
		b.MenuGroup("EC Management").SubItems(
			"ec-dashboard",
			"Order",
			"Product",
			"Category",
		).Icon("shopping_cart"),
		b.MenuGroup("Site Management").SubItems(
			"Setting",
			"QorSEOSetting",
		).Icon("settings"),
		b.MenuGroup("User Management").SubItems(
			"profile",
			"User",
			"Role",
		).Icon("group"),
		b.MenuGroup("Featured Models Management").SubItems(
			"InputHarness",
			"Post",
			"List Editor Example",
			"Customers",
			"ListModels",
			"MicrositeModels",
			"L10nModel",
			"L10nModelWithVersion",
		).Icon("featured_play_list"),
		"Worker",
		"ActivityLogs",
	)

	m := b.Model(&models.Post{})
	slug.Configure(b, m)

	mListing := m.Listing("ID", "Title", "TitleWithSlug", "HeroImage", "Body").
		SearchColumns("title", "body").
		PerPage(10)

	mListing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vuetifyx.FilterItem{
			{
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "posts", "Posts", u.ID, "Posts"),
			},
			{
				Key:          "created",
				Label:        "Create Time",
				ItemType:     vuetifyx.ItemTypeDatetimeRange,
				SQLCondition: `created_at %s ?`,
			},
			{
				Key:          "title",
				Label:        "Title",
				ItemType:     vuetifyx.ItemTypeString,
				SQLCondition: `title %s ?`,
			},
			{
				Key:      "status",
				Label:    "Status",
				ItemType: vuetifyx.ItemTypeSelect,
				Options: []*vuetifyx.SelectItem{
					{Text: publish.StatusDraft, Value: publish.StatusDraft},
					{Text: publish.StatusOnline, Value: publish.StatusOnline},
					{Text: publish.StatusOffline, Value: publish.StatusOffline},
				},
				SQLCondition: `status %s ?`,
			},
			{
				Key:      "multi_statuses",
				Label:    "Multiple Statuses",
				ItemType: vuetifyx.ItemTypeMultipleSelect,
				Options: []*vuetifyx.SelectItem{
					{Text: publish.StatusDraft, Value: publish.StatusDraft},
					{Text: publish.StatusOnline, Value: publish.StatusOnline},
					{Text: publish.StatusOffline, Value: publish.StatusOffline},
				},
				SQLCondition: `status %s ?`,
				Folded:       true,
			},
			{
				Key:          "id",
				Label:        "ID",
				ItemType:     vuetifyx.ItemTypeNumber,
				SQLCondition: `id %s ?`,
				Folded:       true,
			},
		}
	})

	mListing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsAll"),
				ID:    "all",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsHasUnreadNotes"),
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})

	w := worker.New(db)
	defer w.Listen()
	w.Configure(b)
	addJobs(w)

	ed := m.Editing("Status", "Schedule", "Title", "TitleWithSlug", "Seo", "HeroImage", "Body", "BodyImage")
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

	roleB := role.New(db).
		Resources([]*vuetify.DefaultOptionItem{
			{Text: "All", Value: "*"},
			{Text: "InputHarnesses", Value: "*:input_harnesses:*"},
			{Text: "Posts", Value: "*:posts:*"},
			{Text: "Settings", Value: "*:settings:*,*:site_management:"},
			{Text: "SEO", Value: "*:qor_seo_settings:*,*:site_management:"},
			{Text: "Customers", Value: "*:customers:*"},
			{Text: "Products", Value: "*:products:*,*:product_management:"},
			{Text: "Categories", Value: "*:categories:*,*:product_management:"},
			{Text: "Pages", Value: "*:pages:*,*:page_builder:"},
			{Text: "ListModels", Value: "*:list_models:*"},
			{Text: "ActivityLogs", Value: "*:activity_logs:*"},
			{Text: "Workers", Value: "*:workers:*"},
		})
	roleB.Configure(b)
	product := configProduct(b, db, w)
	category := configCategory(b, db)

	// Use m to customize the model, Or config more models here.

	type Setting struct{}
	sm := b.Model(&Setting{})
	sm.RegisterEventFunc(pages.LogInfoEvent, pages.LogInfo)
	sm.Listing().PageFunc(pages.Settings(db))

	// FIXME: list editor does not support use in page func
	// type ListEditorExample struct{}
	// leem := b.Model(&ListEditorExample{}).Label("List Editor Example")
	// pf, sf := pages.ListEditorExample(db, b)
	// leem.Listing().PageFunc(pf)
	// leem.RegisterEventFunc("save", sf)

	configCustomer(b, db)

	pageBuilder := example.ConfigPageBuilder(db, "/page_builder", ``, b.I18n())
	pm := pageBuilder.Configure(b, db)
	pmListing := pm.Listing()
	pmListing.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vuetifyx.FilterItem{
			{
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "page_builder_pages", "Pages", u.ID, "Pages"),
			},
		}
	})

	pmListing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsAll"),
				ID:    "all",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsHasUnreadNotes"),
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})

	tm := pageBuilder.ConfigTemplate(b, db)
	cm := pageBuilder.ConfigCategory(b, db)

	publisher := publish.New(db, PublishStorage).WithPageBuilder(pageBuilder)

	l := b.Model(&models.ListModel{})
	l.Listing("ID", "Title", "Status")
	l.Editing("Status", "Schedule", "Title")

	b.GetWebBuilder().RegisterEventFunc(noteMarkAllAsRead, markAllAsRead(db))

	note.Configure(db, b, m, pm)

	if err := db.AutoMigrate(&UserUnreadNote{}); err != nil {
		panic(err)
	}
	note.AfterCreateFunc = NoteAfterCreateFunc

	// @snippet_begin(ActivityExample)
	ab := activity.New(b, db).SetCreatorContextKey(login.UserKey).SetTabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		})
	_ = ab
	// ab.Model(m).UseDefaultTab()
	// ab.Model(pm).UseDefaultTab()
	// ab.Model(l).SkipDelete().SkipCreate()
	// @snippet_end
	ab.RegisterModel(m).UseDefaultTab()
	ab.RegisterModels(l, pm, tm, cm)
	mm := b.Model(&models.MicrositeModel{})
	mm.Listing("ID", "Name", "PrePath", "Status").
		SearchColumns("ID", "Name").
		PerPage(10)
	mm.Editing("Status", "Schedule", "Name", "Description", "PrePath", "FilesList", "Package")
	microsite_views.Configure(b, db, ab, media_oss.Storage, domain, publisher, mm)
	l10nM, l10nVM := configL10nModel(b)
	_ = l10nM
	publish_view.Configure(b, db, ab, publisher, m, l, pm, product, category, l10nVM)

	initLoginBuilder(db, b, ab)

	configInputHarness(b, db)

	configOrder(b, db)
	configECDashboard(b, db)

	configUser(b, db)
	configProfile(b, db)

	l10n_view.Configure(b, db, l10nBuilder, ab, pm, l10nM, l10nVM)

	return Config{
		pb:          b,
		pageBuilder: pageBuilder,
	}
}

func notifierCount(db *gorm.DB) func(ctx *web.EventContext) int {
	return func(ctx *web.EventContext) int {
		data, err := getUnreadNotesCount(ctx, db)
		if err != nil {
			return 0
		}

		a, b, c := data["Pages"], data["Posts"], data["Users"]
		return a + b + c
	}
}

func notifierComponent(db *gorm.DB) func(ctx *web.EventContext) h.HTMLComponent {
	return func(ctx *web.EventContext) h.HTMLComponent {
		data, err := getUnreadNotesCount(ctx, db)
		if err != nil {
			return nil
		}

		a, b, c := data["Pages"], data["Posts"], data["Users"]

		return vuetify.VList(
			vuetify.VListItem(
				vuetify.VListItemContent(
					vuetify.VListItemTitle(h.Text("Pages")),
					vuetify.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", a))),
				),
			).TwoLine(true).Href("/pages?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			vuetify.VListItem(
				vuetify.VListItemContent(
					vuetify.VListItemTitle(h.Text("Posts")),
					vuetify.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", b))),
				),
			).TwoLine(true).Href("/posts?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			vuetify.VListItem(
				vuetify.VListItemContent(
					vuetify.VListItemTitle(h.Text("Users")),
					vuetify.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", c))),
				),
			).TwoLine(true).Href("/users?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			h.If(a+b+c > 0,
				vuetify.VListItem(
					vuetify.VListItemContent(
						vuetify.VListItemSubtitle(h.Text("Mark all as read")),
					),
				).Attr("@click", web.Plaid().EventFunc(noteMarkAllAsRead).Go()),
			),
		).Class("mx-auto").Attr("max-width", "140")
	}
}
