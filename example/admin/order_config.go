package admin

import (
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/example/models"
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
