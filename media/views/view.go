package views

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"sort"
	"strconv"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor5/cropper"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

type MediaBoxConfigKey int

const MediaBoxConfig MediaBoxConfigKey = iota
const I18nMediaLibraryKey i18n.ModuleKey = "I18nMediaLibraryKey"

func Configure(b *presets.Builder, db *gorm.DB) {
	b.FieldDefaults(presets.WRITE).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(MediaBoxComponentFunc(db)).
		SetterFunc(MediaBoxSetterFunc(db))

	b.FieldDefaults(presets.LIST).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(MediaBoxListFunc())

	registerEventFuncs(b.GetWebBuilder(), db)

	b.I18n().
		RegisterForModule(language.English, I18nMediaLibraryKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nMediaLibraryKey, Messages_zh_CN)
}

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
		mediaBox := media_library.MediaBox{}
		err = mediaBox.Scan(ctx.R.FormValue(jsonValuesField))
		if err != nil {
			return
		}
		descriptionField := fmt.Sprintf("%s.Description", field.Name)
		if len(mediaBox.Files) > 0 {
			mediaBox.Files[0].Description = ctx.R.FormValue(descriptionField)
		}
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

const (
	openFileChooserEvent   = "mediaLibrary_OpenFileChooserEvent"
	deleteFileEvent        = "mediaLibrary_DeleteFileEvent"
	cropImageEvent         = "mediaLibrary_CropImageEvent"
	loadImageCropperEvent  = "mediaLibrary_LoadImageCropperEvent"
	imageSearchEvent       = "mediaLibrary_ImageSearchEvent"
	imageJumpPageEvent     = "mediaLibrary_ImageJumpPageEvent"
	uploadFileEvent        = "mediaLibrary_UploadFileEvent"
	chooseFileEvent        = "mediaLibrary_ChooseFileEvent"
	updateDescriptionEvent = "mediaLibrary_UpdateDescriptionEvent"
)

func registerEventFuncs(hub web.EventFuncHub, db *gorm.DB) {
	hub.RegisterEventFunc(openFileChooserEvent, fileChooser(db))
	hub.RegisterEventFunc(deleteFileEvent, deleteFileField())
	hub.RegisterEventFunc(cropImageEvent, cropImage(db))
	hub.RegisterEventFunc(loadImageCropperEvent, loadImageCropper(db))
	hub.RegisterEventFunc(imageSearchEvent, searchFile(db))
	hub.RegisterEventFunc(imageJumpPageEvent, jumpPage(db))
	hub.RegisterEventFunc(uploadFileEvent, uploadFile(db))
	hub.RegisterEventFunc(chooseFileEvent, chooseFile(db))
	hub.RegisterEventFunc(updateDescriptionEvent, updateDescription(db))
}

