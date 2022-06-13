package pagebuilder

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	. "github.com/goplaid/x/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
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
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	ctx.Injector.HeadHTMLComponent("style", b.pageStyle, true)

	var comps []h.HTMLComponent
	var p *Page
	err = b.db.First(&p, "id = ? and version = ?", id, version).Error
	if err != nil {
		return
	}
	comps, err = b.renderContainers(ctx, p.ID, p.GetVersion(), true)
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
	id := pat.Param(ctx.R, "id")
	version := ctx.R.FormValue("version")
	var comps []h.HTMLComponent
	var p *Page
	err = b.db.First(&p, "id = ? and version = ?", id, version).Error
	if err != nil {
		return
	}
	comps, err = b.renderContainers(ctx, p.ID, p.GetVersion(), false)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", id, p.Title)
	device, _ := b.getDevice(ctx)
	var body h.HTMLComponent
	body = h.Components(comps...)
	if b.pageLayoutFunc != nil {
		input := &PageLayoutInput{
			IsEditor: true,
			Page:     p,
		}
		body = b.pageLayoutFunc(h.Components(comps...), input, ctx)
	}
	r.Body = h.Components(
		VAppBar(
			VSpacer(),

			VBtn("").Icon(true).Children(
				VIcon("phone_iphone"),
			).Attr("@click", web.Plaid().Queries(url.Values{"version": []string{version}, "device": []string{"phone"}}).PushState(true).Go()).
				Class("mr-10").InputValue(device == "phone"),

			VBtn("").Icon(true).Children(
				VIcon("tablet_mac"),
			).Attr("@click", web.Plaid().Queries(url.Values{"version": []string{version}, "device": []string{"tablet"}}).PushState(true).Go()).
				Class("mr-10").InputValue(device == "tablet"),

			VBtn("").Icon(true).Children(
				VIcon("laptop_mac"),
			).Attr("@click", web.Plaid().Queries(url.Values{"version": []string{version}, "device": []string{"laptop"}}).PushState(true).Go()).
				InputValue(device == "laptop"),

			VSpacer(),
			VBtn("Preview").Text(true).Href(b.prefix+fmt.Sprintf("/preview?id=%s&version=%s", id, version)).Target("_blank"),
			b.addContainerMenu(id, version),
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
		style = "width: 600px"
	case "tablet":
		style = "width: 960px"
	case "laptop":
		style = "width: 1264px"
	}

	return
}

func (b *Builder) renderContainers(ctx *web.EventContext, pageID uint, pageVersion string, preview bool) (r []h.HTMLComponent, err error) {
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

		pure := ec.builder.renderFunc(obj, ctx)

		if preview {
			r = append(r, pure)
		} else {
			_, width := b.getDevice(ctx)

			r = append(r, b.containerEditor(obj, ec, pure, width))
		}
	}
	return
}

const AddContainerEvent = "page_builder_AddContainerEvent"
const DeleteContainerEvent = "page_builder_DeleteContainerEvent"
const MoveContainerEvent = "page_builder_MoveContainerEvent"

func (b *Builder) AddContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	containerName := ctx.R.FormValue(paramContainerName)

	var modelID uint
	modelID, err = b.AddContainerToPage(pageID, pageVersion, containerName)

	// r.Location = web.Location(url.Values{})
	r.VarsScript = web.Plaid().
		URL(b.ContainerByName(containerName).mb.Info().ListingHref()).
		EventFunc(actions.Edit).
		Query(presets.ParamID, fmt.Sprint(modelID)).
		Go()
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
	//pageID := ctx.QueryAsInt(paramPageID)
	//pageVersion := ctx.R.FormValue(paramPageVersion)
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
		ModelID:      modelID,
		DisplayOrder: maxOrder.Float64 + 8,
	}).Error
	if err != nil {
		return
	}
	return
}

const (
	paramPageID        = "pageID"
	paramPageVersion   = "pageVersion"
	paramContainerID   = "containerID"
	paramDirection     = "direction"
	paramContainerName = "containerName"
)

func (b *Builder) containerEditor(obj interface{}, ec *editorContainer, c h.HTMLComponent, width string) (r h.HTMLComponent) {

	return VRow(
		VCol(
			h.Div(
				h.Tag("shadow-root").Children(
					h.Div(
						b.pageStyle,
						c,
					).Style("position:relative; z-index: 0;"),
				),
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
							//Query(paramPageID, ec.container.PageID).
							//Query(paramPageVersion, ec.container.PageVersion).
							Query(paramContainerID, ec.container.ID).
							Go(),
					),
				),
			).OffsetY(true),

			VBtn("").Color("primary").Children(
				h.Text(ec.builder.name),
			).Text(true).
				Class("my-2 float-right").Attr("@click",
				web.Plaid().
					URL(ec.builder.mb.Info().ListingHref()).
					EventFunc(actions.Edit).
					Query(presets.ParamID, fmt.Sprint(reflectutils.MustGet(obj, "ID"))).
					Go(),
			),
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

			innerPr.Body.(h.HTMLComponent),
		).Id("vt-app").Attr(web.InitContextVars, `{presetsRightDrawer: false}`)

		return
	}
}

func (b *Builder) addContainerMenu(pageID, pageVersion string) h.HTMLComponent {
	var items []h.HTMLComponent

	for _, builder := range b.containerBuilders {
		items = append(items,
			VCol(
				VCard(

					VCardTitle(h.Text(builder.name)),
					VCardActions(
						VSpacer(),
						VBtn("Select").
							Text(true).
							Color("primary").Attr("@click",
							web.Plaid().EventFunc(AddContainerEvent).
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

	return VMenu(
		web.Slot(
			VBtn("Add Container").Text(true).
				Attr("v-bind", "attrs", "v-on", "on"),
		).Name("activator").Scope("{ on, attrs }"),
		VSheet(
			VContainer(
				VRow(
					items...,
				),
			),
		),
	).OffsetY(true).NudgeWidth(600).CloseOnContentClick(true)
}
