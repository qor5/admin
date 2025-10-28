package publish

import "strings"

type Messages struct {
	StatusDraft                             string
	StatusOnline                            string
	StatusOffline                           string
	StatusNext                              string
	Publish                                 string
	Unpublish                               string
	Republish                               string
	ConfirmPublish                          string
	ConfirmUnpublish                        string
	ConfirmRepublish                        string
	ConfirmDuplicate                        string
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
	NextVersion                             string
	VersionsList                            string
	AllVersions                             string
	NamedVersions                           string
	RenameVersion                           string
	DeleteVersionConfirmationTextTemplate   string
	ToStatusOnlineTemplate                  string
	ToStatusOfflineTemplate                 string

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

func (msgr *Messages) ToStatusOnline(versionName, scheduleTime string) string {
	return strings.NewReplacer(
		"{VersionName}", versionName,
		"{ScheduleTime}", scheduleTime,
	).Replace(msgr.ToStatusOnlineTemplate)
}

func (msgr *Messages) ToStatusOffline(versionName, scheduleTime string) string {
	return strings.NewReplacer(
		"{VersionName}", versionName,
		"{ScheduleTime}", scheduleTime,
	).Replace(msgr.ToStatusOfflineTemplate)
}

var Messages_en_US = &Messages{
	StatusDraft:                             "Draft",
	StatusOnline:                            "Online",
	StatusOffline:                           "Offline",
	StatusNext:                              "Next",
	Publish:                                 "Publish",
	Unpublish:                               "Unpublish",
	Republish:                               "Republish",
	ConfirmPublish:                          "Are you sure you want to publish this page?",
	ConfirmUnpublish:                        "Are you sure you want to unpublish this page?",
	ConfirmRepublish:                        "Are you sure you want to republish this page?",
	ConfirmDuplicate:                        "Are you sure you want to duplicate this page?",
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
	NextVersion:                             "Next Version",
	VersionsList:                            "Versions List",
	AllVersions:                             "All versions",
	NamedVersions:                           "Named versions",
	RenameVersion:                           "Rename Version",
	DeleteVersionConfirmationTextTemplate:   "Are you sure you want to delete version {VersionName} ?",
	ToStatusOnlineTemplate:                  "{VersionName} will be online at {ScheduleTime}",
	ToStatusOfflineTemplate:                 "{VersionName} will be offline at {ScheduleTime}",

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
	StatusNext:                              "下一个",
	Publish:                                 "发布",
	Unpublish:                               "取消发布",
	Republish:                               "重新发布",
	ConfirmPublish:                          "你确定要发布此页面吗?",
	ConfirmUnpublish:                        "你确定要取消发布此页面吗?",
	ConfirmRepublish:                        "你确定要重新发布此页面吗?",
	ConfirmDuplicate:                        "你确定要复制此页面吗?",
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
	NextVersion:                             "下个版本",
	VersionsList:                            "版本列表",
	AllVersions:                             "所有版本",
	NamedVersions:                           "已命名版本",
	RenameVersion:                           "命名版本",
	DeleteVersionConfirmationTextTemplate:   "你确定你要删除此版本 {VersionName} 吗？",
	ToStatusOnlineTemplate:                  "{VersionName} 将在 {ScheduleTime} 上线",
	ToStatusOfflineTemplate:                 "{VersionName} 将在 {ScheduleTime} 下线",

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
	StatusOnline:                            "オンライン",
	StatusOffline:                           "オフライン",
	StatusNext:                              "次",
	Publish:                                 "公開する",
	Unpublish:                               "非公開",
	Republish:                               "再公開",
	ConfirmPublish:                          "このページを公開してもよろしいですか？",
	ConfirmUnpublish:                        "このページを非公開にしてもよろしいですか？",
	ConfirmRepublish:                        "このページを再公開してもよろしいですか？",
	ConfirmDuplicate:                        "このページを複製してもよろしいですか？",
	ScheduledStartAt:                        "開始時刻",
	ScheduledEndAt:                          "終了時刻",
	ScheduledStartAtShouldLaterThanNow:      "開始時刻は現在より遅くなければならない",
	ScheduledEndAtShouldLaterThanNowOrEmpty: "終了時刻は現在よりも遅いか、または空でなければならない",
	ScheduledEndAtShouldLaterThanStartAt:    "終了時刻は開始時刻より遅くする",
	ScheduledStartAtShouldNotEmpty:          "開始時刻は空であってはならない",
	PublishedAt:                             "開始時間",
	UnPublishedAt:                           "公開終了日時",
	ActualPublishTime:                       "実際の公開時間",
	SchedulePublishTime:                     "公開時間設定",
	NotSet:                                  "未設定",
	WhenDoYouWantToPublish:                  "公開日時を設定してください",
	PublishScheduleTip:                      "{SchedulePublishTime}を設定すると、システムが自動的に公開/非公開を行います。",
	DateTimePickerClearText:                 "消去する",
	DateTimePickerOkText:                    "OK",
	SaveAsNewVersion:                        "新しいバージョンとして保存",
	SwitchedToNewVersion:                    "新バージョンに変更しました",
	SuccessfullyCreated:                     "作成に成功しました",
	SuccessfullyRename:                      "名前の変更に成功しました",
	SuccessfullyPublish:                     "公開に成功しました",
	SuccessfullyUnPublish:                   "非公開に成功しました",
	OnlineVersion:                           "オンラインバージョン",
	NextVersion:                             "次のバージョン",
	VersionsList:                            "バージョンリスト",
	AllVersions:                             "すべてのバージョン",
	NamedVersions:                           "指定バージョン",
	RenameVersion:                           "バージョン名の変更",
	DeleteVersionConfirmationTextTemplate:   "本当にバージョン{VersionName}を削除しますか？",
	ToStatusOnlineTemplate:                  "{VersionName} は {ScheduleTime} にオンラインになります",
	ToStatusOfflineTemplate:                 "{VersionName} は {ScheduleTime} にオフラインになります",

	FilterTabAllVersions:   "すべてのバージョン",
	FilterTabOnlineVersion: "オンラインバージョン",
	FilterTabNamedVersions: "名前付きバージョン",
	Rename:                 "名称変更",
	PageOverView:           "ページ概要",
	Duplicate:              "コピー",

	HeaderVersion: "バージョン",
	HeaderStatus:  "ステータス",
	HeaderStartAt: "開始日",
	HeaderEndAt:   "終了時",
	HeaderOption:  "オプション",

	HeaderDraftCount: "下書き数",
	HeaderLive:       "公開ステータス",
}
