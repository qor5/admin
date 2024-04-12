package pagebuilder

import (
	"fmt"
	"net/url"

	"github.com/qor5/admin/publish"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
)

func (b *Builder) PageContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.Form.Get("locale")
	isLocalizable := ctx.R.Form.Has("locale")
	var body h.HTMLComponent
	var containerList h.HTMLComponent
	var device string
	var p *Page
	var previewHref string
	_ = previewHref
	deviceQueries := url.Values{}
	deviceQueries.Add("tab", "content")
	if isTpl {
		previewHref = fmt.Sprintf("/preview?id=%s&tpl=1", id)
		// deviceQueries.Add("tpl", "1")
		if isLocalizable && l10nON {
			previewHref = fmt.Sprintf("/preview?id=%s&tpl=1&locale=%s", id, locale)
			// deviceQueries.Add("locale", locale)
		}
	} else {
		previewHref = fmt.Sprintf("/preview?id=%s&version=%s", id, version)
		// deviceQueries.Add("version", version)

		if isLocalizable && l10nON {
			previewHref = fmt.Sprintf("/preview?id=%s&version=%s&locale=%s", id, version, locale)
			// deviceQueries.Add("locale", locale)
		}
	}
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

	containerList, err = b.renderContainersList(ctx, p.ID, p.GetVersion(), p.GetLocale(), p.GetStatus() != publish.StatusDraft)
	if err != nil {
		return
	}
	// msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	r.Body = h.Components(
		VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
			Class("mt-6").
			Fluid(true),
		web.Scope(
			VNavigationDrawer(
				VContainer(
					VRow(
						VCol(
							web.Scope(
								VBtnToggle(
									VBtn("").Icon(true).Children(
										VIcon("mdi-laptop").Size(SizeSmall),
									).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", DeviceComputer).PushState(true).Go()),
									VBtn("").Icon(true).Children(
										VIcon("mdi-tablet").Size(SizeSmall),
									).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", DeviceTablet).PushState(true).Go()),
									VBtn("").Icon(true).Children(
										VIcon("mdi-cellphone").Size(SizeSmall),
									).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", DevicePhone).PushState(true).Go()),
								).Class("pa-2 rounded-lg").Attr("v-model", "toggleLocals.activeDevice").Density(DensityCompact),
							).VSlot("{ locals : toggleLocals}").Init(fmt.Sprintf(`{activeDevice: %d}`, activeDevice)),
						).Cols(9).Class("pa-2"),
						VCol(
							VBtn("").Icon("mdi-eye").Href(b.prefix+previewHref).To("_blank"),
						).Cols(3).Class("pa-2 d-flex justify-center"),
					),
				),
				containerList,
			).Location(LocationRight).
				Permanent(true).
				Width(420),
		),
	)
	return
}
