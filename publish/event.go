package publish

import (
	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

const (
	EventPublish   = "publish_EventPublish"
	EventRepublish = "publish_EventRepublish"
	EventUnpublish = "publish_EventUnpublish"

	EventDuplicateVersion      = "publish_EventDuplicateVersion"
	eventSchedulePublishDialog = "publish_eventSchedulePublishDialog"
	eventSchedulePublish       = "publish_eventSchedulePublish"

	eventRenameVersionDialog = "publish_eventRenameVersionDialog"
	eventRenameVersion       = "publish_eventRenameVersion"
	eventDeleteVersionDialog = "publish_eventDeleteVersionDialog"
	eventDeleteVersion       = "publish_eventDeleteVersion"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"

	ParamScriptAfterPublish = "publish_param_script_after_publish"
)

func registerEventFuncsForResource(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder) {
	mb.RegisterEventFunc(EventPublish, publishAction(db, mb, publisher, ActivityPublish))
	mb.RegisterEventFunc(EventRepublish, publishAction(db, mb, publisher, ActivityRepublish))
	mb.RegisterEventFunc(EventUnpublish, unpublishAction(db, mb, publisher, ActivityUnPublish))

	mb.RegisterEventFunc(EventDuplicateVersion, duplicateVersionAction(mb, db))
	mb.RegisterEventFunc(eventSchedulePublishDialog, scheduleDialog(db, mb))
	mb.RegisterEventFunc(eventSchedulePublish, schedule(db, mb))
}

func registerEventFuncsForVersion(mb *presets.ModelBuilder, db *gorm.DB) {
	mb.RegisterEventFunc(eventRenameVersionDialog, renameVersionDialog(mb))
	mb.RegisterEventFunc(eventRenameVersion, renameVersion(mb))
	mb.RegisterEventFunc(eventDeleteVersionDialog, deleteVersionDialog(mb))
	mb.RegisterEventFunc(eventDeleteVersion, deleteVersion(mb, db))
}
