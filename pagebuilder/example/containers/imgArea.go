package containers

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

type ImgArea struct {
	Style CommonStyle
	Image media_library.MediaBox `sql:"type:text;"`
}

func (this ImgArea) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *ImgArea) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
}

func SetImgAreaComponent(pb *pagebuilder.Builder, eb *presets.EditingBuilder) {
	fb := pb.GetPresetsBuilder().NewFieldsBuilder(presets.WRITE).Model(&ImgArea{}).Only("Image", "Style")
	// fb.
	// fb.Field("Style").WithContextValue(media.MediaBoxConfig, &media_library.MediaBoxConfig{

	// })

	// eb.Field("Tags").Nested(fb)

	SetCommonStyleComponent(pb, fb.Field("Style"))

	eb.Field("ImgArea").Nested(fb)
}
