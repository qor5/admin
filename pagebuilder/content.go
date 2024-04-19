package pagebuilder

import (
	"fmt"
	"github.com/qor5/admin/v3/presets"
	"net/url"

	"github.com/qor5/admin/v3/publish"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.Form.Get("locale")
	var body h.HTMLComponent
	var containerList h.HTMLComponent
	var device string
	var p *Page
	deviceQueries := url.Values{}
	deviceQueries.Add("tab", "content")
	body, p, err = b.renderPageOrTemplate(ctx, isTpl, id, version, locale, true)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", id, p.Title)
	device, _ = b.getDevice(ctx)
	activeDevice := 0
	_ = activeDevice
	switch device {
	case DeviceTablet:
		activeDevice = 1
	case DevicePhone:
		activeDevice = 2
	}

	containerList = b.renderContainersList(ctx, p.GetStatus() != publish.StatusDraft)
	action := web.Plaid().
		URL(fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, p.ID, p.GetVersion(), locale)).
		EventFunc(AddContainerEvent).
		Query(paramPageID, p.ID).
		Query(paramPageVersion, p.GetVersion()).
		Query(paramLocale, locale).
		Query(paramContainerName, web.Var("$event.start.id")).
		Query(paramSharedContainer, web.Var(`$event.start.getAttribute("shared")`)).
		Query(paramModelID, web.Var(`$event.start.getAttribute("modelid")`)).
		Go()
	r.Body = h.Tag("vx-drag-listener").Attr("@drop", action).Children(
		VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
			Class("mt-6").
			Fluid(true),
		VNavigationDrawer(
			web.Portal(containerList).Name(presets.RightDrawerContentPortalName),
		).Location(LocationRight).
			Permanent(true).
			Width(420),
	)
	return
}

func (b *Builder) previewHref(id, version, locale string) string {
	return b.prefix + fmt.Sprintf("/preview?id=%s&version=%s&locale=%s", id, version, locale)
}
