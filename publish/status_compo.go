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
			modelSchema                *schema.Schema
			count                      int64
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
				return setPrimaryKeysConditionWithoutFields(db.Model(reflect.New(modelSchema.ModelType).Interface()), obj, modelSchema, "Version", "LocaleCode")
			}
			nowTime = db.NowFunc()
		)
		if _, ok = obj.(StatusInterface); !ok {
			err = errors.New("ErrorModel")
			return
		}
		if _, ok = obj.(ScheduleInterface); !ok {
			comp, err = liveStatusColumn(g, msgr, "")
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
			comp, err = liveStatusColumn(g, msgr, "")
			return
		}

		g().Select(startFieldName).Where(fmt.Sprintf(`%s >= ?`, startFieldName), nowTime).Order(startFieldName).Limit(1).Scan(&scheduleStart)
		g().Select(endFieldName).Where(fmt.Sprintf(`%s >= ?`, endFieldName), nowTime).Order(endFieldName).Limit(1).Scan(&scheduleEnd)
		if scheduleStart.ScheduledStartAt == nil && scheduleEnd.ScheduledEndAt == nil {
			err = errors.New("dbError")
			return
		} else if scheduleStart.ScheduledStartAt != nil && scheduleEnd.ScheduledEndAt == nil {
			comp, err = liveStatusColumn(g, msgr, StatusOnline)
			return
		} else if scheduleStart.ScheduledStartAt == nil && scheduleEnd.ScheduledEndAt != nil {
			comp, err = liveStatusColumn(g, msgr, StatusOffline)
			return
		} else {
			if scheduleEnd.ScheduledEndAt.Before(*scheduleStart.ScheduledStartAt) {
				comp, err = liveStatusColumn(g, msgr, StatusOffline)
			} else {
				comp, err = liveStatusColumn(g, msgr, StatusOnline)
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

func liveStatusColumn(g func() *gorm.DB, msgr *Messages, toStatus string) (comp h.HTMLComponent, err error) {
	var count int64
	if err = g().Where("status = ?", StatusOnline).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return liveChips(StatusOnline, toStatus, msgr), nil
	}
	if err = g().Where("status = ?", StatusOffline).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		if toStatus == StatusOffline {
			return liveChips(StatusOffline, "", msgr), nil
		}
		return liveChips(StatusOffline, toStatus, msgr), nil
	}
	if toStatus == StatusOffline {
		return liveChips(StatusDraft, "", msgr), nil
	}
	return liveChips(StatusDraft, toStatus, msgr), nil
}

func liveChip(status string, isScheduled bool, msgr *Messages) h.HTMLComponent {
	label, color := GetStatusLabelColor(status, msgr)
	return VChip(
		h.If(status == StatusOnline,
			web.Slot(
				VRadio().Density(DensityCompact).ModelValue(true).Readonly(true).Ripple(false).Class("ml-n2"),
			).Name(VSlotPrepend),
		),
		h.Span(label),
		h.If(isScheduled, VIcon("mdi-menu-right").Size(SizeSmall).Class("ml-2")),
	).Color(color).Density(DensityCompact).Tile(true).Class("px-2")
}

func liveChips(status string, toStatus string, msgr *Messages) h.HTMLComponent {
	if status == toStatus && status == StatusOnline {
		toStatus = "+1"
	}
	return h.Components(
		liveChip(status, toStatus != "", msgr),
		h.If(toStatus != "", liveChip(toStatus, false, msgr)),
	)
}

func GetStatusLabelColor(status string, msgr *Messages) (label, color string) {
	switch status {
	case StatusOnline:
		return msgr.StatusOnline, ColorSuccess
	case StatusOffline:
		return msgr.StatusOffline, ColorSecondary
	case StatusDraft:
		return msgr.StatusDraft, ColorWarning
	default:
		return status, ColorSuccess
	}
}
