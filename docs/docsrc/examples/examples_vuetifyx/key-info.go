package examples_vuetifyx

import (
	"fmt"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type Event struct {
	ID        uint
	Title     string
	CreatedAt time.Time
}

func KeyInfoDemo(ctx *web.EventContext) (pr web.PageResponse, err error) {
	data := []*Event{
		{
			1,
			"<span><strong>¥5,000</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
		{
			2,
			"<span><strong>¥207,626</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
		{
			3,
			"<span><strong>¥7,848</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
		{
			4,
			"<span><strong>¥5,000</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
		{
			5,
			"<span><strong>¥207,626</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
		{
			6,
			"<span><strong>¥7,848</strong> was refunded from a <strong>¥236,170</strong> payment</span>",
			time.Now(),
		},
	}

	dt := vx.DataTable(data).WithoutHeader(true).LoadMoreAt(3, "Show More")

	dt.Column("Title").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.RawHTML(fmt.Sprint(reflectutils.MustGet(obj, fieldName))))
	})

	dt.Column("CreatedAt").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		t := reflectutils.MustGet(obj, fieldName).(time.Time)
		return h.Td(h.Text(t.Format("01/02/06, 15:04:05 PM"))).Class("text-right")
	})

	logsDt := vx.DataTable(data).
		WithoutHeader(true).
		LoadMoreAt(3, "Show More").
		LoadMoreURL("/e20_vuetify_expansion_panels").
		RowExpandFunc(func(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
			return h.Div().Text(h.JSONString(obj)).Class("pa-5")
		})

	logsDt.Column("Title").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		return h.Td(h.RawHTML(fmt.Sprint(reflectutils.MustGet(obj, fieldName))))
	})

	logsDt.Column("CreatedAt").CellComponentFunc(func(obj interface{}, fieldName string, ctx *web.EventContext) h.HTMLComponent {
		t := reflectutils.MustGet(obj, fieldName).(time.Time)
		return h.Td(h.Text(t.Format("01/02/06, 15:04:05 PM"))).Class("text-right")
	})

	pr.Body = VApp(
		VMain(
			vx.Card(
				vx.KeyInfo(
					vx.KeyField(h.Text(time.Now().Format("Jan _2, 15:04 PM"))).Label("Date"),
					vx.KeyField(h.A().Href("https://google.com").Text("customer0077N52")).Label("Customer"),
					vx.KeyField(h.Text("•••• 4242")).Label("Payment method").Icon(VIcon("credit_card")),
					vx.KeyField(h.Text("Normal")).Label("Risk evaluation").Icon(VChip(h.Text("43")).Size(SizeSmall)),
				),
			).SystemBar(
				VIcon("link"),
				h.Text("Hello"),
				VSpacer(),
				h.Text("ch_1EJtQMAqkzzGorqLtIjCEPU5"),
			).Header(
				h.Text("$100.00USD"),
				VChip(h.Text("Refunded"), VIcon("reply").Size(SizeSmall)).Size(SizeSmall),
			).Actions(
				VBtn("Edit"),
			).Class("mb-4"),

			vx.Card(vx.DetailInfo(
				vx.DetailColumn(
					vx.DetailField(vx.OptionalText("cus_EnUK8WcwQkuKQP")).Label("ID"),
					vx.DetailField(vx.OptionalText(time.Now().Format("2006/01/02 15:04"))).Label("Created"),
					vx.DetailField(vx.OptionalText("hello@example.com")).Label("Email"),
					vx.DetailField(vx.OptionalText("customer0077N52")).Label("Description"),
					vx.DetailField(vx.OptionalText("B0E69DBD")).Label("Invoice prefix"),
					vx.DetailField(vx.OptionalText("").ZeroLabel("No VAT number")).Label("VAT number"),
					vx.DetailField(vx.OptionalText("Normal")).Label("Risk evaluation").Icon(VChip(h.Text("43")).Size(
						SizeSmall)),
				).Header("ACCOUNT INFORMATION"),
				vx.DetailColumn(
					vx.DetailField(vx.OptionalText("").ZeroLabel("No address")).Label("Address"),
					vx.DetailField(vx.OptionalText("").ZeroLabel("No phone number")).Label("Phone number"),
				).Header("BILLING INFORMATION"),
			)).HeaderTitle("Details").
				Actions(VBtn("Update details")).
				Class("mb-4"),

			vx.Card(dt).HeaderTitle("Events").Class("mb-4"),

			vx.Card(logsDt).HeaderTitle("Logs").Class("mb-4"),
		),
	)
	return
}

var KeyInfoDemoPB = web.Page(KeyInfoDemo)

var KeyInfoDemoPath = examples.URLPathByFunc(KeyInfoDemo)
