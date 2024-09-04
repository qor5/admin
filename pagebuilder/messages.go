package pagebuilder

import (
	"fmt"

	"github.com/qor5/x/v3/i18n"
)

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
	BlankPage                      string
	AddBlankPage                   string
	NewPage                        string
	FilterTabAllVersions           string
	FilterTabOnlineVersion         string
	FilterTabNamedVersions         string

	Rename                    string
	PageOverView              string
	Others                    string
	Add                       string
	AddContainer              string
	BuildYourPages            string
	BuildYourTemplates        string
	PlaceAnElementFromLibrary string
	NewElement                string
	Title                     string
	Slug                      string
	EditPage                  string
	ScheduledAt               string
	OnlineHit                 string
	NoContentHit              string
	PageBuilder               string
	PageTemplate              string

	InvalidPathMsg           string
	InvalidTitleMsg          string
	InvalidNameMsg           string
	InvalidSlugMsg           string
	ConflictSlugMsg          string
	ConflictPathMsg          string
	ExistingPathMsg          string
	UnableDeleteCategoryMsg  string
	Versions                 string
	NewContainer             string
	Settings                 string
	SelectElementMsg         string
	StartBuildingMsg         string
	StartBuildingTemplateMsg string
	StartBuildingSubMsg      string

	ListHeaderID          string
	ListHeaderTitle       string
	ListHeaderName        string
	ListHeaderPath        string
	ListHeaderDescription string

	FilterTabAll       string
	FilterTabFilled    string
	FilterTabNotFilled string

	ModelLabelPages            string
	ModelLabelPage             string
	ModelLabelSharedContainers string
	ModelLabelSharedContainer  string
	ModelLabelDemoContainers   string
	ModelLabelDemoContainer    string
	ModelLabelTemplates        string
	ModelLabelTemplate         string
	ModelLabelPageCategories   string
	ModelLabelPageCategory     string
	AreWantDeleteContainer     func(v string) string
	AddPageTemplate            string
	Name                       string
	Description                string
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
	BlankPage:                      "Blank Page",
	AddBlankPage:                   "Add Blank Page",
	NewPage:                        "New Page",
	FilterTabAllVersions:           "All Versions",
	FilterTabOnlineVersion:         "Online Versions",
	FilterTabNamedVersions:         "Named Versions",
	PageBuilder:                    "Page Builder",
	PageTemplate:                   "Page Template",
	Rename:                         "Rename",
	PageOverView:                   "Page Overview",

	Others:                    "Others",
	Add:                       "Add",
	AddContainer:              "Add Container",
	BuildYourPages:            "Build your pages",
	BuildYourTemplates:        "Build your templates",
	PlaceAnElementFromLibrary: "Place an element from  library.",
	NewElement:                "New Element",
	Title:                     "Title",
	Slug:                      "Slug",
	EditPage:                  "Edit Page",
	ScheduledAt:               "Scheduled at",
	OnlineHit:                 "The version cannot be edited directly after it is released. Please copy the version and edit it.",
	NoContentHit:              "This page has no content yet, start to edit in page builder",

	InvalidPathMsg:           "Invalid Path",
	InvalidTitleMsg:          "Invalid Title",
	InvalidNameMsg:           "Invalid Name",
	InvalidSlugMsg:           "Invalid Slug",
	ConflictSlugMsg:          "Conflicting Slug",
	ConflictPathMsg:          "Conflicting Path",
	ExistingPathMsg:          "Existing Path",
	UnableDeleteCategoryMsg:  "To delete this category you need to remove all association to products first",
	Versions:                 "versions",
	NewContainer:             "New Container",
	Settings:                 "settings",
	SelectElementMsg:         "Select an element and change the setting here.",
	StartBuildingMsg:         "Start building a page",
	StartBuildingTemplateMsg: "Start building a template",
	StartBuildingSubMsg:      "By Browsing and selecting container from the library",

	ListHeaderID:          "ID",
	ListHeaderTitle:       "Title",
	ListHeaderName:        "Name",
	ListHeaderPath:        "Path",
	ListHeaderDescription: "Description",

	FilterTabAll:       "All",
	FilterTabFilled:    "Filled",
	FilterTabNotFilled: "Not Filled",

	ModelLabelPages:            "Pages",
	ModelLabelPage:             "Page",
	ModelLabelSharedContainers: "Shared Containers",
	ModelLabelSharedContainer:  "Shared Container",
	ModelLabelDemoContainers:   "Demo Containers",
	ModelLabelDemoContainer:    "Demo Container",
	ModelLabelTemplates:        "Templates",
	ModelLabelTemplate:         "Template",
	ModelLabelPageCategories:   "Page Categories",
	ModelLabelPageCategory:     "Page Category",
	AreWantDeleteContainer: func(v string) string {
		return fmt.Sprintf("Are you sure you want to delete %v?", v)
	},
	AddPageTemplate: "Add Page Template",
	Name:            "Name",
	Description:     "Description",
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
	BlankPage:                      "空白页",
	AddBlankPage:                   "新增空白页",
	NewPage:                        "新页面",
	FilterTabAllVersions:           "所有版本",
	FilterTabOnlineVersion:         "在线版本",
	FilterTabNamedVersions:         "已命名版本",
	Rename:                         "重命名",
	PageOverView:                   "页面概览",

	Others:                    "其他",
	Add:                       "新增",
	AddContainer:              "新增组件",
	BuildYourPages:            "构建你的页面",
	BuildYourTemplates:        "构建你的模版",
	PlaceAnElementFromLibrary: "从你的库从选择一个组件",
	NewElement:                "新的组件",
	Title:                     "编辑",
	Slug:                      "Slug",
	EditPage:                  "编辑页面",
	ScheduledAt:               "安排在",
	OnlineHit:                 "这个版本无法在上线后直接编辑.请拷贝这个版本再编辑.",
	NoContentHit:              "这个页面没有内容，在page builder中开始编辑",

	InvalidPathMsg:           "无效的路径",
	InvalidTitleMsg:          "无效的标题",
	InvalidNameMsg:           "无效的名称",
	InvalidSlugMsg:           "无效的Slug",
	ConflictSlugMsg:          "冲突的Slug",
	ConflictPathMsg:          "冲突的路径",
	ExistingPathMsg:          "已存在的路径",
	UnableDeleteCategoryMsg:  "这个分类没办法被删除,因为已被页面使用",
	Versions:                 "版本",
	NewContainer:             "新增组件",
	Settings:                 "设置",
	SelectElementMsg:         "选择一个组件，这里会变成设置",
	StartBuildingMsg:         "开始构建页面",
	StartBuildingTemplateMsg: "开始构建模版",
	StartBuildingSubMsg:      "从库中选择组件",
	PageBuilder:              "页面构建",
	PageTemplate:             "页面模版",

	ListHeaderID:          "ID",
	ListHeaderTitle:       "标题",
	ListHeaderName:        "名称",
	ListHeaderPath:        "路径",
	ListHeaderDescription: "描述",

	FilterTabAll:       "全部",
	FilterTabFilled:    "已填写",
	FilterTabNotFilled: "未填写",

	ModelLabelPages:            "页面管理",
	ModelLabelPage:             "页面",
	ModelLabelSharedContainers: "公用组件",
	ModelLabelSharedContainer:  "公用组件",
	ModelLabelDemoContainers:   "示例组件",
	ModelLabelDemoContainer:    "示例组件",
	ModelLabelTemplates:        "模板页面",
	ModelLabelTemplate:         "模板页面",
	ModelLabelPageCategories:   "目录管理",
	ModelLabelPageCategory:     "目录",
	AreWantDeleteContainer: func(v string) string {
		return fmt.Sprintf("你确定要删除 %v?", v)
	},
	AddPageTemplate: "添加页面模版",
	Name:            "名称",
	Description:     "说明",
}

