package admin

import (
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	l10n_view "github.com/qor/qor5/l10n/views"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/note"
	"github.com/qor/qor5/pagebuilder"
	publish_view "github.com/qor/qor5/publish/views"
	"github.com/qor/qor5/utils"
)

const I18nExampleKey i18n.ModuleKey = "I18nExampleKey"

type Messages struct {
	FilterTabsAll            string
	FilterTabsHasUnreadNotes string
	FilterTabsActive         string
}

var Messages_en_US = &Messages{
	FilterTabsAll:            "All",
	FilterTabsHasUnreadNotes: "Has Unread Notes",
	FilterTabsActive:         "Active",
}

var Messages_ja_JP = &Messages{
	FilterTabsAll:            "すべて",
	FilterTabsHasUnreadNotes: "未読のノートがあります",
	FilterTabsActive:         "有効",
}

var Messages_zh_CN = &Messages{
	FilterTabsAll:            "全部",
	FilterTabsHasUnreadNotes: "未读备注",
	FilterTabsActive:         "有效",
}

type Messages_ModelsI18nModuleKey struct {
	QOR5Example string
	Roles       string
	Users       string

	Posts          string
	PostsID        string
	PostsTitle     string
	PostsHeroImage string
	PostsBody      string
	Example        string
	Settings       string
	Post           string
	PostsBodyImage string

	SeoPost             string
	SeoVariableTitle    string
	SeoVariableSiteName string

	PageBuilder              string
	Pages                    string
	SharedContainers         string
	DemoContainers           string
	Templates                string
	Categories               string
	ProductManagement        string
	Products                 string
	SiteManagement           string
	SEO                      string
	UserManagement           string
	Profile                  string
	FeaturedModelsManagement string
	InputHarnesses           string
	ListEditorExample        string
	Customers                string
	ListModels               string
	MicrositeModels          string
	Workers                  string
	ActivityLogs             string
	MediaLibrary             string

	PagesID         string
	PagesTitle      string
	PagesSlug       string
	PagesLocale     string
	PagesNotes      string
	PagesDraftCount string
	PagesOnline     string

	Page                   string
	PagesStatus            string
	PagesSchedule          string
	PagesCategoryID        string
	PagesTemplateSelection string
	PagesEditContainer     string

	WebHeader       string
	WebHeadersColor string
	Header          string

	WebFooter             string
	WebFootersEnglishUrl  string
	WebFootersJapaneseUrl string
	Footer                string

	VideoBanner                       string
	VideoBannersAddTopSpace           string
	VideoBannersAddBottomSpace        string
	VideoBannersAnchorID              string
	VideoBannersVideo                 string
	VideoBannersBackgroundVideo       string
	VideoBannersMobileBackgroundVideo string
	VideoBannersVideoCover            string
	VideoBannersMobileVideoCover      string
	VideoBannersHeading               string
	VideoBannersPopupText             string
	VideoBannersText                  string
	VideoBannersLinkText              string
	VideoBannersLink                  string

	Heading                   string
	HeadingsAddTopSpace       string
	HeadingsAddBottomSpace    string
	HeadingsAnchorID          string
	HeadingsHeading           string
	HeadingsFontColor         string
	HeadingsBackgroundColor   string
	HeadingsLink              string
	HeadingsLinkText          string
	HeadingsLinkDisplayOption string
	HeadingsText              string

	BrandGrid                string
	BrandGridsAddTopSpace    string
	BrandGridsAddBottomSpace string
	BrandGridsAnchorID       string
	BrandGridsBrands         string

	ListContent                   string
	ListContentsAddTopSpace       string
	ListContentsAddBottomSpace    string
	ListContentsAnchorID          string
	ListContentsBackgroundColor   string
	ListContentsItems             string
	ListContentsLink              string
	ListContentsLinkText          string
	ListContentsLinkDisplayOption string

	ImageContainer                           string
	ImageContainersAddTopSpace               string
	ImageContainersAddBottomSpace            string
	ImageContainersAnchorID                  string
	ImageContainersBackgroundColor           string
	ImageContainersTransitionBackgroundColor string
	ImageContainersImage                     string
	Image                                    string

	InNumber                string
	InNumbersAddTopSpace    string
	InNumbersAddBottomSpace string
	InNumbersAnchorID       string
	InNumbersHeading        string
	InNumbersItems          string
	InNumbers               string

	ContactForm                    string
	ContactFormsAddTopSpace        string
	ContactFormsAddBottomSpace     string
	ContactFormsAnchorID           string
	ContactFormsHeading            string
	ContactFormsText               string
	ContactFormsSendButtonText     string
	ContactFormsFormButtonText     string
	ContactFormsMessagePlaceholder string
	ContactFormsNamePlaceholder    string
	ContactFormsEmailPlaceholder   string
	ContactFormsThankyouMessage    string
	ContactFormsActionUrl          string
	ContactFormsPrivacyPolicy      string
}

