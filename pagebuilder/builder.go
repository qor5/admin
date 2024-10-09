package pagebuilder

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/richeditor"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/admin/v3/utils"
)

type RenderInput struct {
	IsEditor    bool
	IsReadonly  bool
	Device      string
	ContainerId string
	DisplayName string
	Obj         interface{}
}

type RenderFunc func(obj interface{}, input *RenderInput, ctx *web.EventContext) h.HTMLComponent

type PageLayoutFunc func(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent

type (
	SubPageTitleFunc func(ctx *web.EventContext) string
	WrapCompFunc     func(comps h.HTMLComponents) h.HTMLComponents
	PageLayoutInput  struct {
		SeoTags           h.HTMLComponent
		CanonicalLink     h.HTMLComponent
		StructuredData    h.HTMLComponent
		FreeStyleCss      []string
		FreeStyleTopJs    []string
		FreeStyleBottomJs []string
		WrapHead          WrapCompFunc
		WrapBody          WrapCompFunc
		Hreflang          map[string]string
		Header            h.HTMLComponent
		Footer            h.HTMLComponent
		IsEditor          bool
		LocaleCode        string
		EditorCss         []h.HTMLComponent
		IsPreview         bool
	}
)

type Builder struct {
	prefix                        string
	wb                            *web.Builder
	db                            *gorm.DB
	containerBuilders             []*ContainerBuilder
	ps                            *presets.Builder
	models                        []*ModelBuilder
	templates                     []*TemplateBuilder
	templateModel                 *presets.ModelBuilder
	l10n                          *l10n.Builder
	mediaBuilder                  *media.Builder
	ab                            *activity.Builder
	publisher                     *publish.Builder
	seoBuilder                    *seo.Builder
	pageStyle                     h.HTMLComponent
	pageLayoutFunc                PageLayoutFunc
	subPageTitleFunc              SubPageTitleFunc
	images                        http.Handler
	imagesPrefix                  string
	defaultDevice                 string
	publishBtnColor               string
	duplicateBtnColor             string
	templateEnabled               bool
	expendContainers              bool
	pageEnabled                   bool
	autoSaveReload                bool
	disabledNormalContainersGroup bool
	previewOpenNewTab             bool
	previewContainer              bool
	templateInstall               presets.ModelInstallFunc
	pageInstall                   presets.ModelInstallFunc
	categoryInstall               presets.ModelInstallFunc
	devices                       []Device
	fields                        []string
}

const (
	selectTemplateEvent = "selectTemplateEvent"
	// clearTemplateEvent               = "clearTemplateEvent"
	republishRelatedOnlinePagesEvent = "republish_related_online_pages"

	paramOpenFromSharedContainer = "open_from_shared_container"

	PageBuilderPreviewCard = "PageBuilderPreviewCard"
)

func New(prefix string, db *gorm.DB, b *presets.Builder) *Builder {
	return newBuilder(prefix, db, b)
}

func newBuilder(prefix string, db *gorm.DB, b *presets.Builder) *Builder {
	r := &Builder{
		db:                db,
		wb:                web.New(),
		prefix:            prefix,
		defaultDevice:     DeviceComputer,
		publishBtnColor:   "primary",
		duplicateBtnColor: "primary",
		templateEnabled:   true,
		expendContainers:  true,
		pageEnabled:       true,
		previewContainer:  true,
	}
	r.templateInstall = r.defaultTemplateInstall
	r.categoryInstall = r.defaultCategoryInstall
	r.pageInstall = r.defaultPageInstall
	r.pageLayoutFunc = defaultPageLayoutFunc

	r.ps = presets.New().
		BrandTitle("Page Builder").
		DataOperator(gorm2op.DataOperator(db)).
		URIPrefix(prefix).
		I18n(b.GetI18n()).
		DetailLayoutFunc(r.pageEditorLayout)
	r.ps.Permission(b.GetPermission())
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

func (b *Builder) AutoSaveReload(v bool) (r *Builder) {
	b.autoSaveReload = v
	return b
}

func (b *Builder) AutoMigrate() (r *Builder) {
	err := AutoMigrate(b.db)
	if err != nil {
		panic(err)
	}
	return b
}

func (b *Builder) WrapPageInstall(w func(presets.ModelInstallFunc) presets.ModelInstallFunc) (r *Builder) {
	b.pageInstall = w(b.pageInstall)
	return b
}

func (b *Builder) WrapTemplateInstall(w func(presets.ModelInstallFunc) presets.ModelInstallFunc) (r *Builder) {
	b.templateInstall = w(b.templateInstall)
	return b
}

func (b *Builder) WrapCategoryInstall(w func(presets.ModelInstallFunc) presets.ModelInstallFunc) (r *Builder) {
	b.categoryInstall = w(b.categoryInstall)
	return b
}

func (b *Builder) PageLayout(v PageLayoutFunc) (r *Builder) {
	b.pageLayoutFunc = v
	return b
}

func AutoMigrate(db *gorm.DB) (err error) {
	if err = db.AutoMigrate(
		&Page{},
		&Template{},
		&Container{},
		&Category{},
		&DemoContainer{},
	); err != nil {
		return
	}
	// https://github.com/go-gorm/sqlite/blob/64917553e84d5482e252c7a0c8f798fb672d7668/ddlmod.go#L16
	// fxxk: newline is not allowed
	err = db.Exec(`
create unique index if not exists uidx_page_builder_demo_containers_model_name_locale_code on page_builder_demo_containers (model_name, locale_code) where deleted_at is null;
`).Error
	return
}

func (b *Builder) WrapPageLayout(warp func(v PageLayoutFunc) PageLayoutFunc) (r *Builder) {
	b.pageLayoutFunc = warp(b.pageLayoutFunc)
	return b
}

func (b *Builder) SubPageTitle(v SubPageTitleFunc) (r *Builder) {
	b.subPageTitleFunc = v
	return b
}

func (b *Builder) L10n(v *l10n.Builder) (r *Builder) {
	b.l10n = v
	return b
}

func (b *Builder) Media(v *media.Builder) (r *Builder) {
	b.mediaBuilder = v
	return b
}

func (b *Builder) Activity(v *activity.Builder) (r *Builder) {
	b.ab = v
	return b
}

func (b *Builder) SEO(v *seo.Builder) (r *Builder) {
	b.seoBuilder = v
	return b
}

func (b *Builder) Publisher(v *publish.Builder) (r *Builder) {
	b.publisher = v
	return b
}

func (b *Builder) GetPageTitle() SubPageTitleFunc {
	if b.subPageTitleFunc == nil {
		b.subPageTitleFunc = defaultSubPageTitle
	}
	return b.subPageTitleFunc
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

func (b *Builder) PageEnabled(v bool) (r *Builder) {
	b.pageEnabled = v
	return b
}

func (b *Builder) PreviewContainer(v bool) (r *Builder) {
	b.previewContainer = v
	return b
}

func (b *Builder) ExpendContainers(v bool) (r *Builder) {
	b.expendContainers = v
	return b
}

func (b *Builder) PreviewOpenNewTab(v bool) (r *Builder) {
	b.previewOpenNewTab = v
	return b
}

func (b *Builder) DisabledNormalContainersGroup(v bool) (r *Builder) {
	b.disabledNormalContainersGroup = v
	return b
}

func (b *Builder) Model(mb *presets.ModelBuilder) (r *ModelBuilder) {
	r = &ModelBuilder{
		mb:      mb,
		editor:  b.ps.Model(mb.NewModel()).URIName(mb.Info().URIName()),
		builder: b,
		db:      b.db,
	}
	b.models = append(b.models, r)
	r.setName()
	r.registerFuncs()
	r.configDuplicate(r.mb)
	r.configDuplicate(r.editor)
	return r
}

func (b *Builder) ModelInstall(pb *presets.Builder, mb *presets.ModelBuilder) (err error) {
	defer b.ps.Build()

	r := b.Model(mb)
	b.useAllPlugin(mb)
	// register model editors page
	b.installAsset(pb)
	b.configEditor(r)
	b.configPublish(r)
	b.configDetail(r)
	return nil
}

func (b *Builder) installAsset(pb *presets.Builder) {
	pb.GetI18n().
		RegisterForModule(language.English, I18nPageBuilderKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPageBuilderKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPageBuilderKey, Messages_ja_JP)

	pb.ExtraAsset("/redactor.js", "text/javascript", richeditor.JSComponentsPack())
	pb.ExtraAsset("/redactor.css", "text/css", richeditor.CSSComponentsPack())
}

func (b *Builder) configPageSaver(pb *presets.Builder) (mb *presets.ModelBuilder) {
	mb = pb.Model(&Page{})
	eb := mb.Editing()
	eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			p := obj.(*Page)
			if p.Slug != "" {
				p.Slug = path.Clean(p.Slug)
			}
			funcName := ctx.R.FormValue(web.EventFuncIDName)
			if funcName == publish.EventDuplicateVersion {
				var fromPage Page
				eb.Fetcher(&fromPage, ctx.Param(presets.ParamID), ctx)
				p.SEO = fromPage.SEO
			}
			return in(obj, id, ctx)
		}
	})
	return
}

