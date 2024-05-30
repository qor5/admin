package publish

import (
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
			ok            bool
			err           error
			modelSchema   *schema.Schema
			scheduleStart Schedule
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
		st, ok := obj.(StatusInterface)
		if !ok {
			err = errors.New("ErrorModel")
			return
		}
		sc, ok := obj.(ScheduleInterface)
		if !ok {
			return statusChip(st.GetStatus(), msgr)
		}

		var (
			statusFieldName = modelSchema.FieldsByName["Status"].DBName
			startFieldName  = modelSchema.FieldsByName["ScheduledStartAt"].DBName
		)

		var toStatus string
		if st.GetStatus() != StatusOnline {
			if sc.GetScheduledStartAt() != nil {
				toStatus = StatusOnline
			}
		} else {
			err := g().Select(startFieldName).Where(fmt.Sprintf("%s <> ? AND %s > ?", statusFieldName, startFieldName), StatusOnline, nowTime).Order(startFieldName).Limit(1).Scan(&scheduleStart).Error
			if err != nil {
				return
			}
			currentEndAt := sc.GetScheduledEndAt()
			if scheduleStart.ScheduledStartAt != nil && (currentEndAt == nil || !scheduleStart.ScheduledStartAt.After(*currentEndAt)) {
				toStatus = "+1"
			} else if currentEndAt != nil && !currentEndAt.Before(nowTime) {
				toStatus = StatusOffline
			}
		}
		return liveChips(st.GetStatus(), toStatus, msgr)
	}
}

func StatusListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		if s, ok := obj.(StatusInterface); ok {
			return h.Td(statusChip(s.GetStatus(), msgr))
		}
		return nil
	}
}

func liveChip(status string, isScheduled bool, msgr *Messages) *VChipBuilder {
	label, color := GetStatusLabelColor(status, msgr)
	chip := VChip(
		h.If(status == StatusOnline,
			VIcon("mdi-radiobox-marked").Size(SizeSmall).Class("mr-1"),
		),
		h.Span(label),
		h.If(isScheduled, VIcon("mdi-menu-right").Size(SizeSmall).Class("ml-1")),
	).Color(color).Density(DensityCompact).Tile(true).Class("px-1")
	if !isScheduled {
		return chip
	}
	return chip.Class("rounded-s-lg")
}

func statusChip(status string, msgr *Messages) *VChipBuilder {
	return liveChip(status, false, msgr).Class("rounded")
}

func liveChips(status string, toStatus string, msgr *Messages) h.HTMLComponent {
	if toStatus != "" {
		return h.Components(
			liveChip(status, true, msgr).Class("rounded-s"),
			liveChip(toStatus, false, msgr).Class("rounded-e"),
		)
	}
	return statusChip(status, msgr)
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