var Messages_ja_JP = &Messages{
	Category:                       "カテゴリ",
	Preview:                        "プレビュー",
	Containers:                     "コンテナ",
	AddContainers:                  "コンテナの追加",
	New:                            "作成する",
	Shared:                         "共有",
	Select:                         "選択",
	SelectedTemplateLabel:          "テンプレート",
	CreateFromTemplate:             "テンプレートから作成",
	ChangeTemplate:                 "テンプレートの変更",
	RelatedOnlinePages:             "関連するオンラインページ",
	RepublishAllRelatedOnlinePages: "すべてを再公開する",
	Unnamed:                        "名前なし",
	NotDescribed:                   "説明なし",
	Blank:                          "空白",
	BlankPage:                      "空白ページ",
	AddBlankPage:                   "空白ページを追加",
	NewPage:                        "新しいページ",
	FilterTabAllVersions:           "すべてのバージョン",
	FilterTabOnlineVersion:         "オンラインバージョン",
	FilterTabNamedVersions:         "名前付きバージョン",
	Rename:                         "名前の変更",
	PageOverView:                   "ページの概要",
	PageBuilder:                    "ページビルダー",
	PageTemplate:                   "ページテンプレート",

	Others:                    "その他",
	Add:                       "追加",
	AddContainer:              "コンテナの追加",
	BuildYourPages:            "ページの作成",
	BuildYourTemplates:        "テンプレートを作成する",
	PlaceAnElementFromLibrary: "ライブラリからコンテナを配置します。",
	NewElement:                "新しいコンテナ",
	Title:                     "タイトル",
	Slug:                      "スラッグ",
	EditPage:                  "ページの編集",
	ScheduledAt:               "公開開始日時",
	OnlineHit:                 "バージョンはリリース後直接に編集できません。バージョンをコピーして編集してください。",
	NoContentHit:              "このページにはまだコンテンツがありません。ページビルダーで編集を開始してください",

	InvalidPathMsg:           "無効なパス",
	InvalidTitleMsg:          "無効なタイトル",
	InvalidNameMsg:           "無効な名前",
	InvalidSlugMsg:           "無効なスラッグ",
	ConflictSlugMsg:          "競合するスラッグ",
	ConflictPathMsg:          "競合するパス",
	ExistingPathMsg:          "既存のパス",
	UnableDeleteCategoryMsg:  "このカテゴリを削除するには、まず商品との関連付けをすべて削除する必要があります。",
	Versions:                 "バージョン",
	NewContainer:             "新しいコンテナ",
	Settings:                 "設定",
	SelectElementMsg:         "コンテナを選択後、設定変更してください",
	StartBuildingMsg:         "ページの構築を開始します",
	StartBuildingTemplateMsg: "テンプレートの作成を開始する",
	StartBuildingSubMsg:      "ライブラリからコンテナを参照して選択する",

	ListHeaderID:          "ID",
	ListHeaderTitle:       "タイトル",
	ListHeaderName:        "名前",
	ListHeaderPath:        "パス",
	ListHeaderDescription: "説明",

	FilterTabAll:       "すべて",
	FilterTabFilled:    "入力済み",
	FilterTabNotFilled: "未入力",

	ModelLabelPages:            "ページ",
	ModelLabelPage:             "ページ",
	ModelLabelSharedContainers: "共有コンテナ",
	ModelLabelSharedContainer:  "共有コンテナ",
	ModelLabelDemoContainers:   "デモコンテナ",
	ModelLabelDemoContainer:    "デモコンテナ",
	ModelLabelTemplates:        "テンプレート",
	ModelLabelTemplate:         "テンプレート",
	ModelLabelPageCategories:   "ページカテゴリ",
	ModelLabelPageCategory:     "ページカテゴリ",
	AreWantDeleteContainer: func(v string) string {
		return fmt.Sprintf("%v を削除してもよろしいですか?", v)
	},
	AddPageTemplate: "ページテンプレートを追加",
	Name:            "名前",
	Description:     "説明",
}
