package publish

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/samber/lo"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
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
		maxWidthStr := lo.If(displayStartAtPicker, "480").Else("280")
		maxWidth, err := strconv.Atoi(maxWidthStr)
		if err != nil {
			panic("convert string to int error")
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: web.Scope().VSlot("{locals}").Init("{schedulePublishDialog:true}").Children(
				vx.VXDialog(
					v.VRow().Class("justify-center").Children(
						h.If(displayStartAtPicker, v.VCol().Children(
							vx.VXDatepicker().Type("datetimepicker").
								Format("YYYY-MM-DD HH:mm").
								Clearable(true).
								Attr(web.VField(fieldScheduledStartAt, valStartAt)...).
								Label(msgr.ScheduledStartAt),
						)),
						v.VCol().Children(
							vx.VXDatepicker().Type("datetimepicker").
								Format("YYYY-MM-DD HH:mm").
								Clearable(true).
								Attr(web.VField(fieldScheduledEndAt, valEndAt)...).
								Label(msgr.ScheduledEndAt),
						),
					),
				).Attr("v-model", "locals.schedulePublishDialog").
					Title(msgr.SchedulePublishTime).
					ContentHeight(108).
					CancelText(cmsgr.Cancel).
					OkText(cmsgr.Update).
					Attr(":disable-ok", "isFetching").
					Attr("@click:ok", fmt.Sprintf(`({isLoading}) => {
						isLoading.value = isFetching;
						%s
					}`, web.Plaid().EventFunc(eventSchedulePublish).Query(presets.ParamID, slug).URL(mb.Info().ListingHref()).Go())).
					MaxWidth(maxWidth),
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

		slug := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return r, err
		}
		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, PermPublish, PermUnpublish, PermSchedule) {
			return r, perm.PermissionDenied
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
		r.Emit(mb.NotifModelsUpdated(), presets.PayloadModelsUpdated{
			Ids:    []string{slug},
			Models: map[string]any{slug: obj},
		})
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
