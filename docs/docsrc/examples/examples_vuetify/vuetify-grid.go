package examples_vuetify

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func HelloVuetifyGrid(ctx *web.EventContext) (pr web.PageResponse, err error) {
	row := func(col int, count int, color string) (r h.HTMLComponent) {
		rw := VRow()
		for i := 0; i < count; i++ {
			rw.AppendChildren(
				VCol(
					VCard(
						VCardText(h.Text(fmt.Sprint(col))),
					).Color(color),
				),
			)
		}

		return rw
	}

	var lc []h.HTMLComponent
	lc = append(lc, row(12, 1, "primary"))
	lc = append(lc, row(6, 2, "secondary"))
	lc = append(lc, row(4, 3, "primary"))
	lc = append(lc, row(3, 4, "secondary"))
	lc = append(lc, row(2, 6, "primary"))
	lc = append(lc, row(1, 12, "secondary"))

	pr.Body = VApp(
		VMain(
			VContainer(
				lc...,
			).GridList(Md).TextAlign(Xs, Center),
		),
	)
	return
}

var VuetifyGridPB = web.Page(HelloVuetifyGrid)

var VuetifyGridPath = examples.URLPathByFunc(HelloVuetifyGrid)
