package utils

import (
	"github.com/goplaid/web"
	. "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
)

func ConfirmDialog(msg string, okAction string) h.HTMLComponent {
	return VDialog(
		VCard(
			VCardTitle(h.Text(msg)),
			VCardActions(
				VSpacer(),
				VBtn("Cancel").
					Depressed(true).
					Class("ml-2").
					On("click", "locals.commonConfirmDialog = false"),

				VBtn("OK").
					Color("primary").
					Depressed(true).
					Dark(true).
					Attr("@click", okAction),
			),
		),
	).MaxWidth("600px").
		Attr(web.InitContextLocals, `{commonConfirmDialog: false}`).Attr("v-model", "locals.commonConfirmDialog")

}
