package media

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/presets"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func fileChooser(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))
		portalName := mainPortalName(field)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: portalName,
			Body: web.Scope(
				VDialog(
					VCard(
						VToolbar(
							VToolbarTitle(msgr.ChooseAFile),
							VSpacer(),
							searchComponent(ctx, field, cfg, false),
							VBtn("").
								Icon("mdi-close").
								Theme(ThemeDark).
								Attr("@click", "vars.showFileChooser = false"),
						).Color(ColorBackground).
							// MaxHeight(64).
							Flat(true),
						web.Portal(
							fileChooserDialogContent(mb, field, ctx, cfg),
						).Name(dialogContentPortalName(field)),
					),
				).
					Fullscreen(true).
					// HideOverlay(true).
					Transition("dialog-bottom-transition").
					// Scrollable(true).
					Attr("v-model", "vars.showFileChooser"),
			).VSlot("{form,locals}"),
		})
		r.RunScript = `setTimeout(function(){ vars.showFileChooser = true }, 100)`
		return
	}
}

const (
	paramOrderByKey = "order_by"
	paramTypeKey    = "type"
	paramTab        = "tab"

	orderByCreatedAt     = "created_at"
	orderByCreatedAtDESC = "created_at_desc"

	typeAll   = "all"
	typeImage = "image"
	typeVideo = "video"
	typeFile  = "file"

	tabFiles   = "files"
	tabFolders = "folders"
)

type selectItem struct {
	Text  string
	Value string
}

func fileChooserDialogContent(mb *Builder, field string, ctx *web.EventContext,
	cfg *media_library.MediaBoxConfig,
) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	return h.Div(
		imageDialog(),
		VSnackbar(h.Text(msgr.DescriptionUpdated)).
			Attr("v-model", "vars.snackbarShow").
			Location("top").
			Color("primary").
			Timeout(5000),
		mediaLibraryContent(mb, field, ctx, cfg),
		VOverlay(
			h.Img("").Attr(":src", "vars.isImage? vars.mediaShow: ''").
				Style("max-height: 80vh; max-width: 80vw; background: rgba(0, 0, 0, 0.5)"),
			h.Div(
				h.A(
					VIcon("info").Size(SizeSmall).Class("mb-1"),
					h.Text("{{vars.mediaName}}"),
				).Attr(":href", "vars.mediaShow? vars.mediaShow: ''").Target("_blank").
					Class("white--text").Style("text-decoration: none;"),
			).Class("d-flex align-center justify-center pt-2"),
		).Attr("v-if", "vars.mediaName").Attr("@click", "vars.mediaName = null").ZIndex(10),
	).Attr(web.VAssign("vars",
		`{snackbarShow: false, mediaShow: null, mediaName: null, isImage: false,imagePreview:false,imageSrc:""}`)...)
}

func fileChips(f *media_library.MediaLibrary) h.HTMLComponent {
	text := "original"
	if f.File.Width != 0 && f.File.Height != 0 {
		text = fmt.Sprintf("%s(%dx%d)", "original", f.File.Width, f.File.Height)
	}
	if f.File.FileSizes["original"] != 0 {
		text = fmt.Sprintf("%s %s", text, base.ByteCountSI(f.File.FileSizes["original"]))
	}
	return h.Text(text)
}

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func uploadFile(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))

		if err = mb.uploadIsAllowed(ctx.R); err != nil {
			return
		}

		var uf uploadFiles
		ctx.MustUnmarshalForm(&uf)
		for _, fh := range uf.NewFiles {
			m := media_library.MediaLibrary{}

			if base.IsImageFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_IMAGE
			} else if base.IsVideoFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_VIDEO
			} else {
				m.SelectedType = media_library.ALLOW_TYPE_FILE
			}
			err = m.File.Scan(fh)
			if err != nil {
				panic(err)
			}
			if mb.currentUserID != nil {
				m.UserID = mb.currentUserID(ctx)
			}
			err = base.SaveUploadAndCropImage(mb.db, &m)
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				return r, nil
			}
		}

		renderFileChooserDialogContent(ctx, &r, field, mb, cfg)
		return
	}
}

