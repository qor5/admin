package examples_admin

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func MediaAllowTypesExample(b *presets.Builder, db *gorm.DB) http.Handler {
	mediaBuilder := media.New(db).AutoMigrate().
		AllowTypes(media_library.ALLOW_TYPE_IMAGE, media_library.ALLOW_TYPE_VIDEO).FileAccept("image/*,video/*")
	b.DataOperator(gorm2op.DataOperator(db)).Use(mediaBuilder)
	return b
}
