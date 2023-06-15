package pagebuilder

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/qor5/admin/utils"

	"goji.io/pat"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/l10n"
	l10n_view "github.com/qor5/admin/l10n/views"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	"github.com/qor5/admin/publish/views"
	pv "github.com/qor5/admin/publish/views"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/perm"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type RenderInput struct {
	IsEditor bool
	Device   string
}

type RenderFunc func(obj interface{}, input *RenderInput, ctx *web.EventContext) h.HTMLComponent

type PageLayoutFunc func(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent

type PageLayoutInput struct {
	Page              *Page
	SeoTags           template.HTML
	CanonicalLink     template.HTML
	StructuredData    template.HTML
	FreeStyleCss      []string
	FreeStyleTopJs    []string
	FreeStyleBottomJs []string
	Header            h.HTMLComponent
	Footer            h.HTMLComponent
	IsEditor          bool
	EditorCss         []h.HTMLComponent
	IsPreview         bool
	Locale            string
}

type Builder struct {
	prefix            string
	wb                *web.Builder
	db                *gorm.DB
	containerBuilders []*ContainerBuilder
	ps                *presets.Builder
	mb                *presets.ModelBuilder
	pageStyle         h.HTMLComponent
	pageLayoutFunc    PageLayoutFunc
	preview           http.Handler
	images            http.Handler
	imagesPrefix      string
	defaultDevice     string
}

const (
	openTemplateDialogEvent          = "openTemplateDialogEvent"
	selectTemplateEvent              = "selectTemplateEvent"
	clearTemplateEvent               = "clearTemplateEvent"
	republishRelatedOnlinePagesEvent = "republish_related_online_pages"

	paramOpenFromSharedContainer = "open_from_shared_container"
)

func New(prefix string, db *gorm.DB, i18nB *i18n.Builder) *Builder {
	err := db.AutoMigrate(
		&Page{},
		&Template{},
		&Container{},
		&DemoContainer{},
		&Category{},
	)
	if err != nil {
		panic(err)
	}
	// https://github.com/go-gorm/sqlite/blob/64917553e84d5482e252c7a0c8f798fb672d7668/ddlmod.go#L16
	// fxxk: newline is not allowed
	err = db.Exec(`
create unique index if not exists uidx_page_builder_demo_containers_model_name_locale_code on page_builder_demo_containers (model_name, locale_code) where deleted_at is null;
`).Error
	if err != nil {
		panic(err)
	}

	r := &Builder{
		db:            db,
		wb:            web.New(),
		prefix:        prefix,
		defaultDevice: Device_Computer,
	}
	r.ps = presets.New().
		BrandTitle("Page Builder").
		DataOperator(gorm2op.DataOperator(db)).
		URIPrefix(prefix).
		LayoutFunc(r.pageEditorLayout).
		SetI18n(i18nB)
	type Editor struct {
	}
	r.ps.Model(&Editor{}).
		Detailing().
		PageFunc(r.Editor)
	r.ps.GetWebBuilder().RegisterEventFunc(AddContainerDialogEvent, r.AddContainerDialog)
	r.ps.GetWebBuilder().RegisterEventFunc(AddContainerEvent, r.AddContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(DeleteContainerConfirmationEvent, r.DeleteContainerConfirmation)
	r.ps.GetWebBuilder().RegisterEventFunc(DeleteContainerEvent, r.DeleteContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(MoveContainerEvent, r.MoveContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(ToggleContainerVisibilityEvent, r.ToggleContainerVisibility)
	r.ps.GetWebBuilder().RegisterEventFunc(MarkAsSharedContainerEvent, r.MarkAsSharedContainer)
	r.ps.GetWebBuilder().RegisterEventFunc(RenameCotainerDialogEvent, r.RenameContainerDialog)
	r.ps.GetWebBuilder().RegisterEventFunc(RenameContainerEvent, r.RenameContainer)
	r.preview = r.ps.GetWebBuilder().Page(r.Preview)
	return r
}

func (b *Builder) Prefix(v string) (r *Builder) {
	b.ps.URIPrefix(v)
	b.prefix = v
	return b
}

func (b *Builder) PageStyle(v h.HTMLComponent) (r *Builder) {
	b.pageStyle = v
	return b
}

func (b *Builder) PageLayout(v PageLayoutFunc) (r *Builder) {
	b.pageLayoutFunc = v
	return b
}

func (b *Builder) Images(v http.Handler, imagesPrefix string) (r *Builder) {
	b.images = v
	b.imagesPrefix = imagesPrefix
	return b
}

func (b *Builder) DefaultDevice(v string) (r *Builder) {
	b.defaultDevice = v
	return b
}

func (b *Builder) GetPresetsBuilder() (r *presets.Builder) {
	return b.ps
}

func (b *Builder) Configure(pb *presets.Builder, db *gorm.DB, l10nB *l10n.Builder, activityB *activity.ActivityBuilder) (pm *presets.ModelBuilder) {
	pb.I18n().
		RegisterForModule(language.English, I18nPageBuilderKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPageBuilderKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPageBuilderKey, Messages_ja_JP)
	pm = pb.Model(&Page{})
	b.mb = pm
	pm.Listing("ID", "Title", "Slug")
	dp := pm.Detailing("Overview")
	dp.Field("Overview").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mi := field.ModelInfo
		p := obj.(*Page)
		c := &Category{}
		db.First(c, "id = ?", p.CategoryID)

		overview := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Title).ZeroLabel("No Title")).Label("Title"),
				vx.DetailField(vx.OptionalText(c.Path).ZeroLabel("No Category")).Label("Category"),
			),
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Slug).ZeroLabel("No Slug")).Label("Slug"),
			),
		)
		var start string
		if p.GetScheduledStartAt() != nil {
			start = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		pageState := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.GetStatus()).ZeroLabel("No State")).Label("State"),
				vx.DetailField(vx.OptionalText(start).ZeroLabel("No Set")).Label("SchedulePublishTime"),
			),
		)
		var unpublishBtn h.HTMLComponent
		pvMsgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, utils.Messages_en_US).(*pv.Messages)
		if p.GetStatus() == publish.StatusOnline {
			unpublishBtn = VBtn(pvMsgr.Unpublish).Depressed(true).Class("mr-2").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.UnpublishEvent))
		}
		return VContainer(VRow(VCol(
			vx.Card(overview).HeaderTitle("Overview").
				Actions(
					VBtn("Edit").
						Depressed(true).
						Attr("@click", web.POST().
							EventFunc(actions.Edit).
							Query(presets.ParamOverlay, actions.Dialog).
							Query(presets.ParamID, p.PrimarySlug()).
							URL(mi.PresetsPrefix()+"/pages").
							Go(),
						),
				).Class("mb-4"),
			vx.Card(pageState).HeaderTitle("Page State").
				Actions(
					h.If(unpublishBtn != nil, unpublishBtn),
					VBtn("Edit").
						Depressed(true).
						Attr("@click", web.POST().
							EventFunc(actions.Edit).
							Query(presets.ParamOverlay, actions.Dialog).
							Query(presets.ParamID, p.PrimarySlug()).
							URL(mi.PresetsPrefix()+"/pages").
							Go(),
						),
				).Class("mb-4"),
		).Cols(8)))
	})

	oldDetailLayout := pb.GetDetailLayoutFunc()
	pb.DetailLayoutFunc(func(in web.PageFunc, cfg *presets.LayoutConfig) (out web.PageFunc) {
		return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
			if !strings.Contains(ctx.R.RequestURI, "/"+b.mb.Info().URIName()+"/") {
				pr, err = oldDetailLayout(in, cfg)(ctx)
				return
			}

			pb.InjectAssets(ctx)
			// call createMenus before in(ctx) to fill the menuGroupName for modelBuilders first
			menu := pb.CreateMenus(ctx)

			var profile h.HTMLComponent
			if pb.GetProfileFunc() != nil {
				profile = pb.GetProfileFunc()(ctx)
			}

			//msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)
			pvMsgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, utils.Messages_en_US).(*pv.Messages)
			id := pat.Param(ctx.R, "id")

			if id == "" {
				return pb.DefaultNotFoundPageFunc(ctx)
			}
			var obj = pm.NewModel()
			obj, err = dp.GetFetchFunc()(obj, id, ctx)
			if err != nil {
				if err == presets.ErrRecordNotFound {
					return pb.DefaultNotFoundPageFunc(ctx)
				}
				return
			}

			p := obj.(*Page)

			var tabContent web.PageResponse
			tab := ctx.R.FormValue("tab")
			isContent := tab == "content"
			activeTabIndex := 0
			if isContent {
				activeTabIndex = 1
				ctx.R.Form.Set("id", strconv.Itoa(int(p.ID)))
				ctx.R.Form.Set("version", p.GetVersion())
				ctx.R.Form.Set("locale", p.GetLocale())
				tabContent, err = b.PageContent(ctx)
			} else {
				tabContent, err = in(ctx)
			}
			if err == perm.PermissionDenied {
				pr.Body = h.Text(perm.PermissionDenied.Error())
				return pr, nil
			}
			if err != nil {
				panic(err)
			}
			idVerionStatus := fmt.Sprintf("%d %s | %s", p.ID, p.GetVersionName(), p.GetStatus())
			queries := url.Values{}
			action := web.POST().
				EventFunc(actions.Edit).
				URL(web.Var("\""+b.prefix+"/\"+arr[0]")).
				Query(presets.ParamOverlay, actions.Dialog).
				Query(presets.ParamID, web.Var("arr[1]")).
				Go()

			var publishBtn h.HTMLComponent
			switch p.GetStatus() {
			case publish.StatusDraft, publish.StatusOffline:
				publishBtn = VBtn(pvMsgr.Publish).Small(true).Color("#4F378B").Height(40).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.PublishEvent))
			case publish.StatusOnline:
				publishBtn = VBtn(pvMsgr.Republish).Small(true).Color("#4F378B").Height(40).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.RepublishEvent))
			}
			pr.Body = VApp(
				VNavigationDrawer(
					pb.RunBrandProfileSwitchLanguageDisplayFunc(pb.RunBrandFunc(ctx), profile, pb.RunSwitchLanguageFunc(ctx), ctx),
					VDivider(),
					menu,
				).App(true).
					Fixed(true).
					Value(true).
					Attr("v-model", "vars.navDrawer").
					Attr(web.InitContextVars, `{navDrawer: null}`),

				VAppBar(
					VAppBarNavIcon().On("click.stop", "vars.navDrawer = !vars.navDrawer"),
					VTabs(
						VTab(h.Text("{{item.label}}")).Attr("@click", web.Plaid().Queries(queries).Query("tab", web.Var("item.query")).PushState(true).Go()).
							Attr("v-for", "(item, index) in locals.tabs", ":key", "index"),
					).Class("v-tabs--centered").Attr("v-model", `locals.activeTab`).Attr("style", "width:400px"),
					h.If(isContent, VAppBarNavIcon().On("click.stop", "vars.pbEditorDrawer = !vars.pbEditorDrawer")),
					VSelect().HideDetails(true).Dense(true).Outlined(true).
						Items([]string{idVerionStatus}).Value(idVerionStatus).Class("col col-3").
						AppendIcon("chevron_right"),
					VBtn("Duplicate").Small(true).Color("#235FF8").Height(40).Attr("style", "right:13px;"),
					publishBtn,
				).Dark(true).
					Color(presets.ColorPrimary).
					App(true).
					ClippedRight(true).
					Fixed(true),

				web.Portal().Name(presets.RightDrawerPortalName),
				web.Portal().Name(presets.DialogPortalName),
				web.Portal().Name(presets.DeleteConfirmPortalName),
				web.Portal().Name(presets.DefaultConfirmDialogPortalName),
				web.Portal().Name(presets.ListingDialogPortalName),
				web.Portal().Name(dialogPortalName),
				utils.ConfirmDialog(pvMsgr.Areyousure, web.Plaid().EventFunc(web.Var("locals.action")).
					Query(presets.ParamID, p.PrimarySlug()).Go(),
					utilsMsgr),
				h.If(isContent, h.Script(`
(function(){
	let scrollLeft = 0;
	let scrollTop = 0;
	
	function pause(duration) {
		return new Promise(res => setTimeout(res, duration));
	}
	function backoff(retries, fn, delay = 100) {
		fn().catch(err => retries > 1
			? pause(delay).then(() => backoff(retries - 1, fn, delay * 2)) 
			: Promise.reject(err));
	}

	function restoreScroll() {
		window.scroll({left: scrollLeft, top: scrollTop, behavior: "auto"});
		if (window.scrollX == scrollLeft && window.scrollY == scrollTop) {
			return Promise.resolve();
		}
		return Promise.reject();
	}

	window.addEventListener('fetchStart', (event) => {
		scrollLeft = window.scrollX;
		scrollTop = window.scrollY;
	});
	
	window.addEventListener('fetchEnd', (event) => {
		backoff(5, restoreScroll, 100);
	});
})()

`),
					vx.VXMessageListener().ListenFunc(fmt.Sprintf(`
function(e){
	if (!e.data.split) {
		return
	}
	let arr = e.data.split("_");
	if (arr.length != 2) {
		console.log(arr);
		return
	}
	%s
}`, action))),
				VProgressLinear().
					Attr(":active", "isFetching").
					Attr("style", "position: fixed; z-index: 99").
					Indeterminate(true).
					Height(2).
					Color(pb.GetProgressBarColor()),
				h.Template(
					VSnackbar(h.Text("{{vars.presetsMessage.message}}")).
						Attr("v-model", "vars.presetsMessage.show").
						Attr(":color", "vars.presetsMessage.color").
						Timeout(2000).
						Top(true),
				).Attr("v-if", "vars.presetsMessage"),
				VMain(
					tabContent.Body.(h.HTMLComponent),
				),
			).Id("vt-app").
				Attr(web.InitContextVars, `{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false, presetsListingDialog: false, presetsMessage: {show: false, color: "success", message: ""}}`).
				Attr(web.InitContextLocals, fmt.Sprintf(`{action: "", commonConfirmDialog: false, activeTab:%d, tabs: [{label:"PAGE SETTINGS",query:"settings"},{label:"PAGE CONTENT",query:"content"}]}`, activeTabIndex))
			return
		}
	})

	pm.RegisterEventFunc(openTemplateDialogEvent, openTemplateDialog(db))
	pm.RegisterEventFunc(selectTemplateEvent, selectTemplate(db))
	pm.RegisterEventFunc(clearTemplateEvent, clearTemplate(db))

	eb := pm.Editing("Title", "Slug", "CategoryID")
	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Page)
		err = pageValidator(c, db)
		return
	})
	eb.Field("Slug").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return VTextField().FieldName(field.Name).Label(field.Label).Value(field.Value(obj)).
			ErrorMessages(vErr.GetFieldErrors("Page.Slug")...)
	})
	eb.Field("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		categories := []*Category{}
		if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
			panic(err)
		}
		var showURLComp h.HTMLComponent
		if p.ID != 0 && p.GetStatus() == publish.StatusOnline {
			var u string
			domain := os.Getenv("PUBLISH_URL")
			if p.OnlineUrl != "" {
				u = domain + p.getAccessUrl(p.OnlineUrl)
			} else {
				var c Category
				for _, e := range categories {
					if e.ID == p.CategoryID {
						c = *e
						break
					}
				}

				var localPath string
				if l10nB != nil {
					localPath = l10nB.GetLocalePath(p.LocaleCode)
				}
				u = domain + p.getAccessUrl(p.getPublishUrl(localPath, c.Path))
			}

			showURLComp = h.Div(
				h.A().Text(u).Href(u).Target("_blank"),
			).Class("mb-4")
		}

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		return h.Div(
			showURLComp,
			VAutocomplete().Label(msgr.Category).FieldName(field.Name).MenuProps("top").
				Items(categories).Value(p.CategoryID).ItemText("Path").ItemValue("ID").
				ErrorMessages(vErr.GetFieldErrors("Page.Category")...),
		).ClassIf("mb-4", p.ID != 0)
	})

	eb.Field("TemplateSelection").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		// Display template selection only when creating a new page
		if p.ID == 0 {
			return h.Div(
				web.Portal().Name(templateSelectPortal),
				web.Portal().Name(selectedTemplatePortal),
				VRow(
					VCol(
						VBtn(msgr.CreateFromTemplate).Color("primary").
							Attr("@click", web.Plaid().EventFunc(openTemplateDialogEvent).Go()),
					),
				),
			).Class("my-2").Attr(web.InitContextVars, `{showTemplateDialog: false}`)
		}
		return nil
	})

	eb.Field("EditContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		p := obj.(*Page)
		if p.ID == 0 {
			return nil
		}
		if p.GetStatus() == publish.StatusDraft {
			var href = fmt.Sprintf("%s/editors/%d?version=%s", b.prefix, p.ID, p.GetVersion())
			if locale, isLocalizable := l10n.IsLocalizableFromCtx(ctx.R.Context()); isLocalizable && l10nON {
				href = fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, p.ID, p.GetVersion(), locale)
			}
			return h.Div(
				VBtn(msgr.EditPageContent).
					Target("_blank").
					Href(href).
					Color("secondary"),
			)
		} else {
			var href = fmt.Sprintf("%s/preview?id=%d&version=%s", b.prefix, p.ID, p.GetVersion())
			if locale, isLocalizable := l10n.IsLocalizableFromCtx(ctx.R.Context()); isLocalizable && l10nON {
				href = fmt.Sprintf("%s/preview?id=%d&version=%s&locale=%s", b.prefix, p.ID, p.GetVersion(), locale)
			}
			return h.Div(
				VBtn(msgr.Preview).
					Target("_blank").
					Href(href).
					Color("secondary"),
			)
		}
		return nil
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		localeCode, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		p := obj.(*Page)
		if p.Slug != "" {
			p.Slug = path.Clean(p.Slug)
		}

		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if inerr = gorm2op.DataOperator(tx).Save(obj, id, ctx); inerr != nil {
				return
			}

			if strings.Contains(ctx.R.RequestURI, pv.SaveNewVersionEvent) {
				if inerr = b.copyContainersToNewPageVersion(tx, int(p.ID), p.GetLocale(), p.ParentVersion, p.GetVersion()); inerr != nil {
					return
				}
				return
			}

			if v := ctx.R.FormValue(templateSelectionID); v != "" {
				var tplID int
				tplID, inerr = strconv.Atoi(v)
				if inerr != nil {
					return
				}
				if !l10nON {
					localeCode = ""
				}
				if inerr = b.copyContainersToAnotherPage(tx, tplID, templateVersion, localeCode, int(p.ID), p.GetVersion(), localeCode); inerr != nil {
					panic(inerr)
					return
				}
			}
			if l10nON && strings.Contains(ctx.R.RequestURI, l10n_view.DoLocalize) {
				fromID := ctx.R.Context().Value(l10n_view.FromID).(string)
				fromVersion := ctx.R.Context().Value(l10n_view.FromVersion).(string)
				fromLocale := ctx.R.Context().Value(l10n_view.FromLocale).(string)

				var fromIDInt int
				fromIDInt, err = strconv.Atoi(fromID)
				if err != nil {
					return
				}

				if inerr = b.localizeContainersToAnotherPage(tx, fromIDInt, fromVersion, fromLocale, int(p.ID), p.GetVersion(), p.GetLocale()); inerr != nil {
					panic(inerr)
					return
				}
				return
			}
			return
		})

		return
	})

	sharedContainerM := b.ConfigSharedContainer(pb, db)
	demoContainerM := b.ConfigDemoContainer(pb, db)
	templateM := b.ConfigTemplate(pb, db)
	categoryM := b.ConfigCategory(pb, db)

	if activityB != nil {
		activityB.RegisterModels(pm, sharedContainerM, demoContainerM, templateM, categoryM)
	}
	if l10nB != nil {
		l10n_view.Configure(pb, db, l10nB, activityB, pm, demoContainerM, templateM)
	}
	return
}

