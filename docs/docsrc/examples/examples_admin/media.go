package examples_admin

import (
	"net/http"

	"github.com/qor5/admin/v3/media/media_library"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type MediaRole struct {
	gorm.Model
	RoleName string
	MediaID  uint
}

func MediaExample(b *presets.Builder, db *gorm.DB) http.Handler {
	db.AutoMigrate(&MediaRole{})
	mediaBuilder := media.New(db).AutoMigrate().WrapSaverFunc(func(in media.SaverFunc) media.SaverFunc {
		return func(db *gorm.DB, obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = in(db, obj, id, ctx); err != nil {
				return
			}
			if id == "" {
				p := obj.(*media_library.MediaLibrary)
				roleName := "viewer"
				if p.Folder {
					roleName = "viewer_folder"
				}
				db.Save(&MediaRole{MediaID: p.ID, RoleName: roleName})

			}
			return
		}
	})
	b.DataOperator(gorm2op.DataOperator(db)).Use(mediaBuilder)
	b.MenuOrder("Default", "Simple", "Media Library")
	configDefaultMedia(b, db)
	configSimpleMedia(b, db)

	return b
}

func configDefaultMedia(b *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	db.AutoMigrate(&models.InputDemo{})
	mb := b.Model(&models.InputDemo{}).URIName("default").Label("Default Media")
	mb.Editing("MediaLibrary1")
	return mb
}

func configSimpleMedia(b *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	db.AutoMigrate(&models.InputDemo{})
	mb := b.Model(&models.InputDemo{}).URIName("simple").Label("Simple Media")
	mb.Editing("MediaLibrary1")
	mb.Editing().Field("MediaLibrary1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		in := obj.(*models.InputDemo)
		// If you use SimpleMediaBox, only the URL is required, no other params
		// In other words, you can use a string to open a media_box
		return media.SimpleMediaBox(
			in.MediaLibrary1.URL(),
			"MediaLibrary1",
			false,
			nil,
			db,
		)
	})
	return mb
}