func (b *QMediaBoxBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	if len(b.fieldName) == 0 {
		panic("FieldName required")
	}
	if b.value == nil {
		panic("Value required")
	}

	ctx := web.MustGetEventContext(c)
	registerEventFuncs(ctx.Hub, b.db)

	portalName := createPortalName(b.fieldName)

	return h.Components(
		VSheet(
			h.If(len(b.label) > 0,
				h.Label(b.label).Class("v-label theme--light"),
			),
			web.Portal(
				mediaBoxThumbnails(ctx, b.value, b.fieldName, b.config),
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

func mediaBoxThumbnailsPortalName(field string) string {
	return fmt.Sprintf("%s_portal_thumbnails", field)
}

func mediaBoxThumb(msgr *Messages, cfg *media_library.MediaBoxConfig,
	f media_library.File, field string, thumb string) h.HTMLComponent {
	size := cfg.Sizes[thumb]
	return VCard(
		VImg().Src(fmt.Sprintf("%s?%d", f.URL(thumb), time.Now().UnixNano())).Height(150),
		h.If(size != nil,
			VCardActions(
				VMenu(
					web.Slot(
						VChip(
							thumbName(thumb, size),
						).Small(true).Attr("v-on", "on").Attr("v-bind", "attrs"),
					).Name("activator").Scope("{ on, attrs }"),

					VCard(
						VCardTitle(h.Text(msgr.CropImage)),
						web.Portal().EventFunc(loadImageCropperEvent, field, fmt.Sprint(f.ID), thumb),
						VCardActions(
							VSpacer(),
							VBtn(msgr.Crop).Text(true).Color("primary").
								Attr("@click", web.Plaid().
									EventFunc(cropImageEvent, field, fmt.Sprint(f.ID), thumb, h.JSONString(cfg)).
									Go()),
						),
					).Width(600),
				).CloseOnContentClick(false).OffsetX(true).OffsetY(true),
			),
		),
	)
}

func loadImageCropper(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		//field := ctx.Event.Params[0]
		id := ctx.Event.ParamAsInt(1)
		thumb := ctx.Event.Params[2]

		var m media_library.MediaLibrary
		err = db.Find(&m, id).Error
		if err != nil {
			return
		}

		moption := m.GetMediaOption()

		size := moption.Sizes[thumb]
		if size == nil {
			return
		}

		c := cropper.Cropper().
			Src(m.File.URL("original")).
			AspectRatio(float64(size.Width), float64(size.Height)).
			Attr("@input", web.Plaid().
				FieldValue("CropOption", web.Var("JSON.stringify($event)")).
				String())
		cropOption := moption.CropOptions[thumb]
		if cropOption != nil {
			c.Value(cropper.Value{
				X:      float64(cropOption.X),
				Y:      float64(cropOption.Y),
				Width:  float64(cropOption.Width),
				Height: float64(cropOption.Height),
			})
		}

		r.Body = c
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

		if len(cropOption) == 0 {
			return
		}
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

		mb := &media_library.MediaBox{}
		err = mb.Scan(ctx.R.FormValue(fmt.Sprintf("%s.Values", field)))
		if err != nil {
			panic(err)
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, mb, field, cfg),
		})
		return
	}
}

func mediaBoxThumbnails(ctx *web.EventContext, mediaBox *media_library.MediaBox, field string, cfg *media_library.MediaBoxConfig) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	c := VContainer().Fluid(true)

	for _, f := range mediaBox.Files {
		row := VRow()

		if len(cfg.Sizes) == 0 {
			row.AppendChildren(
				VCol(
					mediaBoxThumb(msgr, cfg, f, field, "original"),
				).Cols(6).Sm(4).Xl(3).Class("pl-0"),
			)
		} else {

			var keys []string
			for k, _ := range cfg.Sizes {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for _, k := range keys {
				row.AppendChildren(
					VCol(
						mediaBoxThumb(msgr, cfg, f, field, k),
					).Cols(6).Sm(4).Xl(3).Class("pl-0"),
				)
			}
		}

		c.AppendChildren(row)
	}

	if len(mediaBox.Files) > 0 {

		fieldName := fmt.Sprintf("%s.Description", field)
		value := ctx.R.FormValue(fieldName)
		if len(value) == 0 {
			value = mediaBox.Files[0].Description
		}
		c.AppendChildren(
			VRow(
				VCol(
					VTextField().
						Value(value).
						Attr(web.VFieldName(fieldName)...).
						Label(msgr.DescriptionForAccessibility).
						Dense(true).
						HideDetails(true).
						Outlined(true),
				).Cols(12).Class("pl-0 pt-0"),
			),
		)
	}

	if field == "richeditor" {
		return h.Components(

			h.Input("").Type("hidden").
				Value(h.JSONString(mediaBox.Files)).
				Attr(web.VFieldName(fmt.Sprintf("%s.Values", field))...),
			VBtn(msgr.ChooseFile).
				Depressed(true).
				OnClick(createPortalName(field)),
		)
	}

	return h.Components(
		c,
		h.Input("").Type("hidden").
			Value(h.JSONString(mediaBox.Files)).
			Attr(web.VFieldName(fmt.Sprintf("%s.Values", field))...),
		VBtn(msgr.ChooseFile).
			Depressed(true).
			OnClick(openFileChooserEvent, field, h.JSONString(cfg)),
		h.If(mediaBox != nil && len(mediaBox.Files) > 0,
			VBtn(msgr.Delete).
				Depressed(true).
				OnClick(deleteFileEvent, field, h.JSONString(cfg)),
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

func deleteFileField() web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Event.Params[0]
		cfg := stringToCfg(ctx.Event.Params[1])
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, &media_library.MediaBox{}, field, cfg),
		})
		return
	}
}

func searchKeywordName(field string) string {
	return fmt.Sprintf("%s_file_chooser_search_keyword", field)
}
func currentPageName(field string) string {
	return fmt.Sprintf("%s_file_chooser_current_page", field)
}

