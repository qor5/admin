package pagebuilder

import "github.com/qor5/x/v3/i18n"

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
	FilterTabAllVersions           string
	FilterTabOnlineVersion         string
	FilterTabNamedVersions         string

	Rename                    string
	PageOverView              string
	Others                    string
	Add                       string
	AddComponent              string
	BuildYourPages            string
	PlaceAnElementFromLibrary string
	NewElement                string
	Title                     string
	Slug                      string
	EditPage                  string
	ScheduledAt               string
	OnlineHit                 string
	NoContentHit              string
	PageBuilder               string

	InvalidPathMsg          string
	InvalidTitleMsg         string
	InvalidNameMsg          string
	InvalidSlugMsg          string
	ConflictSlugMsg         string
	ConflictPathMsg         string
	ExistingPathMsg         string
	UnableDeleteCategoryMsg string
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
	FilterTabAllVersions:           "All Versions",
	FilterTabOnlineVersion:         "Online Versions",
	FilterTabNamedVersions:         "Named Versions",
	PageBuilder:                    "Page Builder",
	Rename:                         "Rename",
	PageOverView:                   "Page Overview",

	Others:                    "Others",
	Add:                       "Add",
	AddComponent:              "Add Component",
	BuildYourPages:            "Build your pages",
	PlaceAnElementFromLibrary: "Place an element from  library.",
	NewElement:                "New Element",
	Title:                     "Title",
	Slug:                      "Slug",
	EditPage:                  "Edit Page",
	ScheduledAt:               "Scheduled at",
	OnlineHit:                 "The version cannot be edited directly after it is released. Please copy the version and edit it.",
	NoContentHit:              "This page has no content yet, start to edit in page builder",

	InvalidPathMsg:          "Invalid Path",
	InvalidTitleMsg:         "Invalid Title",
	InvalidNameMsg:          "Invalid Name",
	InvalidSlugMsg:          "Invalid Slug",
	ConflictSlugMsg:         "Conflicting Slug",
	ConflictPathMsg:         "Conflicting Path",
	ExistingPathMsg:         "Existing Path",
	UnableDeleteCategoryMsg: "this category cannot be deleted because it has used with pages",
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
	FilterTabAllVersions:           "所有版本",
	FilterTabOnlineVersion:         "在线版本",
	FilterTabNamedVersions:         "已命名版本",
	Rename:                         "重命名",
	PageOverView:                   "页面概览",

	Others:                    "其他",
	Add:                       "新增",
	AddComponent:              "新增组件",
	BuildYourPages:            "构建你的页面",
	PlaceAnElementFromLibrary: "从你的库从选择一个组件",
	NewElement:                "新的组件",
	Title:                     "编辑",
	Slug:                      "Slug",
	EditPage:                  "编辑页面",
	ScheduledAt:               "安排在",
	OnlineHit:                 "这个版本无法在上线后直接编辑.请拷贝这个版本再编辑.",
	NoContentHit:              "这个页面没有内容，在page builder中开始编辑",

	InvalidPathMsg:          "无效的路径",
	InvalidTitleMsg:         "无效的辩题",
	InvalidNameMsg:          "无效的名称",
	InvalidSlugMsg:          "无效的Slug",
	ConflictSlugMsg:         "冲突的Slug",
	ConflictPathMsg:         "冲突的路径",
	ExistingPathMsg:         "已存在的路径",
	UnableDeleteCategoryMsg: "这个分类没办法被删除,因为已被页面使用",
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
	Rename:                         "名前の変更",
	PageOverView:                   "ページ概要",

	Others:                    "Others",
	Add:                       "Add",
	AddComponent:              "Add Component",
	BuildYourPages:            "Build your pages",
	PlaceAnElementFromLibrary: "Place an element from  library.",
	NewElement:                "New Element",
	Title:                     "Title",
	Slug:                      "Slug",
	EditPage:                  "Edit Page",
	ScheduledAt:               "Scheduled at",
	OnlineHit:                 "The version cannot be edited directly after it is released. Please copy the version and edit it.",
	NoContentHit:              "This page has no content yet, start to edit in page builder",

	InvalidPathMsg:          "Invalid Path",
	InvalidTitleMsg:         "Invalid Title",
	InvalidNameMsg:          "Invalid Name",
	InvalidSlugMsg:          "Invalid Slug",
	ConflictSlugMsg:         "Conflicting Slug",
	ConflictPathMsg:         "Conflicting Path",
	ExistingPathMsg:         "Existing Path",
	UnableDeleteCategoryMsg: "this category cannot be deleted because it has used with pages",
}
