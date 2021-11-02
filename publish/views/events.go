package views

import (
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/publish"
	"github.com/theplant/jsontyperegistry"
	"gorm.io/gorm"
)

const (
	publishEvent        = "publish_PublishEvent"
	unpublishEvent      = "publish_UnpublishEvent"
	switchVersionEvent  = "publish_SwitchVersionEvent"
	saveNewVersionEvent = "publish_SaveNewVersionEvent"
	removeVersionEvent  = "publish_RemoveVersionEvent"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) {
	mb.RegisterEventFunc(publishEvent, publishAction(db, mb, publisher))
	mb.RegisterEventFunc(unpublishEvent, unpublishAction(db, mb, publisher))
	mb.RegisterEventFunc(switchVersionEvent, switchVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(saveNewVersionEvent, saveNewVersionAction(db, mb, publisher))
}

func publishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var obj interface{}
		obj, err = getCurrentObj(db, ctx)
		if err != nil {
			return
		}
		err = publisher.Publish(obj)
		r.Reload = true
		return
	}
}

func unpublishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var obj interface{}

		obj, err = getCurrentObj(db, ctx)
		if err != nil {
			return
		}

		err = publisher.UnPublish(obj)

		r.Reload = true
		return
	}
}

func getCurrentObj(db *gorm.DB, ctx *web.EventContext) (obj interface{}, err error) {
	objJson := ctx.R.FormValue("objJson")

	obj = jsontyperegistry.MustNewWithJSONString(objJson)
	if err = db.First(obj).Error; err != nil {
		return
	}
	return
}
