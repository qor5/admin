package emailbuilder

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	v "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

// Campaign status constants
const (
	StatusDraft     = "draft"
	StatusSent      = "sent"
	StatusScheduled = "scheduled"
)

// Schedule frequency constants
const (
	FrequencyNone    = "none"
	FrequencyDaily   = "daily"
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
)

type (
	EmailCampaign struct {
		gorm.Model
		EmailDetail
		UTM
		Schedule

		To     string
		Status string // StatusDraft, StatusSent, StatusScheduled
	}

	Schedule struct {
		Enabled    bool
		Frequency  string // FrequencyNone, FrequencyDaily, FrequencyWeekly, FrequencyMonthly
		StartTime  time.Time
		EndTime    time.Time
		RetryCount int   // Number of retries on failure
		JobID      int64 // Job ID for scheduler reference
	}

	UTM struct {
		// UTM Parameters
		Source   string
		Medium   string
		Campaign string
		Term     string
		Content  string
	}
)

func (c *EmailCampaign) PrimarySlug() string {
	return fmt.Sprintf("%d", c.ID)
}

func (c *EmailCampaign) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"id": slug,
	}
}

func DefaultMailCampaign(pb *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	mb := pb.Model(&EmailCampaign{}).Label("Email Campaigns")

	configureListing(mb)

	// Configure detail page
	dp := mb.Detailing("From", "To", "Subject", EmailDetailField, "UTM", "Schedule")
	// dp := mb.Detailing(EmailDetailField, "Recipient", "Schedule")
	// Add sections to detail page in the desired order (UTM section above Schedule section)
	dp.Section(configureFromSection(mb, db))
	dp.Section(configureSegmentSection(mb, db))
	dp.Section(configureSubjectSection(mb, db))
	dp.Section(configureUTMParametersSection(mb, db))
	dp.Section(configureScheduleSection(mb, db))

	configureEditing(mb)
	return mb
}

func configureListing(mb *presets.ModelBuilder) {
	// Configure listing page
	listing := mb.Listing("ID", "Status", "CreatedAt", "UpdatedAt")

	// Customize the listing display
	listing.Field("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		campaign := obj.(*EmailCampaign)
		var color string
		var text string

		switch campaign.Status {
		case StatusDraft:
			color = "warning"
			text = "Draft"
		case StatusSent:
			color = "success"
			text = "Sent"
		case StatusScheduled:
			color = "info"
			text = "Scheduled"
		default:
			color = "warning"
			text = "Draft"
		}

		return h.Td(vx.VXChip(text).Color(color).Size(v.SizeSmall).Class("text-capitalize"))
	})

	listing.Field("CreatedAt").Label("Create On").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.Text(field.Value(obj).(time.Time).Local().Format("15:04 01/02/2006")))
	})

	listing.Field("UpdatedAt").Label("Last Updated").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.Text(field.Value(obj).(time.Time).Local().Format("15:04 01/02/2006")))
	})

	// Configure row menu actions
	rowMenu := listing.RowMenu()
	// Edit action
	rowMenu.RowMenuItem("Edit").Icon("mdi-pencil").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		return v.VListItem().Attr("@click", web.Plaid().
			EventFunc(actions.Edit).
			Query(presets.ParamID, id).
			Go()).Children(
			web.Slot(
				v.VIcon("mdi-pencil"),
			).Name("prepend"),
			v.VListItemTitle(h.Text("Edit")),
		)
	})

	// Report action
	rowMenu.RowMenuItem("Report").Icon("mdi-file-document-outline").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		return v.VListItem().Attr("@click", web.Plaid().
			EventFunc("report").
			Query(presets.ParamID, id).
			Go()).Children(
			web.Slot(
				v.VIcon("mdi-file-document-outline"),
			).Name("prepend"),
			v.VListItemTitle(h.Text("Report")),
		)
	})

	// Delete action
	rowMenu.RowMenuItem("Delete").Icon("mdi-delete").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		return v.VListItem().Class("text-error").Attr("@click", web.Plaid().
			EventFunc(actions.DoDelete).
			Query(presets.ParamID, id).
			Go()).Children(
			web.Slot(
				v.VIcon("mdi-delete"),
			).Name("prepend"),
			v.VListItemTitle(h.Text("Delete")),
		)
	})

	// Configure filter using the correct SearchFunc signature
	listing.WrapSearchFunc(func(in presets.SearchFunc) presets.SearchFunc {
		return func(ctx *web.EventContext, params *presets.SearchParams) (result *presets.SearchResult, err error) {
			if status := ctx.R.URL.Query().Get("status"); status != "" {
				params.SQLConditions = append(params.SQLConditions, &presets.SQLCondition{
					Query: "status = ?",
					Args:  []interface{}{status},
				})
			}
			return in(ctx, params)
		}
	})

	mb.RegisterEventFunc("report", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.Param(presets.ParamID)
		if id == "" {
			return
		}

		// Logic to show campaign report
		// ...

		return
	})
}

