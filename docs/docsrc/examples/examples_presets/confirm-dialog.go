package examples_presets

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	"github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type confirmDialog struct{}

func PresetsConfirmDialog(b *presets.Builder, db *gorm.DB) (
	mb *presets.ModelBuilder,
	cl *presets.ListingBuilder,
	ce *presets.EditingBuilder,
	dp *presets.DetailingBuilder,
) {
	_ = []interface{}{
		// @snippet_begin(OpenConfirmDialog)
		presets.OpenConfirmDialog,
		// @snippet_end
		// @snippet_begin(ConfirmDialogConfirmEvent)
		presets.ConfirmDialogConfirmEvent,
		// @snippet_end
		// @snippet_begin(ConfirmDialogPromptText)
		presets.ConfirmDialogPromptText,
		// @snippet_end
		// @snippet_begin(ConfirmDialogDialogPortalName)
		presets.ConfirmDialogDialogPortalName,
		// @snippet_end
	}

	b.DataOperator(gorm2op.DataOperator(db))

	mb = b.Model(&confirmDialog{}).
		URIName("confirm-dialog").
		Label("Confirm Dialog")

	mb.Listing().PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		r.Body = htmlgo.Div(
			// @snippet_begin(ConfirmDialogSample)
			vuetify.VBtn("Delete File").
				Attr("@click",
					web.Plaid().
						EventFunc(presets.OpenConfirmDialog).
						Query(presets.ConfirmDialogPromptText, `Are you sure you want to delete this file?`).
						Query(presets.ConfirmDialogConfirmEvent,
							`alert("file deleted")`,
						).
						Go(),
				),
			// @snippet_end
		).Class("ma-8")
		return r, nil
	})
	return
}
