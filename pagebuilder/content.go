package pagebuilder

import (
	"fmt"
	"github.com/qor5/admin/v3/presets"
	"net/url"
	"strconv"

	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.Form.Get("locale")
	var (
		body          h.HTMLComponent
		containerList h.HTMLComponent
		device        string
		p             *Page
		isEmpty       bool
	)
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
	ctx.R.Form.Set(paramPageID, strconv.Itoa(int(p.ID)))
	ctx.R.Form.Set(paramStatus, p.GetStatus())
	ctx.R.Form.Set(paramPageVersion, p.GetVersion())
	if containerList, isEmpty, err = b.renderContainersSortedList(ctx); err != nil {
		return
	}
	if isEmpty {
		ctx.R.Form.Set(paramsIsNotEmpty, "1")
		containerList = b.renderContainersList(ctx)
	}
	r.Body = h.Components(
		VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
			Class("mt-6").
			Fluid(true),
		VNavigationDrawer(
			web.Portal(containerList).Name(presets.RightDrawerPortalName),
		).Location(LocationRight).
			Permanent(true).
			Width(420),
	)
	return
}

func (b *Builder) previewHref(id, version, locale string) string {
	return b.prefix + fmt.Sprintf("/preview?id=%s&version=%s&locale=%s", id, version, locale)
}
