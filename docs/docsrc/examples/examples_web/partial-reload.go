package examples_web

// @snippet_begin(PartialReloadSample)
import (
	"fmt"
	"time"

	"github.com/qor5/docs/v3/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func PartialReloadPage(ctx *web.EventContext) (pr web.PageResponse, err error) {
	reloadCount = 0

	ctx.Injector.HeadHTML(`
<style>
.rp {
	float: left;
	width: 200px;
	height: 200px;
	margin-right: 20px;
	background-color: orange;
}
</style>
`,
	)
	pr.Body = Div(
		H1("Portal Reload Automatically"),

		web.Scope(
			web.Portal().Loader(web.POST().EventFunc("autoReload")).AutoReloadInterval("locals.interval"),
			Button("stop").Attr("@click", "locals.interval = 0"),
		).Init(`{interval: 2000}`).VSlot("{ locals, form }"),

		H1("Load Data Only"),
		web.Scope(
			Ul(
				Li(
					Text("{{item}}"),
				).Attr("v-for", "item in locals.items"),
			),
			Button("Fetch Data").Attr("@click", web.GET().EventFunc("loadData").ThenScript(`locals.items = r.data`).Go()),
		).VSlot("{ locals, form }").FormInit("{ items: []}"),

		H1("Partial Load and Reload"),
		Div(
			H2("Product 1"),
		).Style("height: 200px; background-color: grey;"),
		H2("Related Products"),
		web.Portal().Name("related_products").Loader(web.POST().EventFunc("related").Query("productCode", "AH123")),
		A().Href("javascript:;").Text("Reload Related Products").
			Attr("@click", web.POST().EventFunc("reload3").Go()),
	)
	return
}

func related(ctx *web.EventContext) (er web.EventResponse, err error) {
	code := ctx.R.FormValue("productCode")
	er.Body = Div(

		Div(
			H3("Product A (related products of "+code+")"),
			Div().Text(time.Now().Format(time.RFC3339Nano)),
		).Class("rp"),
		Div(
			H3("Product B"),
			Div().Text(time.Now().Format(time.RFC3339Nano)),
		).Class("rp"),
		Div(
			H3("Product C"),
			Div().Text(time.Now().Format(time.RFC3339Nano)),
		).Class("rp"),
	)
	return
}

func reload3(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.ReloadPortals = []string{"related_products"}
	return
}

var reloadCount = 1

func autoReload(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.Body = Span(time.Now().String())
	reloadCount++

	if reloadCount > 5 {
		er.RunScript = `locals.interval = 0;`
	}
	return
}

func loadData(ctx *web.EventContext) (er web.EventResponse, err error) {
	var r []string
	for i := 0; i < 10; i++ {
		r = append(r, fmt.Sprintf("%d-%d", i, time.Now().Nanosecond()))
	}
	er.Data = r
	return
}

var PartialReloadPagePB = web.Page(PartialReloadPage).
	EventFunc("related", related).
	EventFunc("reload3", reload3).
	EventFunc("autoReload", autoReload).
	EventFunc("loadData", loadData)

var PartialReloadPagePath = examples.URLPathByFunc(PartialReloadPage)

// @snippet_end
