package CardList

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/media"
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

type Product struct {
	Image       media_library.MediaBox `sql:"type:text;"`
	Title       string
	Description string
	Href        string
}

type cardListContent struct {
	Title string

	Products []*Product `sql:"type:text;"`
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
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&cardListContent{})
	eb.Field("Content").Nested(fb).PlainFieldBody().HideLabel()

	fb.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic", "Color", "FontFamily", "Clear", "TextAlign",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	fb1 := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&Product{}).Only("Image", "Title", "Description", "Href")

	fb1.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		product := obj.(*Product)
		if product.Title == "" {
			product.Title = "Commerce"
		}
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic", "Color", "FontFamily", "Clear", "TextAlign",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	fb1.Field("Description").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		product := obj.(*Product)
		if product.Description == "" {
			product.Description = "Ultra-reliable Omni-channel software tuned to the needs of any market.Ultra-reliable Ultra-reliable"
		}
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic", "Color", "FontFamily", "Clear", "TextAlign",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	fb1.Field("Image").WithContextValue(media.MediaBoxConfig, &media_library.MediaBoxConfig{
		AllowType: "image",
	})

	fb.Field("Products").Nested(fb1)
}
