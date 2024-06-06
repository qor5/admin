package examples_web

import (
	"fmt"
	"net/url"

	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

var page1Title = "Page 1"

// @snippet_begin(PageTransitionSample)

const (
	Page1Path = "/samples/page_1"
	Page2Path = "/samples/page_2"
)

func Page1(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		H1(page1Title),
		Ul(
			Li(
				A().Href(Page2Path).
					Text("To Page 2 With Normal Link"),
			),
			Li(
				A().Href("javascript:;").
					Text("To Page 2 With Push State Link").
					Attr("@click", web.POST().PushStateURL(Page2Path).Go()),
			),
		),
		fromParam(ctx),
	).Style("color: green; font-size: 24px;")
	return
}

func Page2(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		H1("Page 2"),
		Ul(
			Li(
				A().Href("javascript:;").
					Text("To Page 1 With Normal Link").
					Attr("@click", web.POST().
						PushStateURL(Page1Path).
						Queries(url.Values{"from": []string{"page 2 link 1"}}).
						Go()),
			),
			Li(
				Button("Do an action then go to Page 1 with push state and parameters").
					Attr("@click", web.POST().EventFunc("doAction2").Query("id", "42").Go()),
			),
			Li(
				Button("Do an action then go to Page 1 with redirect url").
					Attr("@click", web.POST().EventFunc("doAction1").Query("id", "41").Go()),
			),
		),
	).Style("color: orange; font-size: 24px;")
	return
}

func fromParam(ctx *web.EventContext) HTMLComponent {
	var from HTMLComponent
	val := ctx.R.FormValue("from")
	if len(val) > 0 {
		from = Components(
			B("from:"),
			Text(val),
		)
	}
	return from
}

func doAction1(ctx *web.EventContext) (er web.EventResponse, err error) {
	updateDatabase(ctx.ParamAsInt("id"))
	er.RedirectURL = Page1Path + "?" + url.Values{"from": []string{"page2 with redirect"}}.Encode()
	return
}

func doAction2(ctx *web.EventContext) (er web.EventResponse, err error) {
	updateDatabase(ctx.ParamAsInt("id"))
	er.PushState = web.Location(url.Values{"from": []string{"page2"}}).
		URL(Page1Path)
	return
}

var Page1PB = web.Page(Page1)

var Page2PB = web.Page(Page2).
	EventFunc("doAction1", doAction1).
	EventFunc("doAction2", doAction2)

// @snippet_end

func updateDatabase(val int) {
	page1Title = fmt.Sprintf("Page 1 (Updated by Page2 to %d)", val)
}
