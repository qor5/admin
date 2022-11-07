package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatus_Pending     OrderStatus = "Pending"
	OrderStatus_Authorised  OrderStatus = "Authorised"
	OrderStatus_Cancelled   OrderStatus = "Cancelled"
	OrderStatus_AuthUnknown OrderStatus = "AuthUnknown"
	OrderStatus_Sending     OrderStatus = "Sending"
	OrderStatus_CheckedIn   OrderStatus = "CheckedIn"
	OrderStatus_Refunded    OrderStatus = "Refunded"
	OrderStatus_Validating  OrderStatus = "Validating"
	OrderStatus_Paid        OrderStatus = "Paid"

	OrderStatusColor_Grey   = "#e0e0e0"
	OrderStatusColor_Orange = "#ffa600"
	OrderStatusColor_Blue   = "#2194f3"
	OrderStatusColor_Green  = "#4caf4f"
	OrderStatusColor_Purple = "#4a148c"
	OrderStatusColor_Brown  = "#9f6d6d"
	OrderStatusColor_Red    = "#f44336"
)

type Order struct {
	gorm.Model

	Source         string
	Status         OrderStatus
	DeliveryMethod string
	PaymentMethod  string
	ConfirmedAt    *time.Time
	OrderItems     []*OrderItem `gorm:"type:text[]"`
}

type OrderItem struct {
	ProductCode     string
	Quantity        int32
	UnitPrice       float64
	TaxUnknown      bool
	TotalPriceExTax *float64
	TotalPrice      float64

	Name  string
	Image string
}

var OrderStatuses = []OrderStatus{
	OrderStatus_Pending,
	OrderStatus_Authorised,
	OrderStatus_Cancelled,
	OrderStatus_AuthUnknown,
	OrderStatus_Sending,
	OrderStatus_CheckedIn,
	OrderStatus_Refunded,
	OrderStatus_Validating,
	OrderStatus_Paid,
}

var OrderStatusColorMap = map[OrderStatus]string{
	OrderStatus_Pending:     OrderStatusColor_Grey,
	OrderStatus_Authorised:  OrderStatusColor_Blue,
	OrderStatus_Cancelled:   OrderStatusColor_Red,
	OrderStatus_AuthUnknown: OrderStatusColor_Red,
	OrderStatus_Sending:     OrderStatusColor_Orange,
	OrderStatus_CheckedIn:   OrderStatusColor_Green,
	OrderStatus_Refunded:    OrderStatusColor_Purple,
	OrderStatus_Validating:  OrderStatusColor_Orange,
	OrderStatus_Paid:        OrderStatusColor_Blue,
}
