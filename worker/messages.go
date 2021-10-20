package worker

import "github.com/goplaid/x/i18n"

const I18nWorkerKey i18n.ModuleKey = "I18nWorkerKey"

type Messages struct {
	StatusNew          string
	StatusScheduled    string
	StatusRunning      string
	StatusCancelled    string
	StatusDone         string
	StatusException    string
	StatusKilled       string
	FilterTabAll       string
	FilterTabRunning   string
	FilterTabScheduled string
	FilterTabDone      string
	FilterTabErrors    string
	ActionCancelJob    string
	ActionAbortJob     string
	ActionUpdateJob    string
	ActionRerunJob     string
	DetailTitleStatus  string
	DetailTitleLog     string
}

var Messages_en_US = &Messages{
	StatusNew:          "New",
	StatusScheduled:    "Scheduled",
	StatusRunning:      "Running",
	StatusCancelled:    "Cancelled",
	StatusDone:         "Done",
	StatusException:    "Exception",
	StatusKilled:       "Killed",
	FilterTabAll:       "All Jobs",
	FilterTabRunning:   "Running",
	FilterTabScheduled: "Scheduled",
	FilterTabDone:      "Done",
	FilterTabErrors:    "Errors",
	ActionCancelJob:    "Cancel Job",
	ActionAbortJob:     "Abort Job",
	ActionUpdateJob:    "Update Job",
	ActionRerunJob:     "Rerun Job",
	DetailTitleStatus:  "Status",
	DetailTitleLog:     "Log",
}

var Messages_zh_CN = &Messages{
	StatusNew:          "新建",
	StatusScheduled:    "计划",
	StatusRunning:      "运行中",
	StatusCancelled:    "取消",
	StatusDone:         "完成",
	StatusException:    "错误",
	StatusKilled:       "中止",
	FilterTabAll:       "全部",
	FilterTabRunning:   "运行中",
	FilterTabScheduled: "计划",
	FilterTabDone:      "完成",
	FilterTabErrors:    "错误",
	ActionCancelJob:    "取消任务",
	ActionAbortJob:     "中止任务",
	ActionUpdateJob:    "更新任务",
	ActionRerunJob:     "重跑任务",
	DetailTitleStatus:  "状态",
	DetailTitleLog:     "日志",
}