func (b *Builder) Install(pb *presets.Builder) (err error) {
	defer b.ps.Build()
	b.ps.I18n(pb.GetI18n())
	if b.pageEnabled {
		var r *ModelBuilder
		r = b.Model(b.configPageSaver(pb))
		b.configEditor(r)
		b.installAsset(pb)
		b.configTemplateAndPage(pb, r)
		b.configSharedContainer(pb, r)
		b.configDetail(r)
		categoryM := pb.Model(&Category{}).URIName("page_categories").Label("Page Categories")
		if err = b.categoryInstall(pb, categoryM); err != nil {
			return
		}
	}
	b.configDemoContainer(pb)
	b.preparePlugins()
	for _, t := range b.templates {
		t.Install()
	}
	for _, t := range b.containerBuilders {
		t.Install()
	}
	return
}

func (b *Builder) configEditor(m *ModelBuilder) {
	b.useAllPlugin(m.editor)
	md := m.editor.Detailing().Drawer(false)
	md.PageFunc(b.Editor(m))
}

func (b *Builder) configDetail(m *ModelBuilder) {
	mb := m.mb
	if mb.HasDetailing() {
		dp := mb.Detailing()
		fb := dp.GetField(PageBuilderPreviewCard)
		if fb != nil && fb.GetCompFunc() == nil {
			fb.ComponentFunc(overview(m))
		}
		if b.ab != nil {
			dp.SidePanelFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
				return b.ab.MustGetModelBuilder(mb).NewTimelineCompo(ctx, obj, "_side")
			})
		}
	}
}

