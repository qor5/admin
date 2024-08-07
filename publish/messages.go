package publish

import "strings"

type Messages struct {
	StatusDraft                             string
	StatusOnline                            string
	StatusOffline                           string
	Publish                                 string
	Unpublish                               string
	Republish                               string
	Areyousure                              string
	ScheduledStartAt                        string
	ScheduledEndAt                          string
	ScheduledStartAtShouldLaterThanNow      string
	ScheduledEndAtShouldLaterThanNowOrEmpty string
	ScheduledEndAtShouldLaterThanStartAt    string
	ScheduledStartAtShouldNotEmpty          string
	PublishedAt                             string
	UnPublishedAt                           string
	ActualPublishTime                       string
	SchedulePublishTime                     string
	NotSet                                  string
	WhenDoYouWantToPublish                  string
	PublishScheduleTip                      string
	DateTimePickerClearText                 string
	DateTimePickerOkText                    string
	SaveAsNewVersion                        string
	SwitchedToNewVersion                    string
	SuccessfullyCreated                     string
	SuccessfullyRename                      string
	SuccessfullyPublish                     string
	SuccessfullyUnPublish                   string
	OnlineVersion                           string
	VersionsList                            string
	AllVersions                             string
	NamedVersions                           string
	RenameVersion                           string
	DeleteVersionConfirmationTextTemplate   string

	FilterTabAllVersions   string
	FilterTabOnlineVersion string
	FilterTabNamedVersions string
	Rename                 string
	PageOverView           string
	Duplicate              string

	HeaderVersion string
	HeaderStatus  string
	HeaderStartAt string
	HeaderEndAt   string
	HeaderOption  string

	HeaderDraftCount string
	HeaderLive       string
}

func (msgr *Messages) DeleteVersionConfirmationText(versionName string) string {
	return strings.NewReplacer("{VersionName}", versionName).
		Replace(msgr.DeleteVersionConfirmationTextTemplate)
}

var Messages_en_US = &Messages{
	StatusDraft:                             "Draft",
	StatusOnline:                            "Online",
	StatusOffline:                           "Offline",
	Publish:                                 "Publish",
	Unpublish:                               "Unpublish",
	Republish:                               "Republish",
	Areyousure:                              "Are you sure?",
	ScheduledStartAt:                        "Start at",
	ScheduledEndAt:                          "End at",
	ScheduledStartAtShouldLaterThanNow:      "Start at should be later than now",
	ScheduledEndAtShouldLaterThanNowOrEmpty: "End at should be later than now or empty",
	ScheduledEndAtShouldLaterThanStartAt:    "End at should be later than start at",
	ScheduledStartAtShouldNotEmpty:          "Start at should not be empty",
	PublishedAt:                             "Start at",
	UnPublishedAt:                           "End at",
	ActualPublishTime:                       "Actual Publish Time",
	SchedulePublishTime:                     "Schedule Publish Time",
	NotSet:                                  "Not set",
	WhenDoYouWantToPublish:                  "When do you want to publish?",
	PublishScheduleTip:                      "After you set the {SchedulePublishTime}, the system will automatically publish/unpublish it.",
	DateTimePickerClearText:                 "Clear",
	DateTimePickerOkText:                    "OK",
	SaveAsNewVersion:                        "Save As New Version",
	SwitchedToNewVersion:                    "Switched To New Version",
	SuccessfullyCreated:                     "Successfully Created",
	SuccessfullyRename:                      "Successfully Rename",
	SuccessfullyPublish:                     "Successfully Publish",
	SuccessfullyUnPublish:                   "Successfully Unpublish",
	OnlineVersion:                           "Online Version",
	VersionsList:                            "Versions List",
	AllVersions:                             "All versions",
	NamedVersions:                           "Named versions",
	RenameVersion:                           "Rename Version",
	DeleteVersionConfirmationTextTemplate:   "Are you sure you want to delete version {VersionName} ?",

	FilterTabAllVersions:   "All Versions",
	FilterTabOnlineVersion: "Online Versions",
	FilterTabNamedVersions: "Named Versions",
	Rename:                 "Rename",
	PageOverView:           "Page Overview",
	Duplicate:              "Duplicate",

	HeaderVersion: "Version",
	HeaderStatus:  "Status",
	HeaderStartAt: "Start At",
	HeaderEndAt:   "End At",
	HeaderOption:  "Option",

	HeaderDraftCount: "Draft Count",
	HeaderLive:       "Live",
}

