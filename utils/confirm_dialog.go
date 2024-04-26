package utils

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
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

func MustGetMessages(r *http.Request) *Messages {
	return i18n.MustGetModuleMessages(r, I18nUtilsKey, Messages_en_US).(*Messages)
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
	return web.Scope(
		VDialog(
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
			Attr("v-model", "locals.deleteConfirmation"),
	).VSlot(" { locals }").Init(`{deleteConfirmation: true}`)
}

const CloseCustomDialog = "locals.customConfirmationDialog = false"

func CustomDialog(title h.HTMLComponent, content h.HTMLComponent, okAction string, msgr *Messages) h.HTMLComponent {
	Vcard := VCard()
	if title != nil {
		Vcard.AppendChildren(VCardTitle(title))
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
				On("click", CloseCustomDialog),

			VBtn(msgr.OK).
				Color("primary").
				Variant(VariantFlat).
				Theme(ThemeDark).
				Attr("@click", okAction),
		),
	)
	return web.Scope(
		VDialog(
			Vcard,
		).MaxWidth("600px").
			Attr("v-model", "locals.customConfirmationDialog"),
	).VSlot(" { locals }").Init(`{ customConfirmationDialog: true }`)
}
