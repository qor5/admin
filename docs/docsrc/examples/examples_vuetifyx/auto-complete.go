package examples_vuetifyx

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

type Category struct {
	Path string
	ID   int
}

func AutoCompleteDemo(ctx *web.EventContext) (pr web.PageResponse, err error) {
	categories := []Category{{ID: 1, Path: "/"}, {ID: 2, Path: "/123"}}
	fd := vuetifyx.VXAutocomplete().Label("Category").
		Attr(web.VField("Category", 2)...).
		Multiple(false).Chips(false).
		Items(categories).ItemTitle("Path").ItemValue("ID")

	pr.Body = VApp(
		VMain(
			fd,
		),
	)
	return
}

var AutoCompleteDemoPB = web.Page(AutoCompleteDemo)

var AutoCompleteDemoPath = examples.URLPathByFunc(AutoCompleteDemo)
