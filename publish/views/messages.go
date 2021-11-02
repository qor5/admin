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
	DateTimePickerClearText string
	DateTimePickerOkText    string
	SaveAsNewVersion        string
	SuccessfullyCreated     string
	OnlineVersion           string
	VersionsList            string
}

var Messages_en_US = &Messages{
	StatusDraft:             "Draft",
	StatusOnline:            "Online",
	StatusOffline:           "Offline",
	Publish:                 "Publish",
	Unpublish:               "Unpublish",
	Republish:               "Republish",
	Areyousure:              "Are you sure?",
	ScheduledStartAt:        "Scheduled start at",
	ScheduledEndAt:          "Scheduled end at",
	DateTimePickerClearText: "Clear",
	DateTimePickerOkText:    "OK",
	SaveAsNewVersion:        "Save As New Version",
	SuccessfullyCreated:     "Successfully Created",
	OnlineVersion:           "Online Version",
	VersionsList:            "Versions List",
}

var Messages_zh_CN = &Messages{
	StatusDraft:             "草稿",
	StatusOnline:            "在线",
	StatusOffline:           "离线",
	Publish:                 "发布",
	Unpublish:               "取消发布",
	Republish:               "重新发布",
	Areyousure:              "你确定吗?",
	ScheduledStartAt:        "预计开始时间",
	ScheduledEndAt:          "预计结束时间",
	DateTimePickerClearText: "清空",
	DateTimePickerOkText:    "确定",
	SaveAsNewVersion:        "保存为一个新版本",
	SuccessfullyCreated:     "成功创建",
	OnlineVersion:           "在线版本",
	VersionsList:            "版本列表",
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
