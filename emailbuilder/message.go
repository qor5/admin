package emailbuilder

import (
	"github.com/qor5/x/v3/i18n"
)

const I18nEmailBuilderKey i18n.ModuleKey = "I18nEmailBuilderKey"

type Messages struct {
	ChangeTemplate     string
	ModelLabelTemplate string
	Blank              string
}

var Messages_en_US = &Messages{
	Blank:              "Blank",
	ModelLabelTemplate: "Template",
	ChangeTemplate:     "ChangeTemplate",
}

var Messages_zh_CN = &Messages{
	ChangeTemplate:     "更改模版",
	Blank:              "空白",
	ModelLabelTemplate: "模板页面",
}

var Messages_ja_JP = &Messages{
	ChangeTemplate: "テンプレートの変更",

	Blank:              "空白",
	ModelLabelTemplate: "テンプレート",
}