func mergeNewSizes(m *media_library.MediaLibrary, cfg *media_library.MediaBoxConfig) (sizes map[string]*base.Size, r bool) {
	sizes = make(map[string]*base.Size)
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

func chooseFile(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := mb.db
		id := ctx.ParamAsInt(ParamMediaIDS)
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))

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

			err = base.SaveUploadAndCropImage(db, &m)
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				return r, nil
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
			Body: mediaBoxThumbnails(ctx, &mediaBox, field, cfg, false, false),
		})
		r.RunScript = `vars.showFileChooser = false`
		return
	}
}

func searchFile(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))

		ctx.R.Form[currentPageName(field)] = []string{"1"}

		renderFileChooserDialogContent(ctx, &r, field, mb, cfg)
		return
	}
}

func jumpPage(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))
		renderFileChooserDialogContent(ctx, &r, field, mb, cfg)
		return
	}
}

func renderFileChooserDialogContent(ctx *web.EventContext, r *web.EventResponse, field string, mb *Builder, cfg *media_library.MediaBoxConfig) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogContentPortalName(field),
		Body: fileChooserDialogContent(mb, field, ctx, cfg),
	})
}

func fileComponent(
	mb *Builder,
	field string,
	tab string,
	ctx *web.EventContext,
	f *media_library.MediaLibrary,
	msgr *Messages,
	cfg *media_library.MediaBoxConfig,
	initCroppingVars []string,
	event *string,
	menus *[]h.HTMLComponent,
) (title h.HTMLComponent, content h.HTMLComponent) {
	_, needCrop := mergeNewSizes(f, cfg)
	croppingVar := fileCroppingVarName(f.ID)
	initCroppingVars = append(initCroppingVars, fmt.Sprintf("%s: false", croppingVar))
	imgClickVars := fmt.Sprintf("vars.mediaShow = '%s'; vars.mediaName = '%s'; vars.isImage = %s", f.File.URL(), f.File.FileName, strconv.FormatBool(base.IsImageFormat(f.File.FileName)))

	src := f.File.URL()
	*menus = append(*menus,
		VListItem(h.Text("Copy")).Attr("@click", web.Plaid().
			EventFunc(CopyFileEvent).
			Query(ParamField, field).
			Query(paramTab, tab).
			Query(ParamCfg, h.JSONString(cfg)).
			Query(ParamParentID, ctx.Param(ParamParentID)).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Go()),
		VListItem(
			h.Text(msgr.DescriptionForAccessibility)).
			Attr("@click", web.Plaid().
				EventFunc(UpdateDescriptionDialogEvent).
				Query(ParamField, field).
				Query(paramTab, tab).
				Query(ParamCfg, h.JSONString(cfg)).
				Query(ParamParentID, ctx.Param(ParamParentID)).
				Query(ParamMediaIDS, fmt.Sprint(f.ID)).
				Go()),
	)
	if base.IsImageFormat(f.File.FileName) {
		*event = fmt.Sprintf(`vars.imageSrc="%s";vars.imagePreview=true;`, src)
	}
	title = h.Div(
		h.If(
			base.IsImageFormat(f.File.FileName),
			VImg(
				h.If(needCrop,
					h.Div(
						VProgressCircular().Indeterminate(true),
						h.Span(msgr.Cropping).Class("text-h6 pl-2"),
					).Class("d-flex align-center justify-center v-card--reveal white--text").
						Style("height: 100%; background: rgba(0, 0, 0, 0.5)").
						Attr("v-if", fmt.Sprintf("locals.%s", croppingVar)),
				),
			).Src(src).Height(120).Cover(true),
		).Else(
			fileThumb(f.File.FileName),
		),
	).AttrIf("role", "button", field != mediaLibraryListField).
		AttrIf("@click", web.Plaid().
			BeforeScript(fmt.Sprintf("locals.%s = true", croppingVar)).
			EventFunc(chooseFileEvent).
			Query(ParamField, field).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Query(ParamCfg, h.JSONString(cfg)).
			Go(), field != mediaLibraryListField).
		AttrIf("@click", imgClickVars, field == mediaLibraryListField)

	content = h.Components(
		web.Slot(
			web.Scope(
				VTextField().Attr(web.VField("name", f.File.FileName)...).Variant(VariantPlain),
			).VSlot(`{form}`),
		).Name("title"),
		web.Slot(h.If(base.IsImageFormat(f.File.FileName),
			fileChips(f))).Name("subtitle"),
	)

	return
}

