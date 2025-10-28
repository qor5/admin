package publish

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/presets"
)

func draftCountFunc(_ *presets.ModelBuilder, db *gorm.DB) presets.FieldComponentFunc {
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
			ok          bool
			err         error
			modelSchema *schema.Schema
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
			return statusChip(st.EmbedStatus().Status, msgr)
		}
		ver, _ := obj.(VersionInterface)

		var (
			statusFieldName = modelSchema.FieldsByName["Status"].DBName
			startFieldName  = modelSchema.FieldsByName["ScheduledStartAt"].DBName
		)

		var toStatus string
		var tooltip string
		if st.EmbedStatus().Status != StatusOnline {
			currentStartAt := sc.EmbedSchedule().ScheduledStartAt
			if currentStartAt != nil {
				toStatus = StatusOnline
				if ver != nil {
					tooltip = msgr.ToStatusOnline(ver.EmbedVersion().VersionName, ScheduleTimeString(currentStartAt))
				}
			}
		} else {
			objNextStart := reflect.New(modelSchema.ModelType).Interface()
			err := g().Where(fmt.Sprintf("%s <> ? AND %s > ?", statusFieldName, startFieldName), StatusOnline, nowTime).
				Order(startFieldName).Limit(1).Scan(&objNextStart).Error
			if err != nil {
				return
			}
			scNext := objNextStart.(ScheduleInterface).EmbedSchedule()

			currentEndAt := sc.EmbedSchedule().ScheduledEndAt
			if scNext.ScheduledStartAt != nil && (currentEndAt == nil || !scNext.ScheduledStartAt.After(*currentEndAt)) {
				toStatus = statusNext

				scNextStartAtFormat := ScheduleTimeString(scNext.ScheduledStartAt)
				if ver != nil {
					tooltip = fmt.Sprintf("%s\n%s",
						msgr.ToStatusOffline(ver.EmbedVersion().VersionName, scNextStartAtFormat),
						msgr.ToStatusOnline(objNextStart.(VersionInterface).EmbedVersion().VersionName, scNextStartAtFormat),
					)
				}
			} else if currentEndAt != nil && !currentEndAt.Before(nowTime) {
				toStatus = StatusOffline
				if ver != nil {
					tooltip = msgr.ToStatusOffline(ver.EmbedVersion().VersionName, ScheduleTimeString(currentEndAt))
				}
			}
		}
		compo := liveChips(st.EmbedStatus().Status, toStatus, msgr)
		if tooltip != "" {
			compo = h.Div().Class("d-flex").Children(
				h.Div().Children(
					VTooltip().Activator("parent").Location(LocationTop).Children(
						h.Div().Class("text-body-2").Style("white-space: pre-wrap").Text(fmt.Sprintf(`{{%q}}`, tooltip)),
					),
					compo,
				),
			)
		}
		return compo
	}
}

func StatusListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		if s, ok := obj.(StatusInterface); ok {
			return h.Td(statusChip(s.EmbedStatus().Status, msgr))
		}
		return nil
	}
}

func liveChip(status string, isScheduled bool, msgr *Messages, forceMarked bool) *VChipBuilder {
	label, color := GetStatusLabelColor(status, msgr)
	chip := VChip(
		h.If(status == StatusOnline || forceMarked,
			VIcon("mdi-radiobox-marked").Size(SizeSmall).Class("mr-1"),
		),
		h.Span(label),
		h.If(isScheduled, VIcon("mdi-menu-right").Size(SizeSmall).Class("ml-1")),
	).Color(color).Density(DensityComfortable).Tile(true).Class("px-1")
	if !isScheduled {
		return chip
	}
	return chip
}

func statusChip(status string, msgr *Messages) *VChipBuilder {
	return liveChip(status, false, msgr, false).Class("rounded")
}

const statusNext = "Next"

func liveChips(status, toStatus string, msgr *Messages) h.HTMLComponent {
	if toStatus != "" {
		return h.Components(
			liveChip(status, true, msgr, false).Class("rounded-s"),
			liveChip(toStatus, false, msgr, toStatus == statusNext).Class("rounded-e"),
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
	case statusNext:
		return msgr.StatusNext, ColorSuccess
	default:
		return status, ColorSuccess
	}
}
