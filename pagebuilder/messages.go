package pagebuilder

import "github.com/qor5/x/i18n"

const I18nPageBuilderKey i18n.ModuleKey = "I18nPageBuilderKey"

type Messages struct {
	Category                       string
	EditPageContent                string
	Preview                        string
	Containers                     string
	AddContainers                  string
	New                            string
	Shared                         string
	Select                         string
	TemplateID                     string
	TemplateName                   string
	CreateFromTemplate             string
	RelatedOnlinePages             string
	RepublishAllRelatedOnlinePages string
	Unnamed                        string
	NotDescribed                   string
	Blank                          string
	NewPage                        string
}

var Messages_en_US = &Messages{
	Category:                       "Category",
	EditPageContent:                "Edit Page Content",
	Preview:                        "Preview",
	Containers:                     "Containers",
	AddContainers:                  "Add Containers",
	New:                            "New",
	Shared:                         "Shared",
	Select:                         "Select",
	TemplateID:                     "Template ID",
	TemplateName:                   "Template Name",
	CreateFromTemplate:             "Create From Template",
	RelatedOnlinePages:             "Related Online Pages",
	RepublishAllRelatedOnlinePages: "Republish All",
	Unnamed:                        "Unnamed",
	NotDescribed:                   "Not Described",
	Blank:                          "Blank",
	NewPage:                        "New Page",
}

var Messages_zh_CN = &Messages{
	Category:                       "目录",
	EditPageContent:                "编辑页面内容",
	Preview:                        "预览",
	Containers:                     "组件",
	AddContainers:                  "增加组件",
	New:                            "新增",
	Shared:                         "公用的",
	Select:                         "选择",
	TemplateID:                     "模板ID",
	TemplateName:                   "模板名",
	CreateFromTemplate:             "从模板中创建",
	RelatedOnlinePages:             "相关在线页面",
	RepublishAllRelatedOnlinePages: "重新发布所有页面",
	Unnamed:                        "未命名",
	NotDescribed:                   "未描述",
	Blank:                          "空白",
	NewPage:                        "新页面",
}

var Messages_ja_JP = &Messages{
	Category:                       "カテゴリー",
	EditPageContent:                "ページコンテナを編集する",
	Preview:                        "プレビュー",
	Containers:                     "コンテナ",
	AddContainers:                  "コンテナを追加する",
	New:                            "新規",
	Shared:                         "共有",
	Select:                         "選択する",
	TemplateID:                     "テンプレートID",
	TemplateName:                   "テンプレート名",
	CreateFromTemplate:             "テンプレートから新規作成する",
	RelatedOnlinePages:             "関連オンラインページ",
	RepublishAllRelatedOnlinePages: "すべて再公開",
	Unnamed:                        "名前なし",
	NotDescribed:                   "記述されていません",
	Blank:                          "空白",
	NewPage:                        "新しいページ",
}
