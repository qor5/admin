package admin

import (
	"github.com/qor5/x/v3/i18n"
)

const I18nExampleKey i18n.ModuleKey = "I18nExampleKey"

type Messages struct {
	FilterTabsAll            string
	FilterTabsHasUnreadNotes string
	FilterTabsActive         string
}

var Messages_en_US = &Messages{
	FilterTabsAll:            "All",
	FilterTabsHasUnreadNotes: "Has Unread Notes",
	FilterTabsActive:         "Active",
}

var Messages_zh_CN = &Messages{
	FilterTabsAll:            "全部",
	FilterTabsHasUnreadNotes: "未读备注",
	FilterTabsActive:         "有效",
}

type MessagesModels struct {
	QOR5Example string
	Roles       string
	Users       string

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

	PageBuilder      string
	Pages            string
	SharedContainers string
	DemoContainers   string
	Templates        string
	PageCategories   string
	SEO              string
	Profile          string
	ActivityLogs     string
	MediaLibrary     string

	PagesID         string
	PagesTitle      string
	PagesSlug       string
	PagesLocale     string
	PagesNotes      string
	PagesDraftCount string
	PagesOnline     string

	Page                   string
	PagesStatus            string
	PagesSchedule          string
	PagesCategoryID        string
	PagesTemplateSelection string
	PagesEditContainer     string
}

var MessagesModels_zh_CN = &MessagesModels{
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

	QOR5Example: "QOR5 示例",
	Roles:       "权限管理",
	Users:       "用户管理",

	PageBuilder:      "页面管理菜单",
	Pages:            "页面管理",
	SharedContainers: "公用组件",
	DemoContainers:   "示例组件",
	Templates:        "模板页面",
	PageCategories:   "目录管理",
	SEO:              "SEO 管理",
	Profile:          "个人页面",
	ActivityLogs:     "操作日志",
	MediaLibrary:     "媒体库",

	PagesID:         "ID",
	PagesTitle:      "标题",
	PagesSlug:       "Slug",
	PagesLocale:     "地区",
	PagesNotes:      "备注",
	PagesDraftCount: "草稿数",
	PagesOnline:     "在线",

	Page:                   "Page",
	PagesStatus:            "PagesStatus",
	PagesSchedule:          "PagesSchedule",
	PagesCategoryID:        "PagesCategoryID",
	PagesTemplateSelection: "PagesTemplateSelection",
	PagesEditContainer:     "PagesEditContainer",
}