func fileOrFolderComponent(
	mb *Builder,
	field string,
	tab string,
	ctx *web.EventContext,
	f *media_library.MediaLibrary,
	msgr *Messages,
	cfg *media_library.MediaBoxConfig,
	initCroppingVars []string,
	inMediaLibrary bool,
) h.HTMLComponent {
	var (
		title, content            h.HTMLComponent
		checkEvent                = fmt.Sprintf(`let arr=locals.select_ids;let find_id=%v; arr.includes(find_id)?arr.splice(arr.indexOf(find_id), 1):arr.push(find_id);`, f.ID)
		clickCardWithoutMoveEvent = "null"
	)
	menus := &[]h.HTMLComponent{
		VListItem(h.Text("Rename")).Attr("@click", web.Plaid().
			EventFunc(RenameDialogEvent).
			Query(ParamField, field).
			Query(paramTab, tab).
			Query(ParamCfg, h.JSONString(cfg)).
			Query(ParamParentID, ctx.Param(ParamParentID)).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Go()),
		VListItem(h.Text("Move to")).Attr("@click", fmt.Sprintf("locals.select_ids=[%v]", f.ID)),
		h.If(mb.deleteIsAllowed(ctx.R, f) == nil, VListItem(h.Text(msgr.Delete)).Attr("@click",
			web.Plaid().
				EventFunc(DeleteConfirmationEvent).
				Query(ParamField, field).
				Query(paramTab, tab).
				Query(ParamCfg, h.JSONString(cfg)).
				Query(ParamParentID, ctx.Param(ParamParentID)).
				Query(ParamMediaIDS, fmt.Sprint(f.ID)).
				Go())),
	}

	if f.Folder {
		title, content = folderComponent(mb, field, ctx, f, msgr)
		clickCardWithoutMoveEvent = web.Plaid().
			EventFunc(imageJumpPageEvent).
			Query(ParamField, field).
			Query(paramTab, tab).
			Query(ParamCfg, h.JSONString(cfg)).
			Query(ParamParentID, f.ID).Go() + fmt.Sprintf(";vars.media_parent_id=%v", f.ID)
		if inMediaLibrary {
			clickCardWithoutMoveEvent += ";" + web.Plaid().PushState(true).MergeQuery(true).Query(ParamParentID, f.ID).RunPushState()
		}
	} else {
		title, content = fileComponent(mb, field, tab, ctx, f, msgr, cfg, initCroppingVars, &clickCardWithoutMoveEvent, menus)

	}

	return VCard(
		VCheckbox().
			Attr(":model-value", fmt.Sprintf(`locals.select_ids.includes(%v)`, f.ID)).
			Attr("@update:model-value", checkEvent).
			Attr("style", "z-index:2").
			Class("position-absolute top-0 right-0").Attr("v-if", "locals.select_ids.length>0"),
		VCardText(
			VCard(
				title,
			).Height(120).Elevation(0),
		).Class("pa-0", W100),
		VCardItem(
			VCard(
				content,
				web.Slot(
					VMenu(
						web.Slot(
							VBtn("").Children(
								VIcon("mdi-dots-horizontal"),
							).Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
						).Name("activator").Scope("{ props }"),
						VList(
							*menus...,
						),
					),
				).Name(VSlotAppend),
			).Color(ColorGreyLighten5),
		).Class("pa-0"),
	).Class("position-relative").
		Hover(true).
		Attr("@click", fmt.Sprintf("if( locals.select_ids.length>0){%s}else{%s}", checkEvent, clickCardWithoutMoveEvent))
}

func folderComponent(
	mb *Builder,
	field string,
	ctx *web.EventContext,
	f *media_library.MediaLibrary,
	msgr *Messages,
) (title h.HTMLComponent, content h.HTMLComponent) {
	var count int64
	mb.db.Model(media_library.MediaLibrary{}).Where("parent_id = ?", f.ID).Count(&count)
	title = VCardText(VIcon("mdi-folder").Size(90).Color(ColorPrimary)).Class("d-flex justify-center align-center")
	content = h.Components(
		web.Slot(
			web.Scope(
				VTextField().Attr(web.VField("name", f.File.FileName)...).
					Variant(VariantPlain),
			).VSlot(`{form}`),
		).Name("title"),
		web.Slot(h.Text(fmt.Sprintf("%v items", count))).Name("subtitle"),
	)

	return
}

