package l10n

import (
	"context"
	"reflect"

	"github.com/qor5/admin/v3/utils"
)

type L10nInterface interface {
	SetLocale(locale string)
	GetLocale() string
}

func IsLocalizable(obj interface{}) (isLocalizable bool) {
	_, isLocalizable = utils.GetStruct(reflect.TypeOf(obj)).(L10nInterface)
	return
}

func IsLocalizableFromCtx(ctx context.Context) (localeCode string, isLocalizable bool) {
	locale := ctx.Value(LocaleCode)
	if locale != nil {
		localeCode = locale.(string)
		isLocalizable = true
	}
	return
}
