package pagebuilder

import (
	"fmt"
	"net/url"
	"strconv"

	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, p *Page, err error) {
	pageID := ctx.R.FormValue(presets.ParamID)
	var (
		body          h.HTMLComponent
		containerList h.HTMLComponent
	)
	deviceQueries := url.Values{}
	deviceQueries.Add("tab", "content")
	body, p, err = b.renderPageOrTemplate(ctx, true)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", pageID, p.Title)
	ctx.R.Form.Set(presets.ParamID, strconv.Itoa(int(p.ID)))
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

func (b *Builder) previewHref(ctx *web.EventContext) string {
	var (
		isTpl         = ctx.R.FormValue(paramsTpl) != ""
		id            = ctx.R.FormValue(paramPageID)
		version       = ctx.R.FormValue(paramPageVersion)
		locale        = ctx.R.Form.Get(paramLocale)
		isLocalizable = ctx.R.Form.Has(paramLocale)
		ur            = url.Values{}
	)
	ur.Add(presets.ParamID, id)
	if isTpl {
		if isLocalizable && l10nON {
			ur.Add(paramsTpl, "1")
			ur.Add(paramLocale, locale)
		}
	} else {
		ur.Add(paramPageVersion, version)
		if isLocalizable && l10nON {
			ur.Add(paramLocale, locale)
		}
	}
	return b.prefix + "/preview?" + ur.Encode()
}
