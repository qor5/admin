package utils

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const I18nUtilsKey i18n.ModuleKey = "I18nUtilsKey"

type UtilDialogPayloadType struct {
	Title        string
	TypeField    vx.VXDialogType
	Size         vx.VXDialogSize
	Text         string
	ContentEl    h.HTMLComponent
	OkAction     string
	CancelAction string
	Width        int
	HideClose    bool
	Msgr         *Messages
}

func Install(b *presets.Builder) {
	b.GetI18n().
		RegisterForModule(language.English, I18nUtilsKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nUtilsKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nUtilsKey, Messages_ja_JP)
}

func MustGetMessages(r *http.Request) *Messages {
	return i18n.MustGetModuleMessages(r, I18nUtilsKey, Messages_en_US).(*Messages)
}

func ConfirmDialog(payload UtilDialogPayloadType) h.HTMLComponent {
	if payload.Title == "" {
		payload.Title = payload.Msgr.ModalTitleConfirm
	}

	return vx.VXDialog().
		Title(payload.Title).
		Text(payload.Text).
		HideClose(true).
		OkText(payload.Msgr.OK).
		CancelText(payload.Msgr.Cancel).
		Attr("@click:ok", payload.OkAction).
		Attr("v-model", "locals.commonConfirmDialog")
}

func DeleteDialog(msg, okAction string, msgr *Messages) h.HTMLComponent {
	return web.Scope(
		vx.VXDialog().
			Title(msgr.ModalTitleConfirm).
			Text(msg).
			HideClose(true).
			OkText(msgr.OK).
			CancelText(msgr.Cancel).
			Attr("@click:ok", okAction).
			Attr("v-model", "locals.deleteConfirmation"),
	).VSlot(" { locals }").Init(`{deleteConfirmation: true}`)
}

func CustomDialog(payload UtilDialogPayloadType) h.HTMLComponent {
	if payload.Size == "" {
		payload.Size = vx.DialogSizeLarge
	}

	return web.Scope(
		vx.VXDialog(
			payload.ContentEl,
		).Size(payload.Size).
			Type(payload.TypeField).
			Title(payload.Title).
			Width(payload.Width).
			OkText(payload.Msgr.OK).
			CancelText(payload.Msgr.Cancel).
			Attr("v-model", "locals.customConfirmationDialog").
			Attr("@click:ok", payload.OkAction),
	).VSlot(" { locals }").Init(`{ customConfirmationDialog: true }`)
}
