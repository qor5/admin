package publish

import (
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
		paramID := ctx.R.FormValue(presets.ParamID)
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

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: web.Scope(
				v.VDialog(
					v.VCard(
						v.VCardTitle(h.Text("Schedule Publish Time")),
						v.VCardText(
							v.VRow(
								v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", start)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
								).Cols(6),
								v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", end)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
								).Cols(6),
							),
						),
						v.VCardActions(
							v.VSpacer(),
							updateBtn,
						),
					),
				).MaxWidth("480px").
					Attr("v-model", "locals.schedulePublishDialogV2"),
			).Init("{schedulePublishDialogV2:true}").VSlot("{locals}"),
		})
		return
	}
}

func schedulePublish(_ *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		err = ScheduleEditSetterFunc(obj, nil, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		err = mb.Editing().Saver(obj, paramID, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PortalSchedulePublishDialog,
			Body: nil,
		})
		return
	}
}
