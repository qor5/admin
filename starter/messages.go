package starter

import (
	"github.com/qor5/x/v3/i18n"
)

const I18nDemoKey i18n.ModuleKey = "I18nDemoKey"

type Messages struct {
	FilterTabsAll                  string
	FilterTabsHasUnreadNotes       string
	FilterTabsActive               string
	DemoTips                       string
	DemoUsernameLabel              string
	DemoPasswordLabel              string
	LoginProviderGoogleText        string
	LoginProviderMicrosoftText     string
	LoginProviderGithubText        string
	OAuthCompleteInfoTitle         string
	OAuthCompleteInfoPositionLabel string
	OAuthCompleteInfoAgreeLabel    string
	OAuthCompleteInfoBackLabel     string
	Demo                           string
	DBResetTipLabel                string
	Name                           string
	Email                          string
	Company                        string
	Role                           string
	Status                         string
	ChangePassword                 string
	LoginSessions                  string
	SystemTitleLabel               string
}

var Messages_en_US = &Messages{
	FilterTabsAll:                  "All",
	FilterTabsHasUnreadNotes:       "Has Unread Notes",
	FilterTabsActive:               "Active",
	DemoTips:                       "Please note that the database would be reset every even hour.",
	DemoUsernameLabel:              "Demo Username: ",
	DemoPasswordLabel:              "Demo Password: ",
	LoginProviderGoogleText:        "Login with Google",
	LoginProviderMicrosoftText:     "Login with Microsoft",
	LoginProviderGithubText:        "Login with Github",
	OAuthCompleteInfoTitle:         "Complete your information",
	OAuthCompleteInfoPositionLabel: "Position(Optional)",
	OAuthCompleteInfoAgreeLabel:    "Subscribe to QOR5 newsletter(Optional)",
	OAuthCompleteInfoBackLabel:     "Back to login",
	Demo:                           "DEMO",
	DBResetTipLabel:                "Database reset countdown",
	Name:                           "Name",
	Email:                          "Email",
	Company:                        "Company",
	Role:                           "Role",
	Status:                         "Status",
	ChangePassword:                 "Change Password",
	LoginSessions:                  "Login Sessions",
	SystemTitleLabel:               "Adex System",
}

var Messages_ja_JP = &Messages{
	FilterTabsAll:                  "すべて",
	FilterTabsHasUnreadNotes:       "未読のノートがあります",
	FilterTabsActive:               "有効",
	DemoTips:                       "データベースは偶数時間ごとにリセットされることに注意してください。",
	DemoUsernameLabel:              "デモのユーザー名: ",
	DemoPasswordLabel:              "デモパスワード: ",
	LoginProviderGoogleText:        "Googleでログイン",
	LoginProviderMicrosoftText:     "Microsoftでログイン",
	LoginProviderGithubText:        "Githubでログイン",
	OAuthCompleteInfoTitle:         "情報を入力してください",
	OAuthCompleteInfoPositionLabel: "役職（任意）",
	OAuthCompleteInfoAgreeLabel:    "QOR5ニュースレターを購読する（任意）",
	OAuthCompleteInfoBackLabel:     "ログインに戻る",
	Demo:                           "デモ",
	DBResetTipLabel:                "データベースリセットのカウントダウン",
	Name:                           "名前",
	Email:                          "メール",
	Company:                        "会社",
	Role:                           "役割",
	Status:                         "ステータス",
	ChangePassword:                 "パスワードを変更する",
	LoginSessions:                  "ログインセッション",
	SystemTitleLabel:               "Adex システム",
}

