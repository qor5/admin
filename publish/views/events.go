package views

import (
	"github.com/goplaid/web"
	"github.com/qor/qor5/publish"
	"github.com/theplant/jsontyperegistry"
	"gorm.io/gorm"
)

const (
	publishEvent   = "publish_PublishEvent"
	unpublishEvent = "publish_UnpublishEvent"
)

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB, publisher *publish.Builder) {
	hub.RegisterEventFunc(publishEvent, publishAction(db, publisher))
	hub.RegisterEventFunc(unpublishEvent, unpublishAction(db, publisher))
}

func publishAction(db *gorm.DB, publisher *publish.Builder) web.EventFunc {
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

func unpublishAction(db *gorm.DB, publisher *publish.Builder) web.EventFunc {
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
	objJson := ctx.Event.Params[0]

	obj = jsontyperegistry.MustNewWithJSONString(objJson)
	if err = db.First(obj).Error; err != nil {
		return
	}
	return
}
