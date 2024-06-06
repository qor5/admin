package examples_web

// @snippet_begin(ShortCutSample)
import (
	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func ShortCutSample(ctx *web.EventContext) (pr web.PageResponse, err error) {
	clickEvent := "locals.count += 1"
	pr.Body = VContainer(
		web.Scope(
			VRow(
				VCol(
					VRow(
						VBtn("count+1").Attr("@click", clickEvent).Class("mr-4"),
						h.Text("Shortcut: enter"),
					).Class("mb-10"),
					VRow(
						VBtn("toggle shortcut").Attr("@click", "locals.shortCutEnabled = !locals.shortCutEnabled"),
					),
				),
				VCol(
					VCard(
						VCardTitle(h.Text("Shortcut Enabled")),
						VCardText().Attr("v-text", "locals.shortCutEnabled"),
					).Class("mb-10"),

					VCard(
						VCardTitle(h.Text("Count")),
						VCardText().Attr("v-text", "locals.count"),
					),
				),
			).Class("mt-10"),
			web.GlobalEvents().Attr(":filter", `(event, handler, eventName) => locals.shortCutEnabled == true`).Attr("@keydown.enter", clickEvent),
		).Init(`{ shortCutEnabled: true, count: 0 }`).
			VSlot("{ locals, form }"),
	)
	return
}

var ShortCutSamplePB = web.Page(ShortCutSample)

var ShortCutSamplePath = examples.URLPathByFunc(ShortCutSample)

// @snippet_end
