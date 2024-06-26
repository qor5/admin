package examples_vuetify

// @snippet_begin(VuetifyNavigationDrawerSample)
import (
	"fmt"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func VuetifyNavigationDrawer(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = VContainer(
		h.H2("A drawer that has close button"),

		VBtn("show").On("click", "vars.drawer1 = !vars.drawer1"),

		VNavigationDrawer(
			h.Text("Hi"),
			VBtn("Close").On("click", "vars.drawer1 = false"),
		).Temporary(true).
			Attr("v-model", "vars.drawer1").
			Location("right").
			Absolute(true).
			Width(600),

		h.H2("Load a drawer from remote and show it").Class("pt-8"),

		VBtn("Show Drawer 2").OnClick("showDrawer"),

		web.Portal().Name("drawer2UpdateContent"),

		web.Portal().Name("drawer2"),
	).Attr(":value", "vars.test1 = 1")

	return
}

func showDrawer(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.UpdatePortals = append(er.UpdatePortals,
		&web.PortalUpdate{
			Name: "drawer2",
			Body: VNavigationDrawer(
				web.Scope(
					h.Text("Drawer 2"),
					web.Portal(
						textField(""),
					).Name("InputPortal"),
					VBtn("Update parent and close").
						OnClick("updateParentAndClose"),
				).VSlot("{ locals, form }"),
			).Location("right").
				Attr("v-model", "vars.drawer2").
				Temporary(true).
				Absolute(true).
				Width(800),
		},
	)

	er.RunScript = `setTimeout(function(){ vars.drawer2 = true }, 100)`
	return
}

func textField(value string, fieldErrors ...string) h.HTMLComponent {
	return VTextField().
		Attr("v-model", "form.Drawer2Input").
		ErrorMessages(fieldErrors...)
}

func updateParentAndClose(ctx *web.EventContext) (er web.EventResponse, err error) {
	if len(ctx.R.FormValue("Drawer2Input")) < 10 {
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "InputPortal",
			Body: textField(ctx.R.FormValue("Drawer2Input"), "input more then 10 characters"),
		})
		return
	}

	er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
		Name: "drawer2UpdateContent",
		Body: h.Text(fmt.Sprintf("Updated content at %s", time.Now())),
	})

	er.RunScript = "vars.drawer2 = false;"
	return
}

var VuetifyNavigationDrawerPB = web.Page(VuetifyNavigationDrawer).
	EventFunc("showDrawer", showDrawer).
	EventFunc("updateParentAndClose", updateParentAndClose)

var VuetifyNavigationDrawerPath = examples.URLPathByFunc(VuetifyNavigationDrawer)

// @snippet_end
