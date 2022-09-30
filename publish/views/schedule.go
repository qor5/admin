package views

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	vx "github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/publish"
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

		var start, end int64
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Unix()
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Unix()
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
							).Bottom(true).MaxWidth(285),
						).Class("v-expansion-panel-header__icon"),
					).Class("v-expansion-panel-header"),
				),

				VExpansionPanels(
					VExpansionPanel(
						VExpansionPanelHeader(
							VRow(
								VCol(
									h.Text(msgr.SchedulePublishTime),
								).Cols(4),
								VCol(
									VFadeTransition(
										h.Span(msgr.WhenDoYouWantToPublish).Attr("v-if", "open"),
										VRow(
											VCol(
												h.If(start <= 0, h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledStartAt, msgr.NotSet))).
													Else(
														h.Text(fmt.Sprintf("%v: ", msgr.ScheduledStartAt)),
														vx.VXDateTimeFormatter().Value(start),
													),
											).Cols(6),
											VCol(
												h.If(end <= 0, h.Text(fmt.Sprintf("%v: %v ", msgr.ScheduledEndAt, msgr.NotSet))).
													Else(
														h.Text(fmt.Sprintf("%v: ", msgr.ScheduledEndAt)),
														vx.VXDateTimeFormatter().Value(end),
													),
											).Cols(6),
										).NoGutters(true).Attr("v-else").Attr(`style="width: 100%"`),
									).LeaveAbsolute(true),
								).Cols(8).Class("text--secondary"),
							).NoGutters(true),
						).Attr("v-slot", "{ open }"),
						VExpansionPanelContent(
							VRow(
								VCol(
									vx.VXDateTimePicker().FieldName("ScheduledStartAt").Label(msgr.ScheduledStartAt).Value(start).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									//h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
								).Cols(6),
								VCol(
									vx.VXDateTimePicker().FieldName("ScheduledEndAt").Label(msgr.ScheduledEndAt).Value(end).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									//h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
								).Cols(6),
							),
						),
					),
				).Flat(true).Hover(true),
			),
			h.Br(),
		)
	}
}

func ScheduleEditSetterFunc(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
	s := ctx.R.FormValue("ScheduledStartAt")
	e := ctx.R.FormValue("ScheduledEndAt")
	if err = setTime(obj, "ScheduledStartAt", s); err != nil {
		return
	}
	if err = setTime(obj, "ScheduledEndAt", e); err != nil {
		return
	}
	return
}

var timeFormat = "2006-01-02 15:04:05 -0700"

func setTime(obj interface{}, fieldName string, val string) (err error) {
	if val == "" {
		err = reflectutils.Set(obj, fieldName, nil)
	} else {
		uts, err1 := strconv.ParseInt(val, 10, 64)
		if err1 != nil {
			return
		}
		startAt := time.Unix(uts, 0)
		if startAt.IsZero() {
			return
		}
		err = reflectutils.Set(obj, fieldName, startAt)
	}
	return
}
