package pagebuilder

import "github.com/qor5/x/i18n"

const I18nPageBuilderKey i18n.ModuleKey = "I18nPageBuilderKey"

type Messages struct {
	Category                       string
	Preview                        string
	Containers                     string
	AddContainers                  string
	New                            string
	Shared                         string
	Select                         string
	SelectedTemplateLabel          string
	CreateFromTemplate             string
	ChangeTemplate                 string
	RelatedOnlinePages             string
	RepublishAllRelatedOnlinePages string
	Unnamed                        string
	NotDescribed                   string
	Blank                          string
	NewPage                        string
	Duplicate                      string
	FilterTabAllVersions           string
	FilterTabOnlineVersion         string
	FilterTabNamedVersions         string
}

var Messages_en_US = &Messages{
	Category:                       "Category",
	Preview:                        "Preview",
	Containers:                     "Containers",
	AddContainers:                  "Add Containers",
	New:                            "New",
	Shared:                         "Shared",
	Select:                         "Select",
	SelectedTemplateLabel:          "Template",
	CreateFromTemplate:             "Create From Template",
	ChangeTemplate:                 "Change Template",
	RelatedOnlinePages:             "Related Online Pages",
	RepublishAllRelatedOnlinePages: "Republish All",
	Unnamed:                        "Unnamed",
	NotDescribed:                   "Not Described",
	Blank:                          "Blank",
	NewPage:                        "New Page",
	Duplicate:                      "Duplicate",
	FilterTabAllVersions:           "All Versions",
	FilterTabOnlineVersion:         "Online Version",
	FilterTabNamedVersions:         "Named Versions",
}

var Messages_zh_CN = &Messages{
	Category:                       "目录",
	Preview:                        "预览",
	Containers:                     "组件",
	AddContainers:                  "增加组件",
	New:                            "新增",
	Shared:                         "公用的",
	Select:                         "选择",
	SelectedTemplateLabel:          "模板",
	CreateFromTemplate:             "从模板中创建",
	ChangeTemplate:                 "更改模版",
	RelatedOnlinePages:             "相关在线页面",
	RepublishAllRelatedOnlinePages: "重新发布所有页面",
	Unnamed:                        "未命名",
	NotDescribed:                   "未描述",
	Blank:                          "空白",
	NewPage:                        "新页面",
	Duplicate:                      "复制",
	FilterTabAllVersions:           "所有版本",
	FilterTabOnlineVersion:         "在线版本",
	FilterTabNamedVersions:         "已命名版本",
}

var Messages_ja_JP = &Messages{
	Category:                       "カテゴリー",
	Preview:                        "プレビュー",
	Containers:                     "コンテナ",
	AddContainers:                  "コンテナを追加する",
	New:                            "新規",
	Shared:                         "共有",
	Select:                         "選択する",
	SelectedTemplateLabel:          "テンプレート",
	CreateFromTemplate:             "テンプレートから新規作成する",
	ChangeTemplate:                 "テンプレートを変更する",
	RelatedOnlinePages:             "関連オンラインページ",
	RepublishAllRelatedOnlinePages: "すべて再公開",
	Unnamed:                        "名前なし",
	NotDescribed:                   "記述されていません",
	Blank:                          "空白",
	NewPage:                        "新しいページ",
	FilterTabAllVersions:           "全てのバージョン",
	FilterTabOnlineVersion:         "オンラインバージョン",
	FilterTabNamedVersions:         "名付け済みバージョン",
}
