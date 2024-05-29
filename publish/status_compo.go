package publish

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
)

func draftCountFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var count int64
		modelSchema, err := schema.Parse(obj, &sync.Map{}, db.NamingStrategy)
		if err != nil {
			return h.Td(h.Text("0"))
		}
		setPrimaryKeysConditionWithoutVersion(db.Model(reflect.New(modelSchema.ModelType).Interface()), obj, modelSchema).
			Where("status = ?", StatusDraft).Count(&count)

		return h.Td(h.Text(fmt.Sprint(count)))
	}
}

func liveFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (comp h.HTMLComponent) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		var (
			ok                         bool
			err                        error
			statusInterface            StatusInterface
			modelSchema                *schema.Schema
			count                      int64
			status                     string
			scheduleStart, scheduleEnd Schedule
		)
		defer func() {
			if err != nil {
				comp = h.Td(h.Text("-"))
				return
			}
			comp = h.Td(comp)
		}()
		if modelSchema, err = schema.Parse(obj, &sync.Map{}, db.NamingStrategy); err != nil {
			return
		}

		var (
			g = func() *gorm.DB {
				return setPrimaryKeysConditionWithoutVersion(db.Model(reflect.New(modelSchema.ModelType).Interface()), obj, modelSchema)
			}
			nowTime = db.NowFunc()
		)
		if statusInterface, ok = obj.(StatusInterface); !ok {
			err = errors.New("ErrorModel")
			return
		}
		status = statusInterface.GetStatus()
		if _, ok = obj.(ScheduleInterface); !ok {
			comp, err = noScheduledLive(g, msgr)
			return
		}
		var (
			startFieldName = modelSchema.FieldsByName["ScheduledStartAt"].DBName
			endFieldName   = modelSchema.FieldsByName["ScheduledEndAt"].DBName
		)
		if err = g().Where(fmt.Sprintf(`%s >= @nowTime or %s >= @nowTime`, startFieldName, endFieldName), sql.Named("nowTime", nowTime)).Count(&count).Error; err != nil {
			return
		}
		if count == 0 {
			comp, err = noScheduledLive(g, msgr)
			return
		}

		g().Select(startFieldName).Where(fmt.Sprintf(`%s >= ?`, startFieldName), nowTime).Order(startFieldName).Limit(1).Scan(&scheduleStart)
		g().Select(endFieldName).Where(fmt.Sprintf(`%s >= ?`, endFieldName), nowTime).Order(endFieldName).Limit(1).Scan(&scheduleEnd)
		if scheduleStart.ScheduledStartAt == nil && scheduleEnd.ScheduledEndAt == nil {
			err = errors.New("dbError")
			return
		} else if scheduleStart.ScheduledStartAt != nil && scheduleEnd.ScheduledEndAt == nil {
			comp = scheduledLive(status, msgr)
			return
		} else if scheduleStart.ScheduledStartAt == nil && scheduleEnd.ScheduledEndAt != nil {
			comp = scheduledLiveOffline(status, msgr)
			return
		} else {
			if scheduleEnd.ScheduledEndAt.Before(*scheduleStart.ScheduledStartAt) {
				comp = scheduledLiveOffline(status, msgr)
			} else {
				comp = scheduledLive(status, msgr)
			}
		}
		return
	}
}

func StatusListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		if s, ok := obj.(StatusInterface); ok {
			return h.Td().Children(
				VChip().Theme(ThemeDark).
					Label(true).
					Color(GetStatusColor(s.GetStatus())).
					Children(
						h.Text(GetStatusText(s.GetStatus(), msgr)),
					),
			)
		}
		return nil
	}
}

func GetStatusColor(status string) string {
	switch status {
	case StatusDraft:
		return "warning"
	case StatusOnline:
		return "success"
	case StatusOffline:
		return "secondary"
	}
	return ""
}

func noScheduledLive(g func() *gorm.DB, msgr *Messages) (comp h.HTMLComponent, err error) {
	var count int64
	if err = g().Where("status = ?", StatusOnline).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return VChip().Text(msgr.StatusOnline).Color(ColorSuccess), nil
	}
	if err = g().Where("status = ?", StatusOffline).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return VChip().Text(msgr.StatusOffline).Color(ColorWarning), nil
	}
	return VChip().Text(msgr.StatusDraft).Color(ColorSecondary), nil
}

func scheduledLive(status string, msgr *Messages) (comp h.HTMLComponent) {
	if status == StatusDraft {
		// draft -> online
		comp = VChip().Text(fmt.Sprintf(`%s->%s`, msgr.StatusDraft, msgr.StatusOnline))
	} else if status == StatusOffline {
		comp = VChip().Text(fmt.Sprintf(`%s->%s`, msgr.StatusOffline, msgr.StatusOnline))
	} else {
		comp = VChip().Text(fmt.Sprintf(`%s->%s`, msgr.StatusOnline, msgr.StatusOnline))
	}
	return
}

func scheduledLiveOffline(status string, msgr *Messages) h.HTMLComponent {
	if status == StatusDraft {
		return VChip().Text(msgr.StatusDraft).Color(ColorSecondary)
	} else if status == StatusOffline {
		return VChip().Text(msgr.StatusOffline).Color(ColorSecondary)
	}
	return VChip().Text(fmt.Sprintf(`%s->%s`, msgr.StatusOnline, msgr.StatusOffline))
}
