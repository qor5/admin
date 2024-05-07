package admin

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/qor/oss"
	"github.com/qor/oss/filesystem"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/l10n"
	l10n_view "github.com/qor5/admin/v3/l10n/views"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	media_oss "github.com/qor5/admin/v3/media/oss"
	media_view "github.com/qor5/admin/v3/media/views"
	microsite_utils "github.com/qor5/admin/v3/microsite/utils"
	microsite_views "github.com/qor5/admin/v3/microsite/views"
	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	publish_view "github.com/qor5/admin/v3/publish/views"
	"github.com/qor5/admin/v3/richeditor"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/admin/v3/slug"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/admin/v3/worker"
	v "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
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
	Publisher   *publish.Builder
}

func NewConfig() Config {
	db := ConnectDB()
	sess := session.Must(session.NewSession())
	media_oss.Storage = s3.New(&s3.Config{
		Bucket:   os.Getenv("S3_Bucket"),
		Region:   os.Getenv("S3_Region"),
		ACL:      s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Endpoint: os.Getenv("S3_Endpoint"),
		Session:  sess,
	})
	PublishStorage = microsite_utils.NewClient(s3.New(&s3.Config{
		Bucket:   os.Getenv("S3_Publish_Bucket"),
		Region:   os.Getenv("S3_Publish_Region"),
		ACL:      s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Session:  sess,
		Endpoint: os.Getenv("PUBLISH_URL"),
	}))
	b := presets.New().RightDrawerWidth("700").VuetifyOptions(`
{
  theme: {
    defaultTheme: 'light',
    themes: {
	  light: {
		dark: false,
		colors: {
		  // primary
		  'primary': '#3E63DD',
		  'on-primary': '#fff',
		  'primary-darken-1': '#3451B2',
		  'primary-darken-2': '#101D46',
		  'primary-lighten-1': '#8DA4EF',
		  'primary-lighten-2': '#E6EDFE',
		  // secondary
		  'secondary': '#5B6471',
		  'on-secondary': '#fff',
		  'secondary-darken-1': '#444C58',
		  'secondary-darken-2': '#19212B',
		  'secondary-lighten-1': '#C2CDDC',
		  'secondary-lighten-2': '#E4ECFE',
		  // info
		  'info': '#0091FF',
		  'on-info': '#fff',
		  'info-darken-1': '#006ADC',
		  'info-darken-2': '#00254D',
		  'info-lighten-1': '#5EB0EF',
		  'info-lighten-2': '#E1F0FF',
		  // success
		  'success': '#30A46C',
		  'on-success': '#fff',
		  'success-darken-1': '#18794E',
		  'success-darken-2': '#153226',
		  'success-lighten-1': '#5BB98C',
		  'success-lighten-2': '#DDF3E4',
		  // warning
		  'warning': '#F76808',
		  'on-warning': '#fff',
		  'warning-darken-1': '#BD4B00',
		  'warning-darken-2': '#451E11',
		  'warning-lighten-1': '#FA934E',
		  'warning-lighten-2': '#FFE8D7',
		  // error
		  'error': '#E5484D',
		  'on-error': '#fff',
		  'error-darken-1': '#CD2B31',
		  'error-darken-2': '#381316',
		  'error-lighten-1': '#EB9091',
		  'error-lighten-2': '#FFE5E5',
		  // background
		  'background': '#fff',
		  'on-background': '#212121',
		  // surface
		  'surface': '#fff',
		  'on-surface': '#424242',
		},
		variables: {
		  'border-color': '#000000',
		  'border-opacity': 0.12,
		  'high-emphasis-opacity': 1,
		  'medium-emphasis-opacity': 0.60,
		  'disabled-opacity': 0.38,
		  'idle-opacity': 0.04,
		  'hover-opacity': 0.04,
		  'focus-opacity': 0.12,
		  'selected-opacity': 0.08,
		  'activated-opacity': 0.12,
		  'pressed-opacity': 0.12,
		  'dragged-opacity': 0.08,
		  'theme-kbd': '#212121',
		  'theme-on-kbd': '#FFFFFF',
		  'theme-code': '#F5F5F5',
		  'theme-on-code': '#000000',
	
		  /* custom variables */
		  // shadows
		  'shadow-key-umbra-color': '#000000',
		  'shadow-xs-opacity': '0.16',
		  'shadow-sm-opacity': '0.18',
		  'shadow-md-opacity': '0.20',
		  'shadow-lg-opacity': '0.22',
		  'shadow-xl-opacity': '0.24',
		},
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
	b.BrandFunc(func(ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)
		logo := "https://qor5.com/img/qor-logo.png"

		now := time.Now()
		nextEvenHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1+(now.Hour()%2), 0, 0, 0, now.Location())
		diff := int(nextEvenHour.Sub(now).Seconds())
		hours := diff / 3600
		minutes := (diff % 3600) / 60
		seconds := diff % 60
		countdown := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)

		return h.Div(
			v.VRow(
				v.VCol(h.A(h.Img(logo).Attr("width", "80")).Href("/")),
				v.VCol(h.H1(msgr.Demo)).Class("pt-4"),
			),
			// ).Density(DensityCompact),
			h.If(os.Getenv("AWS_REGION") != "",
				h.Div(
					h.Span(msgr.DBResetTipLabel),
					v.VIcon("schedule").Size(v.SizeXSmall),
					// .Left(true),
					h.Span(countdown).Id("countdown"),
				).Class("pt-1 pb-2"),
				v.VDivider(),
				h.Script("function updateCountdown(){const now=new Date();const nextEvenHour=new Date(now);nextEvenHour.setHours(nextEvenHour.getHours()+(nextEvenHour.getHours()%2===0?2:1),0,0,0);const timeLeft=nextEvenHour-now;const hours=Math.floor(timeLeft/(60*60*1000));const minutes=Math.floor((timeLeft%(60*60*1000))/(60*1000));const seconds=Math.floor((timeLeft%(60*1000))/1000);const countdownElem=document.getElementById(\"countdown\");countdownElem.innerText=`${hours.toString().padStart(2,\"0\")}:${minutes.toString().padStart(2,\"0\")}:${seconds.toString().padStart(2,\"0\")}`}updateCountdown();setInterval(updateCountdown,1000);"),
			),
		).Class("mb-n4 mt-n2")
	}).ProfileFunc(profile).
		NotificationFunc(notifierComponent(db), notifierCount(db)).
		DataOperator(gorm2op.DataOperator(db)).
		HomePageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r.PageTitle = "Home"
			r.Body = Dashboard()
			return
		}).
		NotFoundPageLayoutConfig(&presets.LayoutConfig{
			SearchBoxInvisible:          true,
			NotificationCenterInvisible: true,
		})

	initPermission(b, db)

	b.I18n().
		SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese).
		RegisterForModule(language.SimplifiedChinese, presets.ModelsI18nModuleKey, Messages_zh_CN_ModelsI18nModuleKey).
		RegisterForModule(language.Japanese, presets.ModelsI18nModuleKey, Messages_ja_JP_ModelsI18nModuleKey).
		RegisterForModule(language.English, I18nExampleKey, Messages_en_US).
		RegisterForModule(language.Japanese, I18nExampleKey, Messages_ja_JP).
		RegisterForModule(language.SimplifiedChinese, I18nExampleKey, Messages_zh_CN).
		GetSupportLanguagesFromRequestFunc(func(r *http.Request) []language.Tag {
			// // Example:
			// user := getCurrentUser(r)
			// var supportedLanguages []language.Tag
			// for _, role := range user.GetRoles() {
			//	switch role {
			//	case "English Group":
			//		supportedLanguages = append(supportedLanguages, language.English)
			//	case "Chinese Group":
			//		supportedLanguages = append(supportedLanguages, language.SimplifiedChinese)
			//	}
			// }
			// return supportedLanguages
			return b.I18n().GetSupportLanguages()
		})

	l10nBuilder := l10n.New()
	l10nBuilder.
		RegisterLocales("International", "international", "International").
		RegisterLocales("China", "cn", "China").
		RegisterLocales("Japan", "jp", "Japan").
		GetSupportLocaleCodesFromRequestFunc(func(R *http.Request) []string {
			return l10nBuilder.GetSupportLocaleCodes()[:]
		})

	utils.Configure(b)

	media_view.Configure(b, db)
	// media_view.MediaLibraryPerPage = 3
	// vips.UseVips(vips.Config{EnableGenerateWebp: true})
	ConfigureSeo(b, db, l10nBuilder.GetSupportLocaleCodes()...)

	b.MenuOrder(
		"profile",
		b.MenuGroup("Page Builder").SubItems(
			"Page",
			"shared_containers",
			"demo_containers",
			"page_templates",
			"page_categories",
		).Icon("mdi-view-quilt"),
		b.MenuGroup("EC Management").SubItems(
			"ec-dashboard",
			"Order",
			"Product",
			"Category",
		).Icon("mdi-cart"),
		// b.MenuGroup("Site Management").SubItems(
		// 	"Setting",
		// 	"QorSEOSetting",
		// ).Icon("settings"),
		b.MenuGroup("User Management").SubItems(
			"User",
			"Role",
		).Icon("mdi-account-multiple"),
		b.MenuGroup("Featured Models Management").SubItems(
			"InputDemo",
			"Post",
			"qor-seo-settings",
			"List Editor Example",
			"nested-field-demos",
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

	mListing.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vx.FilterItem{
			{
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "posts", "Posts", u.ID, "Posts"),
			},
			{
				Key:          "created",
				Label:        "Create Time",
				ItemType:     vx.ItemTypeDatetimeRange,
				SQLCondition: `created_at %s ?`,
			},
			{
				Key:          "title",
				Label:        "Title",
				ItemType:     vx.ItemTypeString,
				SQLCondition: `title %s ?`,
			},
			{
				Key:      "status",
				Label:    "Status",
				ItemType: vx.ItemTypeSelect,
				Options: []*vx.SelectItem{
					{Text: publish.StatusDraft, Value: publish.StatusDraft},
					{Text: publish.StatusOnline, Value: publish.StatusOnline},
					{Text: publish.StatusOffline, Value: publish.StatusOffline},
				},
				SQLCondition: `status %s ?`,
			},
			{
				Key:      "multi_statuses",
				Label:    "Multiple Statuses",
				ItemType: vx.ItemTypeMultipleSelect,
				Options: []*vx.SelectItem{
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
				ItemType:     vx.ItemTypeNumber,
				SQLCondition: `id %s ?`,
				Folded:       true,
			},
		}
	})

	mListing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)

		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabsAll,
				ID:    "all",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: msgr.FilterTabsHasUnreadNotes,
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})

	w := worker.New(db)
	defer w.Listen()
	addJobs(w)

	ed := m.Editing("StatusBar", "ScheduleBar", "Title", "TitleWithSlug", "Seo", "HeroImage", "Body", "BodyImage")
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

	roleBuilder := role.New(db).
		Resources([]*v.DefaultOptionItem{
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
	roleModelBuilder := roleBuilder.Configure(b)
	roleModelBuilder.Listing().Searcher = func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		u := getCurrentUser(ctx.R)
		qdb := db

		// If the current user doesn't has 'admin' role, do not allow them to view admin and manager roles
		// We didn't do this on permission because of we are not supporting the permission on listing page
		if currentRoles := u.GetRoles(); !utils.Contains(currentRoles, models.RoleAdmin) {
			qdb = db.Where("name NOT IN (?)", []string{models.RoleAdmin, models.RoleManager})
		}

		return gorm2op.DataOperator(qdb).Search(model, params, ctx)
	}

	product := configProduct(b, db, w)
	category := configCategory(b, db)

	// Use m to customize the model, Or config more models here.

	// type Setting struct{}
	// sm := b.Model(&Setting{})
	// sm.RegisterEventFunc(pages.LogInfoEvent, pages.LogInfo)
	// sm.Listing().PageFunc(pages.Settings(db))

	// FIXME: list editor does not support use in page func
	// type ListEditorExample struct{}
	// leem := b.Model(&ListEditorExample{}).Label("List Editor Example")
	// pf, sf := pages.ListEditorExample(db, b)
	// leem.Listing().PageFunc(pf)
	// leem.RegisterEventFunc("save", sf)

	configNestedFieldDemo(b, db)

	// @snippet_begin(ActivityExample)
	ab := activity.New(b, db).SetCreatorContextKey(login.UserKey).SetTabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		})
	ab.GetPresetModelBuilder().Listing().SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		u := getCurrentUser(ctx.R)
		qdb := db
		if rs := u.GetRoles(); !utils.Contains(rs, models.RoleAdmin) {
			qdb = db.Where("user_id = ?", u.ID)
		}
		return gorm2op.DataOperator(qdb).Search(model, params, ctx)
	})
	// ab.Model(m).EnableActivityInfoTab()
	// ab.Model(pm).EnableActivityInfoTab()
	// ab.Model(l).SkipDelete().SkipCreate()
	// @snippet_end

	w.Activity(ab).Configure(b)
	publisher := publish.New(db, PublishStorage).WithL10nBuilder(l10nBuilder)

	pageBuilder := example.ConfigPageBuilder(db, "/page_builder", ``, b.I18n())
	pm := pageBuilder.Configure(b, db, l10nBuilder, ab, publisher, seoBuilder)
	pmListing := pm.Listing()
	pmListing.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vx.FilterItem{
			{
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "page_builder_pages", "Pages", u.ID, "Pages"),
			},
		}
	})

	pmListing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nExampleKey, Messages_en_US).(*Messages)

		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabsAll,
				ID:    "all",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: msgr.FilterTabsHasUnreadNotes,
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})

	l := b.Model(&models.ListModel{})
	{
		l.Listing("ID", "Title", "Status")
		ed := l.Editing("StatusBar", "ScheduleBar", "Title", "DetailPath", "ListPath")
		ed.Field("DetailPath").ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
				this := obj.(*models.ListModel)

				if this.GetStatus() != publish.StatusOnline {
					return nil
				}

				var content []h.HTMLComponent

				content = append(content,
					h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, l.Info().Label(), field.Label)).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
				)
				domain := PublishStorage.GetEndpoint()
				if this.OnlineUrl != "" {
					p := this.OnlineUrl
					content = append(content, h.A(h.Text(p)).Href(domain+p))
				}

				return h.Div(
					h.Div(
						h.Div(
							content...,
						).Class("v-text-field__slot").Style("padding: 8px 0;"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted")
			},
		).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			return nil
		})

		ed.Field("ListPath").ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
				this := obj.(*models.ListModel)

				if this.GetStatus() != publish.StatusOnline || this.GetPageNumber() == 0 {
					return nil
				}

				var content []h.HTMLComponent

				content = append(content,
					h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, l.Info().Label(), field.Label)).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
				)
				domain := PublishStorage.GetEndpoint()
				if this.OnlineUrl != "" {
					p := this.GetListUrl(strconv.Itoa(this.GetPageNumber()))
					content = append(content, h.A(h.Text(p)).Href(domain+p))
				}

				return h.Div(
					h.Div(
						h.Div(
							content...,
						).Class("v-text-field__slot").Style("padding: 8px 0;"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted")
			},
		).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			return nil
		})
	}

	b.GetWebBuilder().RegisterEventFunc(noteMarkAllAsRead, markAllAsRead(db))

	note.Configure(db, b, m)

	if err := db.AutoMigrate(&UserUnreadNote{}); err != nil {
		panic(err)
	}
	note.AfterCreateFunc = NoteAfterCreateFunc

	ab.RegisterModel(m).EnableActivityInfoTab()
	ab.RegisterModels(l)
	mm := b.Model(&models.MicrositeModel{})
	mm.Listing("ID", "Name", "PrePath", "Status").
		SearchColumns("ID::text", "Name").
		PerPage(10)
	mm.Editing("StatusBar", "ScheduleBar", "Name", "Description", "PrePath", "FilesList", "Package")
	microsite_views.Configure(b, db, ab, PublishStorage, publisher, mm)
	l10nM, l10nVM := configL10nModel(b)
	_ = l10nM
	publish_view.Configure(b, db, ab, publisher, m, l, product, category, l10nVM)

	initLoginBuilder(db, b, ab)

	configInputDemo(b, db)

	configOrder(b, db)
	configECDashboard(b, db)

	configUser(b, db)
	configProfile(b, db)

	l10n_view.Configure(b, db, l10nBuilder, ab, l10nM, l10nVM)

	if os.Getenv("RESET_AND_IMPORT_INITIAL_DATA") == "true" {
		tbs := GetNonIgnoredTableNames()
		EmptyDB(db, tbs)
		InitDB(db, tbs)
	}

	return Config{
		pb:          b,
		pageBuilder: pageBuilder,
		Publisher:   publisher,
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

		return v.VList(
			v.VListItem(
				v.VListItemTitle(h.Text("Pages")),
				v.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", a))),
			).Lines(2).Href("/pages?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			v.VListItem(
				v.VListItemTitle(h.Text("Posts")),
				v.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", b))),
			).Lines(2).Href("/posts?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			v.VListItem(
				v.VListItemTitle(h.Text("Users")),
				v.VListItemSubtitle(h.Text(fmt.Sprintf("%d unread notes", c))),
			).Lines(2).Href("/users?active_filter_tab=hasUnreadNotes&f_hasUnreadNotes=1"),
			h.If(a+b+c > 0,
				v.VListItem(
					v.VListItemSubtitle(h.Text("Mark all as read")),
				).Attr("@click", web.Plaid().EventFunc(noteMarkAllAsRead).Go()),
			),
		)
		// .Class("mx-auto")
		// .Attr("max-width", "140")
	}
}
