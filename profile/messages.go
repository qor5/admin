package profile

import (
	"fmt"
	"strings"
)

type Messages struct {
	UnreadMessagesTemplate string
	ViewLoginSessions      string
	Logout                 string
	Available              string
	Unavailable            string
	SuccessfullyRename     string
}

func (m *Messages) UnreadMessages(n int) string {
	return strings.NewReplacer("{n}", fmt.Sprint(n)).
		Replace(m.UnreadMessagesTemplate)
}

var Messages_en_US = &Messages{
	UnreadMessagesTemplate: "{n} unread notes",
	ViewLoginSessions:      "View login sessions",
	Logout:                 "Logout",
	Available:              "Available",
	Unavailable:            "Unavailable",
	SuccessfullyRename:     "Successfully renamed",
}

var Messages_zh_CN = &Messages{
	UnreadMessagesTemplate: "未读 {n} 条",
	ViewLoginSessions:      "查看登录会话",
	Logout:                 "登出",
	Available:              "可用",
	Unavailable:            "不可用",
	SuccessfullyRename:     "成功重命名",
}

var Messages_ja_JP = &Messages{
	UnreadMessagesTemplate: "{n} 未読",
	ViewLoginSessions:      "ログインセッションを表示",
	Logout:                 "ログアウト",
	Available:              "利用可能",
	Unavailable:            "利用不可",
	SuccessfullyRename:     "名前が変更されました",
}