func stringToCfg(v string) *media_library.MediaBoxConfig {
	var cfg media_library.MediaBoxConfig
	if len(v) == 0 {
		return &cfg
	}
	err := json.Unmarshal([]byte(v), &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}

func fileChooser(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		field := ctx.Event.Params[0]
		cfg := stringToCfg(ctx.Event.Params[1])

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
						VToolbarTitle(msgr.ChooseAFile),
						VSpacer(),
						VLayout(
							VTextField().
								SoloInverted(true).
								PrependIcon("search").
								Label(msgr.Search).
								Flat(true).
								Clearable(true).
								HideDetails(true).
								Value("").
								Attr("@keyup.enter", web.Plaid().
									EventFunc(imageSearchEvent, field, h.JSONString(cfg)).
									FieldValue(searchKeywordName(field), web.Var("$event")).
									Go()),
						).AlignCenter(true).Attr("style", "max-width: 650px"),
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

var MediaLibraryPerPage int64 = 39

func fileChooserDialogContent(db *gorm.DB, field string, ctx *web.EventContext, cfg *media_library.MediaBoxConfig) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)

	keyword := ctx.R.FormValue(searchKeywordName(field))
	var files []*media_library.MediaLibrary
	wh := db.Model(&media_library.MediaLibrary{}).Order("created_at DESC")
	currentPageInt, _ := strconv.ParseInt(ctx.R.FormValue(currentPageName(field)), 10, 64)
	if currentPageInt == 0 {
		currentPageInt = 1
	}

	if len(keyword) > 0 {
		wh = wh.Where("file ILIKE ?", fmt.Sprintf("%%%s%%", keyword))
	}

	var count int
	err := wh.Count(&count).Error
	if err != nil {
		panic(err)
	}
	perPage := MediaLibraryPerPage
	pagesCount := int(int64(count)/perPage + 1)
	if int64(count)%perPage == 0 {
		pagesCount--
	}

	wh = wh.Limit(perPage).Offset((currentPageInt - 1) * perPage)
	err = wh.Find(&files).Error
	if err != nil {
		panic(err)
	}

	row := VRow(
		VCol(
			h.Label("").Children(
				VCard(
					VCardTitle(h.Text(msgr.UploadFiles)),
					VIcon("backup").XLarge(true),
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
						Style("display:none").
						Attr("@change",
							web.Plaid().
								BeforeScript("vars.fileChooserUploadingFiles = $event.target.files").
								FieldValue("NewFiles", web.Var("$event")).
								EventFunc(uploadFileEvent, field, h.JSONString(cfg)).Go()),
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
					h.Div(
						VImg(
							h.If(needCrop,
								h.Div(
									VProgressCircular().Indeterminate(true),
									h.Span(msgr.Cropping).Class("text-h6 pl-2"),
								).Class("d-flex align-center justify-center v-card--reveal white--text").
									Style("height: 100%; background: rgba(0, 0, 0, 0.5)").
									Attr("v-if", fmt.Sprintf("vars.%s", croppingVar)),
							),
						).Src(f.File.URL("@qor_preview")).Height(200),
					).Attr("role", "button").
						Attr("@click", web.Plaid().
							BeforeScript(fmt.Sprintf("vars.%s = true", croppingVar)).
							EventFunc(chooseFileEvent, field, fmt.Sprint(f.ID), h.JSONString(cfg)).
							Go()),
					VCardText(
						h.Text(f.File.FileName),
						h.Input("").
							Style("width: 100%;").
							Placeholder(msgr.DescriptionForAccessibility).
							Value(f.File.Description).
							Attr("@change", web.Plaid().
								EventFunc(updateDescriptionEvent, field, fmt.Sprint(f.ID)).
								FieldValue("CurrentDescription", web.Var("$event.target.value")).
								Go(),
							),
						fileSizes(f),
					),
				).Attr(web.InitContextVars, fmt.Sprintf(`{%s: false}`, croppingVar)),
			).Cols(3),
		)
	}

	return h.Div(
		VSnackbar(h.Text(msgr.DescriptionUpdated)).
			Attr("v-model", "vars.snackbarShow").
			Top(true).
			Color("teal darken-1").
			Timeout(5000),
		VContainer(
			row,
			VRow(
				VCol().Cols(1),
				VCol(
					VPagination().
						Length(pagesCount).
						Value(int(currentPageInt)).
						Attr("@input", web.Plaid().
							FieldValue(currentPageName(field), web.Var("$event")).
							EventFunc(imageJumpPageEvent, field, h.JSONString(cfg)).
							Go()),
				).Cols(10),
			),
			VCol().Cols(1),
		).Fluid(true),
	).Attr(web.InitContextVars, `{snackbarShow: false}`)
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

func uploadFile(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Event.Params[0]
		cfg := stringToCfg(ctx.Event.Params[1])

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

		renderFileChooserDialogContent(ctx, &r, field, db, cfg)
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

func chooseFile(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Event.Params[0]
		id := ctx.Event.ParamAsInt(1)
		cfg := stringToCfg(ctx.Event.Params[2])

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
			Body: mediaBoxThumbnails(ctx, &mediaBox, field, cfg),
		})
		r.VarsScript = `vars.showFileChooser = false; ` + fmt.Sprintf("vars.%s = false", fileCroppingVarName(m.ID))

		return
	}
}

func updateDescription(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		//field := ctx.Event.Params[0]
		id := ctx.Event.ParamAsInt(1)

		var media media_library.MediaLibrary
		if err = db.Find(&media, id).Error; err != nil {
			return
		}

		media.File.Description = ctx.R.FormValue("CurrentDescription")
		if err = db.Save(&media).Error; err != nil {
			return
		}

		r.VarsScript = `vars.snackbarShow = true;`
		return
	}
}

func searchFile(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Event.Params[0]
		cfg := stringToCfg(ctx.Event.Params[1])
		ctx.R.Form[currentPageName(field)] = []string{"1"}

		renderFileChooserDialogContent(ctx, &r, field, db, cfg)
		return
	}
}

func jumpPage(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Event.Params[0]
		cfg := stringToCfg(ctx.Event.Params[1])
		renderFileChooserDialogContent(ctx, &r, field, db, cfg)
		return
	}
}

func renderFileChooserDialogContent(ctx *web.EventContext, r *web.EventResponse, field string, db *gorm.DB, cfg *media_library.MediaBoxConfig) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogContentPortalName(field),
		Body: fileChooserDialogContent(db, field, ctx, cfg),
	})
}