var Messages_zh_CN_ModelsI18nModuleKey = &Messages_ModelsI18nModuleKey{
	Posts:          "帖子 示例",
	PostsID:        "ID",
	PostsTitle:     "标题",
	PostsHeroImage: "主图",
	PostsBody:      "内容",
	Example:        "QOR5演示",
	Settings:       "SEO 设置",
	Post:           "帖子",
	PostsBodyImage: "内容图片",

	SeoPost:             "帖子",
	SeoVariableTitle:    "标题",
	SeoVariableSiteName: "站点名称",

	QOR5Example: "QOR5 示例",
	Roles:       "权限管理",
	Users:       "用户管理",

	PageBuilder:              "页面管理菜单",
	Pages:                    "页面管理",
	SharedContainers:         "公用组件",
	DemoContainers:           "示例组件",
	Templates:                "模板页面",
	Categories:               "目录管理",
	ProductManagement:        "产品管理菜单",
	Products:                 "产品管理",
	SiteManagement:           "站点管理菜单",
	SEO:                      "SEO 管理",
	UserManagement:           "用户管理菜单",
	Profile:                  "个人页面",
	FeaturedModelsManagement: "特色模块管理菜单",
	InputHarnesses:           "Input 示例",
	ListEditorExample:        "ListEditor 示例",
	Customers:                "Customers 示例",
	ListModels:               "发布带排序及分页模块 示例",
	MicrositeModels:          "Microsite 示例",
	Workers:                  "后台工作进程管理",
	ActivityLogs:             "操作日志",
	MediaLibrary:             "媒体库",

	PagesID:         "ID",
	PagesTitle:      "标题",
	PagesSlug:       "Slug",
	PagesLocale:     "地区",
	PagesNotes:      "备注",
	PagesDraftCount: "草稿数",
	PagesOnline:     "在线",

	Page:                   "Page",
	PagesStatus:            "PagesStatus",
	PagesSchedule:          "PagesSchedule",
	PagesCategoryID:        "PagesCategoryID",
	PagesTemplateSelection: "PagesTemplateSelection",
	PagesEditContainer:     "PagesEditContainer",

	WebHeader:       "WebHeader",
	WebHeadersColor: "WebHeadersColor",
	Header:          "Header",

	WebFooter:             "WebFooter",
	WebFootersEnglishUrl:  "WebFootersEnglishUrl",
	WebFootersJapaneseUrl: "WebFootersJapaneseUrl",
	Footer:                "Footer",

	VideoBanner:                       "VideoBanner",
	VideoBannersAddTopSpace:           "VideoBannersAddTopSpace",
	VideoBannersAddBottomSpace:        "VideoBannersAddBottomSpace",
	VideoBannersAnchorID:              "VideoBannersAnchorID",
	VideoBannersVideo:                 "VideoBannersVideo",
	VideoBannersBackgroundVideo:       "VideoBannersBackgroundVideo",
	VideoBannersMobileBackgroundVideo: "VideoBannersMobileBackgroundVideo",
	VideoBannersVideoCover:            "VideoBannersVideoCover",
	VideoBannersMobileVideoCover:      "VideoBannersMobileVideoCover",
	VideoBannersHeading:               "VideoBannersHeading",
	VideoBannersPopupText:             "VideoBannersPopupText",
	VideoBannersText:                  "VideoBannersText",
	VideoBannersLinkText:              "VideoBannersLinkText",
	VideoBannersLink:                  "VideoBannersLink",

	Heading:                   "Heading",
	HeadingsAddTopSpace:       "HeadingsAddTopSpace",
	HeadingsAddBottomSpace:    "HeadingsAddBottomSpace",
	HeadingsAnchorID:          "HeadingsAnchorID",
	HeadingsHeading:           "HeadingsHeading",
	HeadingsFontColor:         "HeadingsFontColor",
	HeadingsBackgroundColor:   "HeadingsBackgroundColor",
	HeadingsLink:              "HeadingsLink",
	HeadingsLinkText:          "HeadingsLinkText",
	HeadingsLinkDisplayOption: "HeadingsLinkDisplayOption",
	HeadingsText:              "HeadingsText",

	BrandGrid:                "BrandGrid",
	BrandGridsAddTopSpace:    "BrandGridsAddTopSpace",
	BrandGridsAddBottomSpace: "BrandGridsAddBottomSpace",
	BrandGridsAnchorID:       "BrandGridsAnchorID",
	BrandGridsBrands:         "BrandGridsBrands",

	ListContent:                   "ListContent",
	ListContentsAddTopSpace:       "ListContentsAddTopSpace",
	ListContentsAddBottomSpace:    "ListContentsAddBottomSpace",
	ListContentsAnchorID:          "ListContentsAnchorID",
	ListContentsBackgroundColor:   "ListContentsBackgroundColor",
	ListContentsItems:             "ListContentsItems",
	ListContentsLink:              "ListContentsLink",
	ListContentsLinkText:          "ListContentsLinkText",
	ListContentsLinkDisplayOption: "ListContentsLinkDisplayOption",

	ImageContainer:                           "ImageContainer",
	ImageContainersAddTopSpace:               "ImageContainersAddTopSpace",
	ImageContainersAddBottomSpace:            "ImageContainersAddBottomSpace",
	ImageContainersAnchorID:                  "ImageContainersAnchorID",
	ImageContainersBackgroundColor:           "ImageContainersBackgroundColor",
	ImageContainersTransitionBackgroundColor: "ImageContainersTransitionBackgroundColor",
	ImageContainersImage:                     "ImageContainersImage",
	Image:                                    "Image",

	InNumber:                "InNumber",
	InNumbersAddTopSpace:    "InNumbersAddTopSpace",
	InNumbersAddBottomSpace: "InNumbersAddBottomSpace",
	InNumbersAnchorID:       "InNumbersAnchorID",
	InNumbersHeading:        "InNumbersHeading",
	InNumbersItems:          "InNumbersItems",
	InNumbers:               "InNumbers",

	ContactForm:                    "ContactForm",
	ContactFormsAddTopSpace:        "ContactFormsAddTopSpace",
	ContactFormsAddBottomSpace:     "ContactFormsAddBottomSpace",
	ContactFormsAnchorID:           "ContactFormsAnchorID",
	ContactFormsHeading:            "ContactFormsHeading",
	ContactFormsText:               "ContactFormsText",
	ContactFormsSendButtonText:     "ContactFormsSendButtonText",
	ContactFormsFormButtonText:     "ContactFormsFormButtonText",
	ContactFormsMessagePlaceholder: "ContactFormsMessagePlaceholder",
	ContactFormsNamePlaceholder:    "ContactFormsNamePlaceholder",
	ContactFormsEmailPlaceholder:   "ContactFormsEmailPlaceholder",
	ContactFormsThankyouMessage:    "ContactFormsThankyouMessage",
	ContactFormsActionUrl:          "ContactFormsActionUrl",
	ContactFormsPrivacyPolicy:      "ContactFormsPrivacyPolicy",
}

