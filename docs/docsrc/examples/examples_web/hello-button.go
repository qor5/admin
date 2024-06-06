package examples_web

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type mystate struct {
	Message string
}

func HelloButton(ctx *web.EventContext) (pr web.PageResponse, err error) {
	s := &mystate{}
	if ctx.Flash != nil {
		s = ctx.Flash.(*mystate)
	}

	pr.Body = Div(
		Button("Hello").Attr("@click", web.POST().EventFunc("reload").Go()),
		Tag("input").
			Attr("type", "text").
			Attr("value", s.Message).
			Attr("@input", web.POST().
				EventFunc("reload").
				FieldValue("Message", web.Var("$event.target.value")).
				Go()),
		Div().
			Style("font-family: monospace;").
			Text(s.Message),
	)
	return
}

func reload(ctx *web.EventContext) (r web.EventResponse, err error) {
	s := &mystate{}
	ctx.MustUnmarshalForm(s)
	ctx.Flash = s

	r.Reload = true
	return
}

var HelloButtonPB = web.Page(HelloButton).
	EventFunc("reload", reload)

var HelloButtonPath = examples.URLPathByFunc(HelloButton)
