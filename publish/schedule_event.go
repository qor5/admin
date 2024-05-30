package publish

import (
	"errors"
	"time"

	"github.com/qor5/admin/v3/presets"
	v "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const scheduleTimeFormat = "2006-01-02 15:04"

var errInvalidObject = errors.New("invalid object")

func schedulePublishDialog(_ *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		obj := mb.NewModel()
		sc, ok := obj.(ScheduleInterface)
		if !ok {
			return r, errInvalidObject
		}

		slug := ctx.Param(presets.ParamID)
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return
		}

		var valStartAt, valEndAt string
		if sc.GetScheduledStartAt() != nil {
			valStartAt = sc.GetScheduledStartAt().Format(scheduleTimeFormat)
		}
		if sc.GetScheduledEndAt() != nil {
			valEndAt = sc.GetScheduledEndAt().Format(scheduleTimeFormat)
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		dislayStartAtPicker := sc.GetStatus() != StatusOnline
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: web.Scope().VSlot("{locals}").Init("{schedulePublishDialog:true}").Children(
				v.VDialog().Attr("v-model", "locals.schedulePublishDialog").MaxWidth(lo.If(dislayStartAtPicker, "480px").Else("280px")).Children(
					v.VCard().Children(
						v.VCardTitle().Children(
							h.Text(msgr.SchedulePublishTime),
						),
						v.VCardText().Children(
							v.VRow().Class("justify-center").Children(
								h.If(dislayStartAtPicker, v.VCol().Children(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", valStartAt)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								)),
								v.VCol().Children(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", valEndAt)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								),
							),
						),
						v.VCardActions().Children(
							v.VSpacer(),
							v.VBtn(cmsgr.Update).Color("primary").Attr(":disabled", "isFetching").Attr(":loading", "isFetching").Attr("@click", web.Plaid().
								EventFunc(eventSchedulePublish).
								Query(presets.ParamID, slug).
								URL(mb.Info().ListingHref()).
								Go(),
							),
						),
					),
				),
			),
		})
		return
	}
}

func wrapEventFuncWithShowError(f web.EventFunc) web.EventFunc {
	return func(ctx *web.EventContext) (web.EventResponse, error) {
		r, err := f(ctx)
		if err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
		}
		return r, nil
	}
}

func schedulePublish(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return wrapEventFuncWithShowError(func(ctx *web.EventContext) (web.EventResponse, error) {
		var r web.EventResponse

		obj := mb.NewModel()
		sc, ok := obj.(ScheduleInterface)
		if !ok {
			return r, errInvalidObject
		}

		slug := ctx.Param(presets.ParamID)
		obj, err := mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return r, err
		}
		if err := setScheduledTimesFromForm(ctx, sc, db); err != nil {
			return r, err
		}
		// TODO: If there are identical StartAts, fine-tuning should be done to ensure that the times of the different versions are not equal

		if err = mb.Editing().Saver(obj, slug, ctx); err != nil {
			return r, err
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: nil,
		})
		return r, nil
	})
}

func parseScheduleTimeValue(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}
	t, err := time.Parse(scheduleTimeFormat, val)
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		return nil, nil
	}
	return &t, nil
}

func setScheduledTimesFromForm(ctx *web.EventContext, sc ScheduleInterface, db *gorm.DB) error {
	startAt, err := parseScheduleTimeValue(ctx.R.FormValue("ScheduledStartAt"))
	if err != nil {
		return err
	}
	endAt, err := parseScheduleTimeValue(ctx.R.FormValue("ScheduledEndAt"))
	if err != nil {
		return err
	}

	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
	now := db.NowFunc()

	if startAt != nil && startAt.Before(now) {
		return errors.New(msgr.ScheduledStartAtShouldLaterThanNow)
	}

	if startAt != nil && endAt != nil {
		if !startAt.Before(*endAt) {
			return errors.New(msgr.ScheduledEndAtShouldLaterThanStartAt)
		}
	}

	if endAt != nil && endAt.Before(now) {
		return errors.New(msgr.ScheduledEndAtShouldLaterThanNowOrEmpty)
	}

	if sc.GetStatus() != StatusOnline && startAt == nil {
		return errors.New(msgr.ScheduledStartAtShouldNotEmpty)
	}

	sc.SetScheduledStartAt(startAt)
	sc.SetScheduledEndAt(endAt)
	return nil
}
