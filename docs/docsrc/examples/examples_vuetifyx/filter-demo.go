package examples_vuetifyx

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/examples"

	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
)

func FilterDemo(ctx *web.EventContext) (pr web.PageResponse, err error) {
	fd := vuetifyx.FilterData([]*vuetifyx.FilterItem{
		{
			Key:          "invoiceDateTimeRange",
			Label:        "Invoice Datetime Range",
			ItemType:     vuetifyx.ItemTypeDatetimeRange,
			SQLCondition: "InvoiceDate %s datetime(?, 'unixepoch')",
			Selected:     true,
		},
		{
			Key:      "invoiceDate",
			Label:    "Invoice Date",
			ItemType: vuetifyx.ItemTypeDate,
			Selected: false,
			ValueIs:  "",
		},
		{
			Key:      "invoiceDateRange",
			Label:    "Invoice Date Range",
			ItemType: vuetifyx.ItemTypeDateRange,
			Selected: false,
		},
		{
			Key:      "invoiceDatetimeRangePicker",
			Label:    "Invoice DatetimeRangePicker",
			ItemType: vuetifyx.ItemTypeDatetimeRangePicker,
			Selected: false,
		},
		{
			Key:      "invoiceDateRangePicker",
			Label:    "Invoice DateRangePicker",
			ItemType: vuetifyx.ItemTypeDateRangePicker,
			Selected: false,
		},
		{
			Key:      "invoiceDatePicker",
			Label:    "Invoice DatePicker",
			ItemType: vuetifyx.ItemTypeDatePicker,
			Selected: false,
		},
		{
			Key:          "country",
			Label:        "Country",
			ItemType:     vuetifyx.ItemTypeSelect,
			SQLCondition: "upper(BillingCountry) %s upper(?)",
			Options: []*vuetifyx.SelectItem{
				{
					Value: "US",
					Text:  "United States",
				},
				{
					Value: "CN",
					Text:  "China",
				},
			},
		},
		{
			Key:          "totalAmount",
			Label:        "Total Amount",
			ItemType:     vuetifyx.ItemTypeNumber,
			SQLCondition: "Total %s ?",
		},
	})

	fd.SetByQueryString(&web.EventContext{}, ctx.R.URL.RawQuery)

	pr.Body = VApp(
		VMain(
			vuetifyx.VXFilter(fd),
		),
	)
	return
}

var FilterDemoPB = web.Page(FilterDemo)

var FilterDemoPath = examples.URLPathByFunc(FilterDemo)
