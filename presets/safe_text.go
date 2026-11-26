package presets

import (
	h "github.com/theplant/htmlgo"
)

// safeText renders s as plain text that Vue will not compile.
//
// The whole page is compiled as a Vue template in the browser, and h.Text only
// escapes HTML special characters, not "{{ }}". Any object data rendered with
// h.Text (names, titles, field values) would therefore let "{{ ... }}" run as
// JavaScript in the viewer's session (client-side template injection).
// The v-pre goes on an inner span so attributes such as @click on the outer
// element keep working.
func safeText(s string) h.HTMLComponent {
	return h.Span(s).Attr("v-pre", true)
}
