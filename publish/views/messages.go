package views

import (
	"github.com/qor/qor5/publish"
)

type Messages struct {
	StatusDraft             string
	StatusOnline            string
	StatusOffline           string
	Publish                 string
	Unpublish               string
	Republish               string
	Areyousure              string
	ScheduledStartAt        string
	ScheduledEndAt          string
	PublishedAt             string
	UnPublishedAt           string
	ActualPublishTime       string
	SchedulePublishTime     string
	NotSet                  string
	WhenDoYouWantToPublish  string
	PublishScheduleTip      string
	DateTimePickerClearText string
	DateTimePickerOkText    string
	SaveAsNewVersion        string
	SwitchedToNewVersion    string
	SuccessfullyCreated     string
	SuccessfullyRename      string
	OnlineVersion           string
	VersionsList            string
	AllVersions             string
	NamedVersions           string
	RenameVersion           string
}

var Messages_en_US = &Messages{
	StatusDraft:             "Draft",
	StatusOnline:            "Online",
	StatusOffline:           "Offline",
	Publish:                 "Publish",
	Unpublish:               "Unpublish",
	Republish:               "Republish",
	Areyousure:              "Are you sure?",
	ScheduledStartAt:        "Start at",
	ScheduledEndAt:          "End at",
	PublishedAt:             "Start at",
	UnPublishedAt:           "End at",
	ActualPublishTime:       "Actual Publish Time",
	SchedulePublishTime:     "Schedule Publish Time",
	NotSet:                  "Not set",
	WhenDoYouWantToPublish:  "When do you want to publish?",
	PublishScheduleTip:      "After you set the {SchedulePublishTime}, the system will automatically publish/unpublish it.",
	DateTimePickerClearText: "Clear",
	DateTimePickerOkText:    "OK",
	SaveAsNewVersion:        "Save As New Version",
	SwitchedToNewVersion:    "Switched To New Version",
	SuccessfullyCreated:     "Successfully Created",
	SuccessfullyRename:      "Successfully Rename",
	OnlineVersion:           "Online Version",
	VersionsList:            "Versions List",
	AllVersions:             "All versions",
	NamedVersions:           "Named versions",
	RenameVersion:           "Rename Version",
}

var Messages_zh_CN = &Messages{
	StatusDraft:             "草稿",
	StatusOnline:            "在线",
	StatusOffline:           "离线",
	Publish:                 "发布",
	Unpublish:               "取消发布",
	Republish:               "重新发布",
	Areyousure:              "你确定吗?",
	ScheduledStartAt:        "发布时间",
	ScheduledEndAt:          "下线时间",
	PublishedAt:             "发布时间",
	UnPublishedAt:           "下线时间",
	ActualPublishTime:       "实际发布时间",
	SchedulePublishTime:     "计划发布时间",
	NotSet:                  "未设定",
	WhenDoYouWantToPublish:  "你希望什么时候发布？",
	PublishScheduleTip:      "设定好 {SchedulePublishTime} 之后, 系统会按照时间自动将它发布/下线。",
	DateTimePickerClearText: "清空",
	DateTimePickerOkText:    "确定",
	SaveAsNewVersion:        "保存为一个新版本",
	SwitchedToNewVersion:    "切换到新版本",
	SuccessfullyCreated:     "成功创建",
	SuccessfullyRename:      "成功命名",
	OnlineVersion:           "在线版本",
	VersionsList:            "版本列表",
	AllVersions:             "所有版本",
	NamedVersions:           "已命名版本",
	RenameVersion:           "命名版本",
}

func GetStatusText(status string, msgr *Messages) string {
	switch status {
	case publish.StatusDraft:
		return msgr.StatusDraft
	case publish.StatusOnline:
		return msgr.StatusOnline
	case publish.StatusOffline:
		return msgr.StatusOffline
	}
	return ""
}
