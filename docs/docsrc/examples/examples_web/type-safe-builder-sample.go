package examples_web

// @snippet_begin(TypeSafeBuilderSample)
import (
	"github.com/qor5/docs/v3/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func result(args ...HTMLComponent) HTMLComponent {
	var converted []HTMLComponent
	for _, arg := range args {
		converted = append(converted, Div(arg).Class("wrapped"))
	}

	return HTML(
		Head(
			Title("XML encoding with Go"),
		),
		Body(
			H1("XML encoding with Go"),
			P().Text("this format can be used as an alternative markup to XML"),
			A().Href("http://golang.org").Text("Go"),
			P(
				Text("this is some"),
				B("mixed"),
				Text("text. For more see the"),
				A().Href("http://golang.org").Text("Go"),
				Text("project"),
			),
			P().Text("some text"),

			P(converted...),
		),
	)
}

func TypeSafeBuilderExample(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = result(H5("1"), B("2"), Strong("3"))
	return
}

var TypeSafeBuilderSamplePFPB = web.Page(TypeSafeBuilderExample)

var TypeSafeBuilderSamplePath = examples.URLPathByFunc(TypeSafeBuilderExample)

// @snippet_end
