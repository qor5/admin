package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/tiptap"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type PArea struct {
	Style CommonStyle
	Text  string
}

func (this PArea) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *PArea) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetPAreaComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder, db *gorm.DB) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&PArea{}).Only("Text", "Style")

	fb.Field("Text").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		extensions := tiptap.TiptapExtensions()
		return tiptap.TiptapEditor(db, field.Name).
			Extensions(extensions).
			MarkdownTheme("github"). // Match tiptap.ThemeGithubCSSComponentsPack
			Attr(presets.VFieldError(field.FormKey, fmt.Sprint(reflectutils.MustGet(obj, field.Name)), field.Errors)...).
			Label(field.Label).
			Disabled(field.Disabled)
	})

	SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("PArea").Nested(fb)
}