// cats should be ordered by path
func fillCategoryIndentLevels(cats []*Category) {
	for i, cat := range cats {
		if cat.Path == "/" {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			if strings.HasPrefix(cat.Path, cats[j].Path+"/") {
				cat.IndentLevel = cats[j].IndentLevel + 1
				break
			}
		}
	}
}

func (b *Builder) ConfigCategory(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&Category{}).URIName("page_categories").Label("Categories")

	lb := pm.Listing("Name", "Path", "Description")

	oldSearcher := lb.Searcher
	lb.SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		r, totalCount, err = oldSearcher(model, params, ctx)
		cats := r.([]*Category)
		sort.Slice(cats, func(i, j int) bool {
			return cats[i].Path < cats[j].Path
		})
		fillCategoryIndentLevels(cats)
		return
	})

	lb.Field("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cat := obj.(*Category)

		icon := "folder"
		if cat.IndentLevel != 0 {
			icon = "insert_drive_file"
		}

		return h.Td(
			h.Div(
				VIcon(icon).Small(true).Class("mb-1"),
				h.Text(cat.Name),
			).Style(fmt.Sprintf("padding-left: %dpx;", cat.IndentLevel*32)),
		)
	})

	eb := pm.Editing("Name", "Path", "Description")

	eb.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		var count int64
		if err = db.Model(&Page{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
			return
		}
		if count > 0 {
			err = errors.New(unableDeleteCategoryMsg)
			return
		}
		if err = db.Model(&Category{}).Where("id = ?", id).Delete(&Category{}).Error; err != nil {
			return
		}
		return
	})

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Category)
		err = categoryValidator(c, db)
		return
	})

	eb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		category := obj.(*Category)

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return VTextField().Label("Path").Value(category.Path).Class("mb-2").FieldName("Path").
			ErrorMessages(vErr.GetFieldErrors("Category.Category")...)
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		c := obj.(*Category)
		c.Path = path.Clean(c.Path)
		err = db.Save(c).Error
		return
	})

	return
}

