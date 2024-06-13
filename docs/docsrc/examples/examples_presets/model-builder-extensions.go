package examples_presets

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// @snippet_begin(PresetsModelBuilderExtensionsSample)

func PresetsModelBuilderExtensions(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	mb, cl, ce, dp = PresetsHelloWorld(b, db)
	mb.LayoutConfig(&presets.LayoutConfig{})

	eb := mb.Editing("Actions", "Name").ActionsFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
		return h.Components(
			VSpacer(),
			VBtn("Action 1"),
			VBtn("Action 2"),
			VBtn("Update").
				Color("primary").
				Attr("@click", web.POST().
					EventFunc(actions.Update).
					Queries(ctx.Queries()).
					URL(mb.Info().ListingHref()).
					Go()),
		)
	})

	eb.Field("Actions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cust := obj.(*Customer)
		return VBtn("Change Name").Attr("@click",
			web.POST().
				EventFunc("changeName").
				Query(presets.ParamID, fmt.Sprint(cust.ID)).
				Go(),
		)
	})

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		cust := obj.(*Customer)
		if len(cust.Name) < 5 {
			err.GlobalError("Name must be longer than 5")
		}
		return
	})

	mb.RegisterEventFunc("changeName", changeNameEventFunc(mb))

	return
}

func changeNameEventFunc(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		eb := mb.Editing()
		obj := mb.NewModel()
		id := ctx.Param(presets.ParamID)
		obj, err = eb.Fetcher(obj, id, ctx)
		obj.(*Customer).Name = "Darwin"
		err = eb.Saver(obj, id, ctx)
		presets.ShowMessage(&r, "Nicely updated", "")
		eb.UpdateOverlayContent(ctx, &r, obj, "Good work", err)
		return
	}
}

// @snippet_end
