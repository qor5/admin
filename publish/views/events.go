package views

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/activity"
	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

const (
	publishEvent        = "publish_PublishEvent"
	republishEvent      = "publish_republishEvent"
	unpublishEvent      = "publish_UnpublishEvent"
	switchVersionEvent  = "publish_SwitchVersionEvent"
	saveNewVersionEvent = "publish_SaveNewVersionEvent"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder) {
	mb.RegisterEventFunc(publishEvent, publishAction(db, mb, publisher, ab, ActivityPublish))
	mb.RegisterEventFunc(republishEvent, publishAction(db, mb, publisher, ab, ActivityRepublish))
	mb.RegisterEventFunc(unpublishEvent, unpublishAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(switchVersionEvent, switchVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(saveNewVersionEvent, saveNewVersionAction(db, mb, publisher))
}

func publishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}
		err = publisher.Publish(obj)
		if err != nil {
			return
		}
		if ab != nil {
			if _, exist := ab.GetModelBuilder(obj); exist {
				ab.AddCustomizedRecord(actionName, false, ctx.R.Context(), obj)
			}
		}

		presets.ShowMessage(&r, "success", "")
		r.Reload = true
		return
	}
}

func unpublishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}

		err = publisher.UnPublish(obj)
		if err != nil {
			return
		}
		if ab != nil {
			if _, exist := ab.GetModelBuilder(obj); exist {
				ab.AddCustomizedRecord(actionName, false, ctx.R.Context(), obj)
			}
		}

		presets.ShowMessage(&r, "success", "")
		r.Reload = true
		return
	}
}