var Messages_zh_CN = &Messages{
	FilterTabsAll:                  "全部",
	FilterTabsHasUnreadNotes:       "未读备注",
	FilterTabsActive:               "有效",
	DemoTips:                       "请注意，数据库将每隔偶数小时重置一次。",
	DemoUsernameLabel:              "演示账户：",
	DemoPasswordLabel:              "演示密码：",
	LoginProviderGoogleText:        "使用Google登录",
	LoginProviderMicrosoftText:     "使用Microsoft登录",
	LoginProviderGithubText:        "使用Github登录",
	OAuthCompleteInfoTitle:         "请填写您的信息",
	OAuthCompleteInfoPositionLabel: "职位（可选）",
	OAuthCompleteInfoAgreeLabel:    "订阅QOR5新闻（可选）",
	OAuthCompleteInfoBackLabel:     "返回登录",
	Demo:                           "演示",
	DBResetTipLabel:                "数据库重置倒计时",
	Name:                           "姓名",
	Email:                          "邮箱",
	Company:                        "公司",
	Role:                           "角色",
	Status:                         "状态",
	ChangePassword:                 "修改密码",
	LoginSessions:                  "登录会话",
	SystemTitleLabel:               "Adex 系统",
}

type Messages_ModelsI18nModuleKey struct {
	Admin       string
	QOR5Example string
	Roles       string
	Users       string

	// TODO: @iBakuman 有很多是不需要的
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
	PageCategories           string
	ECManagement             string
	ECDashboard              string
	Orders                   string
	InputDemos               string
	Products                 string
	NestedFieldDemos         string
	SiteManagement           string
	SEO                      string
	UserManagement           string
	Profile                  string
	FeaturedModelsManagement string
	Customers                string
	ListModels               string
	MicrositeModels          string
	Workers                  string
	MediaLibrary             string

	PagesID         string
	PagesTitle      string
	PagesSlug       string
	PagesLocale     string
	PagesNotes      string
	PagesDraftCount string
	PagesPath       string
	PagesOnline     string
	PagesVersion    string
	PagesVersions   string
	PagesStartAt    string
	PagesEndAt      string
	PagesOption     string
	PagesLive       string

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

	ActivityActionLogIn         string
	ActivityActionExtendSession string

	PagesPage string
}

