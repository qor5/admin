package footer

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"

	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type footerContent struct {
	Group1      string
	Group1Item1 string
	Group1Item2 string
	Group1Item3 string
	Group1Item4 string

	Group2      string
	Group2Item1 string
	Group2Item2 string
	Group2Item3 string
	Group2Item4 string

	Group3      string
	Group3Item1 string
	Group3Item2 string
	Group3Item3 string
	Group3Item4 string

	Group4 string

	Group5      string
	Group5Item1 string
	Group5Item2 string
	Group5Item3 string
	Group5Item4 string

	ImgInitial  bool
	ImageUpload media_library.MediaBox `sql:"type:text;"`
}

func (this footerContent) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *footerContent) Scan(value interface{}) error {
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
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&footerContent{}).Only("ImageUpload")

	fb.Field("Group1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group4").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group5").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group1Item1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group1Item2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group1Item3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group1Item4").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group2Item1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group2Item2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group2Item3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group2Item4").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group3Item1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group3Item2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group3Item3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group3Item4").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group5Item1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group5Item2").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group5Item3").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
	})
	fb.Field("Group5Item4").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx).Type("input")
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