const (
	templateSelectPortal   = "templateSelectPortal"
	selectedTemplatePortal = "selectedTemplatePortal"

	templateSelectionID     = "TemplateSelectionID"
	templateSelectionLocale = "TemplateSelectionLocale"
	templateUnselectVal     = "unselect"
)

func selectTemplate(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		defer func() {
			web.AppendVarsScripts(&er, "vars.showTemplateDialog=false")
		}()

		id := ctx.R.FormValue(templateSelectionID)
		if id == templateUnselectVal {
			er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
				Name: selectedTemplatePortal,
				Body: h.Input("").Type("hidden").
					Value("").
					Attr(web.VFieldName(templateSelectionID)...),
			})
			return
		}
		locale := ctx.R.FormValue(templateSelectionLocale)

		tpl := Template{}
		if err = db.Model(&Template{}).Where("id = ? AND locale_code = ?", id, locale).First(&tpl).Error; err != nil {
			panic(err)
		}
		name := tpl.Name
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: selectedTemplatePortal,
			Body: VRow(
				VCol(
					h.Input("").Type("hidden").
						Value(id).
						Attr(web.VFieldName(templateSelectionID)...),
					vx.VXReadonlyField().
						Label(msgr.SelectedTemplateLabel).
						Children(
							h.Text(fmt.Sprintf("%v (ID: %v)", name, id)),
							VBtn("").Children(
								VIcon("close"),
							).Text(true).Fab(true).Small(true).
								Class("ml-2").
								Attr("@click", web.Plaid().EventFunc(clearTemplateEvent).Go()),
						),
				),
			).Class("mb-n4"),
		})

		return
	}
}

