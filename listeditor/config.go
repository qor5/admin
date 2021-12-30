package listeditor

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/sunfmin/reflectutils"
)

const (
	addRowEvent    = "listEditor_addRowEvent"
	removeRowEvent = "listEditor_removeRowEvent"
)

const (
	ParamAddRowFormKey    = "listEditor_AddRowFormKey"
	ParamRemoveRowFormKey = "listEditor_RemoveRowFormKey"
)

func Configure(mbs ...*presets.ModelBuilder) {
	for _, mb := range mbs {
		mb.RegisterEventFunc(addRowEvent, addRow(mb))
		mb.RegisterEventFunc(removeRowEvent, removeRow(mb))
	}
}

func addRow(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(presets.ParamID), true, ctx)

		formKey := ctx.R.FormValue(ParamAddRowFormKey)
		t := reflectutils.GetType(obj, formKey+"[0]")
		newVal := reflect.New(t.Elem()).Interface()
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

		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(presets.ParamID), true, ctx)

		formKey := ctx.R.FormValue(ParamRemoveRowFormKey)
		lb := strings.LastIndex(formKey, "[")
		sliceField := formKey[0:lb]
		strIndex := formKey[lb+1 : strings.LastIndex(formKey, "]")]

		var index int
		index, err = strconv.Atoi(strIndex)
		if err != nil {
			return
		}
		presets.ContextDeletedIndexesBuilder(ctx).Append(sliceField, index)
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}
