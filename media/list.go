package media

import (
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

const (
	mediaLibraryListField = "media-library-list"
	MediaLibraryURIName   = "media-library"
)

func configList(b *presets.Builder, mb *Builder) {
	mm := b.Model(&media_library.MediaLibrary{}).Label("Media Library").MenuIcon("mdi-image").URIName(MediaLibraryURIName)
	mm.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var (
			filed = mediaLibraryListField
			cfg   = &media_library.MediaBoxConfig{}
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
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
