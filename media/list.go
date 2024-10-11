package media

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

const (
	mediaLibraryListField = "media-library-list"
)

func configList(b *presets.Builder, mb *Builder) {
	mm := b.Model(&media_library.MediaLibrary{}).Label("Media Library").MenuIcon("mdi-image")
	mb.mb = mm
	oldPageFunc := mm.Listing().GetPageFunc()
	mm.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var (
			filed = mediaLibraryListField
			cfg   = &media_library.MediaBoxConfig{}
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		var pr web.PageResponse
		if pr, err = oldPageFunc(ctx); err != nil {
			return
		}
		r.PageTitle = pr.PageTitle
		ctx.WithContextValue(presets.CtxPageTitleComponent, h.Div(
			VAppBarTitle(h.Text(msgr.MediaLibrary)),
			VSpacer(),
			searchComponent(ctx, filed, cfg, true),
		).Class(W100, "d-flex align-center"))
		r.Body = h.Components(
			h.Div(
				web.Portal(
					fileChooserDialogContent(mb, filed, ctx, cfg),
				).Name(dialogContentPortalName(filed)),
			),
		)
		return
	})
}