func clearTemplate(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: selectedTemplatePortal,
			Body: h.Input("").Type("hidden").
				Value("").
				Attr(web.VFieldName(templateSelectionID)...),
		})
		return
	}
}

func openTemplateDialog(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		gmsgr := presets.MustGetMessages(ctx.R)
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())

		tpls := []*Template{}
		if err := db.Model(&Template{}).Where("locale_code = ?", locale).Find(&tpls).Error; err != nil {
			panic(err)
		}

		var tplHTMLComponents []h.HTMLComponent
		tplHTMLComponents = append(tplHTMLComponents,
			h.Div(
				h.Input(templateSelectionID).Type("radio").
					Value(templateUnselectVal).
					Attr(web.VFieldName(templateSelectionID)...).
					Attr("checked", "checked"),
			).Style("visibility:hidden;width:0;height:0;"),
		)
		for _, tpl := range tpls {
			tplHTMLComponents = append(tplHTMLComponents,
				getTplColComponent(ctx, tpl),
			)
		}
		if len(tpls) == 0 {
			tplHTMLComponents = append(tplHTMLComponents,
				h.Div(h.Text(gmsgr.ListingNoRecordToShow)).Class("pl-4 text-center grey--text text--darken-2"),
			)
		}
		msgrPb := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: templateSelectPortal,
			Body: VDialog(
				VCard(
					VCardTitle(
						h.Text(msgrPb.CreateFromTemplate),
						VSpacer(),
						VBtn("").Icon(true).
							Children(VIcon("close")).
							Large(true).
							On("click", fmt.Sprintf("vars.showTemplateDialog=false")),
					),
					VCardActions(
						VRow(tplHTMLComponents...),
					),
					VCardActions(
						VSpacer(),
						VBtn(gmsgr.Cancel).Attr("@click", "vars.showTemplateDialog=false"),
						VBtn(gmsgr.OK).Color("primary").
							Attr("@click", web.Plaid().EventFunc(selectTemplateEvent).
								Query(templateSelectionLocale, locale).
								Go(),
							),
					).Class("pb-4"),
				).Tile(true),
			).MaxWidth("80%").
				Attr("v-model", fmt.Sprintf("vars.showTemplateDialog")).
				Attr(web.InitContextVars, fmt.Sprintf(`{showTemplateDialog: false}`)),
		})

		er.VarsScript = `setTimeout(function(){ vars.showTemplateDialog = true }, 100)`
		return
	}
}