var Messages_zh_CN = &Messages{
	StatusDraft:                             "草稿",
	StatusOnline:                            "在线",
	StatusOffline:                           "离线",
	Publish:                                 "发布",
	Unpublish:                               "取消发布",
	Republish:                               "重新发布",
	Areyousure:                              "你确定吗?",
	ScheduledStartAt:                        "发布时间",
	ScheduledEndAt:                          "下线时间",
	ScheduledStartAtShouldLaterThanNow:      "计划发布时间应当晚于现在时间",
	ScheduledEndAtShouldLaterThanNowOrEmpty: "计划下线时间应当晚于现在时间",
	ScheduledEndAtShouldLaterThanStartAt:    "计划下线时间应当晚于计划发布时间",
	ScheduledStartAtShouldNotEmpty:          "计划发布时间不得为空",
	PublishedAt:                             "发布时间",
	UnPublishedAt:                           "下线时间",
	ActualPublishTime:                       "实际发布时间",
	SchedulePublishTime:                     "计划发布时间",
	NotSet:                                  "未设定",
	WhenDoYouWantToPublish:                  "你希望什么时候发布？",
	PublishScheduleTip:                      "设定好 {SchedulePublishTime} 之后, 系统会按照时间自动将它发布/下线。",
	DateTimePickerClearText:                 "清空",
	DateTimePickerOkText:                    "确定",
	SaveAsNewVersion:                        "保存为一个新版本",
	SwitchedToNewVersion:                    "切换到新版本",
	SuccessfullyCreated:                     "成功创建",
	SuccessfullyRename:                      "成功命名",
	SuccessfullyPublish:                     "成功发布",
	SuccessfullyUnPublish:                   "已取消发布",
	OnlineVersion:                           "在线版本",
	VersionsList:                            "版本列表",
	AllVersions:                             "所有版本",
	NamedVersions:                           "已命名版本",
	RenameVersion:                           "命名版本",
	DeleteVersionConfirmationTextTemplate:   "你确定你要删除此版本 {VersionName} 吗？",

	FilterTabAllVersions:   "所有版本",
	FilterTabOnlineVersion: "在线版本",
	FilterTabNamedVersions: "已命名版本",
	Rename:                 "重命名",
	PageOverView:           "页面概览",
	Duplicate:              "复制",

	HeaderVersion: "版本",
	HeaderStatus:  "状态",
	HeaderStartAt: "开始时间",
	HeaderEndAt:   "结束时间",
	HeaderOption:  "操作",

	HeaderDraftCount: "草稿数",
	HeaderLive:       "发布状态",
}

var Messages_ja_JP = &Messages{
	StatusDraft:                             "下書き",
	StatusOnline:                            "公開中",
	StatusOffline:                           "非公開中",
	Publish:                                 "公開する",
	Unpublish:                               "非公開",
	Republish:                               "再公開",
	Areyousure:                              "よろしいですか？",
	ScheduledStartAt:                        "公開開始日時",
	ScheduledEndAt:                          "公開終了日時",
	ScheduledStartAtShouldLaterThanNow:      "予定開始時間は現在時刻よりも後でなければなりません",
	ScheduledEndAtShouldLaterThanNowOrEmpty: "予定終了時間は現在時刻よりも後、または空でなければなりません",
	ScheduledEndAtShouldLaterThanStartAt:    "予定終了時間は予定開始時間よりも後でなければなりません",
	ScheduledStartAtShouldNotEmpty:          "予定開始時間は空ではない必要があります",
	PublishedAt:                             "開始日時",
	UnPublishedAt:                           "公開終了日時",
	ActualPublishTime:                       "投稿日時",
	SchedulePublishTime:                     "公開日時を設定する",
	NotSet:                                  "未セット",
	WhenDoYouWantToPublish:                  "公開日時を設定してください",
	PublishScheduleTip:                      "{SchedulePublishTime} 設定後、システムが自動で当該記事の公開・非公開を行います。",
	DateTimePickerClearText:                 "クリア",
	DateTimePickerOkText:                    "OK",
	SaveAsNewVersion:                        "新規バージョンとして保存する",
	SwitchedToNewVersion:                    "新規バージョンに変更する",
	SuccessfullyCreated:                     "作成に成功しました",
	SuccessfullyRename:                      "名付けに成功しました",
	SuccessfullyPublish:                     "公開に成功しました",
	SuccessfullyUnPublish:                   "非公開に成功しました",
	OnlineVersion:                           "オンラインバージョン",
	VersionsList:                            "バージョンリスト",
	AllVersions:                             "全てのバージョン",
	NamedVersions:                           "名付け済みバージョン",
	RenameVersion:                           "バージョンの名前を変更する",
	DeleteVersionConfirmationTextTemplate:   "このバージョン {VersionName} を削除してもよろしいですか？",
	Duplicate:                               "Duplicate(need JP translation)",

	FilterTabAllVersions:   "全てのバージョン",
	FilterTabOnlineVersion: "オンラインバージョン",
	FilterTabNamedVersions: "名付け済みバージョン",
	Rename:                 "名前の変更",
	PageOverView:           "ページ概要",

	HeaderVersion: "バージョン",
	HeaderStatus:  "ステータス",
	HeaderStartAt: "開始日時",
	HeaderEndAt:   "終了日時",
	HeaderOption:  "オプション",

	HeaderDraftCount: "下書き数",
	HeaderLive:       "更新中",
}
