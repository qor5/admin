package pagebuilder

import (
	"embed"
	"fmt"
	"strings"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

//go:embed commonContainer/assets/css
//go:embed commonContainer/assets/js
var theme embed.FS

func WrapDefaultPageLayout(pageLayoutFunc PageLayoutFunc) PageLayoutFunc {
	return func(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent {
		containerThemeCss, err := theme.ReadFile("commonContainer/assets/css/common-container-theme.css")
		if err != nil {
			panic(err)
		}

		containerJs, err := theme.ReadFile("commonContainer/assets/js/common-container-scope.js")
		if err != nil {
			panic(err)
		}

		configJs := []string{
			`window.TwindScope = {style: []};`,
			// https://twind.dev/handbook/configuration.html#preflight
			`window.TwindScope.config = {
						hash: false,
						theme: {
							extend: {
								fontFamily: {
									sans: ["InterVariable", "system-ui", "sans-serif"],
								},
							},
						},
				}`,
			fmt.Sprintf("window.TwindScope.style.push(`%s`)", string(func() []byte {
				css, err := theme.ReadFile("commonContainer/assets/css/common-container.css")
				if err != nil {
					panic(err)
				}
				return css
			}())),
		}

		input.FreeStyleCss = append(input.FreeStyleCss, string(containerThemeCss))
		input.FreeStyleTopJs = append(input.FreeStyleTopJs, strings.Join(configJs, "\n"))
		input.FreeStyleBottomJs = append(input.FreeStyleBottomJs, string(containerJs))

		return pageLayoutFunc(body, input, ctx)
	}
}

func defaultPageLayoutFunc(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent {
	return WrapDefaultPageLayout(pageLayoutFunc)(body, input, ctx)
}

func pageLayoutFunc(body h.HTMLComponent, input *PageLayoutInput, ctx *web.EventContext) h.HTMLComponent {
	var freeStyleCss h.HTMLComponent
	if len(input.FreeStyleCss) > 0 {
		freeStyleCss = h.Style(strings.Join(input.FreeStyleCss, "\n"))
	}

	js := "https://the-plant.com/assets/app/container.4f902c4.js"
	css := "https://the-plant.com/assets/app/container.4f902c4.css"
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
		// scriptWithCodes(twindScopeConfigJs),
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

func styleWithCodes(csscodes []string) h.HTMLComponent {
	var css h.HTMLComponent
	if len(csscodes) > 0 {
		css = h.Style(strings.Join(csscodes, "\n"))
	}
	return css
}
