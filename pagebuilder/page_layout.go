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

	js := "https://the-plant.com/assets/app/container.4f902c4.js"
	css := "https://the-plant.com/assets/app/container.4f902c4.css"
	domain := "https://example.qor5.theplant-dev.com"

	// tailwind ecosystem resources
	tailwindJs := "https://cdn.tailwindcss.com"
	alpineJs := "https://unpkg.com/alpinejs"
	InferCss := "https://rsms.me/inter/inter.css"

	pbThemeCss := []string{`
@import url('https://fonts.googleapis.com/css2?family=Pacifico&display=swap');
@import url('https://fonts.googleapis.com/css2?family=Monoton&family=Pacifico&display=swap');

:root {
/**
	--color-base: 47, 45, 39;
	--color-surface: 255, 255, 255;
	--color-brand: 137, 255, 176;
	--font-family-h1: 'Monoton';
	--font-family-h2: '';
	--font-family-p: '';
**/
/**
	--color-base: 255, 242, 185;
	--color-surface: 153, 93, 15;
	--color-brand: 229, 72, 77;
	--font-family-h1: 'Pacifico';
	--font-family-h2: 'Times New Roman';
	--font-family-p: 'Times New Roman';
**/

	--color-base: 255, 255, 255;
	--color-surface: 51, 51, 51;
	--color-brand: 62, 99, 221;
	--font-family-h1: '';
	--font-family-h2: '';
	--font-family-p: '';


	--font-family-fallback: InterVariable, system-ui, sans-serif;

	--pb-primary: #3e63dd;
	--pb-secondary: #5b6471;
	--pb-success: #30a46c;
	--pb-info: #0091ff;
	--pb-warning: #f76808;
	--pb-error: #e5484d;
}

.tailwind-scope .tw-theme-h1 {
	font-family: var(--font-family-h1), var(--font-family-fallback);
}

.tailwind-scope .tw-theme-h2 {
	font-family: var(--font-family-h2), var(--font-family-fallback);
}

.tailwind-scope .tw-theme-p {
	font-family: var(--font-family-p), var(--font-family-fallback);
}

.tw-theme-filter-container {
	position: relative;
}

.tw-theme-filter-container::before {
	--color-opacity: 0.2;
	content: '';
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	background-color: rgba(var(--color-brand), var(--color-opacity));
	z-index: 1;
	mix-blend-mode: multiply;
	pointer-events: none;
}

.tw-theme-bg-base {
	--color-opacity: 1;
	background-color: rgba(var(--color-base), var(--color-opacity));
}

.tw-theme-text-base {
	--color-opacity: 1;
	color: rgba(var(--color-base), var(--color-opacity));
}

.tw-theme-bg-brand {
	--color-opacity: 1;
	background-color: var(
		--tw-theme-bg-color,
		rgba(var(--color-brand), var(--color-opacity))
	);
}

.tw-theme-bg-brand-20 {
	--color-opacity: 0.2;
	background-color: var(
		--tw-theme-bg-color,
		rgba(var(--color-brand), var(--color-opacity))
	);
}

.tw-theme-text {
	--color-opacity: 1;
	color: rgba(var(--color-surface), var(--color-opacity));
}

.btn-primary {
  color: #fff;
  background-color: var(--pb-primary);
}
.btn-secondary {
  color: #fff;
  background-color: var(--pb-secondary);
}
.btn-info {
  color: #fff;
  background-color: var(--pb-info);
}
.btn-success {
  color: #fff;
  background-color: var(--pb-success);
}
.btn-error {
  color: #fff;
  background-color: var(--pb-error);
}
.btn-warning {
  color: #fff;
  background-color: var(--pb-warning);
}

	`}

	tailwindPluginJs := []string{
		`tailwind.config = {
		theme: {
		extend: {
					fontFamily: {
						sans: ["InterVariable", "sans-serif"],
					},
		}
		}
	}`,
	}
	tailwindOverrides := []string{
		`.tailwind-scope {
        h1,
        h2,
        h3,
        h4,
        h5,
        h6,
        a,
        p {
          font-family: InterVariable, system-ui, sans-serif;
        }
      }
		`,
	}

	head := h.Components(
		input.SeoTags,
		input.CanonicalLink,
		h.Meta().Attr("http-equiv", "X-UA-Compatible").Content("IE=edge"),
		h.Meta().Content("true").Name("HandheldFriendly"),
		h.Meta().Content("yes").Name("apple-mobile-web-app-capable"),
		h.Meta().Content("black").Name("apple-mobile-web-app-status-bar-style"),
		h.Meta().Name("format-detection").Content("telephone=no"),
		h.Link("").Rel("stylesheet").Type("text/css").Href(css),

		// tailwind ecosystem resources
		h.Link("").Rel("stylesheet").Type("text/css").Href(InferCss),
		h.Script("").Src(tailwindJs),
		h.Script("").Src(alpineJs).Attr("defer", "true"),

		h.If(len(input.EditorCss) > 0, input.EditorCss...),
		freeStyleCss,
		// RawHTML(dataLayer),
		input.StructuredData,
		scriptWithCodes(input.FreeStyleTopJs),
		scriptWithCodes(tailwindPluginJs),
		styleWithCodes(pbThemeCss),
		styleWithCodes(tailwindOverrides),
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