var Messages_en_US_ModelsI18nModuleKey = &Messages_ModelsI18nModuleKey{
	Posts:          "Posts",
	PostsID:        "ID",
	PostsTitle:     "Title",
	PostsHeroImage: "Hero Image",
	PostsBody:      "Content",
	Example:        "QOR5 Demo",
	Settings:       "SEO Settings",
	Post:           "Post",
	PostsBodyImage: "Content Image",

	SeoPost:             "Post",
	SeoVariableTitle:    "Title",
	SeoVariableSiteName: "Site Name",

	Admin:       "Admin",
	QOR5Example: "QOR5 Example",
	Roles:       "Role Management",
	Users:       "User Management",

	PageBuilder:              "Page Builder",
	Pages:                    "Pages",
	SharedContainers:         "Shared Containers",
	DemoContainers:           "Demo Containers",
	Templates:                "Templates",
	PageCategories:           "Page Categories",
	ECManagement:             "E-commerce Management",
	ECDashboard:              "E-commerce Dashboard",
	Orders:                   "Orders",
	InputDemos:               "Input Demos",
	Products:                 "Products",
	NestedFieldDemos:         "Nested Field Demos",
	SiteManagement:           "Site Management",
	SEO:                      "SEO",
	UserManagement:           "User Management",
	Profile:                  "Profile",
	FeaturedModelsManagement: "Featured Models Management",
	Customers:                "Customers",
	ListModels:               "List Models",
	MicrositeModels:          "Microsite Models",
	Workers:                  "Workers",
	MediaLibrary:             "Media Library",

	PagesID:         "ID",
	PagesTitle:      "Title",
	PagesSlug:       "Slug",
	PagesLocale:     "Locale",
	PagesNotes:      "Notes",
	PagesDraftCount: "Draft Count",
	PagesPath:       "Path",
	PagesOnline:     "Online",
	PagesVersion:    "Version",
	PagesVersions:   "Versions",
	PagesStartAt:    "Start At",
	PagesEndAt:      "End At",
	PagesOption:     "Option",
	PagesLive:       "Live Status",

	Page:                   "Page",
	PagesStatus:            "Status",
	PagesSchedule:          "Schedule",
	PagesCategoryID:        "Category ID",
	PagesTemplateSelection: "Template Selection",
	PagesEditContainer:     "Edit Container",

	WebHeader:       "Web Header",
	WebHeadersColor: "Color",
	Header:          "Header",

	WebFooter:             "Web Footer",
	WebFootersEnglishUrl:  "English URL",
	WebFootersJapaneseUrl: "Japanese URL",
	Footer:                "Footer",

	VideoBanner:                       "Video Banner",
	VideoBannersAddTopSpace:           "Add Top Space",
	VideoBannersAddBottomSpace:        "Add Bottom Space",
	VideoBannersAnchorID:              "Anchor ID",
	VideoBannersVideo:                 "Video",
	VideoBannersBackgroundVideo:       "Background Video",
	VideoBannersMobileBackgroundVideo: "Mobile Background Video",
	VideoBannersVideoCover:            "Video Cover",
	VideoBannersMobileVideoCover:      "Mobile Video Cover",
	VideoBannersHeading:               "Heading",
	VideoBannersPopupText:             "Popup Text",
	VideoBannersText:                  "Text",
	VideoBannersLinkText:              "Link Text",
	VideoBannersLink:                  "Link",

	Heading:                   "Heading",
	HeadingsAddTopSpace:       "Add Top Space",
	HeadingsAddBottomSpace:    "Add Bottom Space",
	HeadingsAnchorID:          "Anchor ID",
	HeadingsHeading:           "Heading",
	HeadingsFontColor:         "Font Color",
	HeadingsBackgroundColor:   "Background Color",
	HeadingsLink:              "Link",
	HeadingsLinkText:          "Link Text",
	HeadingsLinkDisplayOption: "Link Display Option",
	HeadingsText:              "Text",

	BrandGrid:                "Brand Grid",
	BrandGridsAddTopSpace:    "Add Top Space",
	BrandGridsAddBottomSpace: "Add Bottom Space",
	BrandGridsAnchorID:       "Anchor ID",
	BrandGridsBrands:         "Brands",

	ListContent:                   "List Content",
	ListContentsAddTopSpace:       "Add Top Space",
	ListContentsAddBottomSpace:    "Add Bottom Space",
	ListContentsAnchorID:          "Anchor ID",
	ListContentsBackgroundColor:   "Background Color",
	ListContentsItems:             "Items",
	ListContentsLink:              "Link",
	ListContentsLinkText:          "Link Text",
	ListContentsLinkDisplayOption: "Link Display Option",

	ImageContainer:                           "Image Container",
	ImageContainersAddTopSpace:               "Add Top Space",
	ImageContainersAddBottomSpace:            "Add Bottom Space",
	ImageContainersAnchorID:                  "Anchor ID",
	ImageContainersBackgroundColor:           "Background Color",
	ImageContainersTransitionBackgroundColor: "Transition Background Color",
	ImageContainersImage:                     "Image",
	Image:                                    "Image",

	InNumber:                "In Number",
	InNumbersAddTopSpace:    "Add Top Space",
	InNumbersAddBottomSpace: "Add Bottom Space",
	InNumbersAnchorID:       "Anchor ID",
	InNumbersHeading:        "Heading",
	InNumbersItems:          "Items",
	InNumbers:               "In Numbers",

	ContactForm:                    "Contact Form",
	ContactFormsAddTopSpace:        "Add Top Space",
	ContactFormsAddBottomSpace:     "Add Bottom Space",
	ContactFormsAnchorID:           "Anchor ID",
	ContactFormsHeading:            "Heading",
	ContactFormsText:               "Text",
	ContactFormsSendButtonText:     "Send Button Text",
	ContactFormsFormButtonText:     "Form Button Text",
	ContactFormsMessagePlaceholder: "Message Placeholder",
	ContactFormsNamePlaceholder:    "Name Placeholder",
	ContactFormsEmailPlaceholder:   "Email Placeholder",
	ContactFormsThankyouMessage:    "Thank You Message",
	ContactFormsActionUrl:          "Action URL",
	ContactFormsPrivacyPolicy:      "Privacy Policy",

	ActivityActionLogIn:         "Log In",
	ActivityActionExtendSession: "Extend Session",

	PagesPage: "Page",
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

	Admin:       "管理员",
	QOR5Example: "QOR5 示例",
	Roles:       "权限管理",
	Users:       "用户管理",

	PageBuilder:              "页面管理菜单",
	Pages:                    "页面管理",
	SharedContainers:         "公用组件",
	DemoContainers:           "示例组件",
	Templates:                "模板页面",
	PageCategories:           "目录管理",
	ECManagement:             "电子商务管理",
	ECDashboard:              "电子商务仪表盘",
	Orders:                   "订单管理",
	InputDemos:               "表单 示例",
	Products:                 "产品管理",
	NestedFieldDemos:         "嵌套表单 示例",
	SiteManagement:           "站点管理菜单",
	SEO:                      "SEO 管理",
	UserManagement:           "用户管理菜单",
	Profile:                  "个人页面",
	FeaturedModelsManagement: "特色模块管理菜单",
	Customers:                "Customers 示例",
	ListModels:               "发布带排序及分页模块 示例",
	MicrositeModels:          "Microsite 示例",
	Workers:                  "后台工作进程管理",
	MediaLibrary:             "媒体库",

	PagesID:         "ID",
	PagesTitle:      "标题",
	PagesSlug:       "Slug",
	PagesLocale:     "地区",
	PagesNotes:      "备注",
	PagesDraftCount: "草稿数",
	PagesPath:       "路径",
	PagesOnline:     "在线",
	PagesVersion:    "版本",
	PagesVersions:   "版本",
	PagesStartAt:    "开始时间",
	PagesEndAt:      "结束时间",
	PagesOption:     "选项",
	PagesLive:       "发布状态",

	Page:                   "Page",
	PagesStatus:            "状态",
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

	ActivityActionLogIn:         "登录",
	ActivityActionExtendSession: "延长会话",

	PagesPage: "Page",
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

	Admin:       "管理員",
	QOR5Example: "QOR5サンプル",
	Roles:       "ユーザー権限",
	Users:       "ユーザー",

	PageBuilder:              "ページビルダー",
	Pages:                    "ページ",
	SharedContainers:         "共有コンテナ",
	DemoContainers:           "デモ用コン店た",
	Templates:                "テンプレート",
	PageCategories:           "カテゴリー",
	ECManagement:             "ECマネジメント",
	ECDashboard:              "ECダッシュボード",
	Orders:                   "注文",
	InputDemos:               "入力デモ",
	Products:                 "製品",
	SiteManagement:           "サイト管理",
	NestedFieldDemos:         "ネストフィールドデモ",
	SEO:                      "SEO",
	UserManagement:           "ユーザー管理",
	Profile:                  "プロフィール",
	FeaturedModelsManagement: "モデル管理",
	Customers:                "お客さま",
	ListModels:               "リストモデル",
	MicrositeModels:          "マイクロサイトモデル",
	Workers:                  "ワーカーズ",
	MediaLibrary:             "メディアライブラリ",

	PagesID:         "ID",
	PagesTitle:      "タイトル",
	PagesSlug:       "スラッグ",
	PagesLocale:     "ローカル",
	PagesNotes:      "ノート",
	PagesDraftCount: "カウント下書き",
	PagesPath:       "パス",
	PagesOnline:     "オンライン",
	PagesVersion:    "バージョン",
	PagesVersions:   "バージョン",
	PagesStartAt:    "開始日時",
	PagesEndAt:      "終了日時",
	PagesOption:     "オプション",
	PagesLive:       "ライブ",

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

	ActivityActionLogIn:         "ログイン",
	ActivityActionExtendSession: "セッション延長",

	PagesPage: "ページ",
}
