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

// Days of week constants
const (
	Sunday    = "sunday"
	Monday    = "monday"
	Tuesday   = "tuesday"
	Wednesday = "wednesday"
	Thursday  = "thursday"
	Friday    = "friday"
	Saturday  = "saturday"
)

type (
	EmailCampaign struct {
		gorm.Model
		EmailDetail
		UTM
		Schedule

		Recipient string
		Name      string
		Status    string // StatusDraft, StatusSent, StatusScheduled

	}

	Schedule struct {
		Enabled    bool
		Frequency  string // FrequencyNone, FrequencyDaily, FrequencyWeekly, FrequencyMonthly
		DayOfWeek  string // Only used for weekly frequency
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
	mb := pb.Model(&EmailCampaign{})

	configureListing(mb)

	// Configure detail page
	dp := mb.Detailing(EmailDetailField, "Recipient", "UTM", "Schedule")
	// dp := mb.Detailing(EmailDetailField, "Recipient", "Schedule")
	// Add sections to detail page in the desired order (UTM section above Schedule section)
	dp.Section(configureRecipientSection(mb, db))
	dp.Section(configureUTMParametersSection(mb, db))
	dp.Section(configureScheduleSection(mb, db))

	configureEditing(mb)
	return mb
}

func configureListing(mb *presets.ModelBuilder) {
	// Configure listing page
	listing := mb.Listing("ID", "Name", "Status", "CreatedAt", "UpdatedAt")

	// Customize the listing display
	listing.Field("Name").Label("Name")
	listing.Field("Status").Label("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
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

		return h.Td(v.VChip().Color(color).Size("small").Class("text-capitalize").Children(h.Text(text)))
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

	// Add title and create button
	listing.Title(func(ctx *web.EventContext, style presets.ListingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		titleCompo = h.Div(
			h.H4("Email Campaigns").Class("text-h4"),
			v.VSpacer(),
			v.VBtn("Add New").
				Color("primary").
				Attr("@click", web.Plaid().
					EventFunc(actions.New).
					Go()),
		).Class("d-flex align-center mb-4")

		return defaultTitle, titleCompo, nil
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
	mb.Editing("Name", "Subject", "JSONBody", "HTMLBody").Creating("Subject", TemplateSelectionFiled)
}

func configureRecipientSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create recipient section
	section := presets.NewSectionBuilder(mb, "Recipient").
		Editing("Recipient")
	section.ViewingField("Recipient")

	// Portal name for the recipient info
	const recipientInfoPortal = "recipientInfoPortal"

	// Register fetch recipient info event handler
	mb.RegisterEventFunc("fetchRecipientInfo", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		// Get recipient value from form
		recipient := ctx.R.FormValue("Recipient")
		if recipient == "" {
			return
		}
		// Update the portal with the info banner
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: recipientInfoPortal,
			Body: createRecipientInfoBanner(recipient),
		})
		return
	})

	section.EditingField("Recipient").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
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
				Attr("@update:model-value", web.Plaid().EventFunc("fetchRecipientInfo").Go())

			components = append(components, selectField)

			// Create portal for recipient info
			// For the initial state, we'll directly include the banner in the portal
			var portalContent h.HTMLComponent
			if campaign.Recipient != "" {
				portalContent = createRecipientInfoBanner(campaign.Recipient)
			} else {
				// Empty div if no recipient selected
				portalContent = h.Div()
			}

			// Add portal with initial content
			infoContainer := h.Div().
				Class("recipient-info-container mt-3").
				Children(
					web.Portal(portalContent).Name(recipientInfoPortal),
				)

			components = append(components, infoContainer)

			return h.Div(components...)
		}
	}).HideLabel()

	return section
}

// createRecipientInfoBanner creates an info banner showing user count and estimated time
func createRecipientInfoBanner(recipient string) h.HTMLComponent {
	// Get user count
	userCount := getUserCountForRecipient(recipient)

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
		return v.VTextField().
			Label("Source").
			Attr(web.VField(field.Name, field.Value(obj))...).
			Placeholder("e.g., newsletter, google, twitter")
	})

	// UTM Medium field
	section.EditingField("UTM.Medium").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VTextField().
			Label("Medium").
			Attr(web.VField(field.Name, field.Value(obj))...).
			Placeholder("e.g., email, cpc, banner")
	})

	// UTM Campaign field
	section.EditingField("UTM.Campaign").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VTextField().
			Label("Campaign").
			Attr(web.VField(field.Name, field.Value(obj))...).
			Placeholder("e.g., spring_sale, product_launch")
	})

	// UTM Term field
	section.EditingField("UTM.Term").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VTextField().
			Label("Term").
			Attr(web.VField(field.Name, field.Value(obj))...).
			Placeholder("e.g., running_shoes, marketing")
	})

	// UTM Content field
	section.EditingField("UTM.Content").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return v.VTextField().
			Label("Content").
			Attr(web.VField(field.Name, field.Value(obj))...).
			Placeholder("e.g., top_banner, email_footer")
	})

	// Configure Save function for UTM parameters
	section.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		// Fetch the campaign
		var campaign EmailCampaign
		if err := db.First(&campaign, "id = ?", id).Error; err != nil {
			return errors.Wrap(err, "failed to fetch campaign")
		}
		// Extract form values
		ctx.MustUnmarshalForm(&campaign)

		// Save the updated campaign
		if err := db.Save(&campaign).Error; err != nil {
			return errors.Wrap(err, "failed to save campaign")
		}

		return nil
	})

	return section
}

