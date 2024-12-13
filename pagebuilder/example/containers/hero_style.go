package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type heroStyle struct {
	Layout          string
	TopSpace        int
	BottomSpace     int
	ImgInitial      bool
	ImageBackground media_library.MediaBox `sql:"type:text;"`
}

func (this heroStyle) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *heroStyle) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetHeroStyleComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&heroStyle{}).Only("Layout", "TopSpace", "BottomSpace", "ImageBackground")

	fb.Field("Layout").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items([]string{"left", "right"})
	})

	fb.Field("TopSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx)
	})

	fb.Field("BottomSpace").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.TextField(obj, field, ctx)
	})
	// SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("Style").Nested(fb)
}
