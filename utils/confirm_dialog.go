package utils

import (
	"fmt"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const I18nUtilsKey i18n.ModuleKey = "I18nUtilsKey"

func Configure(b *presets.Builder) {
	b.I18n().
		RegisterForModule(language.English, I18nUtilsKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nUtilsKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nUtilsKey, Messages_ja_JP)
}

func ConfirmDialog(msg string, okAction string, msgr *Messages) h.HTMLComponent {
	return VDialog(
		VCard(
			VCardTitle(h.Text(msg)),
			VCardActions(
				VSpacer(),
				VBtn(msgr.Cancel).
					Variant(VariantFlat).
					Class("ml-2").
					On("click", "locals.commonConfirmDialog = false"),

				VBtn(msgr.OK).
					Color("primary").
					Variant(VariantFlat).
					Theme(ThemeDark).
					Attr("@click", okAction),
			),
		),
	).MaxWidth("600px").
		Attr("v-model", "locals.commonConfirmDialog")
}

func DeleteDialog(msg string, okAction string, msgr *Messages) h.HTMLComponent {
	return VDialog(
		VCard(
			VCardTitle(h.Text(msg)),
			VCardActions(
				VSpacer(),
				VBtn(msgr.Cancel).
					Variant(VariantFlat).
					Class("ml-2").
					On("click", "locals.deleteConfirmation = false"),

				VBtn(msgr.OK).
					Color("primary").
					Variant(VariantFlat).
					Theme(ThemeDark).
					Attr("@click", okAction),
			),
		),
	).MaxWidth("600px").
		Attr("v-model", "locals.deleteConfirmation")
}

func CustomDialog(msg string, content h.HTMLComponent, okAction, VModelKey string, msgr *Messages) h.HTMLComponent {
	Vcard := VCard()
	if msg != "" {
		Vcard.AppendChildren(VCardTitle(h.Text(msg)))
	}
	if content != nil {
		Vcard.AppendChildren(VCardText(content))
	}
	Vcard.AppendChildren(
		VCardActions(
			VSpacer(),
			VBtn(msgr.Cancel).
				Variant(VariantFlat).
				Class("ml-2").
				On("click", fmt.Sprintf("locals.%s = false", VModelKey)),

			VBtn(msgr.OK).
				Color("primary").
				Variant(VariantFlat).
				Theme(ThemeDark).
				Attr("@click", okAction),
		),
	)
	return VDialog(
		Vcard,
	).MaxWidth("600px").
		Attr("v-model", fmt.Sprintf("locals.%s", VModelKey))
}
