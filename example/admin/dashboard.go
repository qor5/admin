package admin

import (
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func Dashboard() h.HTMLComponent {
	return vuetify.VContainer(
		h.H1("Welcome to the QOR5 demo site").Class("mt-8"),

		h.A().Text("QOR5 Website").Href("https://qor5.com").Target("_blank"),
		h.A().Text("QOR5 Documentation").Href("https://docs.qor5.com").Target("_blank").Class("ml-4"),
		h.A().Text("Source Code").Href("https://github.com/qor5/admin/tree/main/example").Target("_blank").Class("ml-4"),
	)
}
