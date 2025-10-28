package admin

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"gorm.io/gorm"
)

type ECDashboard struct{}

func configECDashboard(pb *presets.Builder, db *gorm.DB) {
	b := pb.Model(&ECDashboard{}).Label("EC Dashboard").URIName("ec-dashboard")

	lb := b.Listing()

	lb.PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		// DB query
		var productCount int64
		var orderCount int64
		var orders []*models.Order

		if err = db.Model(&models.Product{}).Count(&productCount).Error; err != nil {
			r.Body = errorBody(err.Error())
			return
		}

		if err = db.Model(&models.Order{}).Find(&orders).Count(&orderCount).Error; err != nil {
			r.Body = errorBody(err.Error())
			return
		}

		// Chart generate
		var pie chart.PieChart
		pieBuffer := bytes.NewBuffer([]byte{})
		if orderCount > 0 {
			iter := linq.From(orders)
			pie = chart.PieChart{
				Width:  1024,
				Height: 1024,
				Values: []chart.Value{
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_Pending
						}).Count()),
						Label: string(models.OrderStatus_Pending),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Grey[1:])},
					},
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_Authorised
						}).Count()),
						Label: string(models.OrderStatus_Authorised),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Blue[1:])},
					},
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_Cancelled
						}).Count()),
						Label: string(models.OrderStatus_Cancelled),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Red[1:])},
					},
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_AuthUnknown
						}).Count()),
						Label: string(models.OrderStatus_AuthUnknown),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Red[1:])},
					},
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_Sending
						}).Count()),
						Label: string(models.OrderStatus_Sending),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Orange[1:])},
					},
					{
						Value: float64(iter.Where(func(i interface{}) bool {
							return i.(*models.Order).Status == models.OrderStatus_CheckedIn
						}).Count()),
						Label: string(models.OrderStatus_CheckedIn),
						Style: chart.Style{FillColor: drawing.ColorFromHex(models.OrderStatusColor_Green[1:])},
					},
					// TODO: add more status
				},
			}
			err = pie.Render(chart.SVG, pieBuffer)
		}

		body := vuetify.VContainer(
			vuetify.VRow(
				h.Div(
					h.Div(h.Strong("Statistics")).Class("mt-2 col col-12"),
					h.Div(
						h.Div(
							vuetify.VCard(
								vuetify.VCardTitle(h.Text(strconv.Itoa(int(productCount)))),
								vuetify.VCardSubtitle(h.Text("Products")),
							).Variant(vuetify.VariantOutlined),
						).Class("pa-4 pt-12"),
						h.Div(
							vuetify.VCard(
								vuetify.VCardTitle(h.Text(strconv.Itoa(int(orderCount)))),
								vuetify.VCardSubtitle(h.Text("Orders")),
							).Variant(vuetify.VariantOutlined),
						).Class("pa-4"),
					).Class("v-card v-sheet theme--light").Style("height: 300px;"),
				).Class("col col-6"),
				h.Div(
					h.Div(h.Strong("Order Status")).Class("mt-2 col col-12"),
					h.Div(
						h.RawHTML(
							strings.ReplaceAll(pieBuffer.String(), `width="1024" height="1024"`, `width="100%" height="100%" viewBox="-85 -80 1200 1200"`)),
					).Class("v-card v-sheet theme--light").Style("height: 300px;"),
				).Class("col col-6"),
			),
		)

		r.Body = body
		r.PageTitle = "EC Dashboard"

		return
	})
}

func errorBody(msg string) h.HTMLComponent {
	return vuetify.VContainer(
		h.P().Text(msg),
	)
}