var Messages_ja_JP_ModelsI18nModuleKey = &Messages_ModelsI18nModuleKey{
	Posts:          "投稿",
	PostsID:        "投稿ID",
	PostsTitle:     "投稿タイトル",
	PostsHeroImage: "メイン画像",
	PostsBody:      "コンテンツ",
	Example:        "QOR5サンプル",
	Settings:       "設定",
	Post:           "投稿",
	PostsBodyImage: "内容イメージ",

	SeoPost:             "SEO 投稿",
	SeoVariableTitle:    "SEO タイトル",
	SeoVariableSiteName: "SEO サイト名",

	QOR5Example: "QOR5サンプル",
	Roles:       "ユーザー権限",
	Users:       "ユーザー",

	PageBuilder:              "ページビルダー",
	Pages:                    "ページ",
	SharedContainers:         "共有コンテナ",
	DemoContainers:           "デモ用コン店た",
	Templates:                "テンプレート",
	Categories:               "カテゴリー",
	ProductManagement:        "製品管理",
	Products:                 "製品",
	SiteManagement:           "サイト管理",
	SEO:                      "SEO",
	UserManagement:           "ユーザー管理",
	Profile:                  "プロフィール",
	FeaturedModelsManagement: "モデル管理",
	InputHarnesses:           "ハーネスを入力",
	ListEditorExample:        "リスト編集サンプル",
	Customers:                "お客さま",
	ListModels:               "リストモデル",
	MicrositeModels:          "マイクロサイトモデル",
	Workers:                  "ワーカーズ",
	ActivityLogs:             "アクティビティ履歴",
	MediaLibrary:             "メディアライブラリ",

	PagesID:         "ID",
	PagesTitle:      "タイトル",
	PagesSlug:       "スラッグ",
	PagesLocale:     "ローカル",
	PagesNotes:      "ノート",
	PagesDraftCount: "カウント下書き",
	PagesOnline:     "オンライン",

	Page:                   "ページ",
	PagesStatus:            "状態",
	PagesSchedule:          "スケジュール",
	PagesCategoryID:        "カテゴリーID",
	PagesTemplateSelection: "テンプレート選択",
	PagesEditContainer:     "コンテナ編集",

	WebHeader:       "ウェブヘッダー",
	WebHeadersColor: "カラー",
	Header:          "ヘッダー",

	WebFooter:             "ウェブ用フッター",
	WebFootersEnglishUrl:  "英語用URL",
	WebFootersJapaneseUrl: "日本語用URL",
	Footer:                "フッター",

	VideoBanner:                       "動画バナー",
	VideoBannersAddTopSpace:           "上方に空白を追加",
	VideoBannersAddBottomSpace:        "下方に空白を追加",
	VideoBannersAnchorID:              "アンカーID",
	VideoBannersVideo:                 "動画",
	VideoBannersBackgroundVideo:       "背景動画",
	VideoBannersMobileBackgroundVideo: "モバイル用背景動画",
	VideoBannersVideoCover:            "動画カバー",
	VideoBannersMobileVideoCover:      "モバイル用動画カバー",
	VideoBannersHeading:               "ヘディング",
	VideoBannersPopupText:             "ポップアップ用テキスト",
	VideoBannersText:                  "テキスト",
	VideoBannersLinkText:              "リンクテキスト",
	VideoBannersLink:                  "リンク",

	Heading:                   "ヘディング",
	HeadingsAddTopSpace:       "上方に空白を追加",
	HeadingsAddBottomSpace:    "下方に空白を追加",
	HeadingsAnchorID:          "アンカーID",
	HeadingsHeading:           "ヘディング",
	HeadingsFontColor:         "フォント色",
	HeadingsBackgroundColor:   "背景色",
	HeadingsLink:              "リンク",
	HeadingsLinkText:          "リンクテキスト",
	HeadingsLinkDisplayOption: "リンク表示オプション",
	HeadingsText:              "テキスト",

	BrandGrid:                "ブランドグリッド",
	BrandGridsAddTopSpace:    "上方に空白を追加",
	BrandGridsAddBottomSpace: "下方に空白を追加",
	BrandGridsAnchorID:       "アンカーID",
	BrandGridsBrands:         "ブランド",

	ListContent:                   "リストコンテンツ",
	ListContentsAddTopSpace:       "上方に空白を追加",
	ListContentsAddBottomSpace:    "下方に空白を追加",
	ListContentsAnchorID:          "アンカーID",
	ListContentsBackgroundColor:   "背景色",
	ListContentsItems:             "アイテム",
	ListContentsLink:              "リンク",
	ListContentsLinkText:          "リンクテキスト",
	ListContentsLinkDisplayOption: "リンク表示オプション",

	ImageContainer:                           "画像コンテナ",
	ImageContainersAddTopSpace:               "上方に空白を追加",
	ImageContainersAddBottomSpace:            "ボタン用空白追加",
	ImageContainersAnchorID:                  "アンカーID",
	ImageContainersBackgroundColor:           "背景色",
	ImageContainersTransitionBackgroundColor: "背景色変更",
	ImageContainersImage:                     "画像",
	Image:                                    "画像",

	InNumber:                "数字",
	InNumbersAddTopSpace:    "上方に空白を追加",
	InNumbersAddBottomSpace: "下方に空白を追加",
	InNumbersAnchorID:       "アンカーID",
	InNumbersHeading:        "ヘディング",
	InNumbersItems:          "アイテム",
	InNumbers:               "数字",

	ContactForm:                    "お問合せフォーム",
	ContactFormsAddTopSpace:        "上方に空白を追加",
	ContactFormsAddBottomSpace:     "下方に空白を追加",
	ContactFormsAnchorID:           "アンカーID",
	ContactFormsHeading:            "ヘディング",
	ContactFormsText:               "テキスト",
	ContactFormsSendButtonText:     "送信ボタン用テキスト",
	ContactFormsFormButtonText:     "ウェブフォームボタン用テキスト",
	ContactFormsMessagePlaceholder: "メッセージ",
	ContactFormsNamePlaceholder:    "名前",
	ContactFormsEmailPlaceholder:   "メールアドレス",
	ContactFormsThankyouMessage:    "サンキューメッセージ",
	ContactFormsActionUrl:          "アクションURL",
	ContactFormsPrivacyPolicy:      "プライバシーポリシー",
}