func (b *Builder) configPublish(r *ModelBuilder) {
	publisher := b.publisher
	if publisher != nil {
		publisher.ContextValueFuncs(r.ContextValueProvider).Activity(b.ab).AfterInstall(func() {
			r.mb.Editing().SidePanelFunc(nil).ActionsFunc(nil).TabsPanels()
		})
	}
}

func (b *Builder) configTemplateAndPage(pb *presets.Builder, r *ModelBuilder) {
	pm := r.mb
	err := b.pageInstall(pb, pm)
	if err != nil {
		panic(err)
	}
	if b.templateEnabled {
		err = b.templateInstall(pb, pm)
		if err != nil {
			panic(err)
		}
	}
	b.configPublish(r)
	b.useAllPlugin(pm)

	// dp.TabsPanels()
}

func (b *Builder) useAllPlugin(pm *presets.ModelBuilder) {
	if b.publisher != nil {
		pm.Use(b.publisher)
	}

	if b.ab != nil {
		pm.Use(b.ab)
	}

	if b.seoBuilder != nil {
		pm.Use(b.seoBuilder)
	}

	if b.l10n != nil {
		pm.Use(b.l10n)
	}
}

func (b *Builder) preparePlugins() {
	l10nB := b.l10n
	// activityB := b.ab
	publisher := b.publisher
	if l10nB != nil {
		l10nB.Activity(b.ab)
	}
	seoBuilder := b.seoBuilder
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

	if b.mediaBuilder == nil {
		b.mediaBuilder = media.New(b.db)
	}
	b.ps.Use(b.mediaBuilder, publisher, seoBuilder)
}

func (b *Builder) configDetailLayoutFunc(
	pb *presets.Builder,
	pm *presets.ModelBuilder,
	db *gorm.DB,
) {
	// change old detail layout
	pm.Detailing().AfterTitleCompFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return publish.DefaultVersionBar(db)(obj, ctx)
	})
	pm.Detailing().Title(func(ctx *web.EventContext, obj any, style presets.DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		var pageAppbarContent []h.HTMLComponent
		pageAppbarContent = h.Components(
			VToolbarTitle(
				b.GetPageTitle()(ctx),
			),
			VSpacer(),
			publish.DefaultVersionBar(db)(obj, ctx),
		)
		title = defaultTitle
		titleCompo = h.Div(
			pageAppbarContent...,
		).Class("d-flex align-center  justify-space-between  w-100")
		return
	})
	return
}

