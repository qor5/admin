package admin

import (
	"time"

	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	OrderCodeAttr      = "ID"
	CreatedDateAttr    = "CreatedAt"
	CheckInDateAttr    = "ConfirmedAt"
	StatusAttr         = "Status"
	PaymentMethodAttr  = "PaymentMethod"
	DeliveryMethodAttr = "DeliveryMethod"
	SourceAttr         = "Source"
	OrderItemsAttr     = "OrderItems"
	ActionsAttr        = "Actions"
)

func configOrder(pb *presets.Builder, db *gorm.DB) {
	b := pb.Model(&models.Order{})

	// listing
	lb := b.Listing(
		OrderCodeAttr,
		CreatedDateAttr,
		CheckInDateAttr,
		PaymentMethodAttr,
		StatusAttr,
		SourceAttr,
	)

	lb.Field(CreatedDateAttr).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.Text(field.Value(obj).(time.Time).Local().Format("2006-01-02 15:04:05")))
	}).Label("Date Created")

	lb.Field(CheckInDateAttr).Label("Check In Date").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		pTime := field.Value(obj)
		if pTime != nil {
			return h.Td(h.Text((*pTime.(*time.Time)).Local().Format("2006-01-02 15:04:05")))
		} else {
			return h.Td(h.Text(""))
		}
	})

	lb.Field(StatusAttr).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		status := field.Value(obj).(models.OrderStatus)
		if status == "" {
			return h.Td()
		}
		return h.Td(GetColoredStatus(status))
	})

	lb.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		statusOptions := []*vuetifyx.SelectItem{}
		for _, status := range models.OrderStatuses {
			statusOptions = append(statusOptions, &vuetifyx.SelectItem{Value: string(status), Text: string(status)})
		}

		return []*vuetifyx.FilterItem{
			{
				Key:          "created_at",
				Label:        "Created At",
				ItemType:     vuetifyx.ItemTypeDate,
				SQLCondition: `created_at %s ?`,
			},
			{
				Key:          "status",
				Label:        "Status",
				ItemType:     vuetifyx.ItemTypeMultipleSelect,
				SQLCondition: `status %s ?`,
				Options:      statusOptions,
			},
		}
	})

	lb.Action("Export").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return vuetify.VBtn("Export").
			Color("primary").
			Depressed(true).
			Class("ml-2").
			Href(exportOrdersURL)
	})

	lb.BulkAction("Change status").ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
		vErr := &web.ValidationErrors{}
		if ctx.Flash != nil {
			vErr = ctx.Flash.(*web.ValidationErrors)
		}

		return h.Div(
			vuetify.VCardText(
				vuetify.VAutocomplete().Label("Status").
					FieldName("status").
					Items(models.OrderStatuses).
					Attach(false).ErrorMessages(vErr.GetFieldErrors("status")...),
			),
		)
	}).UpdateFunc(func(selectedIds []string, ctx *web.EventContext) (err error) {
		vErr := &web.ValidationErrors{}
		status := ctx.R.FormValue("status")
		if status == "" {
			vErr.FieldError("status", "Please select status")
			ctx.Flash = vErr
			return nil
		}

		if err := db.Model(&models.Order{}).Where("id IN (?)", selectedIds).Update("status", status).Error; err != nil {
			return err
		}

		return
	})

	// detailing
	b.RightDrawerWidth("50%")
	orderDetailing := b.Detailing(
		// ActionsAttr,
		&presets.FieldsSection{
			Title: "Basic Information",
			Rows: [][]string{
				{OrderCodeAttr, CreatedDateAttr},
				{StatusAttr, CheckInDateAttr},
				{PaymentMethodAttr, DeliveryMethodAttr},
				{SourceAttr},
			},
		},
	).Drawer(true)

	orderDetailing.Field(OrderCodeAttr).Label("Order ID")

	orderDetailing.Field(CreatedDateAttr).Label("Check In Date").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		order := obj.(*models.Order)
		v := order.CreatedAt.Local().Format("2006-01-02 15:04:05")
		return vuetifyx.VXReadonlyField().
			Label(field.Label).
			Value(v)
	})

	orderDetailing.Field(StatusAttr).ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return vuetifyx.VXReadonlyField(GetColoredStatus(obj.(*models.Order).Status)).
			Label(field.Label)
	})

	orderDetailing.Field(CheckInDateAttr).Label("Check In Date").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		order := obj.(*models.Order)
		var v string
		if order.ConfirmedAt != nil {
			v = order.ConfirmedAt.Local().Format("2006-01-02 15:04:05")
		}
		return vuetifyx.VXReadonlyField().
			Label(field.Label).
			Value(v)
	})

	orderDetailing.Field(PaymentMethodAttr).Label("Payment Method").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		order := obj.(*models.Order)
		return vuetifyx.VXReadonlyField().
			Label(field.Label).
			Value(order.PaymentMethod)
	})

	orderDetailing.Field(DeliveryMethodAttr).Label("Fulfilment").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		order := obj.(*models.Order)
		return vuetifyx.VXReadonlyField().
			Label(field.Label).
			Value(order.DeliveryMethod)
	})
}

func GetColoredStatus(status models.OrderStatus) h.HTMLComponent {
	color, ok := models.OrderStatusColorMap[status]
	if !ok {
		return h.Text(string(status))
	}

	return vuetify.VChip(h.Text(string(status))).Color(color).Dark(true)
}