var Messages_ja_JP_I18nLoginKey = &login.Messages{
	Confirm:                             "確認する",
	Verify:                              "検証",
	AccountLabel:                        "メールアドレス",
	AccountPlaceholder:                  "メールアドレス",
	PasswordLabel:                       "パスワード",
	PasswordPlaceholder:                 "パスワード",
	SignInBtn:                           "ログイン",
	ForgetPasswordLink:                  "パスワードをお忘れですか？",
	ForgotMyPasswordTitle:               "パスワードを忘れました",
	ForgetPasswordEmailLabel:            "メールアドレスを入力してください",
	ForgetPasswordEmailPlaceholder:      "メールアドレス",
	SendResetPasswordEmailBtn:           "パスワードリセット用メールが送信されました",
	ResendResetPasswordEmailBtn:         "パスワードリセット用メールを再送する",
	SendEmailTooFrequentlyNotice:        "メール送信回数が上限を超えています。しばらく経ってから再度お試しください",
	ResetPasswordLinkWasSentTo:          "パスワードリセット用リンクが送信されました",
	ResetPasswordLinkSentPrompt:         "このリンクからパスワードリセット手続きを行い、終了後はページを閉じてください",
	ResetYourPasswordTitle:              "パスワードをリセットしてください",
	ResetPasswordLabel:                  "パスワードを変更する",
	ResetPasswordPlaceholder:            "新しいパスワード",
	ResetPasswordConfirmLabel:           "新しいパスワードを再入力",
	ResetPasswordConfirmPlaceholder:     "新しいパスワードを確認する",
	ChangePasswordTitle:                 "パスワードを変更する",
	ChangePasswordOldLabel:              "古いパスワード",
	ChangePasswordOldPlaceholder:        "古いパスワード",
	ChangePasswordNewLabel:              "新しいパスワード",
	ChangePasswordNewPlaceholder:        "新しいパスワード",
	ChangePasswordNewConfirmLabel:       "新しいパスワードを再入力する",
	ChangePasswordNewConfirmPlaceholder: "新しいパスワード",
	TOTPSetupTitle:                      "二段階認証",
	TOTPSetupScanPrompt:                 "Google認証アプリ(または同等アプリ)を利用してこのQRコードをスキャンしてください",
	TOTPSetupSecretPrompt:               "または、お好きな認証アプリを利用して、以下のコードを入力してください",
	TOTPSetupEnterCodePrompt:            "以下のワンタイムコードを入力してください",
	TOTPSetupCodePlaceholder:            "パスコード",
	TOTPValidateTitle:                   "二段階認証",
	TOTPValidateEnterCodePrompt:         "提供されたワンタイムコードを以下に入力してください",
	TOTPValidateCodeLabel:               "認証パスコード",
	TOTPValidateCodePlaceholder:         "パスコード",
	ErrorSystemError:                    "システムエラー",
	ErrorCompleteUserAuthFailed:         "ユーザー認証に失敗しました",
	ErrorUserNotFound:                   "このユーザーは存在しません",
	ErrorIncorrectAccountNameOrPassword: "メールアドレスまたはパスワードが間違っています",
	ErrorUserLocked:                     "ユーザーがロックされました",
	ErrorAccountIsRequired:              "メールアドレスは必須です",
	ErrorPasswordCannotBeEmpty:          "パスワードは必須です",
	ErrorPasswordNotMatch:               "パスワードが間違っています",
	ErrorIncorrectPassword:              "古いパスワードが間違っています",
	ErrorInvalidToken:                   "このトークンは無効です",
	ErrorTokenExpired:                   "トークンの有効期限が切れています",
	ErrorIncorrectTOTPCode:              "パスコードが間違っています",
	ErrorTOTPCodeReused:                 "このパスコードは既に利用されています",
	ErrorIncorrectRecaptchaToken:        "reCAPTCHAトークンが間違っています",
	WarnPasswordHasBeenChanged:          "パスワードが変更されました。再度ログインしてください",
	InfoPasswordSuccessfullyReset:       "パスワードのリセットに成功しました。再度ログインしてください",
	InfoPasswordSuccessfullyChanged:     "パスワードの変更に成功しました。再度ログインしてください",
}

