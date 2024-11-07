package media

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/qor5/admin/v3/media/base"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/ui/cropper"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

const ParamMediaIDS = "media_ids"

func getParams(ctx *web.EventContext) (field string, id int, thumb string, cfg *media_library.MediaBoxConfig) {
	field = ctx.R.FormValue("field")

	id = ctx.ParamAsInt(ParamMediaIDS)
	thumb = ctx.R.FormValue("thumb")
	cfg = stringToCfg(ctx.R.FormValue("cfg"))
	return
}

func loadImageCropper(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			m                     media_library.MediaLibrary
			db                    = mb.db
			msgr                  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
			field, id, thumb, cfg = getParams(ctx)
		)

		err = db.Find(&m, id).Error
		if err != nil {
			return
		}
		var (
			moption = m.GetMediaOption()
			size    = moption.Sizes[thumb]
		)
		if size == nil && thumb != base.DefaultSizeKey {
			return
		}

		c := cropper.Cropper().
			Src(m.File.URL("original")).
			ViewMode(cropper.VIEW_MODE_FILL_FIT_CONTAINER).
			AutoCropArea(1).
			Attr("@update:model-value", "cropLocals.CropOption=JSON.stringify($event)")
		if size != nil {
			c.AspectRatio(float64(size.Width), float64(size.Height))
		}
		// Attr("style", "max-width: 800px; max-height: 600px;")

		cropOption := moption.CropOptions[thumb]
		if cropOption != nil {
			c.ModelValue(cropper.Value{
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
								Attr(":loading", "cropLocals.cropping").
								Attr("@click", web.Plaid().
									BeforeScript("cropLocals.cropping = true").
									EventFunc(cropImageEvent).
									Query(ParamField, field).
									Query(ParamMediaIDS, fmt.Sprint(id)).
									Query("thumb", thumb).
									Query("CropOption", web.Var("cropLocals.CropOption")).
									Query(ParamCfg, h.JSONString(cfg)).
									Go()),
						).Class("pl-2 pr-2"),
						VCardText(
							c,
						).Class("d-flex justify-center").Attr("style", "max-height: 500px"),
					),
				).ModelValue(true).
					Scrollable(true).
					MaxWidth("800px"),
			).Init(`{cropping: false}`).VSlot("{ locals: cropLocals}"),
		})
		return
	}
}

func cropImage(b *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := b.db
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
				moption.CropOptions = make(map[string]*base.CropOption)
			}
			moption.CropOptions[thumb] = &base.CropOption{
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

			err = b.saverFunc(db, &m, strconv.Itoa(id), ctx)
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				return r, nil
			}

			mb.Url = m.File.Url
			mb.FileSizes = m.File.FileSizes
			if thumb == base.DefaultSizeKey {
				mb.Width = int(cropValue.Width)
				mb.Height = int(cropValue.Height)
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, mb, field, cfg, false, false),
		})
		return
	}
}
