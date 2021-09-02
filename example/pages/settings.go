package pages

import (
	"log"

	"github.com/goplaid/web"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/qor/qor5/cropper"
	"github.com/qor/qor5/media_library_view"
	"github.com/qor/qor5/richeditor"
	h "github.com/theplant/htmlgo"
)

func Settings(db *gorm.DB) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		ctx.Hub.RegisterEventFunc("logInfo", logInfo)
		r.PageTitle = "Settings"
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
				VRow(
					VCol(richeditor.Redactor().Value("text1").Placeholder("text").Attr(web.VFieldName("Body")...)),
				),

				VRow(
					cropper.Cropper().
						Src("https://agontuk.github.io/assets/images/berserk.jpg").
						Value(cropper.Value{X: 1141, Y: 540, Width: 713, Height: 466}).
						AspectRatio(713, 466).
						Attr("@input", web.Plaid().
							FieldValue("CropperEvent", web.Var("JSON.stringify($event)")).EventFunc("logInfo").Go()),
				),
			).Fluid(true),
		)
		return
	}
}

func logInfo(ctx *web.EventContext) (r web.EventResponse, err error) {
	log.Println("CropperEvent", ctx.R.FormValue("CropperEvent"))
	return
}