func getTplColComponent(ctx *web.EventContext, tpl *Template) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	// Avoid layout errors
	var name string
	var desc string
	if tpl.Name == "" {
		name = msgr.Unnamed
	} else {
		name = tpl.Name
	}
	if tpl.Description == "" {
		desc = msgr.NotDescribed
	} else {
		desc = tpl.Description
	}

	return VCol(
		VCard(
			h.Div(
				h.Iframe().Src(fmt.Sprintf("./page_builder/preview?id=%d&tpl=1&locale=%s", tpl.ID, tpl.LocaleCode)).
					Attr("width", "100%", "height", "150", "frameborder", "no").
					Style("transform-origin: left top; transform: scale(1, 1); pointer-events: none;"),
			),
			VCardTitle(h.Text(name)),
			VCardSubtitle(h.Text(desc)),
			VBtn(msgr.Preview).Text(true).XSmall(true).Class("ml-2 mb-4").
				Href(fmt.Sprintf("./page_builder/preview?id=%d&tpl=1&locale=%s", tpl.ID, tpl.LocaleCode)).
				Target("_blank").Color("primary"),
			h.Div(
				h.Input(templateSelectionID).Type("radio").
					Value(fmt.Sprintf("%d", tpl.ID)).
					Attr(web.VFieldName(templateSelectionID)...).
					Style("width: 18px; height: 18px"),
			).Class("mr-4 float-right"),
		).Height(280).Class("text-truncate").Outlined(true),
	).Cols(3)
}

