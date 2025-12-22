package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
			pMsgr                 = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			field, id, thumb, cfg = getParams(ctx)
			mediaBox              = &media_library.MediaBox{}
		)
		err = mediaBox.Scan(ctx.R.FormValue(fmt.Sprintf("%s.Values", field)))
		if err != nil {
			panic(err)
		}

		err = db.First(&m, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			presets.ShowMessage(&r, pMsgr.RecordNotFound, ColorError)
		} else if err != nil {
			return
		}
		var (
			moption = m.GetMediaOption()
			size    = moption.Sizes[thumb]
		)
		if size == nil && thumb != base.DefaultSizeKey {
			return
		} else if thumb != base.DefaultSizeKey {
			base.SaleUpDown(m.File.Width, m.File.Height, size)
		}
		c := cropper.Cropper().
			Src(m.File.URL("original")).
			ViewMode(cropper.VIEW_MODE_FIT_WITHIN_CONTAINER).
			AutoCropArea(1).
			Attr("@update:model-value", "cropLocals.CropOption=JSON.stringify($event)")
		if size != nil {
			// scale up and down keep width/height ratio
			c.AspectRatio(float64(size.Width), float64(size.Height))
		}
		// Attr("style", "max-width: 800px; max-height: 600px;")

		cropOption := mediaBox.CropOptions[thumb]
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
		var (
			db                    = b.db
			cropOption            = ctx.R.FormValue("CropOption")
			field, id, thumb, cfg = getParams(ctx)
			mb                    = &media_library.MediaBox{}
			pMsgr                 = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		)
		err = mb.Scan(ctx.R.FormValue(fmt.Sprintf("%s.Values", field)))
		if err != nil {
			panic(err)
		}
		if cropOption != "" {
			cropValue := cropper.Value{}
			err = json.Unmarshal([]byte(cropOption), &cropValue)
			if err != nil {
				panic(err)
			}

			var (
				old media_library.MediaLibrary
				m   media_library.MediaLibrary
			)
			err = db.First(&m, id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				presets.ShowMessage(&r, pMsgr.RecordNotFound, ColorError)
				return r, nil
			} else if err != nil {
				return
			}
			db.Find(&old, id)

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
			cropID := uuid.New().String()
			// save new file with crop id
			m.File.CropID = map[string]string{thumb: cropID}
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
			b.onEdit(ctx, old, m)

			mb.Url = m.File.Url
			mb.FileSizes = m.File.FileSizes
			mb.CropOptions = m.File.CropOptions
			mb.Sizes = m.File.Sizes
			if mb.CropID == nil {
				mb.CropID = make(map[string]string)
			}
			mb.CropID[thumb] = cropID
			if thumb == base.DefaultSizeKey {
				mb.Width = int(cropValue.Width)
				mb.Height = int(cropValue.Height)
			}
			mb.ID = json.Number(fmt.Sprint(m.ID))

		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, mb, field, cfg, false, false),
		})
		return
	}
}
