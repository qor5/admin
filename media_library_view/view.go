package media_library_view

import (
	"encoding/json"
	"fmt"
	"mime/multipart"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/media"
	"github.com/qor/media/media_library"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type MediaBoxConfigKey int

const MediaBoxConfig MediaBoxConfigKey = iota

func MediaBoxComponentFunc(db *gorm.DB) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		portalName := createPortalName(field)

		cfg := field.ContextValue(MediaBoxConfig).(*media_library.MediaBoxConfig)
		ctx.Hub.RegisterEventFunc(portalName, fileChooser(db, field, cfg))

		mediaBox := field.Value(obj).(media_library.MediaBox)

		return h.Components(
			VSheet(

				h.Label(field.Label).Class("v-label theme--light"),
				web.Portal(
					mediaBoxThumbnails(mediaBox, field),
				).Name(mediaBoxThumbnailsPortalName(field)),
				VBtn("Choose File").
					Depressed(true).
					OnClick(portalName),
				web.Portal().Name(portalName),
			).Class("pb-4").
				Rounded(true).
				Attr(web.InitContextVars, `{showFileChooser: false}`),
		)
	}
}

func MediaBoxSetterFunc(db *gorm.DB) presets.FieldSetterFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		mediaBox := media_library.MediaBox{}
		err = reflectutils.Set(obj, field.Name, mediaBox)
		if err != nil {
			return
		}

		jsonValuesField := fmt.Sprintf("%s.Values", field.Name)

		err = reflectutils.Set(obj, jsonValuesField, ctx.R.FormValue(jsonValuesField))
		if err != nil {
			return
		}

		return
	}
}

func createPortalName(field *presets.FieldContext) string {
	return fmt.Sprintf("%s_portal", field.Name)
}

func mediaBoxThumbnailsPortalName(field *presets.FieldContext) string {
	return fmt.Sprintf("%s_portal_thumbnails", field.Name)
}

func mediaBoxThumbnails(mediaBox media_library.MediaBox, field *presets.FieldContext) h.HTMLComponent {
	row := VRow()
	for _, file := range mediaBox.Files {
		row.AppendChildren(
			VCol(
				VCard(
					VImg().Src(file.Url).Height(150),
				),
			).Cols(4).Class("pl-0"),
		)
	}

	return h.Components(
		VContainer(
			row,
		),
		web.Bind(
			h.Input("").Type("hidden").Value(h.JSONString(mediaBox.Files)),
		).FieldName(fmt.Sprintf("%s.Values", field.Name)),
	)
}

func dialogContentPortalName(field *presets.FieldContext) string {
	return fmt.Sprintf("%s_dialog_content", field.Name)
}

func fileChooser(db *gorm.DB, field *presets.FieldContext, cfg *media_library.MediaBoxConfig) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		//msgr := presets.MustGetMessages(ctx.R)
		portalName := createPortalName(field)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: portalName,
			Body: VDialog(
				VCard(
					VToolbar(
						VBtn("").
							Icon(true).
							Dark(true).
							Attr("@click", "vars.showFileChooser = false").
							Children(
								VIcon("close"),
							),
						VToolbarTitle("Choose a File"),
					).Color("primary").
						//MaxHeight(64).
						Flat(true).
						Dark(true),

					web.Portal(
						fileChooserDialogContent(db, field, ctx, cfg),
					).Name(dialogContentPortalName(field)),
				).Tile(true),
			).
				Fullscreen(true).
				//HideOverlay(true).
				Transition("dialog-bottom-transition").
				//Scrollable(true).
				Attr("v-model", "vars.showFileChooser"),
		})
		r.VarsScript = `setTimeout(function(){ vars.showFileChooser = true }, 100)`
		return
	}
}

