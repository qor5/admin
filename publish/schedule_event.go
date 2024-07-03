package publish

import (
	"errors"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	timeFormatSchedule    = "2006-01-02 15:04"
	fieldScheduledStartAt = "ScheduledStartAt"
	fieldScheduledEndAt   = "ScheduledEndAt"
)

var errInvalidObject = errors.New("invalid object")

func ScheduleTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Local().Format(timeFormatSchedule)
}

func scheduleDialog(_ *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		slug := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return r, err
		}

		sc, ok := obj.(ScheduleInterface)
		if !ok {
			return r, errInvalidObject
		}

		valStartAt := ScheduleTimeString(sc.EmbedSchedule().ScheduledStartAt)
		valEndAt := ScheduleTimeString(sc.EmbedSchedule().ScheduledEndAt)

		displayStartAtPicker := EmbedStatus(sc).Status != StatusOnline
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: web.Scope().VSlot("{locals}").Init("{schedulePublishDialog:true}").Children(
				v.VDialog().Attr("v-model", "locals.schedulePublishDialog").MaxWidth(lo.If(displayStartAtPicker, "480px").Else("280px")).Children(
					v.VCard().Children(
						v.VCardTitle().Children(
							h.Text(msgr.SchedulePublishTime),
						),
						v.VCardText().Children(
							v.VRow().Class("justify-center").Children(
								h.If(displayStartAtPicker, v.VCol().Children(
									vx.VXDateTimePicker().Attr(web.VField(fieldScheduledStartAt, valStartAt)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								)),
								v.VCol().Children(
									vx.VXDateTimePicker().Attr(web.VField(fieldScheduledEndAt, valEndAt)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
								),
							),
						),
						v.VCardActions().Children(
							v.VSpacer(),
							v.VBtn(cmsgr.Cancel).
								Variant(v.VariantFlat).
								On("click", "locals.schedulePublishDialog = false"),
							v.VBtn(cmsgr.Update).Color("primary").Attr(":disabled", "isFetching").Attr(":loading", "isFetching").
								On("click", web.Plaid().
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

func schedule(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		defer func() {
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				err = nil
			}
		}()

		if mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
			return r, perm.PermissionDenied
		}

		slug := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return r, err
		}

		sc, ok := obj.(ScheduleInterface)
		if !ok {
			return r, errInvalidObject
		}
		if err := setScheduledTimesFromForm(ctx, sc, db, mb); err != nil {
			return r, err
		}

		if err = mb.Editing().Saver(obj, slug, ctx); err != nil {
			return r, err
		}

		web.AppendRunScripts(&r, "locals.schedulePublishDialog = false")
		if mb.HasDetailing() && mb.Detailing().GetDrawer() {
			web.AppendRunScripts(&r, web.Plaid().EventFunc(actions.ReloadList).Go())
		}
		return r, nil
	}
}

func parseScheduleTimeValue(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(timeFormatSchedule, val, time.Local)
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		return nil, nil
	}
	return &t, nil
}

func setScheduledTimesFromForm(ctx *web.EventContext, sc ScheduleInterface, db *gorm.DB, mb *presets.ModelBuilder) error {
	startAt, err := parseScheduleTimeValue(ctx.R.FormValue(fieldScheduledStartAt))
	if err != nil {
		return err
	}
	endAt, err := parseScheduleTimeValue(ctx.R.FormValue(fieldScheduledEndAt))
	if err != nil {
		return err
	}

	if EmbedStatus(sc).Status == StatusOnline {
		startAt = nil
	}

	if startAt == nil && endAt == nil {
		sc.EmbedSchedule().ScheduledStartAt = startAt
		sc.EmbedSchedule().ScheduledEndAt = endAt
		return nil
	}

	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
	now := db.NowFunc()

	if startAt != nil && !startAt.After(now) {
		return errors.New(msgr.ScheduledStartAtShouldLaterThanNow)
	}

	if startAt != nil && endAt != nil {
		if !endAt.After(*startAt) {
			return errors.New(msgr.ScheduledEndAtShouldLaterThanStartAt)
		}
	}

	if endAt != nil && !endAt.After(now) {
		return errors.New(msgr.ScheduledEndAtShouldLaterThanNowOrEmpty)
	}

	if EmbedStatus(sc).Status != StatusOnline && startAt == nil {
		return errors.New(msgr.ScheduledStartAtShouldNotEmpty)
	}

	sc.EmbedSchedule().ScheduledEndAt = endAt
	if startAt == nil {
		sc.EmbedSchedule().ScheduledStartAt = startAt
		return nil
	}

	oldStartAt := sc.EmbedSchedule().ScheduledStartAt
	sc.EmbedSchedule().ScheduledStartAt = startAt

	// If there are identical StartAts, fine-tuning should be done to ensure that the times of the different versions are not equal
	if _, ok := sc.(VersionInterface); ok {
		if oldStartAt != nil && oldStartAt.Truncate(time.Minute).Equal(*startAt) {
			sc.EmbedSchedule().ScheduledStartAt = oldStartAt
			return nil
		}

		ver := mb.NewModel()
		err := utils.PrimarySluggerWhere(db, ver, sc.(presets.SlugEncoder).PrimarySlug(), "version").
			Where("scheduled_start_at >= ? AND scheduled_start_at < ?", startAt, startAt.Add(time.Minute)).
			Order("scheduled_start_at DESC").
			First(ver).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ref, _ := ver.(ScheduleInterface)
			t := ref.EmbedSchedule().ScheduledStartAt.Add(time.Microsecond)
			if t.Sub(*startAt) >= time.Minute {
				return errors.New("no enough time space to fine tuning")
			}
			sc.EmbedSchedule().ScheduledStartAt = &t
		}
	}
	return nil
}
