package worker

import (
	"net/http"

	"github.com/qor5/x/i18n"
	"github.com/qor5/admin/presets"
)

const I18nWorkerKey i18n.ModuleKey = "I18nWorkerKey"

type Messages struct {
	StatusNew                string
	StatusScheduled          string
	StatusRunning            string
	StatusCancelled          string
	StatusDone               string
	StatusException          string
	StatusKilled             string
	FilterTabAll             string
	FilterTabRunning         string
	FilterTabScheduled       string
	FilterTabDone            string
	FilterTabErrors          string
	ActionCancelJob          string
	ActionAbortJob           string
	ActionUpdateJob          string
	ActionRerunJob           string
	DetailTitleStatus        string
	DetailTitleLog           string
	NoticeJobCannotBeAborted string
	NoticeJobWontBeExecuted  string
	ScheduleTime             string
	DateTimePickerClearText  string
	DateTimePickerOkText     string
	PleaseSelectJob          string
}

var Messages_en_US = &Messages{
	StatusNew:                "New",
	StatusScheduled:          "Scheduled",
	StatusRunning:            "Running",
	StatusCancelled:          "Cancelled",
	StatusDone:               "Done",
	StatusException:          "Exception",
	StatusKilled:             "Killed",
	FilterTabAll:             "All Jobs",
	FilterTabRunning:         "Running",
	FilterTabScheduled:       "Scheduled",
	FilterTabDone:            "Done",
	FilterTabErrors:          "Errors",
	ActionCancelJob:          "Cancel Job",
	ActionAbortJob:           "Abort Job",
	ActionUpdateJob:          "Update Job",
	ActionRerunJob:           "Rerun Job",
	DetailTitleStatus:        "Status",
	DetailTitleLog:           "Log",
	NoticeJobCannotBeAborted: "This job cannot be aborted/canceled/updated due to its status change",
	NoticeJobWontBeExecuted:  "This job won't be executed due to code being deleted/modified",
	ScheduleTime:             "Schedule Time",
	DateTimePickerClearText:  "Clear",
	DateTimePickerOkText:     "OK",
	PleaseSelectJob:          "Please select job",
}

var Messages_zh_CN = &Messages{
	StatusNew:                "??????",
	StatusScheduled:          "??????",
	StatusRunning:            "?????????",
	StatusCancelled:          "??????",
	StatusDone:               "??????",
	StatusException:          "??????",
	StatusKilled:             "??????",
	FilterTabAll:             "??????",
	FilterTabRunning:         "?????????",
	FilterTabScheduled:       "??????",
	FilterTabDone:            "??????",
	FilterTabErrors:          "??????",
	ActionCancelJob:          "??????Job",
	ActionAbortJob:           "??????Job",
	ActionUpdateJob:          "??????Job",
	ActionRerunJob:           "??????Job",
	DetailTitleStatus:        "??????",
	DetailTitleLog:           "??????",
	NoticeJobCannotBeAborted: "Job????????????????????????????????????/??????/??????",
	NoticeJobWontBeExecuted:  "Job???????????????/??????, ??????Job???????????????",
	ScheduleTime:             "????????????",
	DateTimePickerClearText:  "??????",
	DateTimePickerOkText:     "??????",
	PleaseSelectJob:          "?????????Job",
}

func getTStatus(msgr *Messages, status string) string {
	switch status {
	case JobStatusNew:
		return msgr.StatusNew
	case JobStatusScheduled:
		return msgr.StatusScheduled
	case JobStatusRunning:
		return msgr.StatusRunning
	case JobStatusCancelled:
		return msgr.StatusCancelled
	case JobStatusDone:
		return msgr.StatusDone
	case JobStatusException:
		return msgr.StatusException
	case JobStatusKilled:
		return msgr.StatusKilled
	}
	return status
}

func getTJob(r *http.Request, v string) string {
	return i18n.PT(r, presets.ModelsI18nModuleKey, "WorkerJob", v)
}
