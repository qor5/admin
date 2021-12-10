package pages

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/listeditor"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/publish"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func ListEditorExample(db *gorm.DB) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.PageTitle = "List Editor Example"
		listValue := []*models.Post{
			{Title: "Post 1", Status: publish.Status{Status: "PendingReview"}},
			{Title: "Post 2", Status: publish.Status{Status: "Approved"}},
			{Title: "Post 3"},
		}

		le := listeditor.New().Value(listValue)
		le.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VTextField().FieldName(field.Name).Value(field.Value(obj)).Label(field.Label)
		})

		le.Field("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSelect().Items([]string{"Draft", "PendingReview", "Approved"}).Value(field.Value(obj).(publish.Status).Status).Label(field.Label)
		})

		le.Field("HeroImage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return media_view.QMediaBox(db).
				FieldName(field.Name).
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
				})
		})

		r.Body = VContainer(
			le,
		)
		return
	}
}
