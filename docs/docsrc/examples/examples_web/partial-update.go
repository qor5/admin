package examples_web

// @snippet_begin(PartialUpdateSample)
import (
	"time"

	"github.com/qor5/docs/v3/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func PartialUpdatePage(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		H1("Partial Update"),
		A().Text("Edit").Href("javascript:;").
			Attr("@click", web.POST().EventFunc("edit1").Go()),
		web.Portal(
			If(len(fd.Title) > 0,
				H1(fd.Title),
				H5(fd.Date),
			).Else(
				Text("Default value"),
			),
		).Name("part1"),
		Div().Text(time.Now().Format(time.RFC3339Nano)),
	)
	return
}

type formData struct {
	Title string
	Date  string
}

var fd formData

func edit1(ctx *web.EventContext) (er web.EventResponse, err error) {
	er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
		Name: "part1",
		Body: Div(
			web.Scope(
				Fieldset(
					Legend("Input value"),
					Div(
						Label("Title"),
						Input("").Type("text").Attr("v-model", "form.Title"),
					),

					Div(
						Label("Date"),
						Input("").Type("date").Attr("v-model", "form.Date"),
					),
				),
				Button("Update").
					Attr("@click", web.POST().EventFunc("reload2").Go()),
			).VSlot("{ locals, form }").FormInit(JSONString(fd)),
		),
	})
	return
}

func reload2(ctx *web.EventContext) (er web.EventResponse, err error) {
	ctx.MustUnmarshalForm(&fd)
	er.Reload = true
	return
}

var PartialUpdatePagePB = web.Page(PartialUpdatePage).
	EventFunc("edit1", edit1).
	EventFunc("reload2", reload2)

var PartialUpdatePagePath = examples.URLPathByFunc(PartialUpdatePage)

// @snippet_end