func versionCount(db *gorm.DB, obj interface{}, id string, localCode string) (count int64) {
	db.Model(obj).Where("id = ? and locale_code = ?", id, localCode).Count(&count)
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

func (b *Builder) defaultCategoryInstall(pb *presets.Builder, pm *presets.ModelBuilder) (err error) {
	db := b.db

	lb := pm.Listing("Name", "Path", "Description")
	pm.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ModelLabelPageCategory
		}
		return msgr.ModelLabelPageCategories
	})
	lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"Name":        msgr.ListHeaderName,
			"Path":        msgr.ListHeaderPath,
			"Description": msgr.ListHeaderDescription,
		}, nil
	}))
	lb.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			result, err = in(ctx, params)
			cats := result.Nodes.([]*Category)
			sort.Slice(cats, func(i, j int) bool {
				return cats[i].Path < cats[j].Path
			})
			fillCategoryIndentLevels(cats)
			return
		}
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
	eb.Field("Name").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return vx.VXField().Label(msgr.ListHeaderName).
			Attr(web.VField(field.Name, field.Value(obj))...).
			ErrorMessages(vErr.GetFieldErrors("Category.Name")...)
	})

	eb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return vx.VXField().Label(msgr.ListHeaderPath).
			Attr(web.VField(field.Name, strings.TrimPrefix(field.Value(obj).(string), "/"))...).
			Attr("prefix", "/").
			ErrorMessages(vErr.GetFieldErrors("Category.Category")...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		m := obj.(*Category)
		m.Path = path.Join("/", m.Path)
		return nil
	})

	eb.Field("Description").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return vx.VXField().
			Attr(web.VField(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)))...).
			Label(msgr.ListHeaderDescription).
			ErrorMessages(field.Errors...).
			Disabled(field.Disabled)
	})

	eb.DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		cs := obj.(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)
		ID := cs["id"]
		Locale := cs[l10n.SlugLocaleCode]
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		var count int64
		if err = db.Model(&Page{}).Where("category_id = ? AND locale_code = ?", ID, Locale).Count(&count).Error; err != nil {
			return
		}
		if count > 0 {
			err = errors.New(msgr.UnableDeleteCategoryMsg)
			return
		}
		if err = db.Model(&Category{}).Where("id = ? AND locale_code = ?", ID, Locale).Delete(&Category{}).Error; err != nil {
			return
		}
		return
	})

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Category)
		err = categoryValidator(ctx, c, db, b.l10n)
		return
	})

	eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			c := obj.(*Category)
			c.Path = path.Clean(c.Path)
			return in(obj, id, ctx)
		}
	})
	if b.ab != nil {
		pm.Use(b.ab)
	}
	if b.l10n != nil {
		pm.Use(b.l10n)
	}

	return
}

const (
	templateSelectedID = "TemplateSelectedID"
)

func (b *Builder) configSharedContainer(pb *presets.Builder, r *ModelBuilder) {
	db := b.db

	pm := pb.Model(&Container{}).URIName("shared_containers").Label("Shared Containers")

	pm.RegisterEventFunc(republishRelatedOnlinePagesEvent, republishRelatedOnlinePages(r.mb.Info().ListingHref()))

	listing := pm.Listing("DisplayName").SearchColumns("display_name")
	pm.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ModelLabelSharedContainer
		}
		return msgr.ModelLabelSharedContainers
	})
	listing.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"DisplayName": msgr.ListHeaderName,
		}, nil
	}))
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
	if permB := pb.GetPermission(); permB != nil {
		permB.CreatePolicies(
			perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:shared_containers:*"),
		)
	}
	listing.Field("DisplayName").Label("Name")
	listing.SearchFunc(sharedContainerSearcher(db, r))
	listing.WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
		return func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
			c := obj.(*Container)
			cell.SetAttr("@click",
				web.Plaid().
					EventFunc(actions.Edit).
					URL(b.ContainerByName(c.ModelName).GetModelBuilder().Info().ListingHref()).
					Query(presets.ParamID, c.ModelID).
					Query(paramOpenFromSharedContainer, 1).
					Query(presets.ParamVarCurrentActive, presets.ListingCompo_GetVarCurrentActive(evCtx)).
					Go())
			return in(evCtx, cell, id, obj)
		}
	})

	if b.ab != nil {
		pm.Use(b.ab)
	}
	if b.l10n != nil {
		pm.Use(b.l10n)
	}
	return
}

