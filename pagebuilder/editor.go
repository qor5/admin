package pagebuilder

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
	"gorm.io/gorm"
)

const (
	AddContainerDialogEvent    = "page_builder_AddContainerDialogEvent"
	AddContainerEvent          = "page_builder_AddContainerEvent"
	DeleteContainerEvent       = "page_builder_DeleteContainerEvent"
	MoveContainerEvent         = "page_builder_MoveContainerEvent"
	MarkAsSharedContainerEvent = "page_builder_MarkAsSharedContainerEvent"
	RenameDialogEvent          = "page_builder_RenameDialogEvent"
	RenameContainerEvent       = "page_builder_RenameContainerEvent"

	paramPageID          = "pageID"
	paramPageVersion     = "pageVersion"
	paramContainerID     = "containerID"
	paramDirection       = "direction"
	paramContainerName   = "containerName"
	paramSharedContainer = "sharedContainer"
	paramModelID         = "modelID"
)

//go:embed dist
var box embed.FS

func ShadowDomComponentsPack() web.ComponentsPack {
	v, err := box.ReadFile("dist/shadow.min.js")
	if err != nil {
		panic(err)
	}
	return web.ComponentsPack(v)
}

func (b *Builder) Preview(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	ctx.Injector.HeadHTMLComponent("style", b.pageStyle, true)

	var comps []h.HTMLComponent
	var p *Page
	if isTpl {
		tpl := &Template{}
		err = b.db.First(tpl, "id = ?", id).Error
		if err != nil {
			return
		}
		p = tpl.Page()
		version = p.Version.Version
	} else {
		err = b.db.First(&p, "id = ? and version = ?", id, version).Error
		if err != nil {
			return
		}
	}
	comps, err = b.renderContainers(ctx, p.ID, p.GetVersion(), false)
	if err != nil {
		return
	}

	r.PageTitle = p.Title
	r.Body = h.Components(comps...)
	if b.pageLayoutFunc != nil {
		input := &PageLayoutInput{
			IsPreview: true,
			Page:      p,
		}
		r.Body = b.pageLayoutFunc(h.Components(comps...), input, ctx)
	}
	return
}

func (b *Builder) Editor(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := pat.Param(ctx.R, "id")
	version := ctx.R.FormValue("version")
	var comps []h.HTMLComponent
	var body h.HTMLComponent
	var device string
	var p *Page
	var previewHref string
	deviceQueries := url.Values{}
	if isTpl {
		tpl := &Template{}
		err = b.db.First(tpl, "id = ?", id).Error
		if err != nil {
			return
		}
		p = tpl.Page()
		version = p.Version.Version
		previewHref = fmt.Sprintf("/preview?id=%s&tpl=1", id)
		deviceQueries.Add("tpl", "1")
	} else {
		err = b.db.First(&p, "id = ? and version = ?", id, version).Error
		if err != nil {
			return
		}
		previewHref = fmt.Sprintf("/preview?id=%s&version=%s", id, version)
		deviceQueries.Add("version", version)
	}
	if p.GetStatus() == publish.StatusDraft {
		comps, err = b.renderContainers(ctx, p.ID, p.GetVersion(), true)
		if err != nil {
			return
		}
		r.PageTitle = fmt.Sprintf("Editor for %s: %s", id, p.Title)
		device, _ = b.getDevice(ctx)
		body = h.Components(comps...)
		if b.pageLayoutFunc != nil {
			input := &PageLayoutInput{
				IsEditor: true,
				Page:     p,
			}
			body = b.pageLayoutFunc(h.Components(comps...), input, ctx)
		}
	} else {
		body = h.Text(perm.PermissionDenied.Error())
	}

	r.Body = h.Components(
		VAppBar(
			VSpacer(),

			VBtn("").Icon(true).Children(
				VIcon("phone_iphone"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "phone").PushState(true).Go()).
				Class("mr-10").InputValue(device == "phone"),

			VBtn("").Icon(true).Children(
				VIcon("tablet_mac"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "tablet").PushState(true).Go()).
				Class("mr-10").InputValue(device == "tablet"),

			VBtn("").Icon(true).Children(
				VIcon("laptop_mac"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "laptop").PushState(true).Go()).
				InputValue(device == "laptop"),

			VSpacer(),
			VBtn("Preview").Text(true).Href(b.prefix+previewHref).Target("_blank"),
			VBtn("Add Container").Text(true).Attr("@click",
				web.Plaid().
					EventFunc(AddContainerDialogEvent).
					Query(paramPageID, id).
					Query(paramPageVersion, version).
					Query(presets.ParamOverlay, actions.Dialog).
					Go(),
			),
		).Dark(true).
			Color("primary").
			App(true),

		VMain(
			VContainer(body).Attr("v-keep-scroll", "true").
				Class("mt-6").
				Fluid(true),
		),
	)

	return
}

