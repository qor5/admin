package examples_vuetifyx

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	. "github.com/theplant/htmlgo"
)

func ExpansionPanelDemo(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = VApp(
		VMain(
			VExpansionPanels(
				VExpansionPanel(
					VExpansionPanelTitle(
						Text("VISA •••• 4242	11 / 2028"),
						web.Slot(
							VIcon("search"),
						).Name("actions"),
					),
					VExpansionPanelText(
						VDivider(),
						vx.DetailInfo(
							vx.DetailColumn(
								vx.DetailField(vx.OptionalText("FENGMIN SUN").ZeroLabel("No Name")).Label("Name"),
								vx.DetailField(vx.OptionalText("•••• 4242").ZeroLabel("No Number")).Label("Number"),
								vx.DetailField(vx.OptionalText("QlfGjXhL3I1xfKVV").ZeroLabel("No Fingerprint")).Label("Fingerprint"),
								vx.DetailField(vx.OptionalText("11 / 2028").ZeroLabel("No Expires")).Label("Expires"),
								vx.DetailField(vx.OptionalText("Visa credit card").ZeroLabel("No Type")).Label("Type"),
								vx.DetailField(vx.OptionalText("card_1EJtLGAqkzzGorqLeFb6h2YV").ZeroLabel("No Type")).Label("ID"),
							),
						).Class("pa-0"),
					),
				),

				VExpansionPanel(
					VExpansionPanelTitle(
						Text("VISA •••• 2121	11 / 2028"),
					),
					VExpansionPanelText(
						VDivider(),
						vx.DetailInfo(
							vx.DetailColumn(
								vx.DetailField(vx.OptionalText("FENGMIN SUN").ZeroLabel("No Name")).Label("Name"),
								vx.DetailField(vx.OptionalText("•••• 4242").ZeroLabel("No Number")).Label("Number"),
								vx.DetailField(vx.OptionalText("QlfGjXhL3I1xfKVV").ZeroLabel("No Fingerprint")).Label("Fingerprint"),
								vx.DetailField(vx.OptionalText("11 / 2028").ZeroLabel("No Expires")).Label("Expires"),
								vx.DetailField(vx.OptionalText("Visa credit card").ZeroLabel("No Type")).Label("Type"),
								vx.DetailField(vx.OptionalText("card_1EJtLGAqkzzGorqLeFb6h2YV").ZeroLabel("No Type")).Label("ID"),
							),
						).Class("pa-0"),
					),
				),
			),
		),
	)
	return
}

var ExpansionPanelDemoPB = web.Page(ExpansionPanelDemo)

var ExpansionPanelDemoPath = examples.URLPathByFunc(ExpansionPanelDemo)