func parentFolders(field string, ctx *web.EventContext,
	cfg *media_library.MediaBoxConfig, db *gorm.DB, currentID, parentID uint, existed map[uint]bool, inMediaLibrary bool,
) (comps h.HTMLComponents) {
	if existed == nil {
		existed = make(map[uint]bool)
	}
	var (
		item    *VBreadcrumbsItemBuilder
		current *media_library.MediaLibrary
	)
	if currentID == 0 {
		return
	}
	if err := db.First(&current, currentID).Error; err != nil {
		return
	}
	item = VBreadcrumbsItem().Title(current.File.FileName)
	if currentID == parentID {
		item.Disabled(true)
	} else {
		breadcrumbsItemClickEvent(field, ctx, cfg, currentID, inMediaLibrary, item.Href("#"))
	}
	comps = append(comps, item)
	if current.ParentId == 0 || existed[current.ID] {
		comps = append(h.Components(breadcrumbsItemClickEvent(field, ctx, cfg, 0, inMediaLibrary, VBreadcrumbsItem().Title("/").Href("#"))), comps...)

		return
	}
	comps = append(h.Components(h.Text("/")), comps...)
	existed[currentID] = true
	return append(parentFolders(field, ctx, cfg, db, current.ParentId, parentID, existed, inMediaLibrary), comps...)
}

func breadcrumbsItemClickEvent(field string, ctx *web.EventContext,
	cfg *media_library.MediaBoxConfig, currentID uint, inMediaLibrary bool, item *VBreadcrumbsItemBuilder,
) *VBreadcrumbsItemBuilder {
	var clickEvent string

	if inMediaLibrary {
		clickEvent += web.Plaid().PushState(true).MergeQuery(true).Query(ParamParentID, currentID).RunPushState() + ";"
	}

	clickEvent += web.Plaid().
		EventFunc(imageJumpPageEvent).
		Query(paramTab, ctx.Param(paramTab)).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		Query(ParamParentID, currentID).
		Go()

	item.Attr("@click.prevent", clickEvent)

	return item
}

func imageDialog() h.HTMLComponent {
	return VDialog(
		VCard(
			VBtn("").Icon("mdi-close").
				Variant(VariantPlain).Attr("@click", "vars.imagePreview=false").
				Class("position-absolute right-0 top-0").Attr("style", "z-index:2;"),
			VImg().Attr(":src", "vars.imageSrc").Width(658),
		).Class("position-relative").Color(ColorBlack),
	).MaxWidth(658).Attr("v-model", "vars.imagePreview")
}