func fileChooserDialogContent(db *gorm.DB, field *presets.FieldContext, ctx *web.EventContext, cfg *media_library.MediaBoxConfig) h.HTMLComponent {

	uploadEventName := fmt.Sprintf("%s_upload", field.Name)
	chooseEventName := fmt.Sprintf("%s_choose", field.Name)
	ctx.Hub.RegisterEventFunc(uploadEventName, uploadFile(db, field, cfg))
	ctx.Hub.RegisterEventFunc(chooseEventName, chooseFile(db, field, cfg))

	var files []*media_library.MediaLibrary
	db.Order("created_at DESC").Find(&files)

	row := VRow(
		VCol(
			h.Label("").Children(
				VCard(
					VCardTitle(h.Text("Upload files")),
					VIcon("backup").XLarge(true),
					web.Bind(
						//VFileInput().
						//	Class("justify-center").
						//	Label("New Files").
						//	Multiple(true).
						//	FieldName("NewFiles").
						//	PrependIcon("backup").
						//	Height(50).
						//	HideInput(true),
						h.Input("").
							Type("file").
							Attr("multiple", true).
							Style("display:none"),
					).On("change").
						FieldName("NewFiles").
						EventFunc(uploadEventName).
						EventScript("vars.fileChooserUploadingFiles = $event.target.files"),
				).
					Height(200).
					Class("d-flex align-center justify-center").
					Attr("role", "button").
					Attr("v-ripple", true),
			),
		).
			Cols(3),

		VCol(
			VCard(
				VProgressCircular().
					Color("primary").
					Indeterminate(true),
			).
				Class("d-flex align-center justify-center").
				Height(200),
		).
			Attr("v-for", "f in vars.fileChooserUploadingFiles").
			Cols(3),
	).
		Attr(web.InitContextVars, `{fileChooserUploadingFiles: []}`)

	for _, f := range files {
		_, needCrop := mergeNewSizes(f, cfg)
		croppingVar := fileCroppingVarName(f.ID)
		row.AppendChildren(
			VCol(
				VCard(
					web.Bind(
						h.Div(
							VImg(
								h.If(needCrop,
									h.Div(
										h.Text("Cropping"),
									).Class("d-flex align-center justify-center v-card--reveal text-h3 white--text").
										Style("height: 100%; background: rgba(0, 0, 0, 0.5)").
										Attr("v-if", fmt.Sprintf("vars.%s", croppingVar)),
								),
							).Src(f.File.URL("@qor_preview")).Height(200),
						).Attr("role", "button"),
					).On("click").
						EventFunc(chooseEventName, fmt.Sprint(f.ID)).
						EventScript(fmt.Sprintf("vars.%s = true", croppingVar)),
					VCardText(
						h.Text(f.File.FileName),
						fileSizes(f),
					),
				).Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, croppingVar)),
			).Cols(3),
		)
	}

	return VContainer(row).Fluid(true)
}

func fileCroppingVarName(id uint) string {
	return fmt.Sprintf("fileChooser%d_cropping", id)
}

func fileSizes(f *media_library.MediaLibrary) h.HTMLComponent {
	if len(f.File.Sizes) == 0 {
		return nil
	}
	g := VChipGroup().Column(true)
	for k, size := range f.File.GetSizes() {
		g.AppendChildren(
			VChip(h.Text(fmt.Sprintf("%s:%dx%d", k, size.Width, size.Height))).XSmall(true),
		)
	}
	return g

}

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func uploadFile(db *gorm.DB, field *presets.FieldContext, cfg *media_library.MediaBoxConfig) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var uf uploadFiles
		ctx.MustUnmarshalForm(&uf)
		for _, fh := range uf.NewFiles {
			m := media_library.MediaLibrary{}
			err1 := m.File.Scan(fh)
			if err1 != nil {
				panic(err)
			}
			err1 = db.Save(&m).Error
			if err1 != nil {
				panic(err1)
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName(field),
			Body: fileChooserDialogContent(db, field, ctx, cfg),
		})
		r.VarsScript = `vars.fileChooserUploadingFiles = []`
		return
	}
}

func mergeNewSizes(m *media_library.MediaLibrary, cfg *media_library.MediaBoxConfig) (sizes map[string]*media.Size, r bool) {
	sizes = make(map[string]*media.Size)
	for k, size := range cfg.Sizes {
		if m.File.Sizes[k] != nil {
			sizes[k] = m.File.Sizes[k]
			continue
		}
		sizes[k] = size
		r = true
	}
	return
}

func chooseFile(db *gorm.DB, field *presets.FieldContext, cfg *media_library.MediaBoxConfig) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.Event.ParamAsInt(0)

		var m media_library.MediaLibrary
		err = db.Find(&m, id).Error
		if err != nil {
			return
		}
		sizes, needCrop := mergeNewSizes(&m, cfg)

		if needCrop {
			err = m.ScanMediaOptions(media_library.MediaOption{
				Sizes: sizes,
				Crop:  true,
			})
			if err != nil {
				return
			}
			err = db.Save(&m).Error
			if err != nil {
				return
			}
		}

		mediaBox := media_library.MediaBox{}
		mediaBox.Files = append(mediaBox.Files, media_library.File{
			ID:          json.Number(fmt.Sprint(m.ID)),
			Url:         m.File.Url,
			VideoLink:   "",
			FileName:    m.File.FileName,
			Description: m.File.Description,
		})

		for key, _ := range cfg.Sizes {
			mediaBox.Files = append(mediaBox.Files, media_library.File{
				ID:          json.Number(fmt.Sprint(m.ID)),
				Url:         m.File.URL(key),
				VideoLink:   "",
				FileName:    m.File.FileName,
				Description: m.File.Description,
			})
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(mediaBox, field),
		})
		r.VarsScript = `vars.showFileChooser = false; ` + fmt.Sprintf("vars.%s = false", fileCroppingVarName(m.ID))

		return
	}
}
