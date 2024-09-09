package media

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

const (
	cardHeight             = 146
	cardTitleHeight        = 90
	cardContentHeight      = 56
	cardWidth              = "w-100"
	chooseFileDialogWidth  = 1037
	chooseFileDialogHeight = 692
)

func fileChooser(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))
		portalName := mainPortalName(field)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: portalName,
			Body: web.Scope(
				VDialog(
					VCard(
						web.Portal(
							fileChooserDialogContent(mb, field, ctx, cfg),
						).Name(dialogContentPortalName(field)),
					).Height(chooseFileDialogHeight),
				).Width(chooseFileDialogWidth).Class("pa-6").
					// HideOverlay(true).
					Transition("dialog-bottom-transition").
					Attr("v-model", "vars.showFileChooser"),
			).VSlot("{form,locals}"),
		})
		r.RunScript = `setTimeout(function(){ vars.showFileChooser = true}, 100)`
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
			Color(ColorPrimary).
			Timeout(5000),
		mediaLibraryContent(mb, field, ctx, cfg),
	).Attr(web.VAssign("vars",
		`{snackbarShow: false,imagePreview:false,imageSrc:""}`)...)
}

func fileChips(f *media_library.MediaLibrary) h.HTMLComponent {
	text := "original"
	if f.File.Width != 0 && f.File.Height != 0 {
		text = fmt.Sprintf("%s(%dx%d)", "original", f.File.Width, f.File.Height)
	}
	if f.File.FileSizes["original"] != 0 {
		text = fmt.Sprintf("%s %s", text, base.ByteCountSI(f.File.FileSizes["original"]))
	}
	return h.Span(text).Attr("v-tooltip:bottom", h.JSONString(text))
}

type uploadFiles struct {
	NewFiles []*multipart.FileHeader
}

func uploadFile(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			field    = ctx.Param(ParamField)
			cfg      = stringToCfg(ctx.Param(ParamCfg))
			parentID = ctx.ParamAsInt(ParamParentID)
			msgr     = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		if err = mb.uploadIsAllowed(ctx.R); err != nil {
			return
		}

		var uf uploadFiles
		ctx.MustUnmarshalForm(&uf)
		for _, fh := range uf.NewFiles {
			m := media_library.MediaLibrary{ParentId: uint(parentID)}

			if base.IsImageFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_IMAGE
			} else if base.IsVideoFormat(fh.Filename) {
				m.SelectedType = media_library.ALLOW_TYPE_VIDEO
			} else {
				m.SelectedType = media_library.ALLOW_TYPE_FILE
			}
			if !mb.checkAllowType(m.SelectedType) {
				presets.ShowMessage(&r, msgr.UnSupportFileType, "error")
				return r, nil
			}
			err = m.File.Scan(fh)
			if err != nil {
				panic(err)
			}
			if mb.currentUserID != nil {
				m.UserID = mb.currentUserID(ctx)
			}
			err = mb.saverFunc(mb.db, &m, "", ctx)
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				return r, nil
			}
		}

		renderFileChooserDialogContent(ctx, &r, field, mb, cfg)
		r.RunScript = `vars.searchMsg = ""`
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

			err = mb.saverFunc(db, &m, strconv.Itoa(id), ctx)
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

