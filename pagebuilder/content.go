package pagebuilder

import (
	"fmt"
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
	isLocalizable := ctx.R.Form.Has("locale")
	var body h.HTMLComponent
	var containerList h.HTMLComponent
	var device string
	var p *Page
	var previewHref string
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
	// msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	r.Body = h.Tag("vx-drag-listener").Attr("@drop", action).Children(
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

func (b *Builder) previewHref(id, version, locale string) string {
	return b.prefix + fmt.Sprintf("/preview?id=%s&version=%s&locale=%s", id, version, locale)
}
