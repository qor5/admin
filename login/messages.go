package login

import (
	"fmt"
	"strings"
)

type Messages struct {
	SessionTableHeaderTime           string
	SessionTableHeaderDevice         string
	SessionTableHeaderIPAddress      string
	SessionTableHeaderStatus         string
	SessionsDialogTitle              string
	SessionStatusExpired             string
	SessionStatusActive              string
	SessionStatusCurrent             string
	HideIPAddressTips                string
	ExpireOtherSessions              string
	SuccessfullyExpiredOtherSessions string
	UnreadMessagesTemplate           string
	ViewLoginSessions                string
	Logout                           string
	Available                        string
	Unavailable                      string
	SuccessfullyRename               string
}

func (m *Messages) UnreadMessages(n int) string {
	return strings.NewReplacer("{n}", fmt.Sprint(n)).
		Replace(m.UnreadMessagesTemplate)
}

var Messages_en_US = &Messages{
	SessionTableHeaderTime:           "Time",
	SessionTableHeaderDevice:         "Device",
	SessionTableHeaderIPAddress:      "IP Address",
	SessionTableHeaderStatus:         "Status",
	SessionsDialogTitle:              "Login Sessions",
	SessionStatusExpired:             "Expired",
	SessionStatusActive:              "Active",
	SessionStatusCurrent:             "Current Session",
	HideIPAddressTips:                "Invisible due to security concerns",
	ExpireOtherSessions:              "Sign out all other sessions",
	SuccessfullyExpiredOtherSessions: "All other sessions have successfully been signed out.",
	UnreadMessagesTemplate:           "{n} unread notes",
	ViewLoginSessions:                "View login sessions",
	Logout:                           "Logout",
	Available:                        "Available",
	Unavailable:                      "Unavailable",
	SuccessfullyRename:               "Successfully renamed",
}

var Messages_zh_CN = &Messages{
	SessionTableHeaderTime:           "时间",
	SessionTableHeaderDevice:         "设备",
	SessionTableHeaderIPAddress:      "IP地址",
	SessionTableHeaderStatus:         "状态",
	SessionsDialogTitle:              "登录会话",
	SessionStatusExpired:             "已过期",
	SessionStatusActive:              "有效",
	SessionStatusCurrent:             "当前会话",
	HideIPAddressTips:                "由于安全原因，隐藏",
	ExpireOtherSessions:              "登出所有其他会话",
	SuccessfullyExpiredOtherSessions: "所有其他会话已成功登出。",
	UnreadMessagesTemplate:           "未读 {n} 条",
	ViewLoginSessions:                "查看登录会话",
	Logout:                           "登出",
	Available:                        "可用",
	Unavailable:                      "不可用",
	SuccessfullyRename:               "成功重命名",
}

var Messages_ja_JP = &Messages{
	SessionTableHeaderTime:           "時間",
	SessionTableHeaderDevice:         "デバイス",
	SessionTableHeaderIPAddress:      "IPアドレス",
	SessionTableHeaderStatus:         "ステータス",
	SessionsDialogTitle:              "ログインセッション",
	SessionStatusExpired:             "期限切れ",
	SessionStatusActive:              "有効",
	SessionStatusCurrent:             "現在のセッション",
	HideIPAddressTips:                "セキュリティ保護のため表示できません",
	ExpireOtherSessions:              "他のすべてのセッションからサインアウトする",
	SuccessfullyExpiredOtherSessions: "他のすべてのセッションは正常にサインアウトされました。",
	UnreadMessagesTemplate:           "{n} 件の未読",
	ViewLoginSessions:                "ログインセッションを表示",
	Logout:                           "ログアウト",
	Available:                        "利用可能",
	Unavailable:                      "利用不可",
	SuccessfullyRename:               "名前が変更されました",
}
