package views

import (
	"fmt"
	"time"

	"github.com/sunfmin/reflectutils"

	. "github.com/goplaid/x/vuetify"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/publish"
	h "github.com/theplant/htmlgo"
)

var timeFormat = "2006-01-02 15:04"

func ScheduleEditFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		s, ok := obj.(publish.ScheduleInterface)
		if !ok {
			return nil
		}

		//msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		//utilsMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)

		start := ""
		end := ""
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		fmt.Printf("=========EditFunc %+v  %+v\n", start, end)
		return h.Div(
			VRow(
				VCol(
					h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
				).Cols(6),
				VCol(
					h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
				).Cols(6),
			),
		)
	}
}

func ScheduleEditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	s := ctx.R.FormValue("ScheduledStartAt")
	e := ctx.R.FormValue("ScheduledEndAt")
	fmt.Printf("=========Setter obj%+v \n", obj)
	if s != "" {
		err = reflectutils.Set(obj, "ScheduledStartAt", nil)
	} else {
		startAt, err1 := time.Parse(timeFormat, s)
		if err1 == nil && !startAt.IsZero() {
			err = reflectutils.Set(obj, "ScheduledStartAt", startAt)
		}
	}
	if e != "" {
		err = reflectutils.Set(obj, "ScheduledEndAt", nil)
	} else {
		endAt, err2 := time.Parse(timeFormat, e)
		if err2 == nil && !endAt.IsZero() {
			err = reflectutils.Set(obj, "ScheduledEndAt", endAt)
		}
	}
	fmt.Printf("=========Setter obj after %+v \n", obj)
	return
}
