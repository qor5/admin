package l10n

import (
	"net/http"

	"github.com/qor5/x/v3/i18n"
)

const I18nLocalizeKey i18n.ModuleKey = "I18nLocalizeKey"

type Messages struct {
	Localize              string
	LocalizeFrom          string
	LocalizeTo            string
	SuccessfullyLocalized string
	Location              string
	Colon                 string
	International         string
	China                 string
	Japan                 string
}

var Messages_en_US = &Messages{
	Localize:              "Localize",
	LocalizeFrom:          "From",
	LocalizeTo:            "To",
	SuccessfullyLocalized: "Successfully Localized",
	Location:              "Location",
	Colon:                 ":",
	International:         "International",
	China:                 "China",
	Japan:                 "Japan",
}

var Messages_zh_CN = &Messages{
	Localize:              "本地化",
	LocalizeFrom:          "从",
	LocalizeTo:            "到",
	SuccessfullyLocalized: "本地化成功",
	Location:              "地区",
	Colon:                 "：",
	International:         "全球",
	China:                 "中国",
	Japan:                 "日本",
}

var Messages_ja_JP = &Messages{
	Localize:              "ローカライズ",
	LocalizeFrom:          "から",
	LocalizeTo:            "に",
	SuccessfullyLocalized: "ローカライズに成功しました",
	Location:              "場所",
	Colon:                 ":",
	International:         "国際的",
	China:                 "中国",
	Japan:                 "日本",
}

func MustGetTranslation(r *http.Request, key string) string {
	return i18n.T(r, I18nLocalizeKey, key)
}
