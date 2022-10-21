package l10n

import (
	"reflect"

	"github.com/qor/qor5/utils"
)

type L10nInterface interface {
	SetLocale(locale string)
}

func IsLocalizable(obj interface{}) (IsLocalizable bool) {
	_, IsLocalizable = utils.GetStruct(reflect.TypeOf(obj)).(L10nInterface)
	return
}
