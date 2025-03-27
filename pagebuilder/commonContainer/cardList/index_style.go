package CardList

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/commonContainer/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/theplant/htmlgo"
)

type cardListStyle struct {
	ProductColumns int
	Visibility     []string
	TopSpace       int
	BottomSpace    int
	LeftSpace      int
	RightSpace     int
}

func (this cardListStyle) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *cardListStyle) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetStyleComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&cardListStyle{}).Only("ProductColumns", "Visibility", "TopSpace", "BottomSpace", "LeftSpace", "RightSpace")

	fb.Field("ProductColumns").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Items([]int{1, 2, 3, 4, 5, 6})
	})

	fb.Field("Visibility").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		return presets.SelectField(obj, field, ctx).Chips(true).Items(utils.CardListVisibilityOptions).ItemTitle("Label").ItemValue("Value").Multiple(true)
	})

	eb.Field("Style").Nested(fb).PlainFieldBody().HideLabel()
}
