package layouts

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

func DefaultPageLayoutFunc(body HTMLComponent, input *pagebuilder.PageLayoutInput, ctx *web.EventContext) HTMLComponent {

	var freeStyleCss HTMLComponent
	if len(input.FreeStyleCss) > 0 {
		freeStyleCss = Style(strings.Join(input.FreeStyleCss, "\n"))
	}

	js := "https://the-plant.com/assets/app/container.4f902c4.js"
	css := "https://the-plant.com/assets/app/container.4f902c4.css"
	domain := "https://example.qor5.theplant-dev.com"

	head := Components(
		Meta().Attr("charset", "utf-8"),
		input.SeoTags,
		input.CanonicalLink,
		Meta().Attr("http-equiv", "X-UA-Compatible").Content("IE=edge"),
		Meta().Content("true").Name("HandheldFriendly"),
		Meta().Content("yes").Name("apple-mobile-web-app-capable"),
		Meta().Content("black").Name("apple-mobile-web-app-status-bar-style"),
		Meta().Name("format-detection").Content("telephone=no"),
		Meta().Name("viewport").Content("width=device-width, initial-scale=1"),

		Link("").Rel("stylesheet").Type("text/css").Href(css),
		If(len(input.EditorCss) > 0, input.EditorCss...),
		freeStyleCss,
		// RawHTML(dataLayer),
		input.StructuredData,
		scriptWithCodes(input.FreeStyleTopJs),
	)
	ctx.Injector.HTMLLang(input.Page.LocaleCode)
	ctx.Injector.HeadHTML(MustString(head, nil))

	return Body(
		// It's required as the body first element!
		If(input.Header != nil, input.Header),
		body,
		If(input.Footer != nil, input.Footer),
		Script("").Src(js),
		scriptWithCodes(input.FreeStyleBottomJs),
	).Attr("data-site-domain", domain)

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
