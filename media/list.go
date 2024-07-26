package media

import (
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

const (
	mediaLibraryListField = "media-library-list"
	MediaLibraryURIName   = "media-library"
)

func configList(b *presets.Builder, mb *Builder) {
	mm := b.Model(&media_library.MediaLibrary{}).Label("Media Library").MenuIcon("mdi-image").URIName(MediaLibraryURIName)
	mm.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.PageTitle = "Media Library"
		keyword := ctx.R.FormValue("keyword")
		ctx.R.Form.Set(searchKeywordName(mediaLibraryListField), keyword)
		r.Body = h.Components(
			web.Portal().Name(deleteConfirmPortalName(mediaLibraryListField)),
			web.Portal(
				h.Input("").
					Type("hidden").
					Attr(web.VField(searchKeywordName(mediaLibraryListField), keyword)...),
				fileChooserDialogContent(mb, mediaLibraryListField, ctx, &media_library.MediaBoxConfig{}),
			).Name(dialogContentPortalName(mediaLibraryListField)),
		)
		return
	})
}
