package pages

import (
	"fmt"

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
	"github.com/theplant/testingutils"
	"gorm.io/gorm"
)

func ListEditorExample(db *gorm.DB) (pf web.PageFunc, sf web.EventFunc) {
	var le = presets.NewFieldBuilders()

	var phoneLe = presets.NewFieldBuilders()
	phoneLe.Field(".").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VTextField().
			FieldName(field.KeyPath).
			Value(field.StringValue(obj)).
			Label(field.Label)
	})

	le.Field("ID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Input("").
			Type("hidden").
			Value(fmt.Sprint(field.Value(obj))).
			Attr(web.VFieldName(field.KeyPath)...)
	})

	le.Field("Street").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VTextField().
			FieldName(field.KeyPath).
			Value(field.Value(obj)).
			Label(field.Label)
	})

	le.Field("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return VSelect().
			Items([]string{"Draft", "PendingReview", "Approved"}).
			Value(field.Value(obj).(publish.Status).Status).
			Label(field.Label).
			FieldName(field.KeyPath)
	})

	le.Field("Phones").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		phoneLe := listeditor.New(phoneLe).Value(field.Value(obj)).FieldContext(field)
		return phoneLe
	})

	le.Field("HeroImage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return media_view.QMediaBox(db).
			FieldName(field.KeyPath).
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

	pf = func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.PageTitle = "List Editor Example"
		var listValue = []*models.Address{
			{
				Model:  gorm.Model{ID: 1},
				Street: "Street 1",
				Status: publish.Status{
					Status: "Draft",
				},
				Phones: []string{"123456789", "987654321"},
			},
			{
				Model:  gorm.Model{ID: 2},
				Street: "Street 2",
				Status: publish.Status{
					Status: "PendingReview",
				},
				Phones: []string{"123456789", "987654321"},
			},
			{
				Model:  gorm.Model{ID: 3},
				Street: "Street 3",
				Status: publish.Status{
					Status: "Approved",
				},
				Phones: []string{"123456789", "987654321"},
			},
		}
		// err = db.Find(&listValue).Error
		// if err != nil {
		// 	panic(err)
		// }

		le := listeditor.New(le).Value(listValue)

		r.Body = VContainer(
			le,
			VBtn("Save").Attr("@click", web.Plaid().EventFunc("save").Go()),
		)
		return
	}

	sf = func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var listValue []*models.Address
		ctx.MustUnmarshalForm(&listValue)
		// fmt.Printf("%#+v\n", listValue)
		testingutils.PrintlnJson(listValue)
		r.Reload = true
		return
	}
	return
}
