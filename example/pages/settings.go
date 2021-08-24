package pages

import (
	"github.com/goplaid/web"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/qor5/media_library_view"
	h "github.com/theplant/htmlgo"
)

func Settings(db *gorm.DB) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = h.Div(
			h.H1("Example of use QMediaBox in any page").Class("text-h5 pt-4 pl-2"),
			VContainer(
				VRow(
					VCol(
						media_library_view.QMediaBox(db).
							FieldName("test").
							Value(&media_library.MediaBox{}).
							Config(&media_library.MediaBoxConfig{
								AllowType: "image",
								Sizes: map[string]*media.Size{
									"thumb": {
										Width:  400,
										Height: 300,
									},
									"main": {
										Width:  800,
										Height: 500,
									},
								},
							}),
					).Cols(6),
				),
			).Fluid(true),
		)
		return
	}
}
