package pagebuilder

import (
	"fmt"
	"net/url"

	h "github.com/theplant/htmlgo"
	"goji.io/v3/pat"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
)

const pageBuilderRightContentPortal = "pageBuilderRightContentPortal"

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, p *Page, err error) {
	p = new(Page)
	var (
		body, editContainerDrawer h.HTMLComponent

		primarySlug = p.PrimaryColumnValuesBySlug(pat.Param(ctx.R, presets.ParamID))
		pageID      = primarySlug["id"]
		version     = primarySlug["version"]
		localeCode  = primarySlug["locale_code"]
	)
	deviceQueries := url.Values{}
	deviceQueries.Add("tab", "content")
	body, p, err = b.renderPageOrTemplate(ctx, pageID, version, localeCode, true)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", pageID, p.Title)
	ctx.R.Form.Set(paramStatus, p.GetStatus())
	if editContainerDrawer, err = b.renderContainersSortedList(ctx); err != nil {
		return
	}
	if ctx.R.FormValue(paramsIsNotEmpty) == "" {
		editContainerDrawer = b.renderContainersList(ctx)
	}

	r.Body = web.Scope(
		VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
			Class("mt-6").
			Fluid(true),
		VNavigationDrawer(
			web.Portal(editContainerDrawer).Name(pageBuilderRightContentPortal),
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