var Messages_ja_JP_I18nUtilsKey = &utils.Messages{
	OK:     "OK",
	Cancel: "キャンセル",
}

var Messages_ja_JP_I18nPublishKey = &publish_view.Messages{
	StatusDraft:             "下書き",
	StatusOnline:            "公開中",
	StatusOffline:           "非公開中",
	Publish:                 "公開する",
	Unpublish:               "非公開",
	Republish:               "再公開",
	Areyousure:              "よろしいですか？",
	ScheduledStartAt:        "公開開始日時",
	ScheduledEndAt:          "公開終了日時",
	PublishedAt:             "開始日時",
	UnPublishedAt:           "公開終了日時",
	ActualPublishTime:       "投稿日時",
	SchedulePublishTime:     "公開日時を設定する",
	NotSet:                  "未セット",
	WhenDoYouWantToPublish:  "公開日時を設定してください",
	PublishScheduleTip:      "{SchedulePublishTime} 設定後、システムが自動で当該記事の公開・非公開を行います。",
	DateTimePickerClearText: "クリア",
	DateTimePickerOkText:    "OK",
	SaveAsNewVersion:        "新規バージョンとして保存する",
	SwitchedToNewVersion:    "新規バージョンに変更する",
	SuccessfullyCreated:     "作成に成功しました",
	SuccessfullyRename:      "名付けに成功しました",
	OnlineVersion:           "オンラインバージョン",
	VersionsList:            "バージョンリスト",
	AllVersions:             "全てのバージョン",
	NamedVersions:           "名付け済みバージョン",
	RenameVersion:           "バージョンの名前を変更する",
}

