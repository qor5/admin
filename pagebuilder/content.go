package pagebuilder

import (
	"fmt"
	"net/url"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

const (
	pageBuilderRightContentPortal   = "pageBuilderRightContentPortal"
	pageBuilderLayerContainerPortal = "pageBuilderLayerContainerPortal"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, p *Page, err error) {
	p = new(Page)
	var body h.HTMLComponent

	pageID, version, localeCode := primaryKeys(ctx)
	body, p, err = b.renderPageOrTemplate(ctx, fmt.Sprint(pageID), version, localeCode, true)
	if err != nil {
		return
	}
	r.Body = web.Portal(
		body.(*h.HTMLTagBuilder).Attr(web.VAssign("vars", "{el:$}")...),
	).Name(editorPreviewContentPortal)
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
		if isLocalizable && b.l10n != nil {
			ur.Add(paramsTpl, "1")
			ur.Add(paramLocale, locale)
		}
	} else {
		ur.Add(paramPageVersion, version)
		if isLocalizable && b.l10n != nil {
			ur.Add(paramLocale, locale)
		}
	}
	return b.prefix + "/preview?" + ur.Encode()
}
