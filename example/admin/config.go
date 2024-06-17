package admin

import (
	"embed"
	"fmt"
	"github.com/qor5/admin/v3/activity"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/qor/oss"
	"github.com/qor/oss/filesystem"
	"github.com/qor/oss/s3"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	media_oss "github.com/qor5/admin/v3/media/oss"
	"github.com/qor5/admin/v3/microsite"
	microsite_utils "github.com/qor5/admin/v3/microsite/utils"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/richeditor"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/admin/v3/slug"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/admin/v3/worker"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

//go:embed assets
var assets embed.FS

// PublishStorage is used to storage static pages published by page builder.
var PublishStorage oss.StorageInterface = filesystem.New("publish")

type Config struct {
	pb          *presets.Builder
	pageBuilder *pagebuilder.Builder
	Publisher   *publish.Builder
}

var (
	s3Bucket                  = osenv.Get("S3_Bucket", "s3-bucket for media library storage", "example")
	s3Region                  = osenv.Get("S3_Region", "s3-region for media library storage", "ap-northeast-1")
	s3Endpoint                = osenv.Get("S3_Endpoint", "s3-endpoint for media library storage", "https://s3.ap-northeast-1.amazonaws.com")
	s3PublishBucket           = osenv.Get("S3_Publish_Bucket", "s3-bucket for publish", "example-publish")
	s3PublishRegion           = osenv.Get("S3_Publish_Region", "s3-region for publish", "ap-northeast-1")
	publishURL                = osenv.Get("PUBLISH_URL", "publish url", "")
	awsRegion                 = osenv.Get("AWS_REGION", "aws region for show count down", "")
	resetAndImportInitialData = osenv.GetBool("RESET_AND_IMPORT_INITIAL_DATA",
		"Will reset and import initial data if set to true", false)
)

func NewConfig(db *gorm.DB) Config {
	if err := db.AutoMigrate(
		&models.Post{},
		&models.InputDemo{},
		&models.User{},
		&models.LoginSession{},
		&models.ListModel{},
		&role.Role{},
		&perm.DefaultDBPolicy{},
		&models.Customer{},
		&models.Address{},
		&models.Phone{},
		&models.MembershipCard{},
		&models.Product{},
		&models.Order{},
		&models.Category{},
	); err != nil {
		panic(err)
	}

	sess := session.Must(session.NewSession())
	media_oss.Storage = s3.New(&s3.Config{
		Bucket:   s3Bucket,
		Region:   s3Region,
		ACL:      s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Endpoint: s3Endpoint,
		Session:  sess,
	})
	PublishStorage = microsite_utils.NewClient(s3.New(&s3.Config{
		Bucket:   s3PublishBucket,
		Region:   s3PublishRegion,
		ACL:      s3control.S3CannedAccessControlListBucketOwnerFullControl,
		Session:  sess,
		Endpoint: publishURL,
	}))
	b := presets.New().RightDrawerWidth("700")
	defer b.Build()

	js, _ := assets.ReadFile("assets/fontcolor.min.js")
	richeditor.Plugins = []string{"alignment", "table", "video", "imageinsert", "fontcolor"}
	richeditor.PluginsJS = [][]byte{js}
	b.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	b.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
	configBrand(b, db)

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
	mediab := media.New(db)

	l10nBuilder := l10n.New(db)
	l10nBuilder.
		RegisterLocales("International", "international", "International").
		RegisterLocales("China", "cn", "China").
		RegisterLocales("Japan", "jp", "Japan").
		SupportLocalesFunc(func(R *http.Request) []string {
			return l10nBuilder.GetSupportLocaleCodes()[:]
		})
	publisher := publish.New(db, PublishStorage).
		ContextValueFuncs(l10nBuilder.ContextValueProvider)

	utils.Install(b)

	//var NoteAfterCreateFunc = func(db *gorm.DB) (err error) {
	//	return db.Exec(`DELETE FROM "user_unread_notes";`).Error
	//}

	// @snippet_begin(ActivityExample)
	ab := activity.New(db).CreatorContextKey(login.UserKey).TabHeading(
		func(log activity.ActivityLogInterface) string {
			return fmt.Sprintf("%s %s at %s", log.GetCreator(), strings.ToLower(log.GetAction()), log.GetCreatedAt().Format("2006-01-02 15:04:05"))
		}).
		WrapLogModelInstall(func(in presets.ModelInstallFunc) presets.ModelInstallFunc {
			return func(pb *presets.Builder, mb *presets.ModelBuilder) (err error) {
				err = in(pb, mb)
				if err != nil {
					return
				}
				mb.Listing().SearchFunc(func(model interface{}, params *presets.SearchParams,
					ctx *web.EventContext,
				) (r interface{}, totalCount int, err error) {
					u := getCurrentUser(ctx.R)
					qdb := db
					if rs := u.GetRoles(); !slices.Contains(rs, models.RoleAdmin) {
						qdb = db.Where("user_id = ?", u.ID)
					}
					return gorm2op.DataOperator(qdb).Search(model, params, ctx)
				})
				return
			}
		})
	b.Use(ab)

	// ab.Model(m).EnableActivityInfoTab()
	// ab.Model(pm).EnableActivityInfoTab()
	// ab.Model(l).SkipDelete().SkipCreate()
	// @snippet_end

	// media_view.MediaLibraryPerPage = 3
	// vips.UseVips(vips.Config{EnableGenerateWebp: true})
	configureSeo(b, db, l10nBuilder.GetSupportLocaleCodes()...)

	configMenuOrder(b)

	configPost(b, db, ab, publisher, ab)

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
		}).
		AfterInstall(func(pb *presets.Builder, mb *presets.ModelBuilder) error {
			mb.Listing().SearchFunc(func(
				model interface{},
				params *presets.SearchParams,
				ctx *web.EventContext,
			) (r interface{}, totalCount int, err error) {
				u := getCurrentUser(ctx.R)
				qdb := db
				// If the current user doesn't has 'admin' role, do not allow them to view admin and manager roles
				// We didn't do this on permission because of we are not supporting the permission on listing page
				if currentRoles := u.GetRoles(); !slices.Contains(currentRoles, models.RoleAdmin) {
					qdb = db.Where("name NOT IN (?)", []string{models.RoleAdmin, models.RoleManager})
				}
				return gorm2op.DataOperator(qdb).Search(model, params, ctx)
			})
			return nil
		})

	w := worker.New(db)
	defer w.Listen()
	addJobs(w)
	configProduct(b, db, w, publisher)
	configCategory(b, db, publisher)

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

	b.Use(w.Activity(ab))

	pageBuilder := example.ConfigPageBuilder(db, "/page_builder", ``, b.I18n())
	pageBuilder.
		Media(mediab).
		L10n(l10nBuilder).
		Activity(ab).
		Publisher(publisher).
		SEO(seoBuilder).
		WrapPageInstall(func(in presets.ModelInstallFunc) presets.ModelInstallFunc {
			return func(pb *presets.Builder, pm *presets.ModelBuilder) (err error) {
				err = in(pb, pm)
				if err != nil {
					return
				}
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
				return nil
			}
		})

	b.Use(pageBuilder)

	configListModel(b, ab)

	b.GetWebBuilder().RegisterEventFunc(noteMarkAllAsRead, markAllAsRead(db))

	if err := db.AutoMigrate(&UserUnreadNote{}); err != nil {
		panic(err)
	}

	microb := microsite.New(db).Publisher(publisher)

	l10nBuilder.Activity(ab)
	l10nM, l10nVM := configL10nModel(db, b)
	l10nM.Use(l10nBuilder)
	l10nVM.Use(l10nBuilder)

	publisher.Activity(ab)

	initLoginBuilder(db, b, ab)

	configInputDemo(b, db)

	configOrder(b, db)
	configECDashboard(b, db)

	configUser(b, ab, db, publisher)
	configProfile(b, db)

	b.Use(
		mediab,
		microb,
		ab,
		publisher,
		l10nBuilder,
		roleBuilder,
	)

	if resetAndImportInitialData {
		tbs := GetNonIgnoredTableNames(db)
		EmptyDB(db, tbs)
		InitDB(db, tbs)
	}

	return Config{
		pb:          b,
		pageBuilder: pageBuilder,
		Publisher:   publisher,
	}
}

