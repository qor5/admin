package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

func ScheduleEditFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		s, ok := obj.(publish.ScheduleInterface)
		if !ok {
			return nil
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		var start, end string
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}

		publishedAt, unpublishedAt := "", ""
		if s.GetPublishedAt() != nil {
			publishedAt = s.GetPublishedAt().Format("2006-01-02 15:04")
		}
		if s.GetUnPublishedAt() != nil {
			unpublishedAt = s.GetUnPublishedAt().Format("2006-01-02 15:04")
		}
		return h.Div(
			VCard(
				h.If(s.GetStatus() != "",
					h.Div(
						VRow(
							VCol(
								h.Text(msgr.ActualPublishTime),
							).Cols(4),
							VCol(
								VRow(
									VCol(
										h.If(publishedAt == "", h.Text(fmt.Sprintf("%v: %v ", msgr.PublishedAt, msgr.NotSet))).Else(h.Text(fmt.Sprintf("%v: %v ", msgr.PublishedAt, publishedAt))),
									).Cols(6),
									VCol(
										h.If(unpublishedAt == "", h.Text(fmt.Sprintf("%v: %v ", msgr.UnPublishedAt, msgr.NotSet))).Else(h.Text(fmt.Sprintf("%v: %v ", msgr.UnPublishedAt, unpublishedAt))),
									).Cols(6),
								).NoGutters(true).Attr(`style="width: 100%"`),
							).Cols(8).Class("text--secondary"),
						).NoGutters(true),
						h.Div(
							VTooltip(
								web.Slot(
									VIcon("publish").Attr("v-bind", "attrs", "v-on", "on"),
								).Name("activator").Scope("{ on, attrs }"),
								h.Span(strings.ReplaceAll(msgr.PublishScheduleTip, "{SchedulePublishTime}", msgr.SchedulePublishTime)),
							).Location(LocationBottom).MaxWidth(285),
						).Class("v-expansion-panel-header__icon"),
					).Class("v-expansion-panel-header"),
				),

				VExpansionPanels(
					VExpansionPanel(
						VExpansionPanelTitle(
							VRow(
								VCol(
									h.Text(msgr.SchedulePublishTime),
								).Cols(4),
								VCol(
									VFadeTransition(
										h.Span(msgr.WhenDoYouWantToPublish).Attr("v-if", "open"),
										VRow(
											VCol(
												h.If(start == "", h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledStartAt, msgr.NotSet))).Else(h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledStartAt, start))),
											).Cols(6),
											VCol(
												h.If(end == "", h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledEndAt, msgr.NotSet))).Else(h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledEndAt, end))),
											).Cols(6),
										).NoGutters(true).Attr("v-else").Attr(`style="width: 100%"`),
									).LeaveAbsolute(true),
								).Cols(8).Class("text--secondary"),
							).NoGutters(true),
						).Attr("v-slot", "{ open }"),
						VExpansionPanelText(
							VRow(
								VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", start)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
								).Cols(6),
								VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", end)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
								).Cols(6),
							),
						),
					),
				),
			),
			h.Br(),
		)
	}
}

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
