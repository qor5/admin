package layouts

import (
	"fmt"
	"html/template"
	"strings"

	. "github.com/theplant/htmlgo"
)

type LayoutInput struct {
	SeoTags           template.HTML
	CanonicalLink     template.HTML
	StructuredData    template.HTML
	FreeStyleCss      []string
	FreeStyleTopJs    []string
	FreeStyleBottomJs []string
	Header            HTMLComponent
	HeaderClass       string
	Footer            HTMLComponent
	HeaderColor       string
	IsPreview         bool
	Locale            string
	LangSwitchUrls    []string
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

func DefaultLayout(input *LayoutInput, body HTMLComponent) HTMLComponent {
	if input.Locale == "" {
		input.Locale = "en"
	}

	if input.HeaderColor == "" {
		input.HeaderColor = "back"
	}

	//if input.Header == nil {
	//	input.Header = HeaderTemplate()
	//}

	//if input.Footer == nil {
	//	input.Footer = FooterTemplate()
	//}

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
				//input.Header,
				body,
				//input.Footer,
				Script("").Src(js),
				scriptWithCodes(input.FreeStyleBottomJs),
			).Attr("data-site-domain", domain),
		),
	)
}
