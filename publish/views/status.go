package views

import (
	"fmt"
	"reflect"

	"github.com/sunfmin/reflectutils"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/publish"
	"github.com/qor/qor5/utils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func Configure(b *presets.Builder, db *gorm.DB, publisher *publish.Builder) {
	b.FieldDefaults(presets.LIST).
		FieldType(publish.Status{}).
		ComponentFunc(StatusListFunc())
	b.FieldDefaults(presets.WRITE).
		FieldType(publish.Status{}).
		ComponentFunc(StatusEditFunc())

	registerEventFuncs(b.GetWebBuilder(), db, publisher)
}

func StatusListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		if s, ok := obj.(publish.StatusInterface); ok {
			return h.Td(VChip(h.Text(s.GetStatus())).Color(publish.GetStatusColor(s.GetStatus())))
		}
		return nil
	}
}

func StatusEditFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		if s, ok := obj.(publish.StatusInterface); ok {
			if s.GetStatus() == "" {
				return nil
			}
			var btn h.HTMLComponent

			switch s.GetStatus() {
			case publish.StatusDraft, publish.StatusOffline:
				btn = VBtn("Publish").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, publishEvent))
			case publish.StatusOnline:
				btn = h.Div(
					VBtn("Unpublish").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, unpublishEvent)),
					VBtn("Republish").Attr("@click", fmt.Sprintf(`locals.action="%s";locals.commonConfirmDialog = true`, publishEvent)),
				)
			}
			params := []interface{}{reflect.TypeOf(obj).String(), fmt.Sprint(reflectutils.MustGet(obj, "ID"))}
			if v, ok := obj.(publish.VersionInterface); ok {
				params = append(params, v.GetVersionName())
			}
			return h.Div(
				btn,
				utils.ConfirmDialog("Are you sure?", web.Plaid().EventFuncVar(web.Var("locals.action"), params...).Go())).
				Attr(web.InitContextLocals, `{ action: ""}`)

		}
		return nil
	}
}
