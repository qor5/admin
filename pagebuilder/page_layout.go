package pagebuilder

import (
	"fmt"
	"strings"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

func defaultPageLayoutFunc(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent {
	var freeStyleCss h.HTMLComponent
	if len(input.FreeStyleCss) > 0 {
		freeStyleCss = h.Style(strings.Join(input.FreeStyleCss, "\n"))
	}

	js := "https://the-plant.com/assets/app/container.ebe884e.js"
	css := "https://the-plant.com/assets/app/container.ebe884e.css"
	domain := "https://example.qor5.theplant-dev.com"

	head := h.Components(
		input.SeoTags,
		input.CanonicalLink,
		h.Meta().Attr("http-equiv", "X-UA-Compatible").Content("IE=edge"),
		h.Meta().Content("true").Name("HandheldFriendly"),
		h.Meta().Content("yes").Name("apple-mobile-web-app-capable"),
		h.Meta().Content("black").Name("apple-mobile-web-app-status-bar-style"),
		h.Meta().Name("format-detection").Content("telephone=no"),

		h.Link("").Rel("stylesheet").Type("text/css").Href(css),
		h.If(len(input.EditorCss) > 0, input.EditorCss...),
		freeStyleCss,
		// RawHTML(dataLayer),
		input.StructuredData,
		scriptWithCodes(input.FreeStyleTopJs),
	)
	ctx.Injector.HTMLLang(input.LocaleCode)
	if input.WrapHead != nil {
		head = input.WrapHead(head)
	}
	ctx.Injector.HeadHTML(h.MustString(head, nil))
	bodies := h.Components(
		// It's required as the body first element!
		h.If(input.Header != nil, input.Header),
		body,
		h.If(input.Footer != nil, input.Footer),
		h.Script("").Src(js),
		scriptWithCodes(input.FreeStyleBottomJs),
	)
	if input.WrapBody != nil {
		bodies = input.WrapBody(bodies)
	}

	return h.Body(
		bodies,
	).Attr("data-site-domain", domain)
}

func scriptWithCodes(jscodes []string) h.HTMLComponent {
	var js h.HTMLComponent
	if len(jscodes) > 0 {
		js = h.Script(fmt.Sprintf(`
try {
	%s
} catch (error) {
	console.log(error);
}
`, strings.Join(jscodes, "\n")))
	}
	return js
}
