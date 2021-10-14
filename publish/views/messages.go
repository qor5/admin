package views

import "github.com/qor/qor5/publish"

type Messages struct {
	StatusDraft   string
	StatusOnline  string
	StatusOffline string
	Publish       string
	Unpublish     string
	Republish     string
	Areyousure    string
}

var Messages_en_US = &Messages{
	StatusDraft:   "Draft",
	StatusOnline:  "Online",
	StatusOffline: "Offline",
	Publish:       "Publish",
	Unpublish:     "Unpublish",
	Republish:     "Republish",
	Areyousure:    "Are you sure?",
}

var Messages_zh_CN = &Messages{
	StatusDraft:   "草稿",
	StatusOnline:  "在线",
	StatusOffline: "离线",
	Publish:       "发布",
	Unpublish:     "取消发布",
	Republish:     "重新发布",
	Areyousure:    "你确定吗?",
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
