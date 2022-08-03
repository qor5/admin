package pagebuilder

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	"github.com/goplaid/x/presets/gorm2op"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/publish"
	"github.com/qor/qor5/publish/views"
	h "github.com/theplant/htmlgo"
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
	IsPreview         bool
	Locale            string
}

type Builder struct {
	prefix            string
	wb                *web.Builder
	db                *gorm.DB
	containerBuilders []*ContainerBuilder
	ps                *presets.Builder
	pageStyle         h.HTMLComponent
	pageLayoutFunc    PageLayoutFunc
	preview           http.Handler
	images            http.Handler
	imagesPrefix      string
}

const (
	openTemplateDialogEvent = "openTemplateDialogEvent"
	selectTemplateEvent     = "selectTemplateEvent"
)

func New(db *gorm.DB) *Builder {
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
		ExtraAsset("/vue-shadow-dom.js", "text/javascript", ShadowDomComponentsPack())

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
	pm = pb.Model(&Page{})
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
		err = pageValidator(c)
		return
	})

	eb.Field("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		categories := []*Category{}
		if err := db.Model(&Category{}).Find(&categories).Error; err != nil {
			panic(err)
		}
		var showURL h.HTMLComponent
		if p.CategoryID != 0 {
			var c *Category
			for _, e := range categories {
				if e.ID == p.CategoryID {
					c = e
					break
				}
			}

			u := os.Getenv("BASE_URL") + c.Path + "/" + p.Slug
			showURL = h.Div(
				h.A().Text(u).Href(u).Target("_blank"),
			).Class("mt-n2 mb-4")
		}

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}

		return h.Div(
			showURL,
			VAutocomplete().Label("Category").FieldName(field.Name).
				Items(categories).Value(p.CategoryID).ItemText("Path").ItemValue("ID").
				ErrorMessages(vErr.GetFieldErrors("Page.Category")...),
		).ClassIf("mb-4", p.GetStatus() != "")
	})

	eb.Field("TemplateSelection").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		// Only displayed when create action
		if p.GetStatus() == "" {
			return h.Div(
				web.Portal().Name("TemplateDialog"),
				VRow(
					VCol(
						web.Portal(
							VTextField().Disabled(true).Label("Template ID"),
						).Name("TemplateIDTextField"),
					),
					VCol(
						web.Portal(
							VTextField().Disabled(true).Label("Template Name"),
						).Name("TemplateNameTextField"),
					),
				),
				VRow(
					VCol(
						VBtn("Create From Template").Color("primary").
							Attr("@click", web.Plaid().EventFunc(openTemplateDialogEvent).Go()),
					),
				),
			).Class("my-2").Attr(web.InitContextVars, `{showTemplateDialog: false}`)
		}
		return nil
	})

	eb.Field("EditContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		if p.GetStatus() == publish.StatusDraft {
			return h.Div(
				VBtn("Edit Containers").
					Target("_blank").
					Href(fmt.Sprintf("%s/editors/%d?version=%s", b.prefix, p.ID, p.GetVersion())).
					Color("secondary"),
			)
		}
		return nil
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		p := obj.(*Page)

		err = db.Transaction(func(tx *gorm.DB) (inerr error) {
			if inerr = gorm2op.DataOperator(tx).Save(obj, id, ctx); inerr != nil {
				return
			}

			if strings.Contains(ctx.R.RequestURI, views.SaveNewVersionEvent) {
				if inerr = b.copyContainersToNewPageVersion(tx, int(p.ID), p.ParentVersion, p.GetVersion()); inerr != nil {
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
				if inerr = b.copyContainersToAnotherPage(tx, tplID, templateVersion, int(p.ID), p.GetVersion()); inerr != nil {
					return
				}
			}
			return
		})

		return
	})

	b.configSharedContainer(pb, db)
	b.configDemoContainer(pb, db)
	b.configTemplate(pb, db)
	b.configCategory(pb, db)
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

func (b *Builder) configCategory(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
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

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Category)
		err = categoryValidator(c)
		return
	})

	eb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		category := obj.(*Category)
		u := os.Getenv("BASE_URL") + category.Path

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

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateIDTextField",
			Body: VTextField().Disabled(true).Label("Template ID").Value(ID),
		})
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateNameTextField",
			Body: VTextField().Disabled(true).Label("Template Name").Value(Name),
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

		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "TemplateDialog",
			Body: VDialog(
				VCard(
					VCardTitle(
						h.Text("Create From Template"),
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
	listing.SearchFunc(sharedContainersearcher(db, pm))
	listing.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		tdbind := cell
		c := obj.(*Container)

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

func (b *Builder) configTemplate(pb *presets.Builder, db *gorm.DB) (pm *presets.ModelBuilder) {
	pm = pb.Model(&Template{}).URIName("page_templates").Label("Templates")

	pm.Listing("ID", "Name", "Description")

	eb := pm.Editing("Name", "Description", "EditContainer")
	eb.Field("EditContainer").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		m := obj.(*Template)
		if m.ID == 0 {
			return nil
		}
		return h.Div(
			VBtn("Edit Containers").
				Target("_blank").
				Href(fmt.Sprintf("%s/editors/%d?tpl=1", b.prefix, m.ID)).
				Color("secondary"),
		)
	})

	return
}

func sharedContainersearcher(db *gorm.DB, mb *presets.ModelBuilder) presets.SearchFunc {
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
