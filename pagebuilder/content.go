package pagebuilder

import (
	"fmt"
	"net/url"
	"strconv"

	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	pageID := ctx.R.FormValue(paramPageID)
	var (
		body          h.HTMLComponent
		containerList h.HTMLComponent
		device        string
		p             *Page
	)
	deviceQueries := url.Values{}
	deviceQueries.Add("tab", "content")
	body, p, err = b.renderPageOrTemplate(ctx, true)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", pageID, p.Title)
	device, _ = b.getDevice(ctx)
	activeDevice := 0
	_ = activeDevice
	switch device {
	case DeviceTablet:
		activeDevice = 1
	case DevicePhone:
		activeDevice = 2
	}
	ctx.R.Form.Set(paramPageID, strconv.Itoa(int(p.ID)))
	ctx.R.Form.Set(paramStatus, p.GetStatus())
	ctx.R.Form.Set(paramPageVersion, p.GetVersion())
	if containerList, err = b.renderContainersSortedList(ctx); err != nil {
		return
	}
	if ctx.R.FormValue(paramsIsNotEmpty) == "" {
		containerList = b.renderContainersList(ctx)
	}
	r.Body = web.Scope(
		VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
			Class("mt-6").
			Fluid(true),
		VNavigationDrawer(
			web.Portal(containerList).Name(pageBuilderRightContentPortal),
		).Location(LocationRight).
			Permanent(true).
			Width(420),
	).VSlot("{ locals }").Init(` { el : $ }`)
	return
}

func (b *Builder) previewHref(id, version, locale string) string {
	uv := url.Values{}
	uv.Add(paramPageID, id)
	uv.Add(paramPageVersion, version)
	uv.Add(paramLocale, locale)
	return b.prefix + "/preview?" + uv.Encode()
}