func fileComponent(mb *Builder, field string, tab string, ctx *web.EventContext, f *media_library.MediaLibrary, msgr *Messages, cfg *media_library.MediaBoxConfig, initCroppingVars []string, event *string, menus *[]h.HTMLComponent, inMediaLibrary bool) (title, content h.HTMLComponent) {
	_, needCrop := mergeNewSizes(f, cfg)
	croppingVar := fileCroppingVarName(f.ID)
	initCroppingVars = append(initCroppingVars, fmt.Sprintf("%s: false", croppingVar))

	src := f.File.URL()
	*menus = append(*menus,
		h.If(mb.copyIsAllowed(ctx.R) == nil,
			VListItem(h.Text(msgr.Copy)).Attr("@click", web.Plaid().
				EventFunc(CopyFileEvent).
				Query(ParamField, field).
				Query(paramTab, tab).
				Query(ParamCfg, h.JSONString(cfg)).
				Query(ParamParentID, ctx.Param(ParamParentID)).
				Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
				Query(ParamMediaIDS, fmt.Sprint(f.ID)).
				Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
				Go()),
		),
		h.If(mb.updateDescIsAllowed(ctx.R, f) == nil,
			VListItem(
				h.Text(msgr.DescriptionForAccessibility)).
				Attr("@click", web.Plaid().
					EventFunc(UpdateDescriptionDialogEvent).
					Query(ParamField, field).
					Query(paramTab, tab).
					Query(ParamCfg, h.JSONString(cfg)).
					Query(ParamParentID, ctx.Param(ParamParentID)).
					Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
					Query(ParamMediaIDS, fmt.Sprint(f.ID)).
					Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
					Go()),
		),
	)
	clickEvent := fmt.Sprintf(`vars.imageSrc=%q;vars.imagePreview=true;`, src)
	if base.IsImageFormat(f.File.FileName) && inMediaLibrary {
		*event = clickEvent
	}
	fileNameComp := h.Span(f.File.FileName).Class("text-body-2").Attr("v-tooltip:bottom", h.JSONString(f.File.FileName))
	if !inMediaLibrary {
		fileNameComp.Class("text-"+ColorPrimary, "text-decoration-underline")
		fileNameComp.Attr("@click.stop", clickEvent)
		*event = web.Plaid().
			BeforeScript(fmt.Sprintf("locals.%s = true", croppingVar)).
			EventFunc(chooseFileEvent).
			Query(ParamField, field).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Query(ParamCfg, h.JSONString(cfg)).
			Go()
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
			).Src(src).Height(cardTitleHeight).Cover(true),
		).Else(
			fileThumb(f.File.FileName),
		),
	)

	content = h.Components(
		web.Slot(
			fileNameComp,
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
		menus                     []h.HTMLComponent
		checkEvent                = fmt.Sprintf(`let arr=locals.select_ids;let find_id=%v;arr.includes(find_id)?arr.splice(arr.indexOf(find_id), 1):arr.push(find_id);`, f.ID)
		moveToEvent               = fmt.Sprintf(`let arr=locals.select_ids;let find_id=%v;if(!arr.includes(find_id)){arr.push(find_id)};`, f.ID)
		clickCardWithoutMoveEvent = "null"
	)
	if mb.updateNameIsAllowed(ctx.R, f) == nil {
		menus = append(menus, VListItem(h.Text(msgr.Rename)).Attr("@click", web.Plaid().
			EventFunc(RenameDialogEvent).
			Query(ParamField, field).
			Query(paramTab, tab).
			Query(ParamCfg, h.JSONString(cfg)).
			Query(ParamParentID, ctx.Param(ParamParentID)).
			Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
			Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Go()))
	}

	if mb.moveToIsAllowed(ctx.R) == nil {
		menus = append(menus, VListItem(h.Text(msgr.MoveTo)).Attr("@click", moveToEvent))
	}
	if mb.deleteIsAllowed(ctx.R, f) == nil {
		menus = append(menus, VListItem(h.Text(msgr.Delete)).Attr("@click",
			web.Plaid().
				EventFunc(DeleteConfirmationEvent).
				Query(ParamField, field).
				Query(paramTab, tab).
				Query(ParamCfg, h.JSONString(cfg)).
				Query(ParamParentID, ctx.Param(ParamParentID)).
				Query(ParamMediaIDS, fmt.Sprint(f.ID)).
				Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
				Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
				Go()))
	}

	if f.Folder {
		title, content = folderComponent(mb, field, ctx, f, msgr)
		clickCardWithoutMoveEvent = web.Plaid().
			EventFunc(ImageJumpPageEvent).
			Query(ParamField, field).
			Query(paramTab, tab).
			Query(ParamCfg, h.JSONString(cfg)).
			Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
			Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
			Query(ParamParentID, f.ID).Go() + fmt.Sprintf(";vars.media_parent_id=%v", f.ID)
		if inMediaLibrary {
			clickCardWithoutMoveEvent += ";" + web.Plaid().PushState(true).MergeQuery(true).Query(ParamParentID, f.ID).RunPushState()
		}
	} else {
		title, content = fileComponent(mb, field, tab, ctx, f, msgr, cfg, initCroppingVars, &clickCardWithoutMoveEvent, &menus, inMediaLibrary)
	}

	card := VCard(
		h.If(inMediaLibrary && (mb.moveToIsAllowed(ctx.R) == nil || mb.deleteIsAllowed(ctx.R, nil) == nil),
			VCheckbox().
				Color(ColorPrimary).
				Attr(":model-value", fmt.Sprintf(`locals.select_ids.includes(%v)`, f.ID)).
				Attr("@update:model-value", checkEvent).
				Attr("@click", "$event.stopPropagation()").
				Attr("style", "z-index:2").
				Class("position-absolute top-0 right-0").Attr("v-if", "isHovering || locals.select_ids.length>0"),
		),
		VCardItem(
			VCard(
				title,
			).Height(cardTitleHeight).Elevation(0),
		).Class("pa-0", W100),
		VCardItem(
			VCard(
				VCardItem(
					content,
					h.If(inMediaLibrary && len(menus) != 0,
						web.Slot(
							VMenu(
								web.Slot(
									VBtn("").Children(
										VIcon("mdi-dots-horizontal"),
									).Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
								).Name("activator").Scope("{ props }"),
								VList(
									menus...,
								),
							),
						).Name(VSlotAppend),
					),
				).Class("pa-2"),
			).Color(ColorGreyLighten5).Height(cardContentHeight),
		).Class("pa-0"),
	).Class("position-relative").Attr("v-bind", "props").
		Hover(true).
		Width(cardWidth).Height(cardHeight).
		Attr("@click", fmt.Sprintf("if(locals.select_ids.length>0){%s}else{%s}", checkEvent, clickCardWithoutMoveEvent))

	return VHover(
		web.Slot(
			card,
		).Name("default").Scope(`{ isHovering, props }`),
	)
}