var Messages_ja_JP_CoreI18nModuleKey = &presets.Messages{
	SuccessfullyUpdated:            "更新に成功しました",
	Search:                         "検索する",
	New:                            "新規",
	Update:                         "更新する",
	Delete:                         "削除する",
	Edit:                           "編集する",
	FormTitle:                      "フォーム",
	OK:                             "OK",
	Cancel:                         "キャンセル",
	Create:                         "新規作成",
	DeleteConfirmationTextTemplate: ": {id}を削除して本当によろしいですか？",
	CreatingObjectTitleTemplate:    "{modelName} を作る",
	EditingObjectTitleTemplate:     "{modelName} {id} を編集する",
	ListingObjectTitleTemplate:     "リスティング {modelName} ",
	DetailingObjectTitleTemplate:   "{modelName} {id} ",
	FiltersClear:                   "フィルターをクリアする",
	FiltersAdd:                     "フィルターを追加する",
	FilterApply:                    "適用する",
	FilterByTemplate:               "{filter} でフィルターする",
	FiltersDateInTheLast:           "がサイト",
	FiltersDateEquals:              "と同等",
	FiltersDateBetween:             "の間",
	FiltersDateIsAfter:             "が後",
	FiltersDateIsAfterOrOn:         "が同時または後",
	FiltersDateIsBefore:            "が前",
	FiltersDateIsBeforeOrOn:        "が前または同時",
	FiltersDateDays:                "日間",
	FiltersDateMonths:              "月数",
	FiltersDateAnd:                 "＆",
	FiltersDateTo:                  "から",
	FiltersNumberEquals:            "と同等",
	FiltersNumberBetween:           "間",
	FiltersNumberGreaterThan:       "より大きい",
	FiltersNumberLessThan:          "より少ない",
	FiltersNumberAnd:               "＆",
	FiltersStringEquals:            "と同等",
	FiltersStringContains:          "を含む",
	FiltersMultipleSelectIn:        "中",
	FiltersMultipleSelectNotIn:     "以外",
	PaginationRowsPerPage:          "行 / ページ",
	ListingNoRecordToShow:          "表示できるデータはありません",
	ListingSelectedCountNotice:     "{count} レコードが選択されています",
	ListingClearSelection:          "選択をクリア",
	BulkActionNoAvailableRecords:   "この機能はご利用いただけません",
	BulkActionSelectedIdsProcessNoticeTemplate: "この一部の機能はご利用いただけません: {ids}",
	ConfirmationDialogText:                     "よろしいですか？",
	Language:                                   "言語",
	Colon:                                      ":",
}

var Messages_ja_JP_I18nNoteKey = &note.Messages{
	SuccessfullyCreated: "作成に成功しました",
	Item:                "アイテム",
	Notes:               "ノート",
	NewNote:             "新規ノート",
}

var Messages_ja_JP_I10nLocalizeKey = &l10n_view.Messages{
	Localize:              "ローカライズ",
	LocalizeFrom:          "から",
	LocalizeTo:            "に",
	SuccessfullyLocalized: "ローカライズに成功しました",
	Location:              "場所",
	Colon:                 ":",
	International:         "インターナショナル",
	China:                 "中国",
	Japan:                 "日本",
}

var Messages_ja_JP_I18nPageBuilderKey = &pagebuilder.Messages{
	Category:           "カテゴリー",
	EditPageContent:    "ページコンテナを編集する",
	Preview:            "プレビュー",
	Containers:         "コンテナ",
	AddContainers:      "コンテナを追加する",
	New:                "新規",
	Shared:             "共有",
	Select:             "選択する",
	TemplateID:         "テンプレートID",
	TemplateName:       "テンプレート名",
	CreateFromTemplate: "テンプレートから新規作成する",
}
