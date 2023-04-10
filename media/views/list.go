package views

import (
	"github.com/qor5/admin/media/media_library"
	"github.com/qor5/admin/presets"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	mediaLibraryListField = "media-library-list"
)

func configList(b *presets.Builder, db *gorm.DB) {
	mm := b.Model(&media_library.MediaLibrary{}).Label("Media Library").MenuIcon("image").URIName("media-library")

	mm.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.PageTitle = "Media Library"
		keyword := ctx.R.FormValue("keyword")
		ctx.R.Form.Set(searchKeywordName(mediaLibraryListField), keyword)
		r.Body = h.Components(
			web.Portal().Name(deleteConfirmPortalName(mediaLibraryListField)),
			web.Portal(
				h.Input("").
					Type("hidden").
					Value(keyword).
					Attr(web.VFieldName(searchKeywordName(mediaLibraryListField))...),
				fileChooserDialogContent(db, mediaLibraryListField, ctx, &media_library.MediaBoxConfig{}),
			).Name(dialogContentPortalName(mediaLibraryListField)),
		)
		return
	})
}
