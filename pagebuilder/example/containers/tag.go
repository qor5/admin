package containers

import (
	"github.com/qor5/web/v3"
	v "github.com/qor5/x/v3/ui/vuetify"
	. "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

type tag struct {
	Text            string
	FontColor       string
	BackgroundColor string
	Icon            string
	Link            string
}

type filterTag struct {
	Text string
}

const (
	ICON_NO    = "no"
	ICON_SPEED = "speed"
)

var TagFontColors = []string{White, Blue, Grey}

var TagBackgroundColors = []string{White, Blue, Orange, Grey}

var TagIcons = []string{ICON_NO, ICON_SPEED}

var TAG_ICON_SPEED = RawHTML(`<svg viewBox="0 0 21 21" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M17.833 7.499l-1.076 1.618a7 7 0 01-.193 6.633H4.437a7 7 0 019.196-9.756l1.619-1.077a8.75 8.75 0 00-12.32 11.708 1.75 1.75 0 001.505.875h12.118a1.75 1.75 0 001.523-.875 8.75 8.75 0 00-.236-9.135l-.01.009z" fill="currentColor"/><path d="M9.266 13.484a1.749 1.749 0 002.476 0l4.953-7.43-7.429 4.953a1.751 1.751 0 000 2.477z" fill="currentColor"/></svg>`)

func SetTagComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&tag{}).
		Only("Text", "FontColor", "BackgroundColor", "Icon", "Link")

	fb.Field("FontColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return v.VAutocomplete().
			Variant(v.FieldVariantUnderlined).
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label).
			Items(TagFontColors)
	})

	fb.Field("BackgroundColor").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return v.VAutocomplete().
			Variant(v.FieldVariantUnderlined).
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label).
			Items(TagBackgroundColors)
	})

	fb.Field("Icon").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return v.VAutocomplete().
			Variant(v.FieldVariantUnderlined).
			Attr(presets.VFieldError(field.FormKey, field.Value(obj), field.Errors)...).
			Label(field.Label).
			Items(TagIcons)
	})

	eb.Field("Tags").Nested(fb)
}

func getTagIconSVG(icon string) HTMLComponent {
	if icon == ICON_NO {
		return nil
	}
	switch icon {
	case ICON_SPEED:
		return TAG_ICON_SPEED
	}
	return nil
}
