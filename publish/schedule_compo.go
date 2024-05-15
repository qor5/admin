package publish

import (
	"fmt"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
)

func ScheduleEditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	_, exist := ctx.R.Form["ScheduledStartAt"]
	if exist {
		s := ctx.R.FormValue("ScheduledStartAt")
		if err = setTime(obj, "ScheduledStartAt", s); err != nil {
			return
		}
	}

	_, exist = ctx.R.Form["ScheduledEndAt"]
	if exist {
		e := ctx.R.FormValue("ScheduledEndAt")
		if err = setTime(obj, "ScheduledEndAt", e); err != nil {
			return
		}

	}

	return
}

var timeFormat = "2006-01-02 15:04:05"

func setTime(obj interface{}, fieldName string, val string) (err error) {
	if val == "" {
		err = reflectutils.Set(obj, fieldName, nil)
	} else {
		startAt, err1 := time.ParseInLocation(timeFormat, fmt.Sprintf("%v:00", val), time.Local)
		if err1 == nil && !startAt.IsZero() {
			err = reflectutils.Set(obj, fieldName, startAt)
		}
	}
	return
}
