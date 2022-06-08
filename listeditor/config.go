package listeditor

import (
	"encoding/json"
	"fmt"
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
	sortEvent      = "listEditor_sortEvent"
)

const (
	ParamAddRowFormKey      = "listEditor_AddRowFormKey"
	ParamRemoveRowFormKey   = "listEditor_RemoveRowFormKey"
	ParamIsStartSort        = "listEditor_IsStartSort"
	ParamSortSectionFormKey = "listEditor_SortSectionFormKey"
	ParamSortResultFormKey  = "listEditor_SortResultFormKey"
)

func Configure(mbs ...*presets.ModelBuilder) {
	for _, mb := range mbs {
		mb.RegisterEventFunc(addRowEvent, addRow(mb))
		mb.RegisterEventFunc(removeRowEvent, removeRow(mb))
		mb.RegisterEventFunc(sortEvent, sort(mb))
	}
}

func addRow(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(presets.ParamID), false, ctx)
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
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(presets.ParamID), false, ctx)

		formKey := ctx.R.FormValue(ParamRemoveRowFormKey)
		lb := strings.LastIndex(formKey, "[")
		sliceField := formKey[0:lb]
		strIndex := formKey[lb+1 : strings.LastIndex(formKey, "]")]

		var index int
		index, err = strconv.Atoi(strIndex)
		if err != nil {
			return
		}
		presets.ContextModifiedIndexesBuilder(ctx).AppendDeleted(sliceField, index)
		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}

func sort(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		me := mb.Editing()
		obj, _ := me.FetchAndUnmarshal(ctx.R.FormValue(presets.ParamID), false, ctx)
		sortSectionFormKey := ctx.R.FormValue(ParamSortSectionFormKey)

		isStartSort := ctx.R.FormValue(ParamIsStartSort)
		if isStartSort != "1" {
			sortResult := ctx.R.FormValue(ParamSortResultFormKey)

			var result []SorterItem
			err = json.Unmarshal([]byte(sortResult), &result)
			if err != nil {
				return
			}
			var indexes []string
			for _, i := range result {
				indexes = append(indexes, fmt.Sprint(i.Index))
			}
			presets.ContextModifiedIndexesBuilder(ctx).SetSorted(sortSectionFormKey, indexes)
		}

		me.UpdateOverlayContent(ctx, &r, obj, "", nil)
		return
	}
}
