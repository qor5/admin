package views

import (
	"fmt"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	vx "github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/publish"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

func ScheduleEditFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		s, ok := obj.(publish.ScheduleInterface)
		if !ok {
			return nil
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		start, end := "", ""
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		return h.Div(
			VRow(
				VCol(
					vx.VXDateTimePicker().FieldName("ScheduledStartAt").Label(msgr.ScheduledStartAt).Value(start).
						TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
						ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
					//h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
				).Cols(6),
				VCol(
					vx.VXDateTimePicker().FieldName("ScheduledEndAt").Label(msgr.ScheduledEndAt).Value(end).
						TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
						ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
				//h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
				).Cols(6),
			),
		)
	}
}

func ScheduleEditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	s := ctx.R.FormValue("ScheduledStartAt")
	e := ctx.R.FormValue("ScheduledEndAt")
	if err = setTime(obj, "ScheduledStartAt", s); err != nil {
		return
	}
	if err = setTime(obj, "ScheduledEndAt", e); err != nil {
		return
	}
	return
}

var timeFormat = "2006-01-02 15:04:05 -0700"

func setTime(obj interface{}, fieldName string, val string) (err error) {
	if val == "" {
		err = reflectutils.Set(obj, fieldName, nil)
	} else {
		startAt, err1 := time.Parse(timeFormat, fmt.Sprintf("%v:00 +0900", val))
		if err1 == nil && !startAt.IsZero() {
			err = reflectutils.Set(obj, fieldName, startAt)
		}
	}
	return
}
