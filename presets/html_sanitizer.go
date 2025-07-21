package presets

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sunfmin/reflectutils"

	"github.com/qor5/web/v3"
)

// HTMLSanitizerConfig defines the configuration for HTML sanitization
type HTMLSanitizerConfig struct {
	// Policy is the bluemonday policy to use for sanitization
	Policy *bluemonday.Policy
}

// SanitizerPolicyType defines the base policy type for extension
type HTMLSanitizerPolicyType string

const (
	HTMLSanitizerPolicyTiptap HTMLSanitizerPolicyType = "TIPTAP"
	HTMLSanitizerPolicyUGC    HTMLSanitizerPolicyType = "UGC"
	HTMLSanitizerPolicyStrict HTMLSanitizerPolicyType = "STRICT"
)

// CreateHTMLSanitizerPolicy creates a fresh base policy of the specified type
func CreateHTMLSanitizerPolicy(policyType HTMLSanitizerPolicyType) *bluemonday.Policy {
	switch policyType {
	case HTMLSanitizerPolicyTiptap:
		return CreateDefaultTiptapSanitizerPolicy()
	case HTMLSanitizerPolicyUGC:
		return bluemonday.UGCPolicy()
	case HTMLSanitizerPolicyStrict:
		return bluemonday.StrictPolicy()
	default:
		panic("unknown policy type: " + string(policyType))
	}
}

// CreateTiptapBasePolicy creates a fresh tiptap policy (for extension)
func CreateDefaultTiptapSanitizerPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// Basic text formatting
	p.AllowElements("p", "br", "span", "div")
	p.AllowElements("strong", "b", "em", "i", "u", "s", "mark", "small", "sub", "sup")

	// Headings
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")

	// Lists
	p.AllowElements("ul", "ol", "li")

	// Links - with safe attributes only
	p.AllowElements("a")
	p.AllowAttrs("href", "alt", "title").OnElements("a")
	p.AllowAttrs("target").Matching(regexp.MustCompile(`^(_blank|_self)?$`)).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^((noopener|noreferrer|nofollow)(\s+(noopener|noreferrer|nofollow))*)?$`)).OnElements("a")
	p.RequireNoReferrerOnLinks(false) // Allow links without rel attributes

	// Allow all common URL schemes for href attributes
	p.AllowURLSchemes("http", "https", "mailto", "tel", "ftp", "ftps")
	p.AllowRelativeURLs(true) // Allow relative URLs like "/page" or "../file"

	// Images - with safe attributes
	p.AllowElements("img")
	p.AllowAttrs("src", "alt").OnElements("img")
	p.AllowAttrs("width", "height").Matching(bluemonday.Integer).OnElements("img")
	p.AllowAttrs("style").Matching(regexp.MustCompile(`^(width|height|max-width|max-height):\s*\d+(\.\d+)?(px|em|rem|%|vw|vh)(\s*;\s*(width|height|max-width|max-height):\s*\d+(\.\d+)?(px|em|rem|%|vw|vh))*\s*;?\s*$`)).OnElements("img")
	// Additional image attributes for rich editors
	p.AllowAttrs("lockaspectratio").Matching(regexp.MustCompile(`^(true|false)$`)).OnElements("img")
	p.AllowAttrs("data-display").Matching(regexp.MustCompile(`^(inline|block|flex)$`)).OnElements("img")

	// Media elements - video, audio with safe attributes
	p.AllowElements("video", "audio", "source", "track")
	p.AllowAttrs("src", "controls", "autoplay", "loop", "muted", "preload").OnElements("video", "audio")
	p.AllowAttrs("src", "type").OnElements("source")
	p.AllowAttrs("kind", "src", "srclang", "label", "default").OnElements("track")

	// Figure and figcaption for media with captions
	p.AllowElements("figure", "figcaption")

	// Iframe for embedded content (videos, maps, etc.) - no domain restrictions
	p.AllowElements("iframe")
	p.AllowAttrs("src", "width", "height", "frameborder", "allowfullscreen", "allow", "title", "name", "sandbox").OnElements("iframe")
	p.AllowAttrs("loading").Matching(regexp.MustCompile(`^(lazy|eager)$`)).OnElements("iframe")

	// Code blocks
	p.AllowElements("code", "pre")

	// Blockquotes
	p.AllowElements("blockquote", "cite")

	// Tables
	p.AllowElements("table", "thead", "tbody", "tfoot", "tr", "td", "th", "colgroup", "col")
	p.AllowAttrs("colspan", "rowspan").Matching(bluemonday.Integer).OnElements("td", "th")

	// Horizontal rule
	p.AllowElements("hr")

	// Allow class attributes for styling (but filter dangerous values)
	classRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	p.AllowAttrs("class").Matching(classRegex).Globally()

	// Allow safe CSS styles including table-related and positioning properties
	safeCSSRegex := regexp.MustCompile(`^(color|background-color|font-size|font-weight|font-style|text-align|text-decoration|margin|padding|border|border-radius|line-height|width|height|min-width|min-height|max-width|max-height|position|top|bottom|left|right|z-index|display|float|clear):\s*[^;]+(\s*;\s*(color|background-color|font-size|font-weight|font-style|text-align|text-decoration|margin|padding|border|border-radius|line-height|width|height|min-width|min-height|max-width|max-height|position|top|bottom|left|right|z-index|display|float|clear):\s*[^;]+)*\s*;?\s*$`)
	p.AllowAttrs("style").Matching(safeCSSRegex).Globally()

	// Allow data attributes for tiptap functionality
	p.AllowAttrs("data-type", "data-id").Matching(regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)).Globally()
	// Additional data attributes for media containers
	p.AllowAttrs("data-video").Matching(regexp.MustCompile(`^(|true|false|[a-zA-Z0-9\-_]*)$`)).OnElements("div") // Allow empty or safe values for data-video

	return p
}

// SanitizeHTMLSetter creates a setter function for tiptap fields with default tiptap policy
func CreateHTMLSanitizer(policyConfig *HTMLSanitizerConfig) FieldSetterFunc {
	return func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.Form.Get(field.FormKey)

		v = policyConfig.Policy.Sanitize(v)

		return reflectutils.Set(obj, field.Name, v)
	}
}
