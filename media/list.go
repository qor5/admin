package media

import (
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
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
		filed := mediaLibraryListField
		cfg := &media_library.MediaBoxConfig{}

		ctx.WithContextValue(presets.CtxPageTitleComponent, h.Div(

			VAppBarTitle(h.Text("Media Library")),
			VSpacer(),
			searchComponent(ctx, filed, cfg, true),
		).Class(W100, "d-flex align-center"))
		r.Body = h.Components(
			web.Portal(
				fileChooserDialogContent(mb, filed, ctx, cfg),
			).Name(dialogContentPortalName(filed)),
		)
		return
	})
}
