package examples_vuetify

// @snippet_begin(VuetifySnackBarsSample)
import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func VuetifySnackBars(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = VContainer(
		VBtn("Show Snack Bar").OnClick("showSnackBar"),
		web.Portal().Name("snackbar"),
		snackbar("bottom", "success"),
	)

	return
}

func showSnackBar(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.UpdatePortals = append(er.UpdatePortals,
		&web.PortalUpdate{
			Name: "snackbar",
			Body: snackbar("top", "red"),
		},
	)

	return
}

func snackbar(pos string, color string) *web.ScopeBuilder {
	return web.Scope(
		VSnackbar().Location(pos).Timeout(-1).Color(color).
			Attr("v-model", "locals.show").
			Children(
				h.Text("Hello, I am a snackbar"),
				web.Slot(
					VBtn("").Variant("text").
						Attr("@click", "locals.show = false").
						Children(VIcon("mdi-close")),
				).Name("actions"),
			),
	).VSlot("{ locals }").Init(`{ show: true }`)
}

var VuetifySnackBarsPB = web.Page(VuetifySnackBars).
	EventFunc("showSnackBar", showSnackBar)

var VuetifySnackBarsPath = examples.URLPathByFunc(VuetifySnackBars)

// @snippet_end