func configListModel(b *presets.Builder, ab *activity.Builder) *presets.ModelBuilder {
	l := b.Model(&models.ListModel{}).Use(ab)
	{
		l.Listing("ID", "Title", "Status")
		ed := l.Editing("StatusBar", "ScheduleBar", "Title", "DetailPath", "ListPath")
		ed.Field("DetailPath").ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
				this := obj.(*models.ListModel)

				if this.Status.Status != publish.StatusOnline {
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

				if this.Status.Status != publish.StatusOnline || this.PageNumber == 0 {
					return nil
				}

				var content []h.HTMLComponent

				content = append(content,
					h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, l.Info().Label(), field.Label)).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
				)
				domain := PublishStorage.GetEndpoint()
				if this.OnlineUrl != "" {
					p := this.GetListUrl(strconv.Itoa(this.PageNumber))
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
	return l
}

func configMenuOrder(b *presets.Builder) {
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
}

func configBrand(b *presets.Builder, db *gorm.DB) {
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
			h.If(awsRegion != "",
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
	}).ProfileFunc(profile(db)).
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
}

func configPost(
	b *presets.Builder,
	db *gorm.DB,
	ab *activity.Builder,
	publisher *publish.Builder,
	nb *activity.Builder,
) *presets.ModelBuilder {
	m := b.Model(&models.Post{})
	m.Use(
		slug.New(),
		ab,
		publisher,
		nb,
	)

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

	ed := m.Editing("StatusBar", "ScheduleBar", "Title", "TitleWithSlug", "Seo", "HeroImage", "Body", "BodyImage")
	ed.Field("HeroImage").
		WithContextValue(
			media.MediaBoxConfig,
			&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*base.Size{
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
			media.MediaBoxConfig,
			&media_library.MediaBoxConfig{})

	ed.Field("Body").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return richeditor.RichEditor(db, "Body").Plugins([]string{"alignment", "video", "imageinsert", "fontcolor"}).Value(obj.(*models.Post).Body).Label(field.Label)
	})
	return m
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
