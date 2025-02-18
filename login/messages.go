package login

import (
	"fmt"
	"strings"
)

type Messages struct {
	SessionTableHeaderTime           string
	SessionTableHeaderDevice         string
	SessionTableHeaderLocation       string
	SessionTableHeaderIPAddress      string
	SessionTableHeaderStatus         string
	SessionTableHeaderLastActivedAt  string
	SessionTableHeaderAction         string
	SessionsDialogTitle              string
	SessionStatusExpired             string
	SessionStatusActive              string
	SessionStatusCurrent             string
	HideIPAddressTips                string
	ExpireOtherSessions              string
	SuccessfullyExpiredOtherSessions string
	SuccessfullyExpiredSessions      string
	UnreadMessagesTemplate           string
	ViewLoginSessions                string
	Logout                           string
	Available                        string
	Unavailable                      string
	SuccessfullyRename               string
	LocationUnknown                  string

	LoginWelcomeLabel string
	LoginTitleLabel   string

	LoginAccountLabel         string
	LoginAccountPlaceholder   string
	LoginPasswordLabel        string
	LoginPasswordPlaceholder  string
	LoginSignInButtonLabel    string
	LoginForgetPasswordLabel  string
	LoginQor5DescriptionLabel string

	LoginProviderGoogleText    string
	LoginProviderMicrosoftText string
	LoginProviderGithubText    string
}

func (m *Messages) UnreadMessages(n int) string {
	return strings.NewReplacer("{n}", fmt.Sprint(n)).
		Replace(m.UnreadMessagesTemplate)
}

var Messages_en_US = &Messages{
	SessionTableHeaderTime:           "Time",
	SessionTableHeaderDevice:         "Device",
	SessionTableHeaderLocation:       "Location",
	SessionTableHeaderIPAddress:      "IP Address",
	SessionTableHeaderStatus:         "Status",
	SessionTableHeaderLastActivedAt:  "Last Active Time",
	SessionTableHeaderAction:         "Action",
	SessionsDialogTitle:              "Login Sessions",
	SessionStatusExpired:             "Expired",
	SessionStatusActive:              "Active",
	SessionStatusCurrent:             "Current Session",
	HideIPAddressTips:                "Invisible due to security concerns",
	ExpireOtherSessions:              "Sign out all other sessions",
	SuccessfullyExpiredOtherSessions: "All other sessions have successfully been signed out.",
	SuccessfullyExpiredSessions:      "Session has successfully been signed out.",
	UnreadMessagesTemplate:           "{n} unread notes",
	ViewLoginSessions:                "View login sessions",
	Logout:                           "Logout",
	Available:                        "Available",
	Unavailable:                      "Unavailable",
	SuccessfullyRename:               "Successfully renamed",
	LocationUnknown:                  "Unknown",

	LoginWelcomeLabel: "Welcome",
	LoginTitleLabel:   "Qor Admin System",

	LoginAccountLabel:         "Email",
	LoginAccountPlaceholder:   "Please enter your email",
	LoginPasswordLabel:        "Password",
	LoginPasswordPlaceholder:  "Please enter your password",
	LoginSignInButtonLabel:    "Sign in",
	LoginForgetPasswordLabel:  "Forget your password?",
	LoginQor5DescriptionLabel: "Enterprise Systems, Startup Speed",

	LoginProviderGoogleText:    "Sign in with Google",
	LoginProviderMicrosoftText: "Sign in with Microsoft",
	LoginProviderGithubText:    "Sign in with Github",
}

var Messages_zh_CN = &Messages{
	SessionTableHeaderTime:           "时间",
	SessionTableHeaderDevice:         "设备",
	SessionTableHeaderLocation:       "位置",
	SessionTableHeaderIPAddress:      "IP地址",
	SessionTableHeaderStatus:         "状态",
	SessionTableHeaderLastActivedAt:  "最后活跃时间",
	SessionTableHeaderAction:         "操作",
	SessionsDialogTitle:              "登录会话",
	SessionStatusExpired:             "已过期",
	SessionStatusActive:              "有效",
	SessionStatusCurrent:             "当前会话",
	HideIPAddressTips:                "由于安全原因，隐藏",
	ExpireOtherSessions:              "登出所有其他会话",
	SuccessfullyExpiredOtherSessions: "所有其他会话已成功登出。",
	SuccessfullyExpiredSessions:      "会话已成功登出。",
	UnreadMessagesTemplate:           "未读 {n} 条",
	ViewLoginSessions:                "查看登录会话",
	Logout:                           "登出",
	Available:                        "可用",
	Unavailable:                      "不可用",
	SuccessfullyRename:               "成功重命名",
	LocationUnknown:                  "未知",

	LoginWelcomeLabel: "欢迎",
	LoginTitleLabel:   "Qor 管理系统",

	LoginAccountLabel:         "邮箱",
	LoginAccountPlaceholder:   "请输入您的邮箱",
	LoginPasswordLabel:        "密码",
	LoginPasswordPlaceholder:  "请输入您的密码",
	LoginSignInButtonLabel:    "登录",
	LoginForgetPasswordLabel:  "忘记密码？",
	LoginQor5DescriptionLabel: "企业级系统，创业速度",

	LoginProviderGoogleText:    "使用 Google 登录",
	LoginProviderMicrosoftText: "使用 Microsoft 登录",
	LoginProviderGithubText:    "使用 Github 登录",
}

var Messages_ja_JP = &Messages{
	SessionTableHeaderTime:           "時間",
	SessionTableHeaderDevice:         "デバイス",
	SessionTableHeaderLocation:       "位置",
	SessionTableHeaderIPAddress:      "IPアドレス",
	SessionTableHeaderStatus:         "ステータス",
	SessionTableHeaderLastActivedAt:  "最終アクティブ",
	SessionTableHeaderAction:         "アクション",
	SessionsDialogTitle:              "ログインセッション",
	SessionStatusExpired:             "期限切れ",
	SessionStatusActive:              "有効",
	SessionStatusCurrent:             "現在のセッション",
	HideIPAddressTips:                "セキュリティ保護のため表示できません",
	ExpireOtherSessions:              "他のすべてのセッションからサインアウトする",
	SuccessfullyExpiredOtherSessions: "他のすべてのセッションは正常にサインアウトされました。",
	SuccessfullyExpiredSessions:      "セッションは正常にサインアウトされました。",
	UnreadMessagesTemplate:           "{n} 件の未読",
	ViewLoginSessions:                "ログインセッションを表示",
	Logout:                           "ログアウト",
	Available:                        "利用可能",
	Unavailable:                      "利用不可",
	SuccessfullyRename:               "名前が変更されました",
	LocationUnknown:                  "不明",

	LoginWelcomeLabel: "ようこそ",
	LoginTitleLabel:   "Qor 管理システム",

	LoginAccountLabel:         "メールアドレス",
	LoginAccountPlaceholder:   "メールアドレスを入力してください",
	LoginPasswordLabel:        "パスワード",
	LoginPasswordPlaceholder:  "パスワードを入力してください",
	LoginSignInButtonLabel:    "サインイン",
	LoginForgetPasswordLabel:  "パスワードをお忘れですか？",
	LoginQor5DescriptionLabel: "エンタープライズシステム、スタートアップスピード",

	LoginProviderGoogleText:    "Google でログイン",
	LoginProviderMicrosoftText: "Microsoft でログイン",
	LoginProviderGithubText:    "Github でログイン",
}
