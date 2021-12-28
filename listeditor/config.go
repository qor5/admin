package listeditor

import (
	"encoding/json"
	"reflect"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

const (
	addRowEvent    = "listEditor_addRowEvent"
	removeRowEvent = "listEditor_removeRowEvent"
)

const (
	ParamAddRowFormKey    = "listEditor_AddRowFormKey"
	ParamRemoveRowFormKey = "listEditor_RemoveRowFormKey"
	ParamHiddenObjectName = "listEditor_HiddenObjectName"
)

func Configure(mbs ...*presets.ModelBuilder) {
	for _, mb := range mbs {
		mb.RegisterEventFunc(addRowEvent, addRow(mb))
		mb.RegisterEventFunc(removeRowEvent, removeRow(mb))

		mb.Editing().AppendHiddenFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			return h.Input("").Type("hidden").Value(h.JSONString(obj)).Attr(web.VFieldName(ParamHiddenObjectName)...)
		})

	}
}

func addRow(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		objJson := ctx.R.FormValue(ParamHiddenObjectName)

		me := mb.Editing()
		obj := mb.NewModel()
		err = json.Unmarshal([]byte(objJson), obj)
		if err != nil {
			return
		}
		_ = me.RunSetterFunc(ctx, obj)

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
		objJson := ctx.R.FormValue(ParamHiddenObjectName)

		me := mb.Editing()
		obj := mb.NewModel()
		err = json.Unmarshal([]byte(objJson), obj)
		if err != nil {
			return
		}
		_ = me.RunSetterFunc(ctx, obj)

		formKey := ctx.R.FormValue(ParamRemoveRowFormKey)
		err = reflectutils.Delete(obj, formKey)
		if err != nil {
			panic(err)
		}
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}