func mediaLibraryContent(mb *Builder, field string, ctx *web.EventContext,
	cfg *media_library.MediaBoxConfig,
) h.HTMLComponent {
	var (
		db             = mb.db
		keyword        = ctx.Param(searchKeywordName(field))
		tab            = ctx.Param(paramTab)
		orderByVal     = ctx.Param(paramOrderByKey)
		typeVal        = ctx.Param(paramTypeKey)
		parentID       = ctx.ParamAsInt(ParamParentID)
		msgr           = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		inMediaLibrary = strings.Contains(ctx.R.RequestURI, "/"+MediaLibraryURIName)
		wh             = db.Model(&media_library.MediaLibrary{})
		files          []*media_library.MediaLibrary
		bc             h.HTMLComponent
	)
	if tab == "" {
		tab = tabFiles
	}
	if mb.searcher != nil {
		wh = mb.searcher(wh, ctx)
	} else if mb.currentUserID != nil {
		wh = wh.Where("user_id = ? ", mb.currentUserID(ctx))
	}
	switch orderByVal {
	case orderByCreatedAt:
		wh = wh.Order("created_at")
	default:
		orderByVal = orderByCreatedAtDESC
		wh = wh.Order("created_at DESC")
	}

	switch typeVal {
	case typeImage:
		wh = wh.Where("selected_type = ?", media_library.ALLOW_TYPE_IMAGE)
	case typeVideo:
		wh = wh.Where("selected_type = ?", media_library.ALLOW_TYPE_VIDEO)
	case typeFile:
		wh = wh.Where("selected_type = ?", media_library.ALLOW_TYPE_FILE)
	default:
		typeVal = typeAll
	}
	if tab == tabFiles {
		wh = wh.Where("folder = ?", false)
	} else {
		wh = wh.Where("parent_id = ?", parentID)
		items := parentFolders(field, ctx, cfg, mb.db, uint(parentID), uint(parentID), nil, inMediaLibrary)
		bc = VBreadcrumbs(
			items...,
		)
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
	perPage := mb.mediaLibraryPerPage
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
		h.If(mb.uploadIsAllowed(ctx.R) == nil,
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

	initCroppingVars := []string{fileCroppingVarName(0) + ": false"}
	clickTabEvent := web.Plaid().
		EventFunc(imageJumpPageEvent).
		Query(paramTab, web.Var("$event")).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		Go()
	clickTabEvent += ";vars.media_tab=$event;vars.media_parent_id=0;"
	if inMediaLibrary {
		clickTabEvent += ";" + web.Plaid().PushState(true).MergeQuery(true).ClearMergeQuery([]string{ParamParentID}).Query(paramTab, web.Var("$event")).RunPushState()
	}

	for _, f := range files {
		var fileComp h.HTMLComponent
		fileComp = fileOrFolderComponent(mb, field, tab, ctx, f, msgr, cfg, initCroppingVars, inMediaLibrary)
		row.AppendChildren(
			VCol(fileComp).Cols(6).Sm(4).Md(3),
		)
	}
	return web.Scope(
		web.Portal().Name(deleteConfirmPortalName(field)),
		web.Portal().Name(newFolderDialogPortalName),
		web.Portal().Name(moveToFolderDialogPortalName),
		web.Portal().Name(renameDialogPortalName),
		web.Portal().Name(updateDescriptionDialogPortalName),
		VContainer(
			VRow(
				VCol(
					web.Scope(
						VTabs(
							VTab(h.Text(msgr.Files)).Value(tabFiles),
							VTab(h.Text("Folders")).Value(tabFolders),
						).Attr("v-model", "tabLocals.tab").
							Attr("@update:model-value",
								fmt.Sprintf(`$event=="%s"?null:%v`, tab, clickTabEvent),
							),
					).VSlot(`{locals:tabLocals}`).Init(fmt.Sprintf(`{tab:"%s"}`, tab)),
				),
				VSpacer(),
				VCol(
					VSelect().Items([]selectItem{
						{Text: msgr.All, Value: typeAll},
						{Text: msgr.Images, Value: typeImage},
						{Text: msgr.Videos, Value: typeVideo},
						{Text: msgr.Files, Value: typeFile},
					}).ItemTitle("Text").ItemValue("Value").
						Attr(web.VField(paramTypeKey, typeVal)...).
						Attr("@update:model-value",
							web.Plaid().EventFunc(imageJumpPageEvent).
								Query(paramTab, tab).
								Query(ParamField, field).
								Query(ParamCfg, h.JSONString(cfg)).
								Query(paramTypeKey, web.Var("$event")).
								Go(),
						).
						Density(DensityCompact).Variant(VariantTonal).Flat(true),
				).Cols(2),

				VCol(
					VSelect().Items([]selectItem{
						{Text: msgr.UploadedAtDESC, Value: orderByCreatedAtDESC},
						{Text: msgr.UploadedAt, Value: orderByCreatedAt},
					}).ItemTitle("Text").ItemValue("Value").
						Attr(web.VField(paramOrderByKey, orderByVal)...).
						Attr("@update:model-value",
							web.Plaid().EventFunc(imageJumpPageEvent).
								Query(paramTab, tab).
								Query(ParamField, field).
								Query(ParamCfg, h.JSONString(cfg)).
								Query(paramOrderByKey, web.Var("$event")).Go(),
						).
						Density(DensityCompact).Variant(VariantTonal).Flat(true),
				).Cols(3),
				VCol(
					h.If(
						tab == tabFolders,
						VBtn("New Folder").PrependIcon("mdi-plus").
							Variant(VariantOutlined).Class("mr-2").
							Attr("@click",
								web.Plaid().EventFunc(NewFolderDialogEvent).
									Query(paramTab, tab).
									Query(ParamField, field).
									Query(ParamCfg, h.JSONString(cfg)).
									Query(ParamParentID, ctx.Param(ParamParentID)).Go()),
					),
					h.If(mb.uploadIsAllowed(ctx.R) == nil,
						h.Div(
							VBtn("Upload file").PrependIcon("mdi-upload").Color(ColorSecondary).
								Attr("@click", "$refs.uploadInput.click()"),
							h.Input("").
								Attr("ref", "uploadInput").
								Attr("accept", fileAccept).
								Type("file").
								Attr("multiple", true).
								Style("display:none").
								Attr("@change",
									"form.NewFiles = [...$event.target.files];"+
										web.Plaid().
											BeforeScript("locals.fileChooserUploadingFiles = $event.target.files").
											EventFunc(uploadFileEvent).
											Query(ParamField, field).
											Query(ParamCfg, h.JSONString(cfg)).
											Go()),
						),
					),
				).Class("d-inline-flex"),
			).Justify("end"),
			VRow(
				VCol(bc),
			),
			row,
			VRow(
				VCol().Cols(1),
				VCol(
					VPagination().
						Length(pagesCount).
						ModelValue(int(currentPageInt)).
						Attr("@update:model-value", web.Plaid().
							FieldValue(currentPageName(field), web.Var("$event")).
							EventFunc(imageJumpPageEvent).
							Query(paramTab, tab).
							Query(ParamParentID, parentID).
							Query(ParamField, field).
							Query(ParamCfg, h.JSONString(cfg)).
							Go()),
				).Cols(10),
			),
			VCol().Cols(1),
			VRow(
				VCol(
					VCheckbox().HideDetails(true).
						BaseColor(ColorPrimary).
						ModelValue(true).
						Density(DensityCompact).
						Class("text-"+ColorPrimary).
						Indeterminate(true).
						Attr("@click", "locals.select_ids=[]").Children(
						web.Slot(
							h.Text(
								fmt.Sprintf("{{locals.select_ids.length}} %s", "Selected"),
							)).Name("label"),
					)).Cols(2),
				VCol(
					VBtn("Move to").Size(SizeSmall).Variant(VariantOutlined).
						Attr(":disabled", "locals.select_ids.length==0").
						Color(ColorSecondary).Class("ml-4").
						Attr("@click", web.Plaid().EventFunc(MoveToFolderDialogEvent).
							Query(ParamField, field).
							Query(paramTab, tab).
							Query(ParamCfg, h.JSONString(cfg)).
							Query(ParamSelectIDS, web.Var(`locals.select_ids.join(",")`)).Go()),
					VBtn("Delete").Size(SizeSmall).Variant(VariantOutlined).
						Color(ColorWarning).Class("ml-2").
						Attr("@click", web.Plaid().
							EventFunc(DeleteConfirmationEvent).
							Query(ParamField, field).
							Query(ParamParentID, parentID).
							Query(paramTab, tab).
							Query(ParamCfg, h.JSONString(cfg)).
							Query(ParamMediaIDS, web.Var(`locals.select_ids.join(",")`)).Go()),
				),
			).Class("d-flex align-center").Attr("v-if", "locals.select_ids && locals.select_ids.length>0"),
		).Fluid(true),
	).Init(fmt.Sprintf(`{fileChooserUploadingFiles: [], %s}`, strings.Join(initCroppingVars, ", "))).
		VSlot("{ locals,form}").Init("{select_ids:[]}")
}

func searchComponent(ctx *web.EventContext, field string, cfg *media_library.MediaBoxConfig, inMediaLibrary bool) h.HTMLComponent {

	var (
		msgr = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	)
	clickEvent := web.Plaid().
		EventFunc(imageSearchEvent).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		FieldValue(searchKeywordName(field), web.Var("searchLocals.msg"))
	if inMediaLibrary {
		clickEvent = clickEvent.MergeQuery(true)
	} else {
		clickEvent = clickEvent.
			Query(paramTab, web.Var("vars.media_tab")).
			Query(ParamParentID, web.Var("vars.media_parent_id"))
	}

	return web.Scope(
		VTextField().
			Density(DensityCompact).
			Variant(FieldVariantOutlined).
			Label(msgr.Search).
			Flat(true).
			Clearable(true).
			HideDetails(true).
			SingleLine(true).
			Attr("v-model", "searchLocals.msg").
			Attr("@keyup.enter", clickEvent.Go()).
			Children(
				web.Slot(VIcon("mdi-magnify")).Name("append-inner"),
			).MaxWidth(320),
	).VSlot("{locals:searchLocals}").Init(fmt.Sprintf(`{msg:""}`))
}
