package pagebuilder

import "github.com/qor5/x/i18n"

const I18nPageBuilderKey i18n.ModuleKey = "I18nPageBuilderKey"

type Messages struct {
	Category           string
	EditPageContent    string
	Preview            string
	Containers         string
	AddContainers      string
	New                string
	Shared             string
	Select             string
	TemplateID         string
	TemplateName       string
	CreateFromTemplate string
}

var Messages_en_US = &Messages{
	Category:           "Category",
	EditPageContent:    "Edit Page Content",
	Preview:            "Preview",
	Containers:         "Containers",
	AddContainers:      "Add Containers",
	New:                "New",
	Shared:             "Shared",
	Select:             "Select",
	TemplateID:         "Template ID",
	TemplateName:       "Template Name",
	CreateFromTemplate: "Create From Template",
}

var Messages_zh_CN = &Messages{
	Category:           "目录",
	EditPageContent:    "编辑页面内容",
	Preview:            "预览",
	Containers:         "组件",
	AddContainers:      "增加组件",
	New:                "新增",
	Shared:             "公用的",
	Select:             "选择",
	TemplateID:         "模板ID",
	TemplateName:       "模板名",
	CreateFromTemplate: "从模板中创建",
}