func (b *Builder) getDevice(ctx *web.EventContext) (device string, style string) {
	device = ctx.R.FormValue("device")
	if len(device) == 0 {
		device = "phone"
	}

	switch device {
	case "phone":
		style = "width: 414px"
	case "tablet":
		style = "width: 768px"
	case "laptop":
		style = "width: 1264px"
	}

	return
}

func (b *Builder) renderContainers(ctx *web.EventContext, pageID uint, pageVersion string, isEditor bool) (r []h.HTMLComponent, err error) {
	var cons []*Container
	err = b.db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ?", pageID, pageVersion).Error
	if err != nil {
		return
	}

	cbs := b.getContainerBuilders(cons)
	for _, ec := range cbs {
		obj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(obj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}
		device, width := b.getDevice(ctx)
		input := RenderInput{
			IsEditor: isEditor,
			Device:   device,
		}
		pure := ec.builder.renderFunc(obj, &input, ctx)

		if isEditor {
			r = append(r, b.containerEditor(ctx, obj, ec, pure, width))
		} else {
			r = append(r, pure)
		}
	}
	return
}

func (b *Builder) AddContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	containerName := ctx.R.FormValue(paramContainerName)
	sharedContainer := ctx.R.FormValue(paramSharedContainer)
	modelID := ctx.QueryAsInt(paramModelID)
	var newModelID uint
	if sharedContainer == "true" {
		err = b.AddSharedContainerToPage(pageID, pageVersion, containerName, uint(modelID))
		r.PushState = web.Location(url.Values{})
	} else {
		newModelID, err = b.AddContainerToPage(pageID, pageVersion, containerName)
		r.VarsScript = web.Plaid().
			URL(b.ContainerByName(containerName).mb.Info().ListingHref()).
			EventFunc(actions.Edit).
			Query(presets.ParamID, fmt.Sprint(newModelID)).
			Go()
	}

	return
}

func (b *Builder) MoveContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	direction := ctx.R.FormValue(paramDirection)
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	containerID := ctx.QueryAsInt(paramContainerID)
	err = b.MoveContainerOrder(pageID, pageVersion, containerID, direction)

	r.PushState = web.Location(url.Values{})

	return
}

type moveDirection string

const (
	up   moveDirection = "up"
	down moveDirection = "down"
)

func (b *Builder) MoveContainerOrder(pageID int, pageVersion string, containerID int, direction string) (err error) {

	var current Container
	err = b.db.Find(&current, "id = ?", containerID).Error
	if err != nil {
		return
	}

	var closest []*Container

	if moveDirection(direction) == up {
		b.db.Order("display_order DESC").
			Where("page_id = ? and page_version = ?", pageID, pageVersion).
			Limit(2).
			Find(&closest, "display_order < ?", current.DisplayOrder)

	} else {
		b.db.Order("display_order ASC").
			Where("page_id = ? and page_version = ?", pageID, pageVersion).
			Limit(2).
			Find(&closest, "display_order > ?", current.DisplayOrder)
	}

	if len(closest) > 0 {
		var displayOrder float64 = 0
		if len(closest) == 1 {
			if moveDirection(direction) == up {
				displayOrder = closest[0].DisplayOrder - 8
			} else {
				displayOrder = closest[0].DisplayOrder + 8
			}
		} else {
			displayOrder = (closest[0].DisplayOrder + closest[1].DisplayOrder) / 2
		}

		err = b.db.Model(&Container{}).Where("id = ?", containerID).
			Update("display_order", displayOrder).Error
		if err != nil {
			return
		}
	}
	return
}

