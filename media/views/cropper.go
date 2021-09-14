package views

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor5/cropper"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	h "github.com/theplant/htmlgo"
)

func loadImageCropper(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		field := ctx.Event.Params[0]

		id := ctx.Event.ParamAsInt(1)
		thumb := ctx.Event.Params[2]
		cfg := ctx.Event.Params[3]

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
			Attr("@input", web.Plaid().
				FieldValue("CropOption", web.Var("JSON.stringify($event)")).
				String())
		if size != nil {
			c.AspectRatio(float64(size.Width), float64(size.Height))
		}
		//Attr("style", "max-width: 800px; max-height: 600px;")

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
			Body: VDialog(
				VCard(
					VToolbar(
						VToolbarTitle(msgr.CropImage),
						VSpacer(),
						VBtn(msgr.Crop).Color("primary").
							Attr(":loading", "locals.cropping").
							Attr("@click", web.Plaid().
								BeforeScript("locals.cropping = true").
								EventFunc(cropImageEvent, field, fmt.Sprint(id), thumb, h.JSONString(stringToCfg(cfg))).
								Go()),
					).Class("pl-2 pr-2"),
					VCardText(
						c,
					).Attr("style", "max-height: 500px"),
				),
			).Value(true).
				Scrollable(true).
				MaxWidth("800px").
				Attr(web.InitContextLocals, `{cropping: false}`),
		})
		return
	}

}
func cropImage(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		cropOption := ctx.R.FormValue("CropOption")
		//log.Println(cropOption, ctx.Event.Params)
		field := ctx.Event.Params[0]
		id := ctx.Event.ParamAsInt(1)
		thumb := ctx.Event.Params[2]
		cfg := stringToCfg(ctx.Event.Params[3])
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
			err = db.Save(&m).Error
			if err != nil {
				return
			}
			mb.FileSizes = m.File.FileSizes
			if thumb == media.DefaultSizeKey {
				mb.Width = int(cropValue.Width)
				mb.Height = int(cropValue.Height)
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, mb, field, cfg),
		})
		return
	}
}