func (b *Builder) ConfigSharedContainer(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&Container{}).URIName("shared_containers").Label("Shared Containers")

	pm.RegisterEventFunc(republishRelatedOnlinePagesEvent, republishRelatedOnlinePages(b.mb.Info().ListingHref()))

	listing := pm.Listing("DisplayName").SearchColumns("display_name")
	listing.RowMenu("").Empty()
	// ed := pm.Editing("SelectContainer")
	// ed.Field("SelectContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	//	var containers []h.HTMLComponent
	//	for _, builder := range b.containerBuilders {
	//		cover := builder.cover
	//		if cover == "" {
	//			cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
	//		}
	//		containers = append(containers,
	//			VCol(
	//				VCard(
	//					VImg().Src(cover).Height(200),
	//					VCardActions(
	//						VCardTitle(h.Text(builder.name)),
	//						VSpacer(),
	//						VBtn("Select").
	//							Text(true).
	//							Color("primary").Attr("@click",
	//							web.Plaid().
	//								EventFunc(actions.New).
	//								URL(builder.GetModelBuilder().Info().ListingHref()).
	//								Go()),
	//					),
	//				),
	//			).Cols(6),
	//		)
	//	}
	//	return VSheet(
	//		VContainer(
	//			VRow(
	//				containers...,
	//			),
	//		),
	//	)
	// })
	if permB := pb.GetPermission(); permB != nil {
		permB.CreatePolicies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:shared_containers:*"),
		)
	}
	listing.Field("DisplayName").Label("Name")
	listing.SearchFunc(sharedContainerSearcher(db, pm))
	listing.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		tdbind := cell
		c := obj.(*Container)

		tdbind.SetAttr("@click.self",
			web.Plaid().
				EventFunc(actions.Edit).
				URL(b.ContainerByName(c.ModelName).GetModelBuilder().Info().ListingHref()).
				Query(presets.ParamID, c.ModelID).
				Query(paramOpenFromSharedContainer, 1).
				Go()+fmt.Sprintf(`; vars.currEditingListItemID="%s-%d"`, dataTableID, c.ModelID))

		return tdbind
	})
	return
}

func (b *Builder) ConfigDemoContainer(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&DemoContainer{}).URIName("demo_containers").Label("Demo Containers")

	pm.RegisterEventFunc("addDemoContainer", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		modelID := ctx.QueryAsInt(presets.ParamOverlayUpdateID)
		modelName := ctx.R.FormValue("ModelName")
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		var existID uint
		{
			m := DemoContainer{}
			db.Where("model_name = ?", modelName).First(&m)
			existID = m.ID
		}
		db.Assign(DemoContainer{
			Model: gorm.Model{
				ID: existID,
			},
			ModelID: uint(modelID),
		}).FirstOrCreate(&DemoContainer{}, map[string]interface{}{
			"model_name":  modelName,
			"locale_code": locale,
		})
		r.Reload = true
		return
	})
	listing := pm.Listing("ModelName").SearchColumns("ModelName")
	listing.Field("ModelName").Label("Name")
	ed := pm.Editing("SelectContainer").ActionsFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent { return nil })
	ed.Field("ModelName")
	ed.Field("ModelID")
	ed.Field("SelectContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		locale, localizable := l10n.IsLocalizableFromCtx(ctx.R.Context())

		var demoContainers []DemoContainer
		db.Find(&demoContainers)

		var containers []h.HTMLComponent
		for _, builder := range b.containerBuilders {
			cover := builder.cover
			if cover == "" {
				cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
			}
			c := VCol(
				VCard(
					VImg().Src(cover).Height(200),
					VCardActions(
						VCardTitle(h.Text(builder.name)),
						VSpacer(),
						VBtn("Select").
							Text(true).
							Color("primary").Attr("@click",
							web.Plaid().
								EventFunc(actions.New).
								URL(builder.GetModelBuilder().Info().ListingHref()).
								Query(presets.ParamOverlayAfterUpdateScript, web.POST().Query("ModelName", builder.name).EventFunc("addDemoContainer").Go()).
								Go()),
					),
				),
			).Cols(6)

			var isExists bool
			var modelID uint
			for _, dc := range demoContainers {
				if dc.ModelName == builder.name {
					if localizable && dc.GetLocale() != locale {
						continue
					}
					isExists = true
					modelID = dc.ModelID
					break
				}
			}
			if isExists {
				c = VCol(
					VCard(
						VImg().Src(cover).Height(200),
						VCardActions(
							VCardTitle(h.Text(builder.name)),
							VSpacer(),
							VBtn("Edit").
								Text(true).
								Color("primary").Attr("@click",
								web.Plaid().
									EventFunc(actions.Edit).
									URL(builder.GetModelBuilder().Info().ListingHref()).
									Query(presets.ParamID, fmt.Sprint(modelID)).
									Go()),
						),
					),
				).Cols(6)
			}

			containers = append(containers, c)
		}
		return VSheet(
			VContainer(
				VRow(
					containers...,
				),
			),
		)
	})

	listing.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		tdbind := cell
		c := obj.(*DemoContainer)

		tdbind.SetAttr("@click.self",
			web.Plaid().
				EventFunc(actions.Edit).
				URL(b.ContainerByName(c.ModelName).GetModelBuilder().Info().ListingHref()).
				Query(presets.ParamID, c.ModelID).
				Go()+fmt.Sprintf(`; vars.currEditingListItemID="%s-%d"`, dataTableID, c.ModelID))

		return tdbind
	})

	ed.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		this := obj.(*DemoContainer)
		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if l10nON && strings.Contains(ctx.R.RequestURI, l10n_view.DoLocalize) {
				if inerr = b.createModelAfterLocalizeDemoContainer(tx, this); inerr != nil {
					panic(inerr)
					return
				}
			}

			if inerr = gorm2op.DataOperator(tx).Save(this, id, ctx); inerr != nil {
				return
			}
			return
		})

		return
	})
	return
}

