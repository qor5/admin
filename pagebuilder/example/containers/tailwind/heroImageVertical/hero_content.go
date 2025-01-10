package heroImageVertical

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers/tailwind"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"

	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type heroContent struct {
	Title       string
	Body        string
	Button      string
	ButtonStyle string
	ImgInitial  bool
	ImageUpload media_library.MediaBox `sql:"type:text;"`
}

func (this heroContent) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *heroContent) Scan(value interface{}) error {
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
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&heroContent{}).Only("Title", "Body", "Button", "ButtonStyle", "ImageUpload")

	fb.Field("Body").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("textarea")
	})

	fb.Field("ButtonStyle").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items(tailwind.ButtonPresets)
	})
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
