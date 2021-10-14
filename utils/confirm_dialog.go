package utils

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const I18nUtilsKey i18n.ModuleKey = "I18nUtilsKey"

func Configure(b *presets.Builder) {
	b.I18n().
		RegisterForModule(language.English, I18nUtilsKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nUtilsKey, Messages_zh_CN)
}

func ConfirmDialog(msg string, okAction string, msgr *Messages) h.HTMLComponent {
	return VDialog(
		VCard(
			VCardTitle(h.Text(msg)),
			VCardActions(
				VSpacer(),
				VBtn(msgr.Cancel).
					Depressed(true).
					Class("ml-2").
					On("click", "locals.commonConfirmDialog = false"),

				VBtn(msgr.OK).
					Color("primary").
					Depressed(true).
					Dark(true).
					Attr("@click", okAction),
			),
		),
	).MaxWidth("600px").
		Attr(web.InitContextLocals, `{commonConfirmDialog: false}`).Attr("v-model", "locals.commonConfirmDialog")

}