func configureEditing(mb *presets.ModelBuilder) {
	// Configure editing page for both creation and editing
	mb.Editing("Subject", "JSONBody", "HTMLBody").Creating("Subject", TemplateSelectionFiled)
}

func configureFromSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	section := presets.NewSectionBuilder(mb, "From").Editing("From")
	section.ViewingField("From").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return vx.VXTextField().ReadOnly(true).Text(GetFromAddress())
		}
	})
	section.ComponentEditBtnFunc(func(_ interface{}, _ *web.EventContext) bool {
		return false // return false to disable edit button.
	})
	return section
}

func configureSegmentSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create recipient section
	section := presets.NewSectionBuilder(mb, "To").Editing("To")
	section.ViewingField("To").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			c := obj.(*EmailCampaign)
			if c.To != "" {
				return vx.VXTextField().ReadOnly(true).Text(c.To)
			} else {
				return vx.VXTextField().ReadOnly(true).Text("Please select a segment")
			}
		}
	})

	// Portal name for the recipient info
	const (
		segmentInfoPortal     = "recipientInfoPortal"
		eventFetchSegmentInfo = "fetchSegmentInfo"
	)

	// Register fetch recipient info event handler
	mb.RegisterEventFunc(eventFetchSegmentInfo, func(ctx *web.EventContext) (r web.EventResponse, err error) {
		// Get recipient value from form
		to := ctx.R.FormValue("To")
		if to == "" {
			return
		}
		// Update the portal with the info banner
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: segmentInfoPortal,
			Body: createSegmentInfoBanner(to),
		})
		return
	})

	section.EditingField("To").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			// Get the current selected value
			campaign, ok := obj.(*EmailCampaign)
			if !ok {
				campaign = &EmailCampaign{}
			}

			// Build components
			var components []h.HTMLComponent

			// Add select field
			selectField := presets.SelectField(obj, field, ctx).
				Items([]string{"segmentationA", "segmentationB", "segmentationC", "segmentationD"}).
				Attr("@update:model-value", web.Plaid().EventFunc(eventFetchSegmentInfo).Go()).
				Label("") // explicitly set label to empty string

			components = append(components, selectField)

			// Create portal for recipient info
			// For the initial state, we'll directly include the banner in the portal
			var portalContent h.HTMLComponent
			if campaign.To != "" {
				portalContent = createSegmentInfoBanner(campaign.To)
			} else {
				// Empty div if no recipient selected
				portalContent = h.Div()
			}

			// Add portal with initial content
			infoContainer := h.Div().
				Class("segment-info-container mt-3").
				Children(
					web.Portal(portalContent).Name(segmentInfoPortal),
				)

			components = append(components, infoContainer)
			return h.Div(components...)
		}
	})

	return section
}

// createSegmentInfoBanner creates an info banner showing user count and estimated time
func createSegmentInfoBanner(segment string) h.HTMLComponent {
	// Get user count
	userCount := getUserCountForSegment(segment)

	// Calculate estimated time (approx 1 minute per 400 emails)
	minutes := (userCount + 399) / 400
	timeEstimate := fmt.Sprintf("%d", minutes)

	// Determine if we need to show a warning (if more than 5 minutes)
	alertType := "info"
	if minutes > 5 {
		alertType = "warning"
	}

	// Create info banner
	infoText := fmt.Sprintf("You are about to send %d emails, which typically takes around %s minutes",
		userCount, timeEstimate)

	return v.VAlert().
		Type(alertType).
		Density("compact").
		Icon(fmt.Sprintf("mdi-%s-outline", alertType)).
		Children(
			h.Text(infoText),
		)
}

func configureSubjectSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create subject section
	section := presets.NewSectionBuilder(mb, "Subject").Editing("Subject")

	// Configure viewing mode
	section.ViewingField("Subject").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			c := obj.(*EmailCampaign)
			return vx.VXTextField().ReadOnly(true).Text(c.Subject)
		}
	})

	// Configure editing mode with placeholder support
	section.EditingField("Subject").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return vx.VXTextField().VField(field.Name, field.Value(obj))
		},
	)

	return section
}

func configureUTMParametersSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create UTM Parameters section
	section := presets.NewSectionBuilder(mb, "UTM").
		Label("UTM Parameters").
		Editing("UTM.Source", "UTM.Medium", "UTM.Campaign", "UTM.Term", "UTM.Content")

	section.ViewingField("UTM.Source").Label("Source")
	section.ViewingField("UTM.Medium").Label("Medium")
	section.ViewingField("UTM.Campaign").Label("Campaign")
	section.ViewingField("UTM.Term").Label("Term")
	section.ViewingField("UTM.Content").Label("Content")

	// UTM Source field
	section.EditingField("UTM.Source").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Label("Source").
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Placeholder("e.g., newsletter, google, twitter")
	})

	// UTM Medium field
	section.EditingField("UTM.Medium").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Label("Medium").
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Placeholder("e.g., email, cpc, banner")
	})

	// UTM Campaign field
	section.EditingField("UTM.Campaign").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Label("Campaign").
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Placeholder("e.g., spring_sale, product_launch")
	})

	// UTM Term field
	section.EditingField("UTM.Term").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Label("Term").
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Placeholder("e.g., running_shoes, marketing")
	})

	// UTM Content field
	section.EditingField("UTM.Content").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vx.VXField().Label("Content").
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Placeholder("e.g., top_banner, email_footer")
	})

	return section
}

func configureScheduleSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create schedule section
	section := presets.NewSectionBuilder(mb, "Schedule").
		Label("Schedule").
		Editing("Schedule.Enabled", "TimeRange", "Schedule.Frequency")
	section.ViewingField("Schedule.Enabled").Label("Enabled")
	section.ViewingField("TimeRange").Label("Time Range").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			c := obj.(*EmailCampaign)
			return vx.VXRangePicker().Label("Time Range").Disabled(true).ModelValue([]time.Time{c.StartTime, c.EndTime})
		},
	)
	section.ViewingField("Schedule.Frequency").Label("Frequency")

	// Toggle for enabling scheduling
	section.EditingField("Schedule.Enabled").Label("Enabled").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return v.VSwitch().Label(field.Label).Color("primary").
				Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...)
		}
	})

	// Frequency selector (daily, weekly, monthly)
	section.EditingField("Schedule.Frequency").Label("Frequency").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return presets.SelectField(obj, field, ctx).
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Items([]string{
				FrequencyNone,
				FrequencyDaily,
				FrequencyWeekly,
				FrequencyMonthly,
			})
	})

	section.EditingField("TimeRange").Label("Time Range").
		ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				c, ok := obj.(*EmailCampaign)
				if !ok || !c.Enabled {
					return h.Div()
				}
				return vx.VXRangePicker().Clearable(true).Label("Time Range").
					Type("datetimepicker").
					Placeholder([]string{"Start Time", "End Time"}).
					Attr(web.VField("TimeRange", []int64{c.StartTime.Unix(), c.EndTime.Unix()})...)
			},
		).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			c := obj.(*EmailCampaign)
			strStartTime := ctx.R.FormValue("TimeRange")
			if strStartTime != "" {
				startTime, err := time.Parse(time.DateTime, strStartTime)
				if err != nil {
					return errors.Wrap(err, "failed to parse start time")
				}
				c.StartTime = startTime
			}
			strEndTime := ctx.R.FormValue("Schedule.EndTime")
			if strEndTime != "" {
				endTime, err := time.Parse(time.DateTime, strEndTime)
				if err != nil {
					return errors.Wrap(err, "failed to parse end time")
				}

				c.EndTime = endTime
			}
			return nil
		})

	section.WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
		return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
			return in(obj, ctx)
		}
	})

	return section
}

// buildCronExpression converts campaign schedule settings to a cron expression
func buildCronExpression(campaign *EmailCampaign) string {
	// Default to empty string (invalid cron) if scheduling is disabled
	if !campaign.Enabled {
		return ""
	}

	// Format: second minute hour day-of-month month day-of-week
	switch campaign.Frequency {
	case FrequencyNone:
		// One-time execution, not a recurring job
		return ""
	case FrequencyDaily:
		// Run every day at the specified time
		t := campaign.StartTime
		return fmt.Sprintf("0 %d %d * * *", t.Minute(), t.Hour())
	case FrequencyWeekly:
		// Run weekly on the specified day at the specified time
		t := campaign.StartTime
		// TODO: calculate week according to start time.
		return fmt.Sprintf("0 %d %d * * %d", t.Minute(), t.Hour(), 1)
	case FrequencyMonthly:
		// Run monthly on the same day of month at the specified time
		t := campaign.StartTime
		return fmt.Sprintf("0 %d %d %d * *", t.Minute(), t.Hour(), t.Day())
	default:
		return ""
	}
}

// getUserCountForSegment returns the user count for a specific recipient
// In a real implementation, this would query the database
func getUserCountForSegment(segment string) int {
	// Mock data based on the segment
	switch segment {
	case "segmentationA":
		return 1200
	case "segmentationB":
		return 500
	case "segmentationC":
		return 3000
	case "segmentationD":
		return 800
	default:
		return 0
	}
}