func (b *Builder) DeleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	// pageID := ctx.QueryAsInt(paramPageID)
	// pageVersion := ctx.R.FormValue(paramPageVersion)
	containerID := ctx.QueryAsInt(paramContainerID)

	err = b.db.Delete(&Container{}, "id = ?", containerID).Error
	if err != nil {
		return
	}
	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) AddContainerToPage(pageID int, pageVersion, containerName string) (modelID uint, err error) {
	model := b.ContainerByName(containerName).NewModel()
	var dc DemoContainer
	b.db.Where("model_name = ?", containerName).First(&dc)
	if dc.ID != 0 && dc.ModelID != 0 {
		b.db.Where("id = ?", dc.ModelID).First(model)
		reflectutils.Set(model, "ID", uint(0))
	}

	err = b.db.Create(model).Error
	if err != nil {
		return
	}

	var maxOrder sql.NullFloat64
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ?", pageID, pageVersion).Scan(&maxOrder).Error
	if err != nil {
		return
	}

	modelID = reflectutils.MustGet(model, "ID").(uint)
	err = b.db.Create(&Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		Name:         containerName,
		DisplayName:  containerName,
		ModelID:      modelID,
		DisplayOrder: maxOrder.Float64 + 8,
	}).Error
	if err != nil {
		return
	}
	return
}

func (b *Builder) AddSharedContainerToPage(pageID int, pageVersion, containerName string, modelID uint) (err error) {
	var c Container
	err = b.db.First(&c, "name = ? AND model_id = ? AND shared = true", containerName, modelID).Error
	if err != nil {
		return
	}
	var maxOrder sql.NullFloat64
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ?", pageID, pageVersion).Scan(&maxOrder).Error
	if err != nil {
		return
	}

	err = b.db.Create(&Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		Name:         containerName,
		DisplayName:  c.DisplayName,
		ModelID:      modelID,
		Shared:       true,
		DisplayOrder: maxOrder.Float64 + 8,
	}).Error
	if err != nil {
		return
	}
	return
}

func (b *Builder) copyContainersToNewPageVersion(db *gorm.DB, pageID int, oldPageVersion, newPageVersion string) (err error) {
	return b.copyContainersToAnotherPage(db, pageID, oldPageVersion, pageID, newPageVersion)
}

func (b *Builder) copyContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion string, toPageID int, toPageVersion string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ?", pageID, pageVersion).Error
	if err != nil {
		return
	}

	for _, c := range cons {
		newModelID := c.ModelID
		if !c.Shared {
			model := b.ContainerByName(c.Name).NewModel()
			if err = db.First(model, "id = ?", c.ModelID).Error; err != nil {
				return
			}
			if err = reflectutils.Set(model, "ID", uint(0)); err != nil {
				return
			}
			if err = db.Create(model).Error; err != nil {
				return
			}
			newModelID = reflectutils.MustGet(model, "ID").(uint)
		}

		if err = db.Create(&Container{
			PageID:       uint(toPageID),
			PageVersion:  toPageVersion,
			Name:         c.Name,
			DisplayName:  c.DisplayName,
			ModelID:      newModelID,
			DisplayOrder: c.DisplayOrder,
			Shared:       c.Shared,
		}).Error; err != nil {
			return
		}
	}
	return
}

func (b *Builder) MarkAsSharedContainerEvent(ctx *web.EventContext) (r web.EventResponse, err error) {
	containerID := ctx.QueryAsInt(paramContainerID)
	err = b.db.Model(&Container{}).Where("id = ?", containerID).Update("shared", true).Error
	if err != nil {
		return
	}
	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) RenameContainerEvent(ctx *web.EventContext) (r web.EventResponse, err error) {
	containerID := ctx.QueryAsInt(paramContainerID)
	name := ctx.R.FormValue("DisplayName")
	var c Container
	err = b.db.First(&c, "id = ?  ", containerID).Error
	if err != nil {
		return
	}
	if c.Shared {
		err = b.db.Model(&Container{}).Where("name = ? AND model_id = ?", c.Name, c.ModelID).Update("display_name", name).Error
		if err != nil {
			return
		}
	} else {
		err = b.db.Model(&Container{}).Where("id = ?", containerID).Update("display_name", name).Error
		if err != nil {
			return
		}
	}

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) RenameDialogEvent(ctx *web.EventContext) (r web.EventResponse, err error) {
	containerID := ctx.QueryAsInt(paramContainerID)
	name := ctx.R.FormValue(paramContainerName)
	okAction := web.Plaid().EventFunc(RenameContainerEvent).Query(paramContainerID, containerID).Go()
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogPortalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(h.Text("Rename")),
					VCardText(
						VTextField().FieldName("DisplayName").Value(name),
					),
					VCardActions(
						VSpacer(),
						VBtn("Cancel").
							Depressed(true).
							Class("ml-2").
							On("click", "locals.renameDialog = false"),

						VBtn("OK").
							Color("primary").
							Depressed(true).
							Dark(true).
							Attr("@click", okAction),
					),
				),
			).MaxWidth("400px").
				Attr("v-model", "locals.renameDialog"),
		).Init("{renameDialog:true}").VSlot("{locals}"),
	})
	return
}

