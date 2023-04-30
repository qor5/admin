package views

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/qor5/admin/media"
	"github.com/qor5/admin/media/media_library"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func fileChooser(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		field := ctx.R.FormValue("field")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))

		portalName := mainPortalName(field)
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
									EventFunc(imageSearchEvent).
									Query("field", field).
									FieldValue("cfg", h.JSONString(cfg)).
									FieldValue(searchKeywordName(field), web.Var("$event")).
									Go()),
						).AlignCenter(true).Attr("style", "max-width: 650px"),
					).Color("primary").
						// MaxHeight(64).
						Flat(true).
						Dark(true),
					web.Portal().Name(deleteConfirmPortalName(field)),
					web.Portal(
						fileChooserDialogContent(db, field, ctx, cfg),
					).Name(dialogContentPortalName(field)),
				).Tile(true),
			).
				Fullscreen(true).
				// HideOverlay(true).
				Transition("dialog-bottom-transition").
				// Scrollable(true).
				Attr("v-model", "vars.showFileChooser"),
		})
		r.VarsScript = `setTimeout(function(){ vars.showFileChooser = true }, 100)`
		return
	}
}

func fileChooserDialogContent(db *gorm.DB, field string, ctx *web.EventContext, cfg *media_library.MediaBoxConfig) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)

	keyword := ctx.R.FormValue(searchKeywordName(field))

	type selectItem struct {
		Text  string
		Value string
	}
	const (
		orderByKey           = "order_by"
		orderByCreatedAt     = "created_at"
		orderByCreatedAtDESC = "created_at_desc"
	)
	orderBy := ctx.R.URL.Query().Get(orderByKey)

	var files []*media_library.MediaLibrary
	wh := db.Model(&media_library.MediaLibrary{})
	switch orderBy {
	case orderByCreatedAt:
		wh = wh.Order("created_at")
	default:
		orderBy = orderByCreatedAtDESC
		wh = wh.Order("created_at DESC")
	}
	currentPageInt, _ := strconv.Atoi(ctx.R.FormValue(currentPageName(field)))
	if currentPageInt == 0 {
		currentPageInt = 1
	}

	if len(cfg.Sizes) > 0 {
		cfg.AllowType = media_library.ALLOW_TYPE_IMAGE
	}

	if len(cfg.AllowType) > 0 {
		wh = wh.Where("selected_type = ?", cfg.AllowType)
	}

	if len(keyword) > 0 {
		wh = wh.Where("file ILIKE ?", fmt.Sprintf("%%%s%%", keyword))
	}

	var count int64
	err := wh.Count(&count).Error
	if err != nil {
		panic(err)
	}
	perPage := MediaLibraryPerPage
	pagesCount := int(count/int64(perPage) + 1)
	if count%int64(perPage) == 0 {
		pagesCount--
	}

	wh = wh.Limit(perPage).Offset((currentPageInt - 1) * perPage)
	err = wh.Find(&files).Error
	if err != nil {
		panic(err)
	}

	fileAccept := "*/*"
	if cfg.AllowType == media_library.ALLOW_TYPE_IMAGE {
		fileAccept = "image/*"
	}

	row := VRow(
		h.If(uploadIsAllowed(ctx.R) == nil,
			VCol(
				h.Label("").Children(
					VCard(
						VCardTitle(h.Text(msgr.UploadFiles)),
						VIcon("backup").XLarge(true),
						h.Input("").
							Attr("accept", fileAccept).
							Type("file").
							Attr("multiple", true).
							Style("display:none").
							Attr("@change",
								web.Plaid().
									BeforeScript("locals.fileChooserUploadingFiles = $event.target.files").
									FieldValue("NewFiles", web.Var("$event")).
									EventFunc(uploadFileEvent).
									Query("field", field).
									FieldValue("cfg", h.JSONString(cfg)).
									Go()),
					).
						Height(200).
						Class("d-flex align-center justify-center pa-6").
						Attr("role", "button").
						Attr("v-ripple", true),
				),
			).
				Cols(6).Sm(4).Md(3),

			VCol(
				VCard(
					VProgressCircular().
						Color("primary").
						Indeterminate(true),
				).
					Class("d-flex align-center justify-center").
					Height(200),
			).
				Attr("v-for", "f in locals.fileChooserUploadingFiles").
				Cols(6).Sm(4).Md(3),
		),
	)

	var initCroppingVars = []string{fileCroppingVarName(0) + ": false"}

	for i, f := range files {
		_, needCrop := mergeNewSizes(f, cfg)
		croppingVar := fileCroppingVarName(f.ID)
		initCroppingVars = append(initCroppingVars, fmt.Sprintf("%s: false", croppingVar))
		imgClickVars := fmt.Sprintf("vars.mediaShow = '%s'; vars.mediaName = '%s'; vars.isImage = %s", f.File.URL(), f.File.FileName, strconv.FormatBool(media.IsImageFormat(f.File.FileName)))

		row.AppendChildren(
			VCol(
				VCard(
					h.Div(
						h.If(
							media.IsImageFormat(f.File.FileName),
							VImg(
								h.If(needCrop,
									h.Div(
										VProgressCircular().Indeterminate(true),
										h.Span(msgr.Cropping).Class("text-h6 pl-2"),
									).Class("d-flex align-center justify-center v-card--reveal white--text").
										Style("height: 100%; background: rgba(0, 0, 0, 0.5)").
										Attr("v-if", fmt.Sprintf("locals.%s", croppingVar)),
								),
							).Src(f.File.URL(media_library.QorPreviewSizeName)).Height(200).Contain(true),
						).Else(
							fileThumb(f.File.FileName),
						),
					).AttrIf("role", "button", field != mediaLibraryListField).
						AttrIf("@click", web.Plaid().
							BeforeScript(fmt.Sprintf("locals.%s = true", croppingVar)).
							EventFunc(chooseFileEvent).
							Query("field", field).
							Query("id", fmt.Sprint(f.ID)).
							FieldValue("cfg", h.JSONString(cfg)).
							Go(), field != mediaLibraryListField).
						AttrIf("@click", imgClickVars, field == mediaLibraryListField),
					VCardText(
						h.A().Text(f.File.FileName).
							Attr("@click", imgClickVars),
						h.Input("").
							Style("width: 100%;").
							Placeholder(msgr.DescriptionForAccessibility).
							Value(f.File.Description).
							Attr("@change", web.Plaid().
								EventFunc(updateDescriptionEvent).
								Query("field", field).
								Query("id", fmt.Sprint(f.ID)).
								FieldValue("cfg", h.JSONString(cfg)).
								FieldValue("CurrentDescription", web.Var("$event.target.value")).
								Go(),
							).Readonly(updateDescIsAllowed(ctx.R, files[i]) != nil),
						h.If(media.IsImageFormat(f.File.FileName),
							fileChips(f),
						),
					),
					h.If(deleteIsAllowed(ctx.R, files[i]) == nil,
						VCardActions(
							VSpacer(),
							VBtn(msgr.Delete).
								Text(true).
								Attr("@click",
									web.Plaid().
										EventFunc(deleteConfirmationEvent).
										Query("field", field).
										Query("id", fmt.Sprint(f.ID)).
										FieldValue("cfg", h.JSONString(cfg)).
										Go(),
								),
						),
					),
				),
			).Cols(6).Sm(4).Md(3),
		)
	}

	return h.Div(
		VSnackbar(h.Text(msgr.DescriptionUpdated)).
			Attr("v-model", "vars.snackbarShow").
			Top(true).
			Color("primary").
			Timeout(5000),
		web.Scope(
			VContainer(
				h.If(field == mediaLibraryListField,
					VRow(
						VCol(
							VSelect().Items([]selectItem{
								{Text: msgr.UploadedAtDESC, Value: orderByCreatedAtDESC},
								{Text: msgr.UploadedAt, Value: orderByCreatedAt},
							}).ItemText("Text").ItemValue("Value").
								Label(msgr.OrderBy).
								FieldName(orderByKey).Value(orderBy).
								Attr("@change",
									web.GET().PushState(true).
										Query(orderByKey, web.Var("[$event]")).Go(),
								).
								Dense(true).Solo(true).Class("mb-n8"),
						).Cols(3),
					).Justify("end"),
				),
				row,
				VRow(
					VCol().Cols(1),
					VCol(
						VPagination().
							Length(pagesCount).
							Value(int(currentPageInt)).
							Attr("@input", web.Plaid().
								FieldValue(currentPageName(field), web.Var("$event")).
								EventFunc(imageJumpPageEvent).
								Query("field", field).
								FieldValue("cfg", h.JSONString(cfg)).
								Go()),
					).Cols(10),
				),
				VCol().Cols(1),
			).Fluid(true),
		).Init(fmt.Sprintf(`{fileChooserUploadingFiles: [], %s}`, strings.Join(initCroppingVars, ", "))).VSlot("{ locals }"),
		VOverlay(
			h.Img("").Attr(":src", "vars.isImage? vars.mediaShow: ''").
				Style("max-height: 80vh; max-width: 80vw; background: rgba(0, 0, 0, 0.5)"),
			h.Div(
				h.A(
					VIcon("info").Small(true).Class("mb-1"),
					h.Text("{{vars.mediaName}}"),
				).Attr(":href", "vars.mediaShow? vars.mediaShow: ''").Target("_blank").
					Class("white--text").Style("text-decoration: none;"),
			).Class("d-flex align-center justify-center pt-2"),
		).Attr("v-if", "vars.mediaName").Attr("@click", "vars.mediaName = null").ZIndex(10),
	).Attr(web.InitContextVars, `{snackbarShow: false, mediaShow: null, mediaName: null, isImage: false}`)
}