func (b *Builder) configDemoContainer(pb *presets.Builder) (pm *presets.ModelBuilder) {
	pm = pb.Model(&DemoContainer{}).URIName("demo_containers").Label("Demo Containers")

	listing := pm.Listing("ModelName").SearchColumns("model_name")
	listing.Field("ModelName").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*DemoContainer)
		modelName := p.ModelName
		if b.ps.GetI18n() != nil {
			modelName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, modelName)
		}

		return h.Td(h.Text(modelName))
	})
	listing.WrapRow(func(in presets.RowProcessor) presets.RowProcessor {
		return func(evCtx *web.EventContext, row h.MutableAttrHTMLComponent, id string, obj any) (comp h.MutableAttrHTMLComponent, err error) {
			c := presets.ListingCompoFromEventContext(evCtx)
			p := obj.(*DemoContainer)
			row.SetAttr(":class", fmt.Sprintf(`{ %q: vars.%s === %q }`, presets.ListingCompo_CurrentActiveClass, c.VarCurrentActive(), p.ModelName))
			return in(evCtx, row, id, obj)
		}
	})
	pm.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ModelLabelDemoContainer
		}
		return msgr.ModelLabelDemoContainers
	})
	listing.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"ModelName": msgr.ListHeaderName,
		}, nil
	}))
	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			b.firstOrCreateDemoContainers(ctx)
			return in(ctx, params)
		}
	})
	listing.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		return []*vx.FilterItem{
			{
				Key:          "all",
				Invisible:    true,
				SQLCondition: ``,
			},
			{
				Key:          "Filled",
				Invisible:    true,
				SQLCondition: `filled = true `,
			},
			{
				Key:          "NotFilled",
				Invisible:    true,
				SQLCondition: `filled = false`,
			},
		}
	})
	listing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabAll,
				ID:    "all",
			},
			{
				Label: msgr.FilterTabFilled,
				ID:    "Filled",
				Query: url.Values{"Filled": []string{"true"}},
			},
			{
				Label: msgr.FilterTabNotFilled,
				ID:    "NotFilled",
				Query: url.Values{"NotFilled": []string{"false"}},
			},
		}
	})
	listing.Field("ModelName").Label("Name")
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return nil
	})
	listing.RowMenu().Empty()
	listing.WrapCell(func(in presets.CellProcessor) presets.CellProcessor {
		return func(evCtx *web.EventContext, cell h.MutableAttrHTMLComponent, id string, obj any) (h.MutableAttrHTMLComponent, error) {
			c := obj.(*DemoContainer)
			cell.SetAttr("@click",
				web.Plaid().
					EventFunc(actions.Edit).
					URL(b.ContainerByName(c.ModelName).GetModelBuilder().Info().ListingHref()).
					Query(presets.ParamID, c.ModelID).
					Query(presets.ParamVarCurrentActive, presets.ListingCompo_GetVarCurrentActive(evCtx)).
					Go())
			return in(evCtx, cell, id, obj)
		}
	})
	if b.ab != nil {
		pm.Use(b.ab)
	}
	if b.l10n != nil {
		pm.Use(b.l10n)
	}
	return
}

func (b *Builder) firstOrCreateDemoContainers(ctx *web.EventContext, cons ...*ContainerBuilder) {
	locale, _ := l10n.IsLocalizableFromContext(ctx.R.Context())
	localeCodes := []string{locale}
	if b.l10n != nil {
		localeCodes = b.l10n.GetSupportLocaleCodes()
	}
	if len(cons) == 0 {
		cons = b.containerBuilders
	}
	for _, con := range cons {
		if err := con.firstOrCreate(slices.Concat(localeCodes)); err != nil {
			continue
		}
	}
}

func sharedContainerSearcher(db *gorm.DB, b *ModelBuilder) presets.SearchFunc {
	return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
		ilike := "ILIKE"
		if db.Dialector.Name() == "sqlite" {
			ilike = "LIKE"
		}

		wh := db.Model(params.Model)
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

		locale, _ := l10n.IsLocalizableFromContext(ctx.R.Context())
		var c int64
		if err = wh.Select("count(display_name)").Where("shared = true AND locale_code = ? and page_model_name = ? ", locale, b.name).Group("display_name, model_name, model_id, locale_code").Count(&c).Error; err != nil {
			return nil, err
		}

		totalCount := int(c)

		if params.PerPage > 0 {
			wh = wh.Limit(int(params.PerPage))
			page := params.Page
			if page == 0 {
				page = 1
			}
			offset := (page - 1) * params.PerPage
			wh = wh.Offset(int(offset))
		}

		rtNodes := reflect.New(reflect.SliceOf(reflect.TypeOf(params.Model))).Elem()
		if err = wh.Select("MIN(id) AS id, display_name, model_name, model_id, locale_code").Find(rtNodes.Addr().Interface()).Error; err != nil {
			return nil, err
		}

		return &presets.SearchResult{
			PageInfo: relay.PageInfo{
				TotalCount: totalCount,
			},
			Nodes: rtNodes.Interface(),
		}, nil
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
	builder      *Builder
	name         string
	mb           *presets.ModelBuilder
	modelBuilder *presets.ModelBuilder
	model        interface{}
	modelType    reflect.Type
	renderFunc   RenderFunc
	cover        string
	group        string
	onlyPages    bool
}

func (b *Builder) RegisterContainer(name string) (r *ContainerBuilder) {
	r = &ContainerBuilder{
		name:    name,
		builder: b,
	}
	b.containerBuilders = append(b.containerBuilders, r)
	return
}

func (b *Builder) RegisterModelContainer(name string, mb *presets.ModelBuilder) (r *ContainerBuilder) {
	r = &ContainerBuilder{
		name:         name,
		builder:      b,
		modelBuilder: mb,
	}
	b.containerBuilders = append(b.containerBuilders, r)
	return
}

type TagInterface interface {
	SetAttr(k string, v interface{})
}

