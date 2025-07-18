package example

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
	"github.com/qor5/web/v3"
)

// Example model with HTML content that needs sanitization
type BlogPost struct {
	ID              uint `gorm:"primarykey"`
	Title           string
	Content         string // Will be sanitized with default tiptap policy
	Summary         string // Will be sanitized with strict policy
	Article         string // Will be sanitized with UGC policy
	ExtendedContent string // Will be sanitized with extended tiptap policy (multimedia support)
	ExtendedUGC     string // Will be sanitized with extended UGC policy (custom data attributes)
	ExtendedStrict  string // Will be sanitized with extended strict policy (basic formatting)
	Body            string // No sanitization
}

// TiptapSanitizerExample demonstrates how to configure HTML sanitization
// for tiptap editors using SetterFunc approach
func TiptapSanitizerExample(db *gorm.DB) {
	b := presets.New()

	// Configure pagebuilder (no global configuration needed)
	pb := pagebuilder.New("/admin", db, b)

	// Configure model with different sanitization for different fields
	mb := b.Model(&BlogPost{})
	eb := mb.Editing()

	// Field with default tiptap sanitization (recommended for tiptap editors)
	eb.Field("Content").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetter) // Uses tiptap-specific policy

	// Field with strict sanitization
	eb.Field("Summary").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithPolicy("strict"))

	// Field with UGC sanitization (standard bluemonday UGC policy)
	eb.Field("Article").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithPolicy("ugc"))

	// Field with sanitization disabled
	eb.Field("Body").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		})
		// Note: No SetterFunc means no sanitization - just use default form handling

	// Field with custom sanitization policy
	customPolicy := bluemonday.UGCPolicy()
	customPolicy.AllowAttrs("class").OnElements("p", "div")
	customPolicy.AllowAttrs("style").OnElements("span")

	eb.Field("CustomField").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithConfig(&presets.HTMLSanitizerConfig{
			Enabled:      true,
			Policy:       customPolicy,
			PresetPolicy: "custom",
		}))

	// Field with extended tiptap policy - adds support for additional elements
	eb.Field("ExtendedContent").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithConfig(
			presets.ExtendHTMLSanitizerConfig(presets.SanitizerPolicyTiptap, func(basePolicy *bluemonday.Policy) *bluemonday.Policy {
				// Clone the base policy by creating a new one with all the same rules
				// Note: bluemonday doesn't have direct cloning, so we create a new policy
				extendedPolicy := bluemonday.NewPolicy()

				// Copy all rules from tiptap policy (this is a simplified approach)
				extendedPolicy.AllowElements("p", "br", "span", "div")
				extendedPolicy.AllowElements("strong", "b", "em", "i", "u", "s", "mark", "small", "sub", "sup")
				extendedPolicy.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")
				extendedPolicy.AllowElements("ul", "ol", "li")
				extendedPolicy.AllowElements("a")
				extendedPolicy.AllowAttrs("href").OnElements("a")
				extendedPolicy.AllowElements("img")
				extendedPolicy.AllowAttrs("src", "alt").OnElements("img")
				extendedPolicy.AllowElements("code", "pre")
				extendedPolicy.AllowElements("blockquote", "cite")
				extendedPolicy.AllowElements("table", "thead", "tbody", "tfoot", "tr", "td", "th")
				extendedPolicy.AllowElements("hr")

				// Add custom extensions
				extendedPolicy.AllowElements("video", "audio") // Allow multimedia
				extendedPolicy.AllowAttrs("controls", "autoplay").OnElements("video", "audio")
				extendedPolicy.AllowAttrs("src").OnElements("video", "audio")
				extendedPolicy.AllowElements("iframe") // Allow embeds (be careful!)
				extendedPolicy.AllowAttrs("src", "width", "height", "frameborder").OnElements("iframe")

				return extendedPolicy
			}),
		))

	// Field with extended UGC policy - adds custom data attributes
	eb.Field("ExtendedUGC").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithConfig(
			presets.ExtendHTMLSanitizerConfig(presets.SanitizerPolicyUGC, func(basePolicy *bluemonday.Policy) *bluemonday.Policy {
				// Create a new policy that extends UGC
				extendedPolicy := bluemonday.UGCPolicy() // Start fresh with UGC policy

				// Add custom attributes on top of UGC policy
				extendedPolicy.AllowAttrs("data-custom", "data-widget-id").Matching(regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)).Globally()
				extendedPolicy.AllowAttrs("role", "aria-label").Globally()

				return extendedPolicy
			}),
		))

	// Field with extended strict policy - adds basic formatting
	eb.Field("ExtendedStrict").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return tiptap.TiptapEditor(db, field.FormKey).
				Label(field.Label).
				Value(field.StringValue(obj)).
				ErrorMessages(field.Errors...)
		}).
		SetterFunc(presets.TiptapHTMLSetterWithConfig(
			presets.ExtendHTMLSanitizerConfig(presets.SanitizerPolicyStrict, func(basePolicy *bluemonday.Policy) *bluemonday.Policy {
				// Create a new policy based on strict policy with basic formatting
				extendedPolicy := bluemonday.StrictPolicy()

				// Add basic formatting elements
				extendedPolicy.AllowElements("strong", "em", "u")
				extendedPolicy.AllowElements("ul", "ol", "li")

				return extendedPolicy
			}),
		))

	// Install pagebuilder
	pb.Install(b)
}
