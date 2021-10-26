package views

import (
	"github.com/goplaid/web"
	h "github.com/theplant/htmlgo"
)

func sidePanel(ctx *web.EventContext) h.HTMLComponent {
	return h.Text("versions")
}
