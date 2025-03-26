package CardList

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"

	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type cardListContent struct {
	Title string
	// Body        string
	// Button      string
	// ButtonStyle string
	// ImgInitial bool
	ImageUpload1        media_library.MediaBox `sql:"type:text;"`
	ProductTitle1       string
	ProductDescription1 string
	ImageUpload2        media_library.MediaBox `sql:"type:text;"`
	ProductTitle2       string
	ProductDescription2 string
	ImageUpload3        media_library.MediaBox `sql:"type:text;"`
	ProductTitle3       string
	ProductDescription3 string
}

func (this cardListContent) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *cardListContent) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetContentComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder, db *gorm.DB) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&cardListContent{}).Only("Title", "ImageUpload1", "ImageUpload2", "ImageUpload3")

	fb.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic", "Color", "FontFamily", "Clear",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	fb.Field("ProductTitle1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("ProductTitle2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("ProductTitle3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("ProductDescription1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("ProductDescription2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("ProductDescription3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})

	// fb.Field("ButtonStyle").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	return presets.SelectField(obj, field, ctx).Items(tailwind.ButtonPresets)
	// })
	// fb.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	extensions := tiptap.TiptapExtensions()
	// 	return tiptap.TiptapEditor(db, field.Name).
	// 		Extensions(extensions).
	// 		MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
	// 		Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
	// 		Label(field.Label).
	// 		Disabled(field.Disabled)
	// })

	// SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("Content").Nested(fb).PlainFieldBody().HideLabel()
}