func configureScheduleSection(mb *presets.ModelBuilder, db *gorm.DB) *presets.SectionBuilder {
	// Create schedule section
	section := presets.NewSectionBuilder(mb, "Schedule").
		Label("Schedule").
		Editing("Schedule.Enabled", "Schedule.StartTime", "Schedule.Frequency", "Schedule.DayOfWeek", "Schedule.EndTime")
	section.ViewingField("Schedule.Enabled").Label("Enabled")
	section.ViewingField("Schedule.StartTime").Label("Start Time")
	section.ViewingField("Schedule.Frequency").Label("Frequency")
	section.ViewingField("Schedule.DayOfWeek").Label("Day of Week")
	section.ViewingField("Schedule.EndTime").Label("End Time")

	// Toggle for enabling scheduling
	section.EditingField("Enabled").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
		return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return v.VSwitch().Label(field.Label).Color("primary").
				Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...)
		}
	})

	// Frequency selector (daily, weekly, monthly)
	section.EditingField("Frequency").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return presets.SelectField(obj, field, ctx).
			Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
			Items([]string{
				FrequencyNone,
				FrequencyDaily,
				FrequencyWeekly,
				FrequencyMonthly,
			})
	})

	// Day of week selector (only shown when frequency is weekly)
	section.EditingField("DayOfWeek").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c, ok := obj.(*EmailCampaign)
		if !ok || !c.Enabled || c.Frequency != FrequencyWeekly {
			return h.Div()
		}
		return presets.SelectField(obj, field, ctx).
			Items([]string{
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
				"Sunday",
			})
	})

	// Start time date picker
	section.EditingField("StartTime").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c, ok := obj.(*EmailCampaign)
		if !ok || !c.Enabled {
			return h.Div()
		}
		var dateStr string
		if !c.StartTime.IsZero() {
			dateStr = c.StartTime.Local().Format("2006-01-02 15:04")
		}
		return vx.VXDateTimePicker().
			Label("Start Time").
			Attr(presets.VFieldError(field.Name, dateStr, field.Errors)...).
			TimePickerProps(vx.TimePickerProps{
				Format:     "24hr",
				Scrollable: true,
			})
	})

	// End time date picker
	section.EditingField("EndTime").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		c, ok := obj.(*EmailCampaign)
		if !ok || !c.Enabled {
			return h.Div()
		}

		// Use native date picker with Vuetify styling
		dateStr := ""
		if !c.EndTime.IsZero() {
			dateStr = c.EndTime.Format("2006-01-02 15:04")
		}

		return vx.VXDateTimePicker().
			Label("End Time").
			Attr(presets.VFieldError(field.Name, dateStr, field.Errors)...).
			TimePickerProps(vx.TimePickerProps{
				Format:     "24hr",
				Scrollable: true,
			})
	})

	section.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		// Fetch the campaign
		var campaign EmailCampaign
		if err := db.First(&campaign, id).Error; err != nil {
			return errors.Wrap(err, "failed to fetch campaign")
		}
		// Parse retry count
		// retryCount := ctx.R.FormValue("RetryCount")
		// if retryCount != "" {
		// 	fmt.Sscanf(retryCount, "%d", &campaign.RetryCount)
		// 	if campaign.RetryCount > MaxRetryCount {
		// 		campaign.RetryCount = MaxRetryCount
		// 	}
		// }

		// // Parse start time
		// if startTime != "" {
		// 	// Try standard date format first (YYYY-MM-DD)
		// 	if t, err := time.Parse("2006-01-02 15:04", startTime); err == nil {
		// 		campaign.StartTime = t
		// 	}
		// }
		//
		// // Parse end time
		// endTime := ctx.R.FormValue(formKeyEndTime)
		// if endTime != "" {
		// 	// Try standard date format first (YYYY-MM-DD)
		// 	if t, err := time.Parse("2006-01-02 15:04:05", endTime); err == nil {
		// 		campaign.EndTime = t
		// 	}
		// }

		ctx.MustUnmarshalForm(&campaign)

		// Create scheduler if scheduling is enabled
		if campaign.Enabled {
			// Update the status to scheduled
			campaign.Status = StatusScheduled
		}

		// Save the updated campaign
		if err := db.Save(&campaign).Error; err != nil {
			return errors.Wrap(err, "failed to save campaign")
		}

		return
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
		dayNum := getDayOfWeekNumber(campaign.DayOfWeek)
		return fmt.Sprintf("0 %d %d * * %d", t.Minute(), t.Hour(), dayNum)
	case FrequencyMonthly:
		// Run monthly on the same day of month at the specified time
		t := campaign.StartTime
		return fmt.Sprintf("0 %d %d %d * *", t.Minute(), t.Hour(), t.Day())
	default:
		return ""
	}
}

// getDayOfWeekNumber converts day name to cron day number (0-6, Sunday-Saturday)
func getDayOfWeekNumber(day string) int {
	switch day {
	case Sunday:
		return 0
	case Monday:
		return 1
	case Tuesday:
		return 2
	case Wednesday:
		return 3
	case Thursday:
		return 4
	case Friday:
		return 5
	case Saturday:
		return 6
	default:
		return 0 // Default to Sunday
	}
}

// getUserCountForRecipient returns the user count for a specific recipient
// In a real implementation, this would query the database
func getUserCountForRecipient(recipient string) int {
	// Mock data based on the segment
	switch recipient {
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