func (b *ContainerBuilder) setFieldsLazyWrapComponentFunc(fields *presets.FieldsBuilder) {
	for _, fieldName := range fields.FieldNames() {
		field := fields.GetField(fieldName.(string))
		if field.GetNestedFieldsBuilder() != nil {
			b.setFieldsLazyWrapComponentFunc(field.GetNestedFieldsBuilder())
			continue
		}
		field.LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				comp := in(obj, field, ctx)
				formKey := field.ModelInfo.URIName() + "_" + field.FormKey
				if p, ok := comp.(TagInterface); ok {
					p.SetAttr("ref", formKey)
					return h.Div(comp).Attr("v-on-mounted", fmt.Sprintf(`({el,window})=>{
		const refName = "%s";
		vars.__currentFocusUpdating = false;
		el.__handleFocusIn=()=>{
			vars.__currentFocusRefName = refName;
		};
		el.__handleFocusOut=(event)=>{
			if(vars.__currentFocusUpdating){return}
			vars.__currentFocusRefName ="";
		};

		el.addEventListener("focusin",el.__handleFocusIn);
		el.addEventListener("focusout",el.__handleFocusOut);
	   }`, formKey)).Attr("v-on-unmounted", `({el})=>{
		el.removeEventListener("focusin",el.__handleFocusIn);
		el.removeEventListener("focusout",el.__handleFocusOut);
}`).Attr("v-before-unmount", `({el})=>{
	vars.__currentFocusUpdating = true;
}`)
				}
				return comp
			}
		})
	}
}

func (b *ContainerBuilder) Install() {
	editing := b.mb.Editing()
	editing.WrapIdCurrentActive(func(in presets.IdCurrentActiveProcessor) presets.IdCurrentActiveProcessor {
		return func(ctx *web.EventContext, current string) (s string, err error) {
			s, err = in(ctx, current)
			if err != nil {
				return
			}
			s = b.name
			return
		}
	})
	if b.builder.autoSaveReload {
		b.setFieldsLazyWrapComponentFunc(&editing.FieldsBuilder)
	}
	editing.AppendHiddenFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		if portalName := ctx.Param(presets.ParamPortalName); portalName != pageBuilderRightContentPortal {
			return nil
		}
		var (
			fromKey     = ctx.Param(presets.ParamAddRowFormKey)
			addRowBtnID = ctx.Param(presets.AddRowBtnKey(fromKey))
		)
		return h.Components(
			h.Div().Style("display:none").Attr("v-on-mounted", fmt.Sprintf(`({window}) => {
				if (!!locals.__pageBuilderRightContentKeepScroll) {
					locals.__pageBuilderRightContentKeepScroll();
				}	
				const addRowBtnID = %q;
				if(addRowBtnID){
				const newAddRowBtn = window.document.getElementById(addRowBtnID);
				newAddRowBtn.scrollIntoView({ behavior: 'smooth', block: 'end' });
				}
				 const __currentFocusRefName = $refs[vars.__currentFocusRefName];
                 if(!__currentFocusRefName || typeof __currentFocusRefName.focus != 'function'){return}
				  __currentFocusRefName.focus();	
			}`, addRowBtnID)),
			web.Listen(
				b.mb.NotifRowUpdated(),
				web.Plaid().
					URL(b.mb.Info().ListingHref()).
					EventFunc(actions.Update).
					Query(presets.ParamID, web.Var("payload.id")).
					Query(presets.ParamPortalName, pageBuilderRightContentPortal).
					Query(presets.ParamOverlay, actions.Content).
					ThenScript(web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).
						Query(paramStatus, ctx.Param(paramStatus)).MergeQuery(true).Go()).
					Go(),
			),
			h.If(b.builder.autoSaveReload,
				web.Listen(
					b.mb.NotifModelsUpdated(),
					web.Plaid().
						URL(b.mb.Info().ListingHref()).
						EventFunc(actions.Edit).
						Query(presets.ParamID, web.Var("payload.ids[0]")).
						Query(presets.ParamPortalName, pageBuilderRightContentPortal).
						Query(presets.ParamOverlay, actions.Content).Go(),
				),
			),
		)
	})
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
	b.uRIName(inflection.Plural(strcase.ToKebab(b.name)))
	b.warpSaver()
	return b
}

func (b *ContainerBuilder) OnlyPages(v bool) *ContainerBuilder {
	b.onlyPages = v
	return b
}

func (b *ContainerBuilder) uRIName(uri string) *ContainerBuilder {
	if b.mb == nil {
		return b
	}
	b.mb.URIName(uri)
	return b
}

func (b *ContainerBuilder) GetModelBuilder() *presets.ModelBuilder {
	return b.mb
}

func (b *ContainerBuilder) warpSaver() {
	b.mb.Editing().WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			var demo *DemoContainer
			db := b.builder.db
			db.Where("model_name = ? and model_id = ? ", b.name, id).First(&demo)
			if demo.ID > 0 && !demo.Filled {
				if err = db.Model(&demo).UpdateColumn("filled", true).Error; err != nil {
					return
				}
			}
			return in(obj, id, ctx)
		}
	})
}

