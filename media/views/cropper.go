package views

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/cropper"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func getParams(ctx *web.EventContext) (field string, id int, thumb string, cfg *media_library.MediaBoxConfig) {
	field = ctx.R.FormValue("field")

	id = ctx.QueryAsInt("id")
	thumb = ctx.R.FormValue("thumb")
	cfg = stringToCfg(ctx.R.FormValue("cfg"))
	return
}

func loadImageCropper(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		field, id, thumb, cfg := getParams(ctx)

		var m media_library.MediaLibrary
		err = db.Find(&m, id).Error
		if err != nil {
			return
		}

		moption := m.GetMediaOption()

		size := moption.Sizes[thumb]
		if size == nil && thumb != media.DefaultSizeKey {
			return
		}

		c := cropper.Cropper().
			Src(m.File.URL("original")+"?"+fmt.Sprint(time.Now().Nanosecond())).
			ViewMode(cropper.VIEW_MODE_FILL_FIT_CONTAINER).
			AutoCropArea(1).
			Attr("@input", web.Plaid().
				FieldValue("CropOption", web.Var("JSON.stringify($event)")).
				String())
		if size != nil {
			c.AspectRatio(float64(size.Width), float64(size.Height))
		}
		// Attr("style", "max-width: 800px; max-height: 600px;")

		cropOption := moption.CropOptions[thumb]
		if cropOption != nil {
			c.Value(cropper.Value{
				X:      float64(cropOption.X),
				Y:      float64(cropOption.Y),
				Width:  float64(cropOption.Width),
				Height: float64(cropOption.Height),
			})
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: cropperPortalName(field),
			Body: web.Scope(
				VDialog(
					VCard(
						VToolbar(
							VToolbarTitle(msgr.CropImage),
							VSpacer(),
							VBtn(msgr.Crop).Color("primary").
								Attr(":loading", "locals.cropping").
								Attr("@click", web.Plaid().
									BeforeScript("locals.cropping = true").
									EventFunc(cropImageEvent).
									Query("field", field).
									Query("id", fmt.Sprint(id)).
									Query("thumb", thumb).
									FieldValue("cfg", h.JSONString(cfg)).
									Go()),
						).Class("pl-2 pr-2"),
						VCardText(
							c,
						).Attr("style", "max-height: 500px"),
					),
				).ModelValue(true).
					Scrollable(true).
					MaxWidth("800px"),
			).Init(`{cropping: false}`).VSlot("{ locals }"),
		})
		return
	}

}
func cropImage(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		cropOption := ctx.R.FormValue("CropOption")
		// log.Println(cropOption, ctx.Event.Params)
		field, id, thumb, cfg := getParams(ctx)

		mb := &media_library.MediaBox{}
		err = mb.Scan(ctx.R.FormValue(fmt.Sprintf("%s.Values", field)))
		if err != nil {
			panic(err)
		}
		if len(cropOption) > 0 {
			cropValue := cropper.Value{}
			err = json.Unmarshal([]byte(cropOption), &cropValue)
			if err != nil {
				panic(err)
			}

			var m media_library.MediaLibrary
			err = db.Find(&m, id).Error
			if err != nil {
				return
			}

			moption := m.GetMediaOption()
			if moption.CropOptions == nil {
				moption.CropOptions = make(map[string]*media.CropOption)
			}
			moption.CropOptions[thumb] = &media.CropOption{
				X:      int(cropValue.X),
				Y:      int(cropValue.Y),
				Width:  int(cropValue.Width),
				Height: int(cropValue.Height),
			}
			moption.Crop = true
			err = m.ScanMediaOptions(moption)
			if err != nil {
				return
			}

			err = media.SaveUploadAndCropImage(db, &m)
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				return r, nil
			}

			mb.Url = m.File.Url
			mb.FileSizes = m.File.FileSizes
			if thumb == media.DefaultSizeKey {
				mb.Width = int(cropValue.Width)
				mb.Height = int(cropValue.Height)
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, mb, field, cfg, false),
		})
		return
	}
}
