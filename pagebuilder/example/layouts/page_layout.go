package layouts

import (
	"fmt"
	"strings"

	"github.com/goplaid/web"
	"github.com/qor/qor5/pagebuilder"
	. "github.com/theplant/htmlgo"
)

func DefaultPageLayoutFunc(body HTMLComponent, input *pagebuilder.PageLayoutInput, ctx *web.EventContext) HTMLComponent {
	if input.IsEditor {
		return body
	}
	var seoTags HTMLComponent
	if len(input.SeoTags) > 0 {
		seoTags = RawHTML(input.SeoTags)
	}

	var canonicalLink HTMLComponent
	if len(input.CanonicalLink) > 0 {
		canonicalLink = Link(string(input.CanonicalLink)).Attr("ref", "canonical")
	}

	var structureData HTMLComponent
	if len(input.StructuredData) > 0 {
		structureData = RawHTML(input.StructuredData)
	}

	var freeStyleCss HTMLComponent
	if len(input.FreeStyleCss) > 0 {
		freeStyleCss = Style(strings.Join(input.FreeStyleCss, "\n"))
	}

	js := "https://the-plant.com/assets/app/container.9506d40.js"
	css := "https://the-plant.com/assets/app/container.9506d40.css"
	domain := "https://example.qor5.theplant-dev.com"

	return Components(
		RawHTML("<!DOCTYPE html>\n"),
		Tag("html").Attr("lang", input.Locale).Children(
			Head(
				Meta().Attr("charset", "utf-8"),
				seoTags,
				canonicalLink,
				Meta().Attr("http-equiv", "X-UA-Compatible").Content("IE=edge"),
				Meta().Content("true").Name("HandheldFriendly"),
				Meta().Content("yes").Name("apple-mobile-web-app-capable"),
				Meta().Content("black").Name("apple-mobile-web-app-status-bar-style"),
				Meta().Name("format-detection").Content("telephone=no"),
				Meta().Name("viewport").Content("width=device-width, initial-scale=1"),

				Link("").Rel("stylesheet").Type("text/css").Href(css),
				freeStyleCss,
				//RawHTML(dataLayer),
				structureData,
				scriptWithCodes(input.FreeStyleTopJs),
			),

			Body(
				//It's required as the body first element!
				If(input.Header != nil, input.Header),
				body,
				If(input.Footer != nil, input.Footer),
				Script("").Src(js),
				scriptWithCodes(input.FreeStyleBottomJs),
			).Attr("data-site-domain", domain),
		),
	)
}

func scriptWithCodes(jscodes []string) HTMLComponent {
	var js HTMLComponent
	if len(jscodes) > 0 {
		js = Script(fmt.Sprintf(`
try {
	%s
} catch (error) {
	console.log(error);
}
`, strings.Join(jscodes, "\n")))
	}
	return js
}
