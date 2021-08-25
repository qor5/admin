package media_library_view

import (
	"context"
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
		cfg := field.ContextValue(MediaBoxConfig).(*media_library.MediaBoxConfig)
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return QMediaBox(db).
			FieldName(field.Name).
			Value(&mediaBox).
			Label(field.Label).
			Config(cfg)
	}
}

func MediaBoxSetterFunc(db *gorm.DB) presets.FieldSetterFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		jsonValuesField := fmt.Sprintf("%s.Values", field.Name)
		mediaBox := media_library.MediaBox{Values: ctx.R.FormValue(jsonValuesField)}
		err = reflectutils.Set(obj, field.Name, mediaBox)
		if err != nil {
			return
		}

		return
	}
}

type QMediaBoxBuilder struct {
	fieldName string
	label     string
	value     *media_library.MediaBox
	config    *media_library.MediaBoxConfig
	db        *gorm.DB
}

func QMediaBox(db *gorm.DB) (r *QMediaBoxBuilder) {
	r = &QMediaBoxBuilder{
		db: db,
	}
	return
}

func (b *QMediaBoxBuilder) FieldName(v string) (r *QMediaBoxBuilder) {
	b.fieldName = v
	return b
}

func (b *QMediaBoxBuilder) Value(v *media_library.MediaBox) (r *QMediaBoxBuilder) {
	b.value = v
	return b
}

func (b *QMediaBoxBuilder) Label(v string) (r *QMediaBoxBuilder) {
	b.label = v
	return b
}

func (b *QMediaBoxBuilder) Config(v *media_library.MediaBoxConfig) (r *QMediaBoxBuilder) {
	b.config = v
	return b
}

func (b *QMediaBoxBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	if len(b.fieldName) == 0 {
		panic("FieldName required")
	}
	if b.value == nil {
		panic("Value required")
	}

	ctx := web.MustGetEventContext(c)
	portalName := createPortalName(b.fieldName)

	ctx.Hub.RegisterEventFunc(portalName, fileChooser(b.db, b.fieldName, b.config))
	ctx.Hub.RegisterEventFunc(deleteEventName(b.fieldName), deleteFileField(b.db, b.fieldName, b.config))

	return h.Components(
		VSheet(
			h.If(len(b.label) > 0,
				h.Label(b.label).Class("v-label theme--light"),
			),
			web.Portal(
				mediaBoxThumbnails(b.value, b.fieldName, b.config),
			).Name(mediaBoxThumbnailsPortalName(b.fieldName)),
			web.Portal().Name(portalName),
		).Class("pb-4").
			Rounded(true).
			Attr(web.InitContextVars, `{showFileChooser: false}`),
	).MarshalHTML(c)

}

func createPortalName(field string) string {
	return fmt.Sprintf("%s_portal", field)
}

func deleteEventName(field string) string {
	return fmt.Sprintf("%s_delete", field)
}

func mediaBoxThumbnailsPortalName(field string) string {
	return fmt.Sprintf("%s_portal_thumbnails", field)
}

func mediaBoxThumb(f media_library.File, thumb string, size *media.Size) h.HTMLComponent {
	return VCard(
		VImg().Src(f.URL(thumb)).Height(150),
		h.If(size != nil,
			VCardActions(
				VBtn("").Children(
					thumbName(thumb, size),
				).Text(true).Small(true),
			),
		),
	)
}

func mediaBoxThumbnails(mediaBox *media_library.MediaBox, field string, cfg *media_library.MediaBoxConfig) h.HTMLComponent {
	c := VContainer().Fluid(true)

	for _, f := range mediaBox.Files {
		row := VRow()

		if len(cfg.Sizes) == 0 {
			row.AppendChildren(
				VCol(
					mediaBoxThumb(f, "original", nil),
				).Cols(4).Class("pl-0"),
			)
		} else {
			for k, size := range cfg.Sizes {
				row.AppendChildren(
					VCol(
						mediaBoxThumb(f, k, size),
					).Cols(4).Class("pl-0"),
				)
			}
		}

		c.AppendChildren(row)
	}

	return h.Components(
		c,
		web.Bind(
			h.Input("").Type("hidden").Value(h.JSONString(mediaBox.Files)),
		).FieldName(fmt.Sprintf("%s.Values", field)),
		VBtn("Choose File").
			Depressed(true).
			OnClick(createPortalName(field)),
		h.If(mediaBox != nil && len(mediaBox.Files) > 0,
			VBtn("Delete").
				Depressed(true).
				OnClick(deleteEventName(field)),
		),
	)
}

func MediaBoxListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return h.Td(h.Img("").Src(mediaBox.URL("@qor_preview")).Style("height: 48px;"))
	}
}

func dialogContentPortalName(field string) string {
	return fmt.Sprintf("%s_dialog_content", field)
}

func deleteFileField(db *gorm.DB, field string, config *media_library.MediaBoxConfig) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(&media_library.MediaBox{}, field, config),
		})
		return
	}
}

func fileChooser(db *gorm.DB, field string, cfg *media_library.MediaBoxConfig) web.EventFunc {
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

func fileChooserDialogContent(db *gorm.DB, field string, ctx *web.EventContext, cfg *media_library.MediaBoxConfig) h.HTMLComponent {
	uploadEventName := fmt.Sprintf("%s_upload", field)
	chooseEventName := fmt.Sprintf("%s_choose", field)
	updateMediaDescription := fmt.Sprintf("%s_update", field)

	ctx.Hub.RegisterEventFunc(uploadEventName, uploadFile(db, field, cfg))
	ctx.Hub.RegisterEventFunc(chooseEventName, chooseFile(db, field, cfg))
	ctx.Hub.RegisterEventFunc(updateMediaDescription, updateDescription(db, field, cfg))

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
										VProgressCircular().Indeterminate(true),
										h.Span("Cropping").Class("text-h6 pl-2"),
									).Class("d-flex align-center justify-center v-card--reveal white--text").
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
						web.Bind(
							h.Input("").
								Style("width: 100%;").
								Placeholder("description for accessibility").
								Value(f.File.Description),
						).On("change").
							EventFunc(updateMediaDescription, fmt.Sprint(f.ID)).
							FieldName(fmt.Sprintf("%v[%v].description", field, f.ID)),
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
			VChip(thumbName(k, size)).XSmall(true),
		)
	}
	return g

}

func thumbName(name string, size *media.Size) h.HTMLComponent {
	if size == nil {
		return h.Text(fmt.Sprintf("%s", name))
	}
	return h.Text(fmt.Sprintf("%s(%dx%d)", name, size.Width, size.Height))
}

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func uploadFile(db *gorm.DB, field string, cfg *media_library.MediaBoxConfig) web.EventFunc {
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

func chooseFile(db *gorm.DB, field string, cfg *media_library.MediaBoxConfig) web.EventFunc {
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

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(&mediaBox, field, cfg),
		})
		r.VarsScript = `vars.showFileChooser = false; ` + fmt.Sprintf("vars.%s = false", fileCroppingVarName(m.ID))

		return
	}
}

func updateDescription(db *gorm.DB, field string, cfg *media_library.MediaBoxConfig) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.Event.ParamAsInt(0)

		var media media_library.MediaLibrary
		if err = db.Find(&media, id).Error; err != nil {
			return
		}

		media.File.Description = ctx.R.FormValue(fmt.Sprintf("%v[%v].description", field, id))
		if err = db.Save(&media).Error; err != nil {
			return
		}

		return
	}
}
