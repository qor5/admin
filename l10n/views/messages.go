package views

import (
	"net/http"

	"github.com/goplaid/x/i18n"
)

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

const I10nLocalizeKey i18n.ModuleKey = "I10nLocalizeKey"

func MustGetMessages(r *http.Request) *Messages {
	return i18n.MustGetModuleMessages(r, I10nLocalizeKey, Messages_en_US).(*Messages)
}
