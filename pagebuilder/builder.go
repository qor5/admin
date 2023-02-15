package pagebuilder

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/qor5/admin/l10n"
	l10n_view "github.com/qor5/admin/l10n/views"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	"github.com/qor5/admin/publish/views"
	. "github.com/qor5/ui/vuetify"
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
}

const (
	openTemplateDialogEvent          = "openTemplateDialogEvent"
	selectTemplateEvent              = "selectTemplateEvent"
	republishRelatedOnlinePagesEvent = "republish_related_online_pages"

	paramOpenFromSharedContainer = "open_from_shared_container"
)

func New(db *gorm.DB, i18nB *i18n.Builder) *Builder {
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

	r := &Builder{
		db:     db,
		wb:     web.New(),
		prefix: "/page_builder",
	}

	r.ps = presets.New().
		BrandTitle("Page Builder").
		DataOperator(gorm2op.DataOperator(db)).
		URIPrefix(r.prefix).
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

func (b *Builder) GetPresetsBuilder() (r *presets.Builder) {
	return b.ps
}

func (b *Builder) Configure(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pb.I18n().
		RegisterForModule(language.English, I18nPageBuilderKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nPageBuilderKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nPageBuilderKey, Messages_ja_JP)
	pm = pb.Model(&Page{})
	b.mb = pm
	pm.Listing("ID", "Title", "Slug")
	pm.RegisterEventFunc(openTemplateDialogEvent, openTemplateDialog(db))
	pm.RegisterEventFunc(selectTemplateEvent, selectTemplate(db))

	// list.Field("ID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	//	p := obj.(*Page)
	//	return h.Td(
	//		h.A().Children(
	//			h.Text(fmt.Sprintf("Editor for %d", p.ID)),
	//		).Href(fmt.Sprintf("%s/editors/%d?version=%s", b.prefix, p.ID, p.GetVersion())).
	//			Target("_blank"),
	//		VIcon("open_in_new").Size(16).Class("ml-1"),
	//	)
	// })

	eb := pm.Editing("Status", "Schedule", "Title", "Slug", "CategoryID", "TemplateSelection", "EditContainer")

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
		var showURL h.HTMLComponent
		if p.ID != 0 {
			var c Category
			for _, e := range categories {
				if e.ID == p.CategoryID {
					c = *e
					break
				}
			}

			u := os.Getenv("PUBLISH_URL") + c.Path + p.Slug
			showURL = h.Div(
				h.A().Text(u).Href(u).Target("_blank"),
			).Class("mb-4")
		}

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		return h.Div(
			showURL,
			VAutocomplete().Label(msgr.Category).FieldName(field.Name).
				Items(categories).Value(p.CategoryID).ItemText("Path").ItemValue("ID").
				ErrorMessages(vErr.GetFieldErrors("Page.Category")...),
		).ClassIf("mb-4", p.GetStatus() != "")
	})

	eb.Field("TemplateSelection").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		// Only displayed when create action
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if p.GetStatus() == "" {
			return h.Div(
				web.Portal().Name("TemplateDialog"),
				VRow(
					VCol(
						web.Portal(
							VTextField().Disabled(true).Label(msgr.TemplateID),
						).Name("TemplateIDTextField"),
					),
					VCol(
						web.Portal(
							VTextField().Disabled(true).Label(msgr.TemplateName),
						).Name("TemplateNameTextField"),
					),
				),
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
		if p.GetStatus() == publish.StatusDraft {
			var href = fmt.Sprintf("%s/editors/%d?version=%s", b.prefix, p.ID, p.GetVersion())
			if locale, isLocalizable := l10n.IsLocalizableFromCtx(ctx); isLocalizable && l10nON {
				href = fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, p.ID, p.GetVersion(), locale)
			}
			return h.Div(
				VBtn(msgr.EditPageContent).
					Target("_blank").
					Href(href).
					Color("secondary"),
			)
		}
		return nil
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		localeCode, _ := l10n.IsLocalizableFromCtx(ctx)
		p := obj.(*Page)
		if p.Slug != "" {
			p.Slug = path.Clean(p.Slug)
		}

		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if inerr = gorm2op.DataOperator(tx).Save(obj, id, ctx); inerr != nil {
				return
			}

			if strings.Contains(ctx.R.RequestURI, views.SaveNewVersionEvent) {
				if inerr = b.copyContainersToNewPageVersion(tx, int(p.ID), p.GetLocale(), p.ParentVersion, p.GetVersion()); inerr != nil {
					return
				}
				return
			}

			if v := ctx.R.FormValue("TemplateSelectionID"); v != "" {
				var tplID int
				tplID, inerr = strconv.Atoi(v)
				if inerr != nil {
					return
				}
				if !l10nON {
					localeCode = ""
				}
				if inerr = b.copyContainersToAnotherPage(tx, tplID, templateVersion, "", int(p.ID), p.GetVersion(), localeCode); inerr != nil {
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

				if inerr = b.copyContainersToAnotherPage(tx, fromIDInt, fromVersion, fromLocale, int(p.ID), p.GetVersion(), p.GetLocale()); inerr != nil {
					panic(inerr)
					return
				}
				return
			}
			return
		})

		return
	})

	b.configSharedContainer(pb, db)
	b.configDemoContainer(pb, db)
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
		u := os.Getenv("PUBLISH_URL") + category.Path

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return h.Div(
			VTextField().Label("Path").Value(category.Path).Class("mb-n4").
				FieldName("Path").
				ErrorMessages(vErr.GetFieldErrors("Category.Category")...),
			h.Div(
				h.A().Text(u).Href(u).Target("_blank").ClassIf("d-none", category.ID == 0),
			).Class("mt-4"),
		).Class("mb-2")
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		c := obj.(*Category)
		c.Path = path.Clean(c.Path)
		err = db.Save(c).Error
		return
	})

	return
}

