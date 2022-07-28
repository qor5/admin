package admin

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/gorm2op"
	"github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
	"github.com/qor/oss/s3"
	"github.com/qor/qor5/activity"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/example/pages"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/media/oss"
	media_view "github.com/qor/qor5/media/views"
	microsite_views "github.com/qor/qor5/microsite/views"
	"github.com/qor/qor5/note"
	"github.com/qor/qor5/pagebuilder"
	"github.com/qor/qor5/pagebuilder/example"
	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"
	"github.com/qor/qor5/richeditor"
	"github.com/qor/qor5/role"
	"github.com/qor/qor5/slug"
	"github.com/qor/qor5/utils"
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
	domain := os.Getenv("Site_Domain")

	sess := session.Must(session.NewSession())

	oss.Storage = s3.New(&s3.Config{
		Bucket:  os.Getenv("S3_Bucket"),
		Region:  os.Getenv("S3_Region"),
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
	b.URIPrefix("/admin").
		BrandTitle("QOR5 Example").
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
	// perm.Verbose = true
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete, presets.PermGet, presets.PermList).On("*:roles:*", "*:users:*"),
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
		}).EnableDBPolicy(db, perm.DefaultDBPolicy{}, time.Minute),
	)

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			u := getCurrentUser(r)
			if u.Name == "中文" {
				return b.I18n().GetSupportLanguages()[1:]
			}
			return b.I18n().GetSupportLanguages()
		})
	utils.Configure(b)

	media_view.Configure(b, db)
	// media_view.MediaLibraryPerPage = 3
	// vips.UseVips(vips.Config{EnableGenerateWebp: true})
	ConfigureSeo(b, db)

	b.MenuOrder(
		"InputHarness",
		"Post",
		"User",
		"Role",
		b.MenuGroup("Site Management").SubItems(
			"Setting",
			"QorSEOSetting",
		).Icon("settings"),
		b.MenuGroup("Product Management").SubItems(
			"Product",
			"Category",
		).Icon("shopping_cart"),
		b.MenuGroup("Page Builder").SubItems(
			"Page",
			"shared_containers",
			"demo_containers",
			"page_templates",
			"page_categories",
		).Icon("view_quilt"),
	)

	m := b.Model(&models.Post{})
	slug.Configure(b, m)

	m.Listing("ID", "Title", "TitleWithSlug", "HeroImage", "Body").
		SearchColumns("title", "body").
		PerPage(10)

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

	configInputHarness(b, db)
	configUser(b, db)
	role.Configure(b, db, role.DefaultActions, []vuetify.DefaultOptionItem{
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
	configProduct(b, db, w)
	configCategory(b, db)

	// Use m to customize the model, Or config more models here.

	type Setting struct{}
	sm := b.Model(&Setting{})
	sm.RegisterEventFunc(pages.LogInfoEvent, pages.LogInfo)
	sm.Listing().PageFunc(pages.Settings(db))

	type ListEditorExample struct{}
	leem := b.Model(&ListEditorExample{}).Label("List Editor Example")
	pf, sf := pages.ListEditorExample(db, b)
	leem.Listing().PageFunc(pf)
	leem.RegisterEventFunc("save", sf)

	configCustomer(b, db)

	pageBuilder := example.ConfigPageBuilder(db, "/admin/page_builder", `<link rel="stylesheet" href="https://the-plant.com/assets/app/container.9506d40.css">`)
	pm := pageBuilder.Configure(b, db)

	publisher := publish.New(db, oss.Storage).WithPageBuilder(pageBuilder)

	l := b.Model(&models.ListModel{})
	l.Listing("ID", "Title", "Status")
	l.Editing("Status", "Schedule", "Title")

	note.Configure(db, b, m, pm)

	// @snippet_begin(ActivityExample)
	ab := activity.New(b, db).SetCreatorContextKey(_userKey).SetTabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		})
	_ = ab
	// ab.Model(m).UseDefaultTab()
	// ab.Model(pm).UseDefaultTab()
	// ab.Model(l).SkipDelete().SkipCreate()
	// @snippet_end
	ab.RegisterModels(m, l, pm)
	ab.GetPresetModelBuilder().Listing().FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		var (
			logs         = ab.NewLogModelSlice()
			activityMsgr = i18n.MustGetModuleMessages(ctx.R, activity.I18nActivityKey, activity.Messages_en_US).(*activity.Messages)
			publishMsgr  = i18n.MustGetModuleMessages(ctx.R, publish_view.I18nPublishKey, publish_view.Messages_en_US).(*publish_view.Messages)
			contextDB    = db
		)

		contextDB.Select("creator").Group("creator").Find(logs)
		reflectValue := reflect.Indirect(reflect.ValueOf(logs))
		var creatorOptions []*vuetifyx.SelectItem
		for i := 0; i < reflectValue.Len(); i++ {
			creator := reflect.Indirect(reflectValue.Index(i)).FieldByName("Creator").String()
			creatorOptions = append(creatorOptions, &vuetifyx.SelectItem{
				Text:  creator,
				Value: creator,
			})
		}

		var modelOptions []*vuetifyx.SelectItem
		for _, m := range ab.GetModelBuilders() {
			modelOptions = append(modelOptions, &vuetifyx.SelectItem{
				Text:  m.GetType().Name(),
				Value: m.GetType().Name(),
			})
		}

		return []*vuetifyx.FilterItem{
			{
				Key:          "action",
				Label:        activityMsgr.FilterAction,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `action %s ?`,
				Options: []*vuetifyx.SelectItem{
					{Text: activityMsgr.ActionEdit, Value: activity.ActivityEdit},
					{Text: activityMsgr.ActionCreate, Value: activity.ActivityCreate},
					{Text: activityMsgr.ActionDelete, Value: activity.ActivityDelete},
					{Text: publishMsgr.Publish, Value: publish_view.ActivityPublish},
					{Text: publishMsgr.Republish, Value: publish_view.ActivityRepublish},
					{Text: publishMsgr.Unpublish, Value: publish_view.ActivityUnPublish},
				},
			},
			{
				Key:          "created",
				Label:        activityMsgr.FilterCreatedAt,
				ItemType:     vuetifyx.ItemTypeDate,
				SQLCondition: `created_at %s ?`,
			},
			{
				Key:          "creator",
				Label:        activityMsgr.FilterCreator,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `creator %s ?`,
				Options:      creatorOptions,
			},
			{
				Key:          "model",
				Label:        activityMsgr.FilterModel,
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `model_name %s ?`,
				Options:      modelOptions,
			},
		}
	})

	mm := b.Model(&models.MicrositeModel{})
	mm.Listing("ID", "Name", "PrePath", "Status").
		SearchColumns("ID", "Name").
		PerPage(10)
	mm.Editing("Status", "Schedule", "Name", "Description", "PrePath", "FilesList", "Package")
	microsite_views.Configure(b, db, ab, oss.Storage, domain, publisher, mm)

	publish_view.Configure(b, db, ab, publisher, m, l, pm)

	return Config{
		pb:          b,
		lb:          newLoginBuilder(db),
		pageBuilder: pageBuilder,
	}
}