func (b *ContainerBuilder) RenderFunc(v RenderFunc) *ContainerBuilder {
	b.renderFunc = v
	return b
}

func (b *ContainerBuilder) Cover(v string) *ContainerBuilder {
	b.cover = v
	return b
}

func (b *ContainerBuilder) Group(v string) *ContainerBuilder {
	b.group = v
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
	eb.OnChangeActionFunc(func(id string, ctx *web.EventContext) (s string) {
		return web.Emit(b.mb.NotifRowUpdated(), presets.PayloadRowUpdated{Id: id})
	})
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
							Color(ColorPrimary).
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

func (b *ContainerBuilder) getContainerDataID(id int) string {
	return fmt.Sprintf(inflection.Plural(strcase.ToKebab(b.name))+"_%v", id)
}

func (b *ContainerBuilder) firstOrCreate(localeCodes []string) (err error) {
	var (
		db   = b.builder.db
		obj  = b.mb.NewModel()
		cons []*DemoContainer
		m    = &DemoContainer{}
	)
	if len(localeCodes) == 0 {
		return
	}
	return db.Transaction(func(tx *gorm.DB) (vErr error) {
		tx.Where("model_name = ? and locale_code in ? ", b.name, localeCodes).Find(&cons)
		if len(cons) == 0 {
			if vErr = tx.Create(obj).Error; vErr != nil {
				return
			}
			modelID := reflectutils.MustGet(obj, "ID").(uint)
			m = &DemoContainer{
				ModelName: b.name,
				ModelID:   modelID,
				Filled:    false,
				Locale:    l10n.Locale{LocaleCode: localeCodes[0]},
			}
			if vErr = tx.Create(m).Error; vErr != nil {
				return
			}
			slices.Delete(localeCodes, 0, 1)

		} else {
			m = cons[0]
			slices.DeleteFunc(localeCodes, func(s string) bool {
				for _, con := range cons {
					if con.LocaleCode == s {
						return true
					}
				}
				return false
			})
		}
		for _, localeCode := range localeCodes {
			if localeCode == "" {
				continue
			}
			if vErr = tx.Create(&DemoContainer{
				Model:     gorm.Model{ID: m.ID},
				ModelName: b.name,
				ModelID:   m.ModelID,
				Filled:    false,
				Locale:    l10n.Locale{LocaleCode: localeCode},
			}).Error; vErr != nil {
				return
			}
		}
		return
	})

	return
}

func republishRelatedOnlinePages(pageURL string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		ids := strings.Split(ctx.R.FormValue("ids"), ",")
		for _, id := range ids {
			statusVar := fmt.Sprintf(`republish_status_%s`, strings.Replace(id, "-", "_", -1))
			web.AppendRunScripts(&r,
				web.Plaid().
					URL(pageURL).
					EventFunc(publish.EventRepublish).
					Query("id", id).
					Query(publish.ParamScriptAfterPublish, fmt.Sprintf(`vars.%s = "done"`, statusVar)).
					Query("status_var", statusVar).
					Go(),
				fmt.Sprintf(`vars.%s = "pending"`, statusVar),
			)
		}
		return r, nil
	}
}

func (b *Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, mb := range b.models {
		if strings.Index(r.RequestURI, b.prefix+"/"+mb.mb.Info().URIName()+"/preview") >= 0 {
			mb.preview.ServeHTTP(w, r)
			return
		}
	}
	if b.images != nil {
		if strings.Index(r.RequestURI, path.Join(b.prefix, b.imagesPrefix)) >= 0 {
			b.images.ServeHTTP(w, r)
			return
		}
	}
	b.ps.ServeHTTP(w, r)
}

