package imageWithText

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

type imageWithTextContent struct {
	Title       string
	Content     string
	Button      string
	ButtonHref  string
	ImgInitial  bool
	ImageUpload media_library.MediaBox `sql:"type:text;"`
}

func (this imageWithTextContent) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *imageWithTextContent) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetHeroContentComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder, db *gorm.DB) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&imageWithTextContent{}).Only("Title", "Content", "Button", "ButtonHref", "ImageUpload")

	fb.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	// fb.Field("ButtonStyle").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
	// 	return presets.SelectField(obj, field, ctx).Items(utils.ButtonPresets)
	// })

	fb.Field("Content").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return Div(
			tiptap.TiptapEditor(db, field.Name).
				Extensions(utils.TiptapExtensions(
					"Bold", "Italic",
					"Heading", "BulletList", "OrderedList",
					"Link",
				)).
				MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
				Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
				Label(field.Label).
				Disabled(field.Disabled),
		).Class("mb-5")
	})

	// SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("Content").Nested(fb).PlainFieldBody().HideLabel()
}
