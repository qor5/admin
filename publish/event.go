package publish

import (
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

const (
	EventPublish   = "publish_EventPublish"
	EventRepublish = "publish_EventRepublish"
	EventUnpublish = "publish_EventUnpublish"

	EventDuplicateVersion    = "publish_EventDuplicateVersion"
	eventSelectVersion       = "publish_eventSelectVersion"
	eventRenameVersionDialog = "publish_eventRenameVersionDialog"
	eventRenameVersion       = "publish_eventRenameVersion"
	eventDeleteVersionDialog = "publish_eventDeleteVersionDialog"

	eventSchedulePublishDialog = "publish_eventSchedulePublishDialog"
	eventSchedulePublish       = "publish_eventSchedulePublish"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"

	ParamScriptAfterPublish = "publish_param_script_after_publish"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder) {
	mb.RegisterEventFunc(EventPublish, publishAction(db, mb, publisher, ab, ActivityPublish))
	mb.RegisterEventFunc(EventRepublish, publishAction(db, mb, publisher, ab, ActivityRepublish))
	mb.RegisterEventFunc(EventUnpublish, unpublishAction(db, mb, publisher, ab, ActivityUnPublish))

	mb.RegisterEventFunc(EventDuplicateVersion, duplicateVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(eventRenameVersionDialog, renameVersionDialog(mb))
	mb.RegisterEventFunc(eventRenameVersion, renameVersion(mb))
	mb.RegisterEventFunc(eventDeleteVersionDialog, deleteVersionDialog(mb))

	mb.RegisterEventFunc(eventSchedulePublishDialog, schedulePublishDialog(db, mb))
	mb.RegisterEventFunc(eventSchedulePublish, schedulePublish(db, mb))
}