func (b *Builder) generateEditorBarJsFunction(ctx *web.EventContext) string {
	editAction := fmt.Sprintf(`vars.%s=container_data_id;`, paramContainerDataID) +
		web.Plaid().
			PushState(true).
			MergeQuery(true).
			Query(paramContainerDataID, web.Var("container_data_id")).
			Query(paramContainerID, web.Var("container_id")).
			RunPushState() + ";" +
		web.POST().
			EventFunc(actions.Edit).
			URL(web.Var(fmt.Sprintf(`"%s/"+arr[0]`, b.prefix))).
			Query(presets.ParamID, web.Var("arr[1]")).
			Query(presets.ParamOverlay, actions.Content).
			Query(presets.ParamPortalName, pageBuilderRightContentPortal).
			Go()
	addAction := web.Plaid().MergeQuery(true).PushState(true).Query(paramContainerID, web.Var("container_id")).RunPushState() +
		`;vars.containerPreview=false;vars.overlayEl.refs.overlay.showByIframe(vars.el.refs.scrollIframe,rect);vars.overlay=true;` +
		addVirtualELeToContainer(web.Var("container_data_id"))
	deleteAction := web.POST().
		EventFunc(DeleteContainerConfirmationEvent).
		Query(paramContainerID, web.Var("container_id")).
		Query(paramContainerName, web.Var("display_name")).
		Go()
	moveAction := web.Plaid().
		EventFunc(MoveUpDownContainerEvent).
		MergeQuery(true).
		Query(paramContainerID, web.Var("container_id")).
		Query(paramMoveDirection, web.Var("msg_type")).
		Query(paramModelID, web.Var("model_id")).
		Query(paramStatus, ctx.Param(paramStatus)).
		Go()
	return fmt.Sprintf(`
function(e){
	const { msg_type,container_data_id, container_id,display_name,rect } = e.data
	if (!msg_type || !container_data_id.split) {
		return
	} 
	let arr = container_data_id.split("_");
	if (arr.length != 2) {
		console.log(arr);
		return
	}
    switch (msg_type) {
	  case '%s':
		%s;
		break
      case '%s':
        %s;
        break
	  case '%s':
	  case '%s':
		%s;
		break
	  case '%s':
        %s;
        break
    }
	
}`,
		EventEdit, editAction,
		EventDelete, deleteAction,
		EventUp, EventDown, moveAction, EventAdd, addAction,
	)
}

func defaultSubPageTitle(ctx *web.EventContext) string {
	return i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages).PageOverView
}

type (
	Device struct {
		Name     string
		Width    string
		Icon     string
		Disabled bool
	}
)

func (b *Builder) PreviewDevices(devices ...Device) *Builder {
	b.devices = devices
	return b
}

func (b *Builder) getDevices() []Device {
	if len(b.devices) == 0 {
		b.setDefaultDevices()
	}
	return b.devices
}

func (b *Builder) setDefaultDevices() {
	b.devices = []Device{
		// {Name: DeviceComputer, Width: "", Icon: "mdi-desktop-mac"},
		// {Name: DevicePhone, Width: "414px", Icon: "mdi-tablet-android"},
		// {Name: DeviceTablet, Width: "768px", Icon: "mdi-tablet"},
		{Name: DeviceComputer, Width: "", Icon: "mdi-monitor"},
		{Name: DevicePhone, Width: "414px", Icon: "mdi-cellphone"},
		{Name: DeviceTablet, Width: "768px", Icon: "mdi-tablet"},
	}
}

func (b *Builder) deviceToggle(ctx *web.EventContext) h.HTMLComponent {
	var comps []h.HTMLComponent
	ctx.R.Form.Del(web.EventFuncIDName)
	device := ctx.Param(paramsDevice)
	for _, d := range b.getDevices() {
		if device == "" && d.Name == b.defaultDevice {
			device = d.Name
		}
		comps = append(comps,
			VBtn("").Icon(d.Icon).Color(ColorPrimary).Value(d.Name).
				Disabled(d.Disabled).
				BaseColor(ColorPrimary).Variant(VariantText).Class("mr-2"),
		)
	}
	if device == "" {
		device = b.getDevices()[0].Name
	}
	return web.Scope(
		VBtnToggle(
			comps...,
		).Class("pa-2 rounded-lg ").
			Mandatory(true).
			Attr("v-model", "toggleLocals.activeDevice").
			Attr("@update:model-value", web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).
				PushState(true).MergeQuery(true).Query(paramsDevice, web.Var("toggleLocals.activeDevice")).Go()),
	).VSlot("{ locals : toggleLocals}").Init(fmt.Sprintf(`{activeDevice: %q}`, device))
}

func (b *Builder) GetModelBuilder(mb *presets.ModelBuilder) *ModelBuilder {
	for _, modelBuilder := range b.models {
		if modelBuilder.mb == mb {
			return modelBuilder
		}
	}
	return nil
}

func (b *Builder) GetPageModelBuilder() *ModelBuilder {
	for _, modelBuilder := range b.models {
		if modelBuilder.name == utils.GetObjectName(&Page{}) {
			return modelBuilder
		}
	}
	return nil
}

func (b *Builder) GetTemplateModel() *presets.ModelBuilder {
	return b.templateModel
}

func (b *Builder) Only(vs ...string) *Builder {
	b.fields = vs
	return b
}

func (b *Builder) filterFields(names []interface{}) []interface{} {
	return utils.Filter(names, func(i interface{}) bool {
		if len(b.fields) == 0 {
			return true
		}
		return slices.Contains(b.fields, i.(string))
	})
}

func (b *Builder) expectField(name string) bool {
	if len(b.fields) == 0 {
		return true
	}
	return slices.Contains(b.fields, name)
}
