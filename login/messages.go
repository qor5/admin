package login

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
	HideIPAddressTips:                "セキュリティ上の理由から非表示",
	ExpireOtherSessions:              "他のすべてのセッションをサインアウトする",
	SuccessfullyExpiredOtherSessions: "他のすべてのセッションは正常にサインアウトされました。",
}