func (b *Builder) AddContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	// okAction := web.Plaid().EventFunc(RenameContainerEvent).Query(paramContainerID, containerID).Go()

	var containers []h.HTMLComponent
	for _, builder := range b.containerBuilders {
		cover := builder.cover
		if cover == "" {
			cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
		}
		containers = append(containers,
			VCol(
				VCard(
					VImg().Src(cover).Height(200),
					VCardActions(
						VCardTitle(h.Text(builder.name)),
						VSpacer(),
						VBtn("Select").
							Text(true).
							Color("primary").Attr("@click",
							"locals.addContainerDialog = false;"+web.Plaid().EventFunc(AddContainerEvent).
								Query(paramPageID, pageID).
								Query(paramPageVersion, pageVersion).
								Query(paramContainerName, builder.name).
								Go(),
						),
					),
				),
			).Cols(4),
		)
	}

	var cons []*Container
	err = b.db.Select("display_name,name,model_id").Where("shared = true").Group("display_name,name,model_id").Find(&cons).Error
	if err != nil {
		return
	}

	var sharedContainers []h.HTMLComponent
	for _, sharedC := range cons {
		c := b.ContainerByName(sharedC.Name)
		cover := c.cover
		if cover == "" {
			cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(c.name, " ", "")+".png")
		}
		sharedContainers = append(sharedContainers,
			VCol(
				VCard(
					VImg().Src(cover).Height(200),
					VCardActions(
						VCardTitle(h.Text(sharedC.DisplayName)),
						VSpacer(),
						VBtn("Select").
							Text(true).
							Color("primary").Attr("@click",
							"locals.addContainerDialog = false;"+web.Plaid().EventFunc(AddContainerEvent).
								Query(paramPageID, pageID).
								Query(paramPageVersion, pageVersion).
								Query(paramContainerName, sharedC.Name).
								Query(paramModelID, sharedC.ModelID).
								Query(paramSharedContainer, "true").
								Go(),
						),
					),
				),
			).Cols(4),
		)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogPortalName,
		Body: web.Scope(
			VDialog(
				VTabs(
					VTab(h.Text("New")),
					VTabItem(
						VSheet(
							VContainer(
								VRow(
									containers...,
								),
							),
						),
					).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
					VTab(h.Text("Shared")),
					VTabItem(
						VSheet(
							VContainer(
								VRow(
									sharedContainers...,
								),
							),
						),
					).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
				),
			).Width("1200px").Attr("v-model", "locals.addContainerDialog"),
		).Init("{addContainerDialog:true}").VSlot("{locals}"),
	})

	return
}

