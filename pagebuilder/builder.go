package pagebuilder

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/qor5/admin/richeditor"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/l10n"
	l10n_view "github.com/qor5/admin/l10n/views"
	mediav "github.com/qor5/admin/media/views"
	"github.com/qor5/admin/note"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	"github.com/qor5/admin/publish/views"
	pv "github.com/qor5/admin/publish/views"
	"github.com/qor5/admin/seo"
	"github.com/qor5/admin/utils"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/perm"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type RenderInput struct {
	Page       *Page
	IsEditor   bool
	IsReadonly bool
	Device     string
}

type RenderFunc func(obj interface{}, input *RenderInput, ctx *web.EventContext) h.HTMLComponent

type PageLayoutFunc func(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent

type PageLayoutInput struct {
	Page              *Page
	SeoTags           h.HTMLComponent
	CanonicalLink     h.HTMLComponent
	StructuredData    h.HTMLComponent
	FreeStyleCss      []string
	FreeStyleTopJs    []string
	FreeStyleBottomJs []string
	Hreflang          map[string]string
	Header            h.HTMLComponent
	Footer            h.HTMLComponent
	IsEditor          bool
	EditorCss         []h.HTMLComponent
	IsPreview         bool
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
	seoBuilder        *seo.Builder
	imagesPrefix      string
	defaultDevice     string
	publishBtnColor   string
	duplicateBtnColor string
	templateEnabled   bool
}

const (
	openTemplateDialogEvent = "openTemplateDialogEvent"
	selectTemplateEvent     = "selectTemplateEvent"
	// clearTemplateEvent               = "clearTemplateEvent"
	republishRelatedOnlinePagesEvent = "republish_related_online_pages"

	schedulePublishDialogEvent = "schedulePublishDialogEvent"
	schedulePublishEvent       = "schedulePublishEvent"

	createNoteDialogEvent = "createNoteDialogEvent"
	createNoteEvent       = "createNoteEvent"

	editSEODialogEvent = "editSEODialogEvent"
	updateSEOEvent     = "updateSEOEvent"

	selectVersionEvent       = "selectVersionEvent"
	renameVersionDialogEvent = "renameVersionDialogEvent"
	renameVersionEvent       = "renameVersionEvent"
	deleteVersionDialogEvent = "deleteVersionDialogEvent"

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
		db:                db,
		wb:                web.New(),
		prefix:            prefix,
		defaultDevice:     DeviceComputer,
		publishBtnColor:   "primary",
		duplicateBtnColor: "primary",
		templateEnabled:   true,
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
	r.ps.GetWebBuilder().RegisterEventFunc(RenameContainerDialogEvent, r.RenameContainerDialog)
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

func (b *Builder) PublishBtnColor(v string) (r *Builder) {
	b.publishBtnColor = v
	return b
}

func (b *Builder) DuplicateBtnColor(v string) (r *Builder) {
	b.duplicateBtnColor = v
	return b
}

func (b *Builder) TemplateEnabled(v bool) (r *Builder) {
	b.templateEnabled = v
	return b
}

func (b *Builder) Configure(pb *presets.Builder, db *gorm.DB, l10nB *l10n.Builder, activityB *activity.ActivityBuilder, publisher *publish.Builder, seoBuilder *seo.Builder) (pm *presets.ModelBuilder) {
	pb.I18n().
		RegisterForModule(language.English, I18nPageBuilderKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPageBuilderKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPageBuilderKey, Messages_ja_JP)

	pb.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	pb.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
	pm = pb.Model(&Page{})
	b.seoBuilder = seoBuilder

	templateM := presets.NewModelBuilder(pb, &Template{})
	if b.templateEnabled {
		templateM = b.ConfigTemplate(pb, db)
	}

	b.mb = pm
	lb := pm.Listing("ID", "Online", "Title", "Path")
	lb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		page := obj.(*Page)
		category, err := page.GetCategory(db)
		if err != nil {
			panic(err)
		}
		return h.Td(h.Text(page.getAccessUrl(page.getPublishUrl(l10nB.GetLocalePath(page.LocaleCode), category.Path))))
	})

	dp := pm.Detailing("Overview")
	dp.Field("Overview").ComponentFunc(settings(db, pm))

	oldDetailLayout := pb.GetDetailLayoutFunc()
	pb.DetailLayoutFunc(func(in web.PageFunc, cfg *presets.LayoutConfig) (out web.PageFunc) {
		return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
			if !strings.Contains(ctx.R.RequestURI, "/"+pm.Info().URIName()+"/") && !strings.Contains(ctx.R.RequestURI, "/"+templateM.Info().URIName()+"/") {
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

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, utils.Messages_en_US).(*utils.Messages)
			pvMsgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, pv.Messages_en_US).(*pv.Messages)
			id := pat.Param(ctx.R, "id")

			if id == "" {
				return pb.DefaultNotFoundPageFunc(ctx)
			}

			isPage := strings.Contains(ctx.R.RequestURI, "/"+pm.Info().URIName()+"/")
			isTemplate := strings.Contains(ctx.R.RequestURI, "/"+templateM.Info().URIName()+"/")
			var obj interface{}
			var dmb *presets.ModelBuilder
			if isPage {
				dmb = pm
			} else {
				dmb = templateM
			}
			obj = dmb.NewModel()
			if sd, ok := obj.(presets.SlugDecoder); ok {
				vs, err := presets.RecoverPrimaryColumnValuesBySlug(sd, id)
				if err != nil {
					return pb.DefaultNotFoundPageFunc(ctx)
				}
				if _, err := strconv.Atoi(vs["id"]); err != nil {
					return pb.DefaultNotFoundPageFunc(ctx)
				}
			} else {
				if _, err := strconv.Atoi(id); err != nil {
					return pb.DefaultNotFoundPageFunc(ctx)
				}
			}
			obj, err = dmb.Detailing().GetFetchFunc()(obj, id, ctx)
			if err != nil {
				if errors.Is(err, presets.ErrRecordNotFound) {
					return pb.DefaultNotFoundPageFunc(ctx)
				}
				return
			}

			if l, ok := obj.(l10n.L10nInterface); ok {
				locale := ctx.R.FormValue("locale")
				if ctx.R.FormValue(web.EventFuncIDName) == "__reload__" && locale != "" && locale != l.GetLocale() {
					// redirect to list page when change locale
					http.Redirect(ctx.W, ctx.R, dmb.Info().ListingHref(), http.StatusSeeOther)
					return
				}
			}

			var tabContent web.PageResponse
			tab := ctx.R.FormValue("tab")
			isContent := tab == "content"
			activeTabIndex := 0
			isVersion := false
			if _, ok := obj.(publish.VersionInterface); ok {
				isVersion = true
			}
			if isContent {
				activeTabIndex = 1
				if v, ok := obj.(interface {
					GetID() uint
				}); ok {
					ctx.R.Form.Set("id", strconv.Itoa(int(v.GetID())))
				}
				if v, ok := obj.(publish.VersionInterface); ok {
					ctx.R.Form.Set("version", v.GetVersion())
				}
				if l, ok := obj.(l10n.L10nInterface); ok {
					ctx.R.Form.Set("locale", l.GetLocale())
				}
				if isTemplate {
					ctx.R.Form.Set("tpl", "1")
				}
				tabContent, err = b.PageContent(ctx)
			} else {
				tabContent, err = in(ctx)
			}
			if errors.Is(err, perm.PermissionDenied) {
				pr.Body = h.Text(perm.PermissionDenied.Error())
				return pr, nil
			}
			if err != nil {
				panic(err)
			}
			device := ctx.R.FormValue("device")
			editAction := web.POST().
				EventFunc(actions.Edit).
				URL(web.Var("\""+b.prefix+"/\"+arr[0]")).
				Query(presets.ParamOverlay, actions.Drawer).
				Query(presets.ParamID, web.Var("arr[1]")).
				Go()

			var publishBtn h.HTMLComponent
			var duplicateBtn h.HTMLComponent
			var versionSwitch h.HTMLComponent
			primarySlug := ""
			if v, ok := obj.(presets.SlugEncoder); ok {
				primarySlug = v.PrimarySlug()
			}
			if isVersion {
				p := obj.(*Page)
				versionCount := versionCount(db, p)
				switch p.GetStatus() {
				case publish.StatusDraft, publish.StatusOffline:
					publishBtn = VBtn(pvMsgr.Publish).Size(SizeSmall).Variant(VariantElevated).Color(b.publishBtnColor).Height(40).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.PublishEvent))
				case publish.StatusOnline:
					publishBtn = VBtn(pvMsgr.Republish).Size(SizeSmall).Variant(VariantElevated).Color(b.publishBtnColor).Height(40).Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, pv.RepublishEvent))
				}
				duplicateBtn = VBtn(msgr.Duplicate).
					Size(SizeSmall).Color(b.duplicateBtnColor).Height(40).Class("rounded-l-0 mr-3").
					Attr("@click", web.Plaid().
						EventFunc(pv.DuplicateVersionEvent).
						URL(pm.Info().ListingHref()).
						Query(presets.ParamID, primarySlug).
						Query("tab", tab).
						QueryIf("device", device, device != "").
						FieldValue("Title", p.Title).FieldValue("Slug", p.Slug).FieldValue("CategoryID", p.CategoryID).
						Go(),
					)
				versionSwitch = VChip(
					VChip(h.Text(fmt.Sprintf("%d", versionCount))).Label(true).Color("#E0E0E0").Size(SizeSmall).Class("px-1 mx-1 text-black").Attr("style", "height:20px"),
					h.Text(p.GetVersionName()+" | "),
					VChip(h.Text(pv.GetStatusText(p.GetStatus(), pvMsgr))).Label(true).Color(pv.GetStatusColor(p.GetStatus())).Size(SizeSmall).Class("px-1  mx-1 text-black").Attr("style", "height:20px"),
					VIcon("chevron_right"),
				).Label(true).Variant(VariantOutlined).Class("px-1 ml-8 rounded-r-0 text-black").Attr("style", "height:40px;background-color:#FFFFFF!important;").
					Attr("@click", web.Plaid().EventFunc(actions.OpenListingDialog).
						URL(b.ps.GetURIPrefix()+"/version-list-dialog").
						Query("select_id", primarySlug).
						Go())
			}

			var innerPr web.PageResponse
			innerPr, err = in(ctx)
			if err == perm.PermissionDenied {
				pr.Body = h.Text(perm.PermissionDenied.Error())
				return pr, nil
			}
			if err != nil {
				panic(err)
			}

			toolbar := VContainer(
				VRow(
					VCol(pb.RunBrandFunc(ctx)).Cols(8),
					VCol(
						pb.RunSwitchLanguageFunc(ctx),
					).Cols(2),

					VCol(
						VAppBarNavIcon().Attr("icon", "mdi-menu").
							Density("compact").
							Class("text-grey-darken-1").
							Attr("@click", "vars.navDrawer = !vars.navDrawer").Density(DensityCompact),
					).Cols(2),
				).Attr("align", "center").Attr("justify", "center"),
			)

			pr.Body = web.Scope(
				VApp(
					VNavigationDrawer(
						VLayout(
							VMain(
								toolbar,
								VCard(
									menu,
								).Class("ma-4").Variant(VariantText),
							),
							//ModelValue(true).
							//	Attr("v-model", "vars.navDrawer").Attr(web.ObjectAssign("vars", `{navDrawer: null}`)...),
							VAppBar(
								profile,
								//VAppBarNavIcon().On("click.stop", "vars.navDrawer = !vars.navDrawer"),
							).Location("bottom").Class("border-t-sm border-b-0").Elevation(0),
						).Class("ma-2 border-sm rounded elevation-1").Attr("style", "height: calc(100% - 16px);"),
					).Width(320).
						ModelValue(true).
						Attr("v-model", "vars.navDrawer").
						Permanent(true).
						Floating(true).
						Elevation(0),

					VAppBar(
						h.Div(
							VProgressLinear().
								Attr(":active", "isFetching").
								Class("ml-4").
								Attr("style", "position: fixed; z-index: 99;").
								Indeterminate(true).
								Height(2).
								Color(pb.GetProgressBarColor()),

							VAppBarNavIcon().
								Density("compact").
								Class("mr-2").
								Attr("v-if", "!vars.navDrawer").
								On("click.stop", "vars.navDrawer = !vars.navDrawer"),
							h.Div(
								VToolbarTitle(innerPr.PageTitle), // Class("text-h6 font-weight-regular"),
							).Class("mr-auto"),
							VTabs(
								VTab(h.Text("{{item.label}}")).Attr("@click", web.Plaid().Query("tab", web.Var("item.query")).PushState(true).Go()).
									Attr("v-for", "(item, index) in locals.tabs", ":key", "index"),
							).CenterActive(true).FixedTabs(true).Attr("v-model", `locals.activeTab`).Attr("style", "width:400px"),
							h.If(isVersion, versionSwitch, duplicateBtn, publishBtn),
						).Class("d-flex align-center mx-4 border-b w-100").Style("height: 48px"),
					).
						Elevation(0).
						Density("compact"),
					web.Portal().Name(presets.RightDrawerPortalName),
					web.Portal().Name(presets.DialogPortalName),
					web.Portal().Name(presets.DeleteConfirmPortalName),
					web.Portal().Name(presets.DefaultConfirmDialogPortalName),
					web.Portal().Name(presets.ListingDialogPortalName),
					web.Portal().Name(dialogPortalName),
					utils.ConfirmDialog(pvMsgr.Areyousure, web.Plaid().EventFunc(web.Var("locals.action")).
						Query(presets.ParamID, primarySlug).Go(),
						utilsMsgr),
					h.If(isContent,
						h.Tag("vx-restore-scroll-listener"),
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
							}`, editAction))),
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
							Timeout(1000),
					).Attr("v-if", "vars.presetsMessage"),
					VMain(
						tabContent.Body.(h.HTMLComponent),
					),
				).Attr("id", "vt-app").Attr(web.ObjectAssign("vars", `{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false, presetsListingDialog: false, presetsMessage: {show: false, color: "success", message: ""}}`)...),
			).VSlot(" { locals } ").Init(fmt.Sprintf(`{action: "", commonConfirmDialog: false, activeTab:%d, tabs: [{label:"SETTINGS",query:"settings"},{label:"CONTENT",query:"content"}]}`, activeTabIndex))
			return
		}
	})

	configureVersionListDialog(db, b.ps, pm)

	if b.templateEnabled {
		pm.RegisterEventFunc(openTemplateDialogEvent, openTemplateDialog(db, b.prefix))
		pm.RegisterEventFunc(selectTemplateEvent, selectTemplate(db))
		// pm.RegisterEventFunc(clearTemplateEvent, clearTemplate(db))
	}
	pm.RegisterEventFunc(schedulePublishDialogEvent, schedulePublishDialog(db, pm))
	pm.RegisterEventFunc(schedulePublishEvent, schedulePublish(db, pm))
	pm.RegisterEventFunc(createNoteDialogEvent, createNoteDialog(db, pm))
	pm.RegisterEventFunc(createNoteEvent, createNote(db, pm))
	pm.RegisterEventFunc(editSEODialogEvent, editSEODialog(db, pm, seoBuilder))
	pm.RegisterEventFunc(updateSEOEvent, updateSEO(db, pm))
	eb := pm.Editing("TemplateSelection", "Title", "CategoryID", "Slug")
	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Page)
		err = pageValidator(ctx.R.Context(), c, db, l10nB)
		return
	})
	eb.Field("Slug").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return VTextField().
			Variant(FieldVariantUnderlined).
			Attr(web.VField(field.Name, strings.TrimPrefix(field.Value(obj).(string), "/"))...).
			Prefix("/").
			ErrorMessages(vErr.GetFieldErrors("Page.Slug")...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		m := obj.(*Page)
		m.Slug = path.Join("/", m.Slug)
		return nil
	})
	eb.Field("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		categories := []*Category{}
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
			panic(err)
		}

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		return vx.VXAutocomplete().Label(msgr.Category).
			Attr(web.VField(field.Name, p.CategoryID)...).
			Multiple(false).Chips(false).
			Items(categories).ItemText("Path").ItemValue("ID").
			ErrorMessages(vErr.GetFieldErrors("Page.Category")...)
	})

	eb.Field("TemplateSelection").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		if !b.templateEnabled {
			return nil
		}

		p := obj.(*Page)

		selectedID := ctx.R.FormValue(templateSelectedID)
		body, err := getTplPortalComp(ctx, db, selectedID)
		if err != nil {
			panic(err)
		}

		// Display template selection only when creating a new page
		if p.ID == 0 {
			return h.Div(
				web.Portal().Name(templateSelectPortal),
				web.Portal(
					body,
				).Name(selectedTemplatePortal),
			).Class("my-2").
				Attr(web.ObjectAssign("vars", `{showTemplateDialog: false}`)...)
		}
		return nil
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		localeCode, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		p := obj.(*Page)
		if p.Slug != "" {
			p.Slug = path.Clean(p.Slug)
		}
		funcName := ctx.R.FormValue(web.EventFuncIDName)
		if funcName == pv.DuplicateVersionEvent {
			id := ctx.R.FormValue(presets.ParamID)
			var fromPage Page
			eb.Fetcher(&fromPage, id, ctx)
			p.SEO = fromPage.SEO
		}

		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if inerr = gorm2op.DataOperator(tx).Save(obj, id, ctx); inerr != nil {
				return
			}

			if strings.Contains(ctx.R.RequestURI, pv.SaveNewVersionEvent) || strings.Contains(ctx.R.RequestURI, pv.DuplicateVersionEvent) {
				if inerr = b.copyContainersToNewPageVersion(tx, int(p.ID), p.GetLocale(), p.ParentVersion, p.GetVersion()); inerr != nil {
					return
				}
				return
			}

			if v := ctx.R.FormValue(templateSelectedID); v != "" {
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

				if inerr = b.localizeCategory(tx, p.CategoryID, fromLocale, p.GetLocale()); inerr != nil {
					panic(inerr)
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
	categoryM := b.ConfigCategory(pb, db, l10nB)

	if activityB != nil {
		activityB.RegisterModels(pm, sharedContainerM, demoContainerM, templateM, categoryM)
	}
	if l10nB != nil {
		l10n_view.Configure(pb, db, l10nB, activityB, pm, demoContainerM, templateM, categoryM)
	}
	if publisher != nil {
		publisher.WithPageBuilder(b)
		pv.Configure(pb, db, activityB, publisher, pm)
		pm.Editing().SidePanelFunc(nil).ActionsFunc(nil)
	}
	if seoBuilder != nil {
		seoBuilder.RegisterSEO("Page", &Page{}).RegisterContextVariable(
			"Title",
			func(object interface{}, _ *seo.Setting, _ *http.Request) string {
				if p, ok := object.(*Page); ok {
					return p.Title
				}
				return ""
			},
		)
	}
	note.Configure(db, pb, pm)
	eb.CleanTabsPanels()
	dp.CleanTabsPanels()
	mediav.Configure(b.GetPresetsBuilder(), db)
	return
}

func versionCount(db *gorm.DB, p *Page) (count int64) {
	db.Model(&Page{}).Where("id = ? and locale_code = ?", p.ID, p.LocaleCode).Count(&count)
	return
}

func configureVersionListDialog(db *gorm.DB, pb *presets.Builder, pm *presets.ModelBuilder) {
	mb := pb.Model(&Page{}).
		URIName("version-list-dialog").
		InMenu(false)
	searcher := mb.Listing().Searcher
	lb := mb.Listing("Version", "State", "Notes", "Option").
		SearchColumns("version", "version_name").
		PerPage(10).
		SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
			id := ctx.R.FormValue("select_id")
			if id == "" {
				id = ctx.R.FormValue("f_select_id")
			}
			if id != "" {
				cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)
				con := presets.SQLCondition{
					Query: "id = ? and locale_code = ?",
					Args:  []interface{}{cs["id"], cs["locale_code"]},
				}
				params.SQLConditions = append(params.SQLConditions, &con)
			}
			params.OrderBy = "created_at DESC"

			return searcher(model, params, ctx)
		})
	lb.Field("Version").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		versionName := obj.(publish.VersionInterface).GetVersionName()
		return h.Td(
			h.Text(versionName),
			VBtn("").Icon("mdi-pencil").Variant(VariantText).Attr("@click", web.Plaid().
				URL(pb.GetURIPrefix()+"/version-list-dialog").
				EventFunc(renameVersionDialogEvent).
				Queries(ctx.Queries()).
				Query(presets.ParamOverlay, actions.Dialog).
				Query("rename_id", obj.(presets.SlugEncoder).PrimarySlug()).
				Query("version_name", versionName).
				Go()),
		)
	})
	lb.Field("State").ComponentFunc(pv.StatusListFunc())
	lb.Field("Notes").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		rt := pm.Info().Label()
		ri := p.PrimarySlug()
		userID, _ := note.GetUserData(ctx)
		count := note.GetUnreadNotesCount(db, userID, rt, ri)

		return h.Td(
			h.If(count > 0,
				VBadge().Content(count).Color("red"),
			).Else(
				h.Text(""),
			),
		)
	}).Label("Unread Notes")
	lb.Field("Option").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}

		versionName := p.GetVersionName()
		if p.PrimarySlug() == id {
			return h.Td(h.Text("current"))
		}
		if p.GetStatus() != publish.StatusDraft {
			return h.Td()
		}

		return h.Td(VBtn("").Icon(true).Children(VIcon("delete")).Attr("@click", web.Plaid().
			URL(pb.GetURIPrefix()+"/version-list-dialog").
			EventFunc(deleteVersionDialogEvent).
			Queries(ctx.Queries()).
			Query(presets.ParamOverlay, actions.Dialog).
			Query("delete_id", obj.(presets.SlugEncoder).PrimarySlug()).
			Query("version_name", versionName).
			Go()))
	})
	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.RowMenu().Empty()
	mb.RegisterEventFunc(selectVersionEvent, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("select_id")
		refer, _ := url.Parse(ctx.R.Referer())
		newQueries := refer.Query()
		r.PushState = web.Location(newQueries).URL(pm.Info().DetailingHref(id))
		return
	})
	mb.RegisterEventFunc(renameVersionDialogEvent, renameVersionDialog(mb))
	mb.RegisterEventFunc(renameVersionEvent, renameVersion(mb))
	mb.RegisterEventFunc(deleteVersionDialogEvent, deleteVersionDialog(mb))

	lb.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		cell.SetAttr("@click.self", web.Plaid().
			Query("select_id", obj.(*Page).PrimarySlug()).
			URL(pb.GetURIPrefix()+"/version-list-dialog").
			EventFunc(selectVersionEvent).
			Go(),
		)
		return cell
	})
	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:          "all",
				Invisible:    true,
				SQLCondition: ``,
			},
			{

				Key:          "online_version",
				Invisible:    true,
				SQLCondition: `status = 'online'`,
			},
			{
				Key:          "named_versions",
				Invisible:    true,
				SQLCondition: `version <> version_name`,
			},
		}
	})

	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		id := ctx.R.FormValue("select_id")
		if id == "" {
			id = ctx.R.FormValue("f_select_id")
		}
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabAllVersions,
				ID:    "all",
				Query: url.Values{"all": []string{"1"}, "select_id": []string{id}},
			},
			{
				Label: msgr.FilterTabOnlineVersion,
				ID:    "online_version",
				Query: url.Values{"online_version": []string{"1"}, "select_id": []string{id}},
			},
			{
				Label: msgr.FilterTabNamedVersions,
				ID:    "named_versions",
				Query: url.Values{"named_versions": []string{"1"}, "select_id": []string{id}},
			},
		}
	})

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

func (b *Builder) ConfigCategory(pb *presets.Builder, db *gorm.DB, l10nB *l10n.Builder) (pm *presets.ModelBuilder) {
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

		icon := "mdi-folder"
		if cat.IndentLevel != 0 {
			icon = "mdi-file"
		}

		return h.Td(
			h.Div(
				VIcon(icon).Size(SizeSmall).Class("mb-1"),
				h.Text(cat.Name),
			).Style(fmt.Sprintf("padding-left: %dpx;", cat.IndentLevel*32)),
		)
	})

	eb := pm.Editing("Name", "Path", "Description")

	eb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return VTextField().Label("Path").
			Variant(FieldVariantUnderlined).
			Attr(web.VField("Path", strings.TrimPrefix(field.Value(obj).(string), "/"))...).
			Prefix("/").
			ErrorMessages(vErr.GetFieldErrors("Category.Category")...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		m := obj.(*Category)
		m.Path = path.Join("/", m.Path)
		return nil
	})

	eb.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		cs := obj.(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)
		ID := cs["id"]
		Locale := cs["locale_code"]

		var count int64
		if err = db.Model(&Page{}).Where("category_id = ? AND locale_code = ?", ID, Locale).Count(&count).Error; err != nil {
			return
		}
		if count > 0 {
			err = errors.New(unableDeleteCategoryMsg)
			return
		}
		if err = db.Model(&Category{}).Where("id = ? AND locale_code = ?", ID, Locale).Delete(&Category{}).Error; err != nil {
			return
		}
		return
	})

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Category)
		err = categoryValidator(c, db, l10nB)
		return
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

	templateSelectedID = "TemplateSelectedID"
	templateID         = "TemplateID"
	templateBlankVal   = "blank"
)

func selectTemplate(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		defer func() {
			web.AppendRunScripts(&er, "vars.showTemplateDialog=false")
		}()

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		id := ctx.R.FormValue(templateID)
		if id == templateBlankVal {
			er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
				Name: selectedTemplatePortal,
				Body: VRow(
					VCol(
						h.Input("").Type("hidden").Attr(web.VField(templateSelectedID, "")...),
						VTextField().Readonly(true).Label(msgr.SelectedTemplateLabel).ModelValue(msgr.Blank).Density(DensityCompact).Variant(VariantOutlined),
					).Cols(5),
					VCol(
						VBtn(msgr.ChangeTemplate).Color("primary").
							Attr("@click", web.Plaid().Query(templateSelectedID, "").EventFunc(openTemplateDialogEvent).Go()),
					).Cols(5),
				),
			})
			return
		}

		var body h.HTMLComponent
		if body, err = getTplPortalComp(ctx, db, id); err != nil {
			return
		}

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: selectedTemplatePortal,
			Body: body,
		})

		return
	}
}

func getTplPortalComp(ctx *web.EventContext, db *gorm.DB, selectedID string) (h.HTMLComponent, error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())

	name := msgr.Blank
	if selectedID != "" {
		if err := db.Model(&Template{}).Where("id = ? AND locale_code = ?", selectedID, locale).Pluck("name", &name).Error; err != nil {
			return nil, err
		}
	}

	return VRow(
		VCol(
			h.Input("").Type("hidden").Attr(web.VField(templateSelectedID, selectedID)...),
			VTextField().Readonly(true).Label(msgr.SelectedTemplateLabel).ModelValue(name).Density(DensityCompact).Variant(VariantOutlined),
		).Cols(5),
		VCol(
			VBtn(msgr.ChangeTemplate).Color("primary").
				Attr("@click", web.Plaid().Query(templateSelectedID, selectedID).EventFunc(openTemplateDialogEvent).Go()),
		).Cols(5),
	), nil
}

// Unused
func clearTemplate(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: selectedTemplatePortal,
			Body: VRow(
				VCol(
					h.Input("").Type("hidden").Attr(web.VField(templateSelectedID, "")...),
					VTextField().Readonly(true).Label(msgr.SelectedTemplateLabel).ModelValue(msgr.Blank).Density(DensityCompact).Variant(VariantOutlined),
				).Cols(5),
				VCol(
					VBtn(msgr.ChangeTemplate).Color("primary").
						Attr("@click", web.Plaid().Query(templateSelectedID, "").EventFunc(openTemplateDialogEvent).Go()),
				).Cols(5),
			),
		})
		return
	}
}

func openTemplateDialog(db *gorm.DB, prefix string) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		gmsgr := presets.MustGetMessages(ctx.R)
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		selectedID := ctx.R.FormValue(templateSelectedID)
		if selectedID == "" {
			selectedID = templateBlankVal
		}

		tpls := []*Template{}
		if err := db.Model(&Template{}).Where("locale_code = ?", locale).Find(&tpls).Error; err != nil {
			panic(err)
		}
		msgrPb := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		var tplHTMLComponents []h.HTMLComponent
		tplHTMLComponents = append(tplHTMLComponents,
			VCol(
				VCard(
					h.Div(
						h.Iframe().
							Attr("width", "100%", "height", "150", "frameborder", "no").
							Style("transform-origin: left top; transform: scale(1, 1); pointer-events: none;"),
					),
					VCardTitle(h.Text(msgrPb.Blank)),
					VCardSubtitle(h.Text("")),
					h.Div(
						h.Input(templateID).Type("radio").
							Value(templateBlankVal).
							Attr(web.VField(templateID, selectedID)...).
							AttrIf("checked", "checked", selectedID == "").
							Style("width: 18px; height: 18px"),
					).Class("mr-4 float-right"),
				).Height(280).Class("text-truncate").Variant(VariantOutlined),
			).Cols(3),
		)
		for _, tpl := range tpls {
			tplHTMLComponents = append(tplHTMLComponents,
				getTplColComponent(ctx, prefix, tpl, selectedID),
			)
		}
		if len(tpls) == 0 {
			tplHTMLComponents = append(tplHTMLComponents,
				h.Div(h.Text(gmsgr.ListingNoRecordToShow)).Class("pl-4 text-center grey--text text--darken-2"),
			)
		}

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: templateSelectPortal,
			Body: VDialog(
				VCard(
					VCardTitle(
						h.Text(msgrPb.CreateFromTemplate),
						VSpacer(),
						VBtn("").Icon("mdi-close").
							Size(SizeLarge).
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
								Go(),
							),
					).Class("pb-4"),
				).Tile(true),
			).MaxWidth("80%").
				Attr(web.ObjectAssign("vars", `{showTemplateDialog: false}`)...).
				Attr("v-model", fmt.Sprintf("vars.showTemplateDialog")),
		})
		er.RunScript = `setTimeout(function(){ vars.showTemplateDialog = true }, 100)`
		return
	}
}

func getTplColComponent(ctx *web.EventContext, prefix string, tpl *Template, selectedID string) h.HTMLComponent {
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
	id := fmt.Sprintf("%d", tpl.ID)

	src := fmt.Sprintf("./%s/preview?id=%d&tpl=1&locale=%s", prefix, tpl.ID, tpl.LocaleCode)

	return VCol(
		VCard(
			h.Div(
				h.Iframe().Src(src).
					Attr("width", "100%", "height", "150", "frameborder", "no").
					Style("transform-origin: left top; transform: scale(1, 1); pointer-events: none;"),
			),
			VCardTitle(h.Text(name)),
			VCardSubtitle(h.Text(desc)),
			VBtn(msgr.Preview).Variant(VariantText).Size(SizeSmall).Class("ml-2 mb-4").
				Href(src).
				Attr("target", "_blank").Color("primary"),
			h.Div(
				h.Input(templateID).Type("radio").
					Value(id).
					Attr(web.VField(templateID, selectedID)...).
					AttrIf("checked", "checked", selectedID == id).
					Style("width: 18px; height: 18px"),
			).Class("mr-4 float-right"),
		).Height(280).Class("text-truncate").Variant(VariantOutlined),
	).Cols(3)
}

func schedulePublishDialog(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		s, ok := obj.(publish.ScheduleInterface)
		if !ok {
			return
		}

		var start, end string
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, Messages_en_US).(*pv.Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		updateBtn := VBtn(cmsgr.Update).
			Color("primary").
			Attr(":disabled", "isFetching").
			Attr(":loading", "isFetching").
			Attr("@click", web.Plaid().
				EventFunc(schedulePublishEvent).
				// Queries(queries).
				Query(presets.ParamID, paramID).
				Query(presets.ParamOverlay, actions.Dialog).
				URL(mb.Info().ListingHref()).
				Go())

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VCardTitle(h.Text("Schedule Publish Time")),
						VCardText(
							VRow(
								VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", start)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
								).Cols(6),
								VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", end)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
								).Cols(6),
							),
						),
						VCardActions(
							VSpacer(),
							updateBtn,
						),
					),
				).MaxWidth("480px").
					Attr("v-model", "locals.schedulePublishDialog"),
			).Init("{schedulePublishDialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func schedulePublish(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		err = pv.ScheduleEditSetterFunc(obj, nil, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		err = mb.Editing().Saver(obj, paramID, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		r.PushState = web.Location(nil)
		return
	}
}

func createNoteDialog(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)

		okAction := web.Plaid().
			URL(mb.Info().ListingHref()).
			EventFunc(createNoteEvent).
			Query("resource_id", paramID).
			Query("resource_type", mb.Info().Label()).
			Query(presets.ParamOverlay, actions.Dialog).Go()

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VCardTitle(h.Text("Note")),
						VCardText(
							VTextField().Variant(FieldVariantUnderlined).Attr(web.VField("Content", "")...),
						),
						VCardActions(
							VSpacer(),
							VBtn("Cancel").
								Variant(VariantFlat).
								Class("ml-2").
								On("click", "locals.createNoteDialog = false"),

							VBtn("OK").
								Color("primary").
								Variant(VariantFlat).
								Theme(ThemeDark).
								Attr("@click", okAction),
						),
					),
				).MaxWidth("420px").Attr("v-model", "locals.createNoteDialog"),
			).Init("{createNoteDialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func createNote(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ri := ctx.R.FormValue("resource_id")
		rt := ctx.R.FormValue("resource_type")
		content := ctx.R.FormValue("Content")

		userID, creator := note.GetUserData(ctx)
		nt := note.QorNote{
			UserID:       userID,
			Creator:      creator,
			ResourceID:   ri,
			ResourceType: rt,
			Content:      content,
		}

		if err = db.Save(&nt).Error; err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			err = nil
			return
		}

		userNote := note.UserNote{UserID: userID, ResourceType: rt, ResourceID: ri}
		db.Where(userNote).FirstOrCreate(&userNote)

		var total int64
		db.Model(&note.QorNote{}).Where("resource_type = ? AND resource_id = ?", rt, ri).Count(&total)
		db.Model(&userNote).UpdateColumn("Number", total)
		r.PushState = web.Location(nil)
		return
	}
}

func editSEODialog(db *gorm.DB, mb *presets.ModelBuilder, seoBuilder *seo.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		// msgr := i18n.MustGetModuleMessages(ctx.R, pv.I18nPublishKey, Messages_en_US).(*pv.Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		updateBtn := VBtn(cmsgr.Update).
			Color("primary").
			Attr(":disabled", "isFetching").
			Attr(":loading", "isFetching").
			Attr("@click", web.Plaid().
				EventFunc(updateSEOEvent).
				// Queries(queries).
				Query(presets.ParamID, paramID).
				// Query(presets.ParamOverlay, actions.Dialog).
				URL(mb.Info().ListingHref()).
				Go())
		ctx.R.Form.Set("hideActionsIconForSEOForm", "true")
		seoForm := seoBuilder.EditingComponentFunc(obj, nil, ctx)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VCardTitle(h.Text("")),
						VCardText(
							seoForm,
						),
						VCardActions(
							VSpacer(),
							updateBtn,
						),
					),
				).MaxWidth("650px").
					Attr("v-model", "locals.editSEODialog"),
			).Init("{editSEODialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func updateSEO(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		err = seo.EditSetterFunc(obj, &presets.FieldContext{Name: "SEO"}, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		err = mb.Editing().Saver(obj, paramID, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		r.PushState = web.Location(nil)
		return
	}
}

func renameVersionDialog(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("rename_id")
		versionName := ctx.R.FormValue("version_name")
		okAction := web.Plaid().
			URL(mb.Info().ListingHref()).
			EventFunc(renameVersionEvent).
			Queries(ctx.Queries()).
			Query("rename_id", id).Go()

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VCardTitle(h.Text("Version")),
						VCardText(
							VTextField().Attr(web.VField("VersionName", versionName)...).Variant(FieldVariantUnderlined),
						),
						VCardActions(
							VSpacer(),
							VBtn("Cancel").
								Variant(VariantFlat).
								Class("ml-2").
								On("click", "locals.renameVersionDialog = false"),

							VBtn("OK").
								Color("primary").
								Variant(VariantFlat).
								Theme(ThemeDark).
								Attr("@click", "locals.renameVersionDialog = false; "+okAction),
						),
					),
				).MaxWidth("420px").Attr("v-model", "locals.renameVersionDialog"),
			).Init("{renameVersionDialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func renameVersion(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue("rename_id")
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		name := ctx.R.FormValue("VersionName")
		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, paramID, ctx); err != nil {
			return
		}
		qs := ctx.Queries()
		delete(qs, "version_name")
		delete(qs, "rename_id")

		r.RunScript = web.Plaid().URL(ctx.R.RequestURI).Queries(qs).EventFunc(actions.UpdateListingDialog).Go()
		return
	}
}

func deleteVersionDialog(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("delete_id")
		versionName := ctx.R.FormValue("version_name")
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DeleteConfirmPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VCardTitle(h.Text(fmt.Sprintf("Are you sure you want to delete %s?", versionName))),
						VCardActions(
							VSpacer(),
							VBtn("Cancel").
								Variant(VariantFlat).
								Class("ml-2").
								On("click", "dialogLocals.deleteConfirmation = false"),

							VBtn("Delete").
								Color("primary").
								Variant(VariantFlat).
								Theme(ThemeDark).
								Attr("@click", web.Plaid().
									URL(mb.Info().ListingHref()).
									EventFunc(actions.DoDelete).
									Queries(ctx.Queries()).
									Query(presets.ParamInDialog, true).
									Query(presets.ParamID, id).Go()),
						),
					),
				).MaxWidth("580px").
					Attr("v-model", "dialogLocals.deleteConfirmation"),
			).VSlot(" { locals: dialogLocals }").Init(`{deleteConfirmation: false}`),
		})

		r.RunScript = "setTimeout(function(){ dialogLocals.deleteConfirmation = true }, 100)"
		return
	}
}

func (b *Builder) ConfigSharedContainer(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&Container{}).URIName("shared_containers").Label("Shared Containers")

	pm.RegisterEventFunc(republishRelatedOnlinePagesEvent, republishRelatedOnlinePages(b.mb.Info().ListingHref()))

	listing := pm.Listing("DisplayName").SearchColumns("display_name")
	listing.RowMenu("Rename").RowMenuItem("Rename").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		c := obj.(*Container)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return VListItem().PrependIcon("mdi-pencil-outline").Title(msgr.Rename).Attr("@click",
			web.Plaid().
				URL(b.ContainerByName(c.ModelName).mb.Info().ListingHref()).
				EventFunc(RenameContainerDialogEvent).
				Query(paramContainerID, c.PrimarySlug()).
				Query(paramContainerName, c.DisplayName).
				Query("portal", "presets").
				Go(),
		)
	})

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
	//							Variant(VariantText).
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
							Variant(VariantText).
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
								Variant(VariantText).
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

	dp := pm.Detailing("Overview")
	dp.Field("Overview").ComponentFunc(templateSettings(db, pm))

	eb := pm.Editing("Name", "Description")

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
		if err = wh.Select("count(display_name)").Where("shared = true AND locale_code = ?", locale).Group("display_name, model_name, model_id, locale_code").Count(&c).Error; err != nil {
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

		if err = wh.Select("MIN(id) AS id, display_name, model_name, model_id, locale_code").Find(obj).Error; err != nil {
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
	eb.AppendTabsPanelFunc(func(obj interface{}, ctx *web.EventContext) (tab h.HTMLComponent, content h.HTMLComponent) {

		if ctx.R.FormValue(paramOpenFromSharedContainer) != "1" {
			return nil, nil
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
			pageListComps = append(pageListComps, web.Scope(
				VListItem(
					h.Text(fmt.Sprintf("%s (%s)", p.Title, pid)),
					VSpacer(),
					VIcon(fmt.Sprintf(`{{itemLocals.%s}}`, statusVar)),
				).
					Density(DensityCompact),
			).VSlot(" { locals : itemLocals }").Init(fmt.Sprintf(`{%s: ""}`, statusVar)),
			)

			tab = VTab(h.Text(msgr.RelatedOnlinePages))
			content = VWindowItem(
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
			)
		}
		return
	})
}

func republishRelatedOnlinePages(pageURL string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ids := strings.Split(ctx.R.FormValue("ids"), ",")
		for _, id := range ids {
			statusVar := fmt.Sprintf(`republish_status_%s`, strings.Replace(id, "-", "_", -1))
			web.AppendRunScripts(&r,
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
