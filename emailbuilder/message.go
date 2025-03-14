package emailbuilder

import (
	"github.com/qor5/x/v3/i18n"
)

const I18nEmailBuilderKey i18n.ModuleKey = "I18nEmailBuilderKey"

type Messages struct {
	ChangeTemplate               string
	ModelLabelTemplate           string
	Blank                        string
	AddBlankPage                 string
	BlankPage                    string
	SendEmail                    string
	EnterEmailAddressPlaceholder string
}

var Messages_en_US = &Messages{
	Blank:                        "Blank",
	ModelLabelTemplate:           "Template",
	ChangeTemplate:               "ChangeTemplate",
	AddBlankPage:                 "Add Blank Page",
	BlankPage:                    "Blank Page",
	SendEmail:                    "Send Email",
	EnterEmailAddressPlaceholder: "Please Enter a Email Address",
}

var Messages_zh_CN = &Messages{
	ChangeTemplate:               "更改模版",
	Blank:                        "空白",
	ModelLabelTemplate:           "模板页面",
	AddBlankPage:                 "新增空白页",
	BlankPage:                    "空白页",
	SendEmail:                    "发送邮件",
	EnterEmailAddressPlaceholder: "请输入邮箱地址",
}

var Messages_ja_JP = &Messages{
	ChangeTemplate:               "テンプレートの変更",
	Blank:                        "空白",
	ModelLabelTemplate:           "テンプレート",
	AddBlankPage:                 "空白ページを追加",
	BlankPage:                    "空白ページ",
	SendEmail:                    "メール送信",
	EnterEmailAddressPlaceholder: "メールアドレスを入力してください",
}
