package publish

import (
	"errors"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	v "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func schedulePublishDialog(_ *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		s, ok := obj.(ScheduleInterface)
		if !ok {
			return
		}

		var start, end string
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		updateBtn := v.VBtn(cmsgr.Update).
			Color("primary").
			Attr(":disabled", "isFetching").
			Attr(":loading", "isFetching").
			Attr("@click", web.Plaid().
				EventFunc(eventSchedulePublish).
				// Queries(queries).
				Query(presets.ParamID, paramID).
				Query(presets.ParamOverlay, actions.Dialog).
				URL(mb.Info().ListingHref()).
				Go())

		dialogWidth := "480px"
		dislayStartAtPicker := s.GetStatus() != StatusOnline
		if !dislayStartAtPicker {
			dialogWidth = "280px"
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: web.Scope(
				v.VDialog(
					v.VCard(
						v.VCardTitle(h.Text("Schedule Publish Time")),
						v.VCardText(
							v.VRow(
								h.If(dislayStartAtPicker, v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", start)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								)),
								v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", end)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								),
							).Class("justify-center"),
						),
						v.VCardActions(
							v.VSpacer(),
							updateBtn,
						),
					),
				).MaxWidth(dialogWidth).
					Attr("v-model", "locals.schedulePublishDialog"),
			).Init("{schedulePublishDialog:true}").VSlot("{locals}"),
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
			return r, errors.New("invalid object")
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

var scheduleTimeFormat = "2006-01-02 15:04"

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
