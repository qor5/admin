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
	Hide                      string
	Show                      string
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
	ViewPage                  string
	EditPage                  string
	EditLastDraft             string
	ScheduledAt               string
	OnlineHit                 string
	NoContentHit              string
	PageBuilder               string
	PageTemplate              string

	InvalidPathMsg            string
	InvalidTitleMsg           string
	InvalidNameMsg            string
	InvalidSlugMsg            string
	ConflictSlugMsg           string
	ConflictPathMsg           string
	ExistingPathMsg           string
	UnableDeleteCategoryMsg   string
	WouldCausePageConflictMsg string
	Versions                  string
	NewContainer              string
	Settings                  string
	SelectElementMsg          string
	StartBuildingMsg          string
	StartBuildingTemplateMsg  string
	StartBuildingSubMsg       string

	ListHeaderID          string
	ListHeaderTitle       string
	ListHeaderName        string
	ListHeaderPath        string
	ListHeaderDescription string

	FilterTabAll       string
	FilterTabFilled    string
	FilterTabNotFilled string

	ModalTitleConfirm          string
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

	CategoryDeleteConfirmationText     string
	TheResourceCanNotBeModified        string
	MarkAsShared                       string
	Copy                               string
	SharedContainerHasBeenUpdated      string
	TemplateFixedAreaMessage           string
	SharedContainerModificationWarning string
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
	Hide:                           "Hide",
	Show:                           "Show",
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
	ViewPage:                  "View Page",
	EditPage:                  "Edit Page",
	EditLastDraft:             "Edit Last Draft",
	ScheduledAt:               "Scheduled at",
	OnlineHit:                 "This version cannot be edited once published. To make changes, please duplicate the current version",
	NoContentHit:              "This page has no content yet, start to edit in page builder",

	InvalidPathMsg:            "Invalid Path",
	InvalidTitleMsg:           "Invalid Title",
	InvalidNameMsg:            "Invalid Name",
	InvalidSlugMsg:            "Invalid Slug",
	ConflictSlugMsg:           "Conflicting Slug",
	ConflictPathMsg:           "Conflicting Path",
	ExistingPathMsg:           "Existing Path",
	UnableDeleteCategoryMsg:   "To delete this category you need to remove all association to products first",
	WouldCausePageConflictMsg: "This path would cause URL conflicts with existing pages",
	Versions:                  "versions",
	NewContainer:              "New Container",
	Settings:                  "settings",
	SelectElementMsg:          "Select an element and change the setting here.",
	StartBuildingMsg:          "Start building a page",
	StartBuildingTemplateMsg:  "Start building a template",
	StartBuildingSubMsg:       "By Browsing and selecting container from the library",

	ListHeaderID:          "ID",
	ListHeaderTitle:       "Title",
	ListHeaderName:        "Name",
	ListHeaderPath:        "Path",
	ListHeaderDescription: "Description",

	FilterTabAll:       "All",
	FilterTabFilled:    "Filled",
	FilterTabNotFilled: "Not Filled",

	ModalTitleConfirm:          "Confirm",
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
	AddPageTemplate:                    "Add Page Template",
	Name:                               "Name",
	Description:                        "Description",
	CategoryDeleteConfirmationText:     "this will remove all the records in all localized languages",
	TheResourceCanNotBeModified:        "The resource can not be modified",
	MarkAsShared:                       "Mark As Shared",
	Copy:                               "Copy",
	SharedContainerHasBeenUpdated:      "The shared container on this page has been updated. You may notice differences between the preview and the live page.",
	TemplateFixedAreaMessage:           "This container is fixed and cannot be updated",
	SharedContainerModificationWarning: "This is a shared container. Any modifications you make will apply to all pages that use it",
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
	Hide:                           "隐藏",
	Show:                           "显示",
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
	ViewPage:                  "查看页面",
	EditPage:                  "编辑页面",
	EditLastDraft:             "编辑最后草稿",
	ScheduledAt:               "安排在",
	OnlineHit:                 "这个版本发布后不能再编辑。要进行更改，请复制当前版本",
	NoContentHit:              "这个页面没有内容，在page builder中开始编辑",

	InvalidPathMsg:            "无效的路径",
	InvalidTitleMsg:           "无效的标题",
	InvalidNameMsg:            "无效的名称",
	InvalidSlugMsg:            "无效的Slug",
	ConflictSlugMsg:           "冲突的Slug",
	ConflictPathMsg:           "冲突的路径",
	ExistingPathMsg:           "已存在的路径",
	UnableDeleteCategoryMsg:   "这个分类没办法被删除,因为已被页面使用",
	WouldCausePageConflictMsg: "此路径会导致与现有页面的URL冲突",
	Versions:                  "版本",
	NewContainer:              "新增组件",
	Settings:                  "设置",
	SelectElementMsg:          "选择一个组件，这里会变成设置",
	StartBuildingMsg:          "开始构建页面",
	StartBuildingTemplateMsg:  "开始构建模版",
	StartBuildingSubMsg:       "从库中选择组件",
	PageBuilder:               "页面构建",
	PageTemplate:              "页面模版",

	ListHeaderID:          "ID",
	ListHeaderTitle:       "标题",
	ListHeaderName:        "名称",
	ListHeaderPath:        "路径",
	ListHeaderDescription: "描述",

	FilterTabAll:       "全部",
	FilterTabFilled:    "已填写",
	FilterTabNotFilled: "未填写",

	ModalTitleConfirm:          "确认",
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

	CategoryDeleteConfirmationText:     "这将删除所有本地化语言中的所有记录",
	TheResourceCanNotBeModified:        "该资源无法被修改",
	MarkAsShared:                       "标记为已共享",
	Copy:                               "复制",
	SharedContainerHasBeenUpdated:      "此页面上的共享容器已更新。您可能会注意到预览和实时页面之间的差异。",
	TemplateFixedAreaMessage:           "此区域由模板固定，无法编辑。",
	SharedContainerModificationWarning: "这是一个共享容器。您所做的任何修改都将应用于使用它的所有页面",
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
	Hide:                           "隠す",
	Show:                           "表示",
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
	ViewPage:                  "ページを表示",
	EditPage:                  "ページの編集",
	EditLastDraft:             "最後の下書きを編集",
	ScheduledAt:               "公開開始日時",
	OnlineHit:                 "このバージョンは一度公開されると編集できなくなります。変更を加えるには、現在のバージョンを複製してください",
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

	ModalTitleConfirm:          "確認",
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
	AddPageTemplate:                    "ページテンプレートを追加",
	Name:                               "名前",
	Description:                        "説明",
	CategoryDeleteConfirmationText:     "これは、すべてのローカライズされた言語のすべてのレコードを削除します",
	TheResourceCanNotBeModified:        "このリソースは変更できません",
	MarkAsShared:                       "共有済みとしてマーク",
	Copy:                               "コピー",
	SharedContainerHasBeenUpdated:      "このページの共有コンテナが更新されました。プレビューとライブページの間に違いがあるかもしれません。",
	TemplateFixedAreaMessage:           "この領域はテンプレートによって固定されており、編集できません。",
	SharedContainerModificationWarning: "これは共有コンテナです。行った変更は、それを使用するすべてのページに適用されます",
}

type ModelsI18nModulePage struct {
	PagesPage string
}

var ModelsI18nModulePage_EN = ModelsI18nModulePage{
	PagesPage: "Page",
}

var ModelsI18nModulePage_Zh = ModelsI18nModulePage{
	PagesPage: "Page",
}

var ModelsI18nModulePage_JP = ModelsI18nModulePage{
	PagesPage: "ページ",
}