func selectTemplate(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		templateSelectionID := ctx.R.FormValue("TemplateSelectionID")

		tpl := Template{}
		isBlank := true
		if templateSelectionID != "0" {
			if err = db.Model(&Template{}).Where("id = ?", templateSelectionID).First(&tpl).Error; err != nil {
				panic(err)
			}
			isBlank = false
		}

		var ID string
		var Name string
		if isBlank {
			ID = ""
			Name = "Blank"
		} else {
			ID = strconv.Itoa(int(tpl.ID))
			Name = tpl.Name
		}
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateIDTextField",
			Body: VTextField().Disabled(true).Label(msgr.TemplateID).Value(ID),
		})
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateNameTextField",
			Body: VTextField().Disabled(true).Label(msgr.TemplateName).Value(Name),
		})

		return
	}
}

func openTemplateDialog(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (er web.EventResponse, err error) {
		msgr := presets.MustGetMessages(ctx.R)
		tpls := []*Template{}

		if err := db.Model(&Template{}).Find(&tpls).Error; err != nil {
			panic(err)
		}

		var tplHTMLComponents []h.HTMLComponent

		if len(tpls) == 0 {
			tplHTMLComponents = append(tplHTMLComponents,
				h.Div(h.Text(msgr.ListingNoRecordToShow)).Class("text-center grey--text text--darken-2"),
			)
		} else {
			tplHTMLComponents = append(tplHTMLComponents,
				getTplColComponent(&Template{
					Model:       gorm.Model{},
					Name:        "Blank",
					Description: "New page",
				}, true),
			)
			for _, tpl := range tpls {
				tplHTMLComponents = append(tplHTMLComponents,
					getTplColComponent(tpl, false),
				)
			}
		}
		msgrPb := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateDialog",
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
						VRow(tplHTMLComponents...).ClassIf("d-none", len(tpls) == 0),
						h.Div(tplHTMLComponents...).ClassIf("d-none", len(tpls) != 0),
					),
					VCardActions(
						VSpacer(),
						VBtn(msgr.Cancel).Attr("@click", "vars.showTemplateDialog=false"),
						VBtn(msgr.OK).Color("primary").
							Attr("@click", fmt.Sprintf("%s;vars.showTemplateDialog=false",
								web.Plaid().EventFunc(selectTemplateEvent).
									Query("TemplateSelectionID", ctx.R.Form).Go()),
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

func getTplColComponent(tpl *Template, isBlank bool) h.HTMLComponent {
	// Avoid layout errors
	var name string
	var desc string
	if tpl.Name == "" {
		name = "Unnamed"
	} else {
		name = tpl.Name
	}
	if tpl.Description == "" {
		desc = "Not described"
	} else {
		desc = tpl.Description
	}

	return VCol(
		VCard(
			h.Div(
				h.Iframe().Src(fmt.Sprintf("./page_builder/preview?id=%d&tpl=1", tpl.ID)).
					Attr("width", "100%", "height", "150", "frameborder", "no").
					Style("transform-origin: left top; transform: scale(1, 1);"),
			),
			VCardTitle(h.Text(name)),
			VCardSubtitle(h.Text(desc)),
			VBtn("Preview").Text(true).XSmall(true).Class("ml-2 mb-4").
				Href(fmt.Sprintf("./page_builder/preview?id=%d&tpl=1", tpl.ID)).
				Target("_blank").Color("primary").ClassIf("d-none", isBlank),
			h.Div(
				h.Input("").Type("radio").Checked(isBlank).
					Value(fmt.Sprintf("%d", tpl.ID)).
					Attr(web.VFieldName("TemplateSelectionID")...).
					Name("TemplateSelectionID").Style("width: 18px; height: 18px"),
			).Class("mr-4 float-right"),
		).Height(280).Class("text-truncate").Outlined(true),
	).Cols(3)
}

func (b *Builder) configSharedContainer(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
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
	pb.GetPermission().Policies(
		perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:shared_containers:*"),
	)
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

func (b *Builder) configDemoContainer(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&DemoContainer{}).URIName("demo_containers").Label("Demo Containers")

	pm.RegisterEventFunc("addDemoContainer", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		modelID := ctx.QueryAsInt(presets.ParamOverlayUpdateID)
		modelName := ctx.R.FormValue("ModelName")
		db.Where(DemoContainer{ModelName: modelName}).FirstOrCreate(&DemoContainer{
			ModelName: modelName,
			ModelID:   uint(modelID),
		})
		r.Reload = true
		return
	})
	listing := pm.Listing("ModelName").SearchColumns("ModelName")
	listing.Field("ModelName").Label("Name")
	ed := pm.Editing("SelectContainer").ActionsFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent { return nil })
	ed.Field("SelectContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
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
		return h.Div(
			VBtn(msgr.EditPageContent).
				Target("_blank").
				Href(fmt.Sprintf("%s/editors/%d?tpl=1", b.prefix, m.ID)).
				Color("secondary"),
		)
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

		var c int64
		if err = wh.Select("count(display_name)").Where("shared = true").Group("display_name,model_name,model_id").Count(&c).Error; err != nil {
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
        and %s.locale_code = %s.page_locale`,
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
									EventFunc(presets.OpenConfirmationDialogEvent).
									Query(presets.ConfirmationDialogConfirmEventKey,
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