func fileChips(f *media_library.MediaLibrary) h.HTMLComponent {
	g := VChipGroup().Column(true)
	text := "original"
	if f.File.Width != 0 && f.File.Height != 0 {
		text = fmt.Sprintf("%s(%dx%d)", "original", f.File.Width, f.File.Height)
	}
	if f.File.FileSizes["original"] != 0 {
		text = fmt.Sprintf("%s %s", text, media.ByteCountSI(f.File.FileSizes["original"]))
	}
	g.AppendChildren(
		VChip(h.Text(text)).XSmall(true),
	)
	// if len(f.File.Sizes) == 0 {
	//	return g
	// }

	// for k, size := range f.File.GetSizes() {
	//	g.AppendChildren(
	//		VChip(thumbName(k, size)).XSmall(true),
	//	)
	// }
	return g

}

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func uploadFile(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))

		if err = uploadIsAllowed(ctx.R); err != nil {
			return
		}

		var uf uploadFiles
		ctx.MustUnmarshalForm(&uf)
		for _, fh := range uf.NewFiles {
			m := media_library.MediaLibrary{}

			if media.IsImageFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_IMAGE
			} else if media.IsVideoFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_VIDEO
			} else {
				m.SelectedType = media_library.ALLOW_TYPE_FILE
			}
			err1 := m.File.Scan(fh)
			if err1 != nil {
				panic(err)
			}

			err1 = media.SaveUploadAndCropImage(db, &m)
			if err1 != nil {
				panic(err)
			}
		}

		renderFileChooserDialogContent(ctx, &r, field, db, cfg)
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
		field := ctx.R.FormValue("field")
		id := ctx.QueryAsInt("id")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))

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
				panic(err)
			}
			err = db.Save(&m).Error
			if err != nil {
				panic(err)
			}

			err = media.SaveUploadAndCropImage(db, &m)
			if err != nil {
				panic(err)
			}
		}

		mediaBox := media_library.MediaBox{
			ID:          json.Number(fmt.Sprint(m.ID)),
			Url:         m.File.Url,
			VideoLink:   "",
			FileName:    m.File.FileName,
			Description: m.File.Description,
			FileSizes:   m.File.FileSizes,
			Width:       m.File.Width,
			Height:      m.File.Height,
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, &mediaBox, field, cfg, false),
		})
		r.VarsScript = `vars.showFileChooser = false`
		return
	}
}

func searchFile(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))

		ctx.R.Form[currentPageName(field)] = []string{"1"}

		renderFileChooserDialogContent(ctx, &r, field, db, cfg)
		return
	}
}

func jumpPage(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))
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