func (b *Builder) ConfigTemplate(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&Template{}).URIName("page_templates").Label("Templates")

	pm.Listing("ID", "Name", "Description")

	eb := pm.Editing("Name", "Description", "EditContainer")
	eb.Field("EditContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		m := obj.(*Template)
		if m.ID == 0 {
			return nil
		}

		var href = fmt.Sprintf("%s/editors/%d?tpl=1", b.prefix, m.ID)
		if locale, isLocalizable := l10n.IsLocalizableFromCtx(ctx.R.Context()); isLocalizable && l10nON {
			href = fmt.Sprintf("%s/editors/%d?tpl=1&locale=%s", b.prefix, m.ID, locale)
		}
		return h.Div(
			VBtn(msgr.EditPageContent).
				Target("_blank").
				Href(href).
				Color("secondary"),
		)
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		this := obj.(*Template)
		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if inerr = gorm2op.DataOperator(tx).Save(obj, id, ctx); inerr != nil {
				return
			}

			if l10nON && strings.Contains(ctx.R.RequestURI, l10n_view.DoLocalize) {
				fromID := ctx.R.Context().Value(l10n_view.FromID).(string)
				fromLocale := ctx.R.Context().Value(l10n_view.FromLocale).(string)

				var fromIDInt int
				fromIDInt, err = strconv.Atoi(fromID)
				if err != nil {
					return
				}

				if inerr = b.localizeContainersToAnotherPage(tx, fromIDInt, "tpl", fromLocale, int(this.ID), "tpl", this.GetLocale()); inerr != nil {
					panic(inerr)
					return
				}
				return
			}
			return
		})

		return
	})

	return
}

func sharedContainerSearcher(db *gorm.DB, mb *presets.ModelBuilder) presets.SearchFunc {
	return func(obj interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		ilike := "ILIKE"
		if db.Dialector.Name() == "sqlite" {
			ilike = "LIKE"
		}

		wh := db.Model(obj)
		if len(params.KeywordColumns) > 0 && len(params.Keyword) > 0 {
			var segs []string
			var args []interface{}
			for _, c := range params.KeywordColumns {
				segs = append(segs, fmt.Sprintf("%s %s ?", c, ilike))
				args = append(args, fmt.Sprintf("%%%s%%", params.Keyword))
			}
			wh = wh.Where(strings.Join(segs, " OR "), args...)
		}

		for _, cond := range params.SQLConditions {
			wh = wh.Where(strings.Replace(cond.Query, " ILIKE ", " "+ilike+" ", -1), cond.Args...)
		}

		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		var c int64
		if err = wh.Select("count(display_name)").Where("shared = true AND locale_code = ?", locale).Group("display_name,model_name,model_id").Count(&c).Error; err != nil {
			return
		}
		totalCount = int(c)

		if params.PerPage > 0 {
			wh = wh.Limit(int(params.PerPage))
			page := params.Page
			if page == 0 {
				page = 1
			}
			offset := (page - 1) * params.PerPage
			wh = wh.Offset(int(offset))
		}

		if err = wh.Select("display_name,model_name,model_id").Find(obj).Error; err != nil {
			return
		}
		r = reflect.ValueOf(obj).Elem().Interface()
		return
	}
}

func (b *Builder) ContainerByName(name string) (r *ContainerBuilder) {
	for _, cb := range b.containerBuilders {
		if cb.name == name {
			return cb
		}
	}
	panic(fmt.Sprintf("No container: %s", name))
}

type ContainerBuilder struct {
	builder    *Builder
	name       string
	mb         *presets.ModelBuilder
	model      interface{}
	modelType  reflect.Type
	renderFunc RenderFunc
	cover      string
}

