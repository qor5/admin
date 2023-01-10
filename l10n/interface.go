package l10n

import (
	"reflect"

	"github.com/qor5/admin/utils"
	"github.com/qor5/web"
)

type L10nInterface interface {
	SetLocale(locale string)
}

func IsLocalizable(obj interface{}) (isLocalizable bool) {
	_, isLocalizable = utils.GetStruct(reflect.TypeOf(obj)).(L10nInterface)
	return
}

func IsLocalizableFromCtx(ctx *web.EventContext) (localeCode string, isLocalizable bool) {
	locale := ctx.R.Context().Value(LocaleCode)
	if locale != nil {
		localeCode = locale.(string)
		isLocalizable = true
	}
	return
}

type L10nONInterface interface {
	L10nON()
}
