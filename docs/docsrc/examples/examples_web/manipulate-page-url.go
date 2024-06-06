package examples_web

// @snippet_begin(MultiStatePageSample)
import (
	"net/url"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type multiStateFormData struct {
	Name string
	Date string
}

var multiStateFd = multiStateFormData{
	Name: "Felix",
	Date: "2021-01-01",
}

func MultiStatePage(ctx *web.EventContext) (pr web.PageResponse, err error) {
	title := "Multi State Page"
	if len(ctx.R.URL.Query().Get("title")) > 0 {
		title = ctx.R.URL.Query().Get("title")
	}
	var panel HTMLComponent
	if len(ctx.R.URL.Query().Get("panel")) > 0 {
		panel = Div(
			web.Scope(
				Fieldset(
					Div(
						Label("Name"),
						Input("").Type("text").Attr("v-model", "form.Name"),
					),
					Div(
						Label("Date"),
						Input("").Type("date").Attr("v-model", "form.Date"),
					),
				),
				Button("Update").Attr("@click", web.POST().EventFunc("update5").Go()),
			).VSlot("{ locals, form }").FormInit(JSONString(multiStateFd)),
		).Style("border: 5px solid orange; height: 200px;")
	}

	pr.Body = Div(
		H1(title),
		Ol(
			Li(
				A().Text("change page title").Href("javascript:;").
					Attr("@click", web.POST().Queries(url.Values{"title": []string{"Hello"}}).Go()),
			),
		),
		panel,

		Table(
			Thead(
				Th("Name"),
				Th("Date"),
			),
			Tbody(
				Tr(
					Td(Text(multiStateFd.Name)),
					Td(Text(multiStateFd.Date)),
					Td(A().Text("Edit with Panel").Href("javascript:;").Attr("@click",
						web.POST().EventFunc("openPanel").Go())),
				),
			),
		),
	)
	return
}

func openPanel(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.PushState = web.Location(url.Values{"panel": []string{"1"}}).MergeQuery(true)
	return
}

func update5(ctx *web.EventContext) (er web.EventResponse, err error) {
	ctx.MustUnmarshalForm(&multiStateFd)
	er.PushState = web.Location(url.Values{"panel": []string{""}}).MergeQuery(true)
	return
}

var MultiStatePagePB = web.Page(MultiStatePage).
	EventFunc("openPanel", openPanel).
	EventFunc("update5", update5)

var MultiStatePagePath = examples.URLPathByFunc(MultiStatePage)

// @snippet_end