func (b *Builder) RegisterContainer(name string) (r *ContainerBuilder) {
	r = &ContainerBuilder{
		name:    name,
		builder: b,
	}
	b.containerBuilders = append(b.containerBuilders, r)
	return
}

func (b *ContainerBuilder) Model(m interface{}) *ContainerBuilder {
	b.model = m
	b.mb = b.builder.ps.Model(m)

	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Ptr {
		panic("model pointer type required")
	}

	b.modelType = val.Elem().Type()

	b.configureRelatedOnlinePagesTab()
	return b
}

func (b *ContainerBuilder) URIName(uri string) *ContainerBuilder {
	if b.mb == nil {
		return b
	}
	b.mb.URIName(uri)
	return b
}

func (b *ContainerBuilder) GetModelBuilder() *presets.ModelBuilder {
	return b.mb
}

func (b *ContainerBuilder) RenderFunc(v RenderFunc) *ContainerBuilder {
	b.renderFunc = v
	return b
}

func (b *ContainerBuilder) Cover(v string) *ContainerBuilder {
	b.cover = v
	return b
}

func (b *ContainerBuilder) NewModel() interface{} {
	return reflect.New(b.modelType).Interface()
}

func (b *ContainerBuilder) ModelTypeName() string {
	return b.modelType.String()
}

func (b *ContainerBuilder) Editing(vs ...interface{}) *presets.EditingBuilder {
	return b.mb.Editing(vs...)
}

func (b *ContainerBuilder) configureRelatedOnlinePagesTab() {
	eb := b.mb.Editing()
	eb.AppendTabsPanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		if ctx.R.FormValue(paramOpenFromSharedContainer) != "1" {
			return nil
		}

		pmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		id, err := reflectutils.Get(obj, "id")
		if err != nil {
			panic(err)
		}

		pages := []*Page{}
		pageTable := (&Page{}).TableName()
		containerTable := (&Container{}).TableName()
		err = b.builder.db.Model(&Page{}).
			Joins(fmt.Sprintf(`inner join %s on 
        %s.id = %s.page_id
        and %s.version = %s.page_version
        and %s.locale_code = %s.locale_code`,
				containerTable,
				pageTable, containerTable,
				pageTable, containerTable,
				pageTable, containerTable,
			)).
			// FIXME: add container locale condition after container supports l10n
			Where(fmt.Sprintf(`%s.status = ? and %s.model_id = ? and %s.model_name = ?`,
				pageTable,
				containerTable,
				containerTable,
			), publish.StatusOnline, id, b.name).
			Group(fmt.Sprintf(`%s.id,%s.version,%s.locale_code`, pageTable, pageTable, pageTable)).
			Find(&pages).
			Error
		if err != nil {
			panic(err)
		}

		var pageIDs []string
		var pageListComps h.HTMLComponents
		for _, p := range pages {
			pid := p.PrimarySlug()
			pageIDs = append(pageIDs, pid)
			statusVar := fmt.Sprintf(`republish_status_%s`, strings.Replace(pid, "-", "_", -1))
			pageListComps = append(pageListComps, VListItem(
				h.Text(fmt.Sprintf("%s (%s)", p.Title, pid)),
				VSpacer(),
				VIcon(fmt.Sprintf(`{{vars.%s}}`, statusVar)),
			).
				Dense(true).
				Attr(web.InitContextVars, fmt.Sprintf(`{%s: ""}`, statusVar)))
		}

		return h.Components(
			VTab(h.Text(msgr.RelatedOnlinePages)),
			VTabItem(
				h.If(len(pages) > 0,
					VList(pageListComps),
					h.Div(
						VSpacer(),
						VBtn(msgr.RepublishAllRelatedOnlinePages).
							Color("primary").
							Attr("@click",
								web.Plaid().
									EventFunc(presets.OpenConfirmDialog).
									Query(presets.ConfirmDialogConfirmEvent,
										web.Plaid().
											EventFunc(republishRelatedOnlinePagesEvent).
											Query("ids", strings.Join(pageIDs, ",")).
											Go(),
									).
									Go(),
							),
					).Class("d-flex"),
				).Else(
					h.Div(h.Text(pmsgr.ListingNoRecordToShow)).Class("text-center grey--text text--darken-2 mt-8"),
				),
			),
		)
	})
}

func republishRelatedOnlinePages(pageURL string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ids := strings.Split(ctx.R.FormValue("ids"), ",")
		for _, id := range ids {
			statusVar := fmt.Sprintf(`republish_status_%s`, strings.Replace(id, "-", "_", -1))
			web.AppendVarsScripts(&r,
				web.Plaid().
					URL(pageURL).
					EventFunc(views.RepublishEvent).
					Query("id", id).
					Query(views.ParamScriptAfterPublish, fmt.Sprintf(`vars.%s = "done"`, statusVar)).
					Query("status_var", statusVar).
					Go(),
				fmt.Sprintf(`vars.%s = "pending"`, statusVar),
			)
		}
		return r, nil
	}
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Index(r.RequestURI, b.prefix+"/preview") >= 0 {
		b.preview.ServeHTTP(w, r)
		return
	}

	if strings.Index(r.RequestURI, path.Join(b.prefix, b.imagesPrefix)) >= 0 {
		b.images.ServeHTTP(w, r)
		return
	}
	b.ps.ServeHTTP(w, r)
}
