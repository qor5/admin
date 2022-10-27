package pagebuilder

import "github.com/goplaid/x/i18n"

const I18nPageBuilderKey i18n.ModuleKey = "I18nPageBuilderKey"

type Messages struct {
	Category        string
	EditPageContent string
	Preview         string
	Containers      string
	AddContainers   string
	New             string
	Shared          string
	Select          string
}

var Messages_en_US = &Messages{
	Category:        "Category",
	EditPageContent: "Edit Page Content",
	Preview:         "Preview",
	Containers:      "Containers",
	AddContainers:   "Add Containers",
	New:             "New",
	Shared:          "Shared",
	Select:          "Select",
}

var Messages_zh_CN = &Messages{
	Category:        "类别",
	EditPageContent: "编辑页面内容",
	Preview:         "预览",
	Containers:      "组件",
	AddContainers:   "增加组件",
	New:             "新增",
	Shared:          "公用的",
	Select:          "选择",
}