func folderComponent(mb *Builder, field string, ctx *web.EventContext, f *media_library.MediaLibrary, msgr *Messages) (title, content h.HTMLComponent) {
	var count int64
	fileNameComp := h.Span(f.File.FileName).Class("text-body-2").Attr("v-tooltip:bottom", h.JSONString(f.File.FileName))

	mb.db.Model(media_library.MediaLibrary{}).Where("parent_id = ?", f.ID).Count(&count)
	title = VCardText(h.RawHTML(folderSvg)).Class("d-flex justify-center align-center")
	content = h.Components(
		web.Slot(
			fileNameComp,
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
		EventFunc(ImageJumpPageEvent).
		Query(paramTab, ctx.Param(paramTab)).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
		Query(ParamParentID, currentID).
		AfterScript(fmt.Sprintf("vars.media_parent_id=%v", currentID)).
		Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
		Go()

	item.Attr("@click.prevent", clickEvent)

	return item
}

func imageDialog() h.HTMLComponent {
	return VDialog(
		VCard(
			VImg().Attr(":src", "vars.imageSrc").Width(658),
		).Class("position-relative").Color(ColorBlack).Attr("@click", "vars.imagePreview=false"),
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
		hasFolders     = false
		hasFiles       = false
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
	selected_type := ""
	switch typeVal {
	case typeImage:
		selected_type = media_library.ALLOW_TYPE_IMAGE
	case typeVideo:
		selected_type = media_library.ALLOW_TYPE_VIDEO
	case typeFile:
		selected_type = media_library.ALLOW_TYPE_FILE
	default:
		typeVal = typeAll
	}
	if tab == tabFiles {
		wh = wh.Where("folder = ?", false)
		if selected_type != "" {
			wh = wh.Where("selected_type = ?", selected_type)
		} else if len(mb.allowTypes) > 0 {
			wh = wh.Where("selected_type in ?", mb.allowTypes)
		}

	} else {
		wh = wh.Where("parent_id = ? ", parentID)
		if selected_type != "" {
			wh = wh.Where("folder = true or (folder = false and selected_type = ? ) ", selected_type)
		} else if len(mb.allowTypes) > 0 {
			wh = wh.Where("folder = true or (folder = false and selected_type in ? ) ", mb.allowTypes)
		}
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
		if tab == tabFiles {
			wh = wh.Where("selected_type = ?", cfg.AllowType)
		} else {
			wh = wh.Where("folder = true or (folder = false and selected_type = ? ) ", cfg.AllowType)
		}
	}

	if len(keyword) > 0 {
		wh = wh.Where("file::json->>'FileName' ILIKE ?", fmt.Sprintf("%%%s%%", keyword))
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
	rowFolder := VRow()
	rowFile := VRow(
		h.If(mb.uploadIsAllowed(ctx.R) == nil,
			VCol(
				VCard(
					VProgressCircular().
						Color(ColorPrimary).
						Indeterminate(true),
				).
					Class("d-flex align-center justify-center").
					Height(cardHeight).Width(cardWidth),
			).
				Attr("v-for", "f in locals.fileChooserUploadingFiles").Cols("auto"),
		),
	)

	initCroppingVars := []string{fileCroppingVarName(0) + ": false"}
	clickTabEvent := web.Plaid().
		EventFunc(ImageJumpPageEvent).
		Query(paramTab, web.Var("$event")).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
		Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
		Go()
	clickTabEvent += ";vars.media_tab=$event;vars.media_parent_id=0;"
	if inMediaLibrary {
		clickTabEvent += ";" + web.Plaid().PushState(true).MergeQuery(true).ClearMergeQuery([]string{ParamParentID}).Query(paramTab, web.Var("$event")).RunPushState()
	}
	for _, f := range files {
		fileComp := fileOrFolderComponent(mb, field, tab, ctx, f, msgr, cfg, initCroppingVars, inMediaLibrary)
		col := VCol(fileComp).Attr("style", "flex: 0 0 calc(100% / 5); max-width: calc(100% / 5);")
		if !f.Folder {
			hasFiles = true
			rowFile.AppendChildren(col)
		} else {
			hasFolders = true
			rowFolder.AppendChildren(col)
		}
	}

	return web.Scope(
		web.Portal().Name(deleteConfirmPortalName(field)),
		web.Portal().Name(newFolderDialogPortalName),
		web.Portal().Name(moveToFolderDialogPortalName),
		web.Portal().Name(renameDialogPortalName),
		web.Portal().Name(updateDescriptionDialogPortalName),
		VContainer(
			VRow(
				h.If(!inMediaLibrary,
					VCol(
						h.Div(VAppBarTitle().Text(msgr.ChooseFile),
							searchComponent(ctx, field, cfg, false),
							VBtn("").
								Icon("mdi-close").
								Variant(VariantText).
								Attr("@click", "vars.showFileChooser = false")).Class("d-flex justify-space-between align-center"),
					).Cols(12),
				),
				VCol(
					h.Div(
						h.If(mb.listFoldersIsAllowed(ctx.R) == nil,
							VCol(
								web.Scope(
									VTabs(
										VTab(h.Text(msgr.Files)).Value(tabFiles),
										VTab(h.Text(msgr.Folders)).Value(tabFolders),
									).Attr("v-model", "tabLocals.tab").
										Attr("@update:model-value",
											fmt.Sprintf(`$event==%q?null:%v`, tab, clickTabEvent),
										),
								).VSlot(`{locals:tabLocals}`).Init(fmt.Sprintf(`{tab:%q}`, tab)),
							),
						),
					),
					h.Div(
						VSelect().Items(mb.allowTypeSelectOptions(msgr)).ItemTitle("Text").ItemValue("Value").
							Attr(web.VField(paramTypeKey, typeVal)...).
							Attr("@update:model-value",
								web.Plaid().EventFunc(ImageJumpPageEvent).
									Query(paramTab, tab).
									Query(ParamField, field).
									Query(ParamCfg, h.JSONString(cfg)).
									Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
									Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
									Query(paramTypeKey, web.Var("$event")).
									Go(),
							).
							Density(DensityCompact).Variant(VariantTonal).Flat(true),
						VSelect().Items([]selectItem{
							{Text: msgr.UploadedAtDESC, Value: orderByCreatedAtDESC},
							{Text: msgr.UploadedAt, Value: orderByCreatedAt},
						}).ItemTitle("Text").ItemValue("Value").
							Attr(web.VField(paramOrderByKey, orderByVal)...).
							Attr("@update:model-value",
								web.Plaid().EventFunc(ImageJumpPageEvent).
									Query(paramTab, tab).
									Query(ParamField, field).
									Query(ParamCfg, h.JSONString(cfg)).
									Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
									Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
									Query(paramOrderByKey, web.Var("$event")).Go(),
							).
							Density(DensityCompact).Variant(VariantTonal).Flat(true),
						h.If(
							tab == tabFolders && mb.newFolderIsAllowed(ctx.R) == nil,
							VBtn(msgr.NewFolder).PrependIcon("mdi-plus").
								Variant(VariantOutlined).Class("mr-2").
								Attr("@click",
									web.Plaid().EventFunc(NewFolderDialogEvent).
										Query(paramTab, tab).
										Query(ParamField, field).
										Query(ParamCfg, h.JSONString(cfg)).
										Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
										Query(ParamParentID, ctx.Param(ParamParentID)).
										Go()),
						),
						h.If(mb.uploadIsAllowed(ctx.R) == nil,
							h.Div(
								VBtn(msgr.UploadFile).PrependIcon("mdi-upload").Color(ColorPrimary).
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
												EventFunc(UploadFileEvent).
												Query(paramTab, tab).
												Query(ParamParentID, parentID).
												Query(ParamField, field).
												Query(ParamCfg, h.JSONString(cfg)).
												Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
												Go()),
							),
						),
					).Class("d-inline-flex"),
				).Cols(12).Class("d-flex justify-space-between"),
			).Class("position-sticky top-0", "bg-"+ColorBackground).Attr("style", "z-index:2"),
			VRow(
				VCol(bc),
			),
			rowFolder,
			h.If(hasFiles && hasFolders, VDivider().Class("my-4")),
			rowFile,
			VRow(
				VCol(
					h.Div(
						VCheckbox().HideDetails(true).
							BaseColor(ColorPrimary).
							ModelValue(true).
							Density(DensityCompact).
							Class("text-"+ColorPrimary).
							Indeterminate(true).
							Class("mr-2").
							Attr("@click", "locals.select_ids=[]").Children(
							web.Slot(
								h.Text(
									fmt.Sprintf("{{locals.select_ids.length}} %s", "Selected"),
								)).Name("label"),
						),
						h.If(mb.moveToIsAllowed(ctx.R) == nil,
							VBtn(msgr.MoveTo).Size(SizeSmall).Variant(VariantOutlined).
								Attr(":disabled", "locals.select_ids.length==0").
								Color(ColorSecondary).Class("ml-2").
								Attr("@click", web.Plaid().EventFunc(MoveToFolderDialogEvent).
									Query(ParamField, field).
									Query(paramTab, tab).
									Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
									Query(ParamCfg, h.JSONString(cfg)).
									Query(ParamSelectIDS, web.Var(`locals.select_ids.join(",")`)).Go()),
						),
						h.If(mb.deleteIsAllowed(ctx.R, nil) == nil,
							VBtn(msgr.Delete).Size(SizeSmall).Variant(VariantOutlined).
								Color(ColorWarning).Class("ml-2").
								Attr("@click", web.Plaid().
									EventFunc(DeleteConfirmationEvent).
									Query(ParamField, field).
									Query(ParamParentID, parentID).
									Query(paramTab, tab).
									Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
									Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
									Query(ParamCfg, h.JSONString(cfg)).
									Query(ParamMediaIDS, web.Var(`locals.select_ids.join(",")`)).Go()),
						),
					).Class("d-flex align-center float-left").Attr("v-if", "locals.select_ids && locals.select_ids.length>0"),
					h.If(len(files) > 0,
						VPagination().
							Length(pagesCount).
							TotalVisible(5).
							NextIcon("mdi-page-last").
							PrevIcon("mdi-page-first").
							ModelValue(int(currentPageInt)).
							Attr("@update:model-value", web.Plaid().
								FieldValue(currentPageName(field), web.Var("$event")).
								EventFunc(ImageJumpPageEvent).
								Query(paramTab, tab).
								Query(ParamParentID, parentID).
								Query(ParamField, field).
								Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
								Query(searchKeywordName(field), ctx.Param(searchKeywordName(field))).
								Query(ParamCfg, h.JSONString(cfg)).
								Go()).Class("float-right"),
					),
				).Cols(12),
			).Class("position-sticky bottom-0", "bg-"+ColorBackground),
		).Fluid(true),
	).Init(fmt.Sprintf(`{fileChooserUploadingFiles: [], %s}`, strings.Join(initCroppingVars, ", "))).
		VSlot("{ locals,form}").Init(`{select_ids:[]}`)
}

func searchComponent(ctx *web.EventContext, field string, cfg *media_library.MediaBoxConfig, inMediaLibrary bool) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	clickEvent := web.Plaid().
		EventFunc(ImageSearchEvent).
		Query(ParamField, field).
		Query(ParamCfg, h.JSONString(cfg)).
		Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
		Query(searchKeywordName(field), web.Var("vars.searchMsg"))
	if inMediaLibrary {
		clickEvent = clickEvent.MergeQuery(true)
	} else {
		clickEvent = clickEvent.
			Query(paramTab, web.Var("vars.media_tab")).
			Query(ParamParentID, web.Var("vars.media_parent_id"))
	}
	event := clickEvent.Go()

	return VTextField().
		Density(DensityCompact).
		Variant(FieldVariantOutlined).
		Label(msgr.Search).
		Flat(true).
		Clearable(true).
		HideDetails(true).
		SingleLine(true).
		Attr("v-model", "vars.searchMsg").
		Attr(web.VAssign("vars", fmt.Sprintf(`{searchMsg:%q}`, ctx.Param(searchKeywordName(field))))...).
		Attr("@click:clear", `vars.searchMsg="";`+event).
		Attr("@keyup.enter", event).
		Children(
			web.Slot(VIcon("mdi-magnify").Attr("@click", event)).Name("append-inner"),
		).MaxWidth(320)
}
