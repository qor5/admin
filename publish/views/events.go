package views

import (
	"fmt"
	"reflect"

	"github.com/goplaid/web"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

const (
	publishEvent   = "publish_PublishEvent"
	unpublishEvent = "publish_UnpublishEvent"
)

var modelsMap = make(map[string]interface{})

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB, publisher *publish.Builder) {
	hub.RegisterEventFunc(publishEvent, publishAction(db, publisher))
	hub.RegisterEventFunc(unpublishEvent, unpublishAction(db, publisher))
}

func RegisterPublishModels(models ...interface{}) {
	for _, m := range models {
		modelsMap[reflect.TypeOf(m).String()] = m
	}
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
	modelType := ctx.Event.Params[0]
	m, ok := modelsMap[modelType]
	if !ok {
		err = fmt.Errorf("unregistered type %s", modelType)
		return
	}
	obj = reflect.New(reflect.TypeOf(m).Elem()).Interface()

	id := ctx.Event.Params[1]
	reflectutils.Set(obj, "ID", id)
	if len(ctx.Event.Params) > 2 {
		version := ctx.Event.Params[2]
		if o, ok := obj.(publish.VersionInterface); ok {
			o.SetVersionName(version)
		}
	}

	if err = db.First(obj).Error; err != nil {
		return
	}
	return
}
