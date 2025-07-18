package presets

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
	"github.com/sunfmin/reflectutils"

	"github.com/qor5/web/v3"
)

// HTMLSanitizerConfig defines the configuration for HTML sanitization
type HTMLSanitizerConfig struct {
	// Enabled determines if HTML sanitization is enabled
	Enabled bool
	// Policy is the bluemonday policy to use for sanitization
	Policy *bluemonday.Policy
	// PresetPolicy allows using predefined policies
	PresetPolicy string // "strict", "ugc", "tiptap"
	// ExtendPolicy is a function that can extend the base policy with additional rules
	ExtendPolicy func(*bluemonday.Policy) *bluemonday.Policy
}

// TiptapHTMLSanitizerConfig returns a configuration specifically designed for Tiptap editor
// This allows common rich text elements while filtering dangerous attributes
func TiptapHTMLSanitizerConfig() *HTMLSanitizerConfig {
	return &HTMLSanitizerConfig{
		Enabled:      true,
		Policy:       CreateTiptapBasePolicy(),
		PresetPolicy: "tiptap",
	}
}

// UGCHTMLSanitizerConfig returns the standard UGC policy configuration
func UGCHTMLSanitizerConfig() *HTMLSanitizerConfig {
	return &HTMLSanitizerConfig{
		Enabled:      true,
		Policy:       bluemonday.UGCPolicy(),
		PresetPolicy: "ugc",
	}
}

// StrictHTMLSanitizerConfig returns a strict configuration
func StrictHTMLSanitizerConfig() *HTMLSanitizerConfig {
	return &HTMLSanitizerConfig{
		Enabled:      true,
		Policy:       bluemonday.StrictPolicy(),
		PresetPolicy: "strict",
	}
}

// getEffectivePolicy returns the policy to use, applying extensions if needed
func (config *HTMLSanitizerConfig) getEffectivePolicy() *bluemonday.Policy {
	if config.ExtendPolicy != nil {
		// Create a fresh base policy and apply extensions
		basePolicy := CreateBasePolicyByType(config.PresetPolicy)
		return config.ExtendPolicy(basePolicy)
	}
	return config.Policy
}

// SanitizeHTML applies the sanitization policy to the input HTML
func (config *HTMLSanitizerConfig) SanitizeHTML(input string) string {
	if !config.Enabled {
		return input
	}

	policy := config.getEffectivePolicy()
	if policy == nil {
		return input
	}

	return policy.Sanitize(input)
}

// SanitizerPolicyType defines the base policy type for extension
type SanitizerPolicyType string

const (
	SanitizerPolicyTiptap SanitizerPolicyType = "tiptap"
	SanitizerPolicyUGC    SanitizerPolicyType = "ugc"
	SanitizerPolicyStrict SanitizerPolicyType = "strict"
)

// CreateBasePolicyByType creates a fresh base policy of the specified type
func CreateBasePolicyByType(policyType string) *bluemonday.Policy {
	switch policyType {
	case "tiptap":
		return CreateTiptapBasePolicy()
	case "ugc":
		return bluemonday.UGCPolicy()
	case "strict":
		return bluemonday.StrictPolicy()
	default:
		return CreateTiptapBasePolicy()
	}
}

// CreateTiptapBasePolicy creates a fresh tiptap policy (for extension)
func CreateTiptapBasePolicy() *bluemonday.Policy {
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

// ExtendHTMLSanitizerConfig creates an extended sanitizer config based on the specified policy type
func ExtendHTMLSanitizerConfig(policyType SanitizerPolicyType, extendFunc func(*bluemonday.Policy) *bluemonday.Policy) *HTMLSanitizerConfig {
	return &HTMLSanitizerConfig{
		Enabled:      true,
		Policy:       nil, // Will be created by getEffectivePolicy
		PresetPolicy: string(policyType),
		ExtendPolicy: extendFunc,
	}
}

// ExtendTiptapHTMLSanitizerConfig creates an extended tiptap sanitizer config
// Deprecated: Use ExtendHTMLSanitizerConfig(SanitizerPolicyTiptap, extendFunc) instead
func ExtendTiptapHTMLSanitizerConfig(extendFunc func(*bluemonday.Policy) *bluemonday.Policy) *HTMLSanitizerConfig {
	return ExtendHTMLSanitizerConfig(SanitizerPolicyTiptap, extendFunc)
}

// ExtendUGCHTMLSanitizerConfig creates an extended UGC sanitizer config
// Deprecated: Use ExtendHTMLSanitizerConfig(SanitizerPolicyUGC, extendFunc) instead
func ExtendUGCHTMLSanitizerConfig(extendFunc func(*bluemonday.Policy) *bluemonday.Policy) *HTMLSanitizerConfig {
	return ExtendHTMLSanitizerConfig(SanitizerPolicyUGC, extendFunc)
}

// ExtendStrictHTMLSanitizerConfig creates an extended strict sanitizer config
// Deprecated: Use ExtendHTMLSanitizerConfig(SanitizerPolicyStrict, extendFunc) instead
func ExtendStrictHTMLSanitizerConfig(extendFunc func(*bluemonday.Policy) *bluemonday.Policy) *HTMLSanitizerConfig {
	return ExtendHTMLSanitizerConfig(SanitizerPolicyStrict, extendFunc)
}

// TiptapHTMLSetterWithConfig creates a setter function for tiptap fields with custom HTML sanitizer configuration
func TiptapHTMLSetterWithConfig(config *HTMLSanitizerConfig) FieldSetterFunc {
	return func(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.Form.Get(field.FormKey)

		if config != nil {
			v = config.SanitizeHTML(v)
		}

		return reflectutils.Set(obj, field.Name, v)
	}
}

// TiptapHTMLSetterWithPolicy creates a setter function for tiptap fields with a specific policy
func TiptapHTMLSetterWithPolicy(policyType string) FieldSetterFunc {
	var config *HTMLSanitizerConfig
	switch policyType {
	case "strict":
		config = StrictHTMLSanitizerConfig()
	case "ugc":
		config = UGCHTMLSanitizerConfig()
	case "tiptap":
		config = TiptapHTMLSanitizerConfig()
	default:
		config = TiptapHTMLSanitizerConfig() // Default to tiptap policy
	}

	return TiptapHTMLSetterWithConfig(config)
}

// TiptapHTMLSetter creates a setter function for tiptap fields with default tiptap policy
func TiptapHTMLSetter(obj interface{}, field *FieldContext, ctx *web.EventContext) (err error) {
	v := ctx.R.Form.Get(field.FormKey)

	// Apply tiptap-specific sanitization
	config := TiptapHTMLSanitizerConfig()
	v = config.SanitizeHTML(v)

	return reflectutils.Set(obj, field.Name, v)
}
