package examples_web

// @snippet_begin(HelloWorldReloadSample)
import (
	"time"

	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func HelloWorldReload(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		H1("Hello World"),
		Text(time.Now().Format(time.RFC3339Nano)),
		Button("Reload Page").Attr("@click", web.GET().
			EventFunc(reloadEvent).
			Go()),
	)
	return
}

func update(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.Reload = true
	return
}

const reloadEvent = "reload"

var HelloWorldReloadPB = web.Page(HelloWorldReload).
	EventFunc(reloadEvent, update)

var HelloWorldReloadPath = examples.URLPathByFunc(HelloWorldReload)

// @snippet_end
