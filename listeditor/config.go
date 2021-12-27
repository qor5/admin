package listeditor

import (
	"fmt"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/sunfmin/reflectutils"
)

const (
	addRowEvent    = "listEditor_addRowEvent"
	removeRowEvent = "listEditor_removeRowEvent"
)

const (
	ParamOpFormKey = "listEditor_ParamOpFormKey"
)

func Configure(mbs ...*presets.ModelBuilder) {
	for _, mb := range mbs {
		mb.RegisterEventFunc(addRowEvent, addRow(mb))
		mb.RegisterEventFunc(removeRowEvent, removeRow(mb))
	}
}

func addRow(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue(presets.ParamID)

		me := mb.Editing()
		obj, vErr := me.FetchAndUnmarshal(id, ctx)
		if vErr.HaveErrors() {
			me.UpdateOverlayContent(ctx, &r, obj, "", &vErr)
			return
		}

		formKey := ctx.R.FormValue(ParamOpFormKey)
		t := reflectutils.GetType(obj, formKey)
		if t.Kind() != reflect.Slice {
			panic(fmt.Sprintf("%s is not slice", formKey))
		}

		newVal := reflect.New(t.Elem().Elem()).Interface()

		err = reflectutils.Set(obj, formKey+"[]", newVal)
		if err != nil {
			panic(err)
		}
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}

func removeRow(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue(presets.ParamID)

		me := mb.Editing()
		obj, vErr := me.FetchAndUnmarshal(id, ctx)
		if vErr.HaveErrors() {
			me.UpdateOverlayContent(ctx, &r, obj, "", &vErr)
			return
		}

		formKey := ctx.R.FormValue(ParamOpFormKey)

		reflectutils.Get(obj, formKey)

		return
	}
}