func (b *Builder) containerEditor(ctx *web.EventContext, obj interface{}, ec *editorContainer, c h.HTMLComponent, width string) (r h.HTMLComponent) {
	containerContent := h.Div(
		b.pageStyle,
		c,
	)
	containerName := ec.container.DisplayName
	if containerName == "" {
		containerName = ec.container.Name
	}
	return VRow(
		VCol(
			h.Div(
				h.RawHTML(fmt.Sprintf("<iframe frameborder='0' scrolling='no' srcdoc=\"%s\" @load='$event.target.style.height=$event.target.contentWindow.document.body.parentElement.offsetHeight+\"px\"' style='width:100%%; display:block; border:none; padding:0; margin:0'></iframe>",
					strings.ReplaceAll(
						h.MustString(containerContent, ctx.R.Context()),
						"\"",
						"&quot;"),
				)),
			).Class("page-builder-container mx-auto").Attr("style", width),
		).Cols(10).Class("pa-0"),

		VCol(
			VMenu(
				web.Slot(
					VBtn("").Color("secondary").Children(
						VIcon("settings"),
					).Icon(true).Class("my-2 mr-4 float-right").
						Attr("v-bind", "attrs", "v-on", "on"),
				).Name("activator").Scope("{ on, attrs }"),

				VList(
					VListItem(
						VListItemTitle(h.Text("Edit")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(actions.Edit).
							Query(presets.ParamID, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
							Go(),
					),
					VListItem(
						VListItemTitle(h.Text("Rename")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(RenameDialogEvent).
							Query(paramContainerID, ec.container.ID).
							Query(paramContainerName, containerName).
							Query(presets.ParamOverlay, actions.Dialog).
							Go(),
					),
					h.If(!ec.container.Shared,
						VListItem(
							VListItemTitle(h.Text("Mark As Shared Container")),
						).Attr("@click",
							web.Plaid().
								URL(ec.builder.mb.Info().ListingHref()).
								EventFunc(MarkAsSharedContainerEvent).
								// Query(paramPageID, ec.container.PageID).
								// Query(paramPageVersion, ec.container.PageVersion).
								Query(paramContainerID, ec.container.ID).
								Go(),
						),
					),
					VListItem(
						VListItemTitle(h.Text("Move Up")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(MoveContainerEvent).
							Query(paramDirection, string(up)).
							Query(paramPageID, ec.container.PageID).
							Query(paramPageVersion, ec.container.PageVersion).
							Query(paramContainerID, ec.container.ID).
							Go(),
					),

					VListItem(
						VListItemTitle(h.Text("Move Down")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(MoveContainerEvent).
							Query(paramDirection, string(down)).
							Query(paramPageID, ec.container.PageID).
							Query(paramPageVersion, ec.container.PageVersion).
							Query(paramContainerID, ec.container.ID).
							Go(),
					),
					VDivider(),

					VListItem(
						VListItemTitle(h.Text("Delete")),
					).Attr("@click",
						web.Plaid().
							URL(ec.builder.mb.Info().ListingHref()).
							EventFunc(DeleteContainerEvent).
							// Query(paramPageID, ec.container.PageID).
							// Query(paramPageVersion, ec.container.PageVersion).
							Query(paramContainerID, ec.container.ID).
							Go(),
					),
				),
			).OffsetY(true),

			VBtn("").Color("primary").Children(
				h.If(ec.container.Shared, VIcon("shared").Small(true).Right(true)),
				h.Text(containerName),
			).Text(true).
				Class("my-2 float-right text-capitalize").Attr("@click",
				web.Plaid().
					URL(ec.builder.mb.Info().ListingHref()).
					EventFunc(actions.Edit).
					Query(presets.ParamID, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
					Go(),
			).Class("pa-0"),
		).Cols(2).Class("pa-0"),
	).Attr("style", "border-top: 0.5px dashed gray")

}

type editorContainer struct {
	builder   *ContainerBuilder
	container *Container
}

func (b *Builder) getContainerBuilders(cs []*Container) (r []*editorContainer) {
	for _, c := range cs {
		for _, cb := range b.containerBuilders {
			if cb.name == c.Name {
				r = append(r, &editorContainer{
					builder:   cb,
					container: c,
				})
			}
		}
	}
	return
}

const (
	dialogPortalName = "pagebuilder_DialogPortalName"
)

func (b *Builder) pageEditorLayout(in web.PageFunc, config *presets.LayoutConfig) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {

		ctx.Injector.HeadHTML(strings.Replace(`
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto+Mono">
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500">
			<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
			<link rel="stylesheet" href="{{prefix}}/assets/main.css">
			<script src='{{prefix}}/assets/vue.js'></script>

<style>
	.page-builder-container {
		overflow: hidden;
		box-shadow: -10px 0px 13px -7px rgba(0,0,0,.3), 10px 0px 13px -7px rgba(0,0,0,.18), 5px 0px 15px 5px rgba(0,0,0,.12);	
}
</style>

			<style>
				[v-cloak] {
					display: none;
				}
			</style>
		`, "{{prefix}}", b.prefix, -1))

		b.ps.InjectExtraAssets(ctx)

		if len(os.Getenv("DEV_PRESETS")) > 0 {
			ctx.Injector.TailHTML(`
<script src='http://localhost:3080/js/chunk-vendors.js'></script>
<script src='http://localhost:3080/js/app.js'></script>
<script src='http://localhost:3100/js/chunk-vendors.js'></script>
<script src='http://localhost:3100/js/app.js'></script>
			`)

		} else {
			ctx.Injector.TailHTML(strings.Replace(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix, -1))
		}

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}

		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, "Page Builder")
		pr.Body = VApp(

			web.Portal().Name(presets.RightDrawerPortalName),
			web.Portal().Name(dialogPortalName),

			innerPr.Body.(h.HTMLComponent),
		).Id("vt-app").Attr(web.InitContextVars, `{presetsRightDrawer: false, dialogPortalName: false}`)

		return
	}
}
