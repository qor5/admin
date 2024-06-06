package examples_web

// @snippet_begin(CompositeComponentSample1)
import (
	"fmt"

	"github.com/qor5/docs/v3/docsrc/examples"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func Navbar(title string, activeIndex int, items ...HTMLComponent) HTMLComponent {
	ul := Ul().Class("navbar-nav mr-auto")

	for i, item := range items {
		ul.AppendChildren(
			Li(
				item,
			).Class("nav-item").ClassIf("active", activeIndex == i),
		)
	}

	return Nav(
		A(Text(title)).Class("navbar-brand").
			Href("#"),

		Button("").Class("navbar-toggler").
			Type("button").
			Attr("data-toggle", "collapse").
			Attr("data-target", "#navbarNav").
			Attr("aria-controls", "navbarNav").
			Attr("aria-expanded", "false").
			Attr("aria-label", "Toggle navigation").
			Children(
				Span("").Class("navbar-toggler-icon"),
			),

		Div(
			ul,
			Form(
				Input("").Class("form-control mr-sm-2").
					Type("search").
					Placeholder("Search").
					Attr("aria-label", "Search"),
				Button("Search").Class("btn btn-outline-light my-2 my-sm-0").
					Type("submit"),
			).Class("form-inline my-2 my-lg-0"),
		).Class("collapse navbar-collapse").
			Id("navbarNav"),
	).Class("navbar navbar-expand-lg navbar-dark bg-primary")
}

type CarouselItem struct {
	ImageSrc string
	ImageAlt string
}

func Carousel(carouselId string, activeIndex int, items []*CarouselItem) HTMLComponent {
	indicators := Ol().Class("carousel-indicators")
	carouselInners := Div().Class("carousel-inner")

	for i, item := range items {
		indicators.AppendChildren(
			Li().Attr("data-target", "#"+carouselId).
				Attr("data-slide-to", fmt.Sprint(i)).
				ClassIf("active", activeIndex == i),
		)

		carouselInners.AppendChildren(
			Div(
				fakeImage(item.ImageAlt),
			).Class("carousel-item").ClassIf("active", activeIndex == i).Style("font-size: 3.5rem;"),
		)
	}

	return Div(
		indicators,
		carouselInners,
		A(
			Span("").Class("carousel-control-prev-icon").
				Attr("aria-hidden", "true"),
			Span("Previous").Class("sr-only"),
		).Class("carousel-control-prev").
			Href("#"+carouselId).
			Role("button").
			Attr("data-slide", "prev"),
		A(
			Span("").Class("carousel-control-next-icon").
				Attr("aria-hidden", "true"),
			Span("Next").Class("sr-only"),
		).Class("carousel-control-next").
			Href("#"+carouselId).
			Role("button").
			Attr("data-slide", "next"),
	).Id(carouselId).
		Class("carousel slide").
		Attr("data-ride", "carousel")
}

func CompositeComponentSample1Page(ctx *web.EventContext) (pr web.PageResponse, err error) {
	pr.Body = Div(
		Navbar(
			"Hello",
			1,

			A(
				Text("Home"),
			).Class("nav-link").
				Href("#"),

			A(
				Text("Features"),
			).Class("nav-link").
				Href("#"),

			A(
				Text("Pricing"),
			).Class("nav-link").
				Href("#"),

			A(
				Text("Disabled"),
			).Class("nav-link disabled").
				Href("#").
				TabIndex(-1).
				Attr("aria-disabled", "true"),
		),

		Div(
			Div(
				Div(
					Carousel("hello1", 1, []*CarouselItem{
						{
							ImageAlt: "First slide",
						},
						{
							ImageAlt: "Second slide",
						},
						{
							ImageAlt: "Third slide",
						},
					}),
				).Class("col-12 py-md-3 pl-md-3"),
			).Class("row"),
		).Class("container-fluid"),
	)
	return
}

var CompositeComponentSample1PagePB = web.Page(CompositeComponentSample1Page)

var CompositeComponentSample1PagePath = examples.URLPathByFunc(CompositeComponentSample1Page)

// @snippet_end

func fakeImage(title string) HTMLComponent {
	return RawHTML(fmt.Sprintf(`
<svg class="bd-placeholder-img bd-placeholder-img-lg d-block w-100" width="800" height="400" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid slice" focusable="false" role="img" aria-label="Placeholder: %s"><title>Placeholder</title><rect width="100%%" height="100%%" fill="#666"></rect><text x="40%%" y="50%%" fill="#444" dy=".3em">%s</text></svg>
`, title, title))
}
