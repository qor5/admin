package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/ui/cropper"
	"github.com/qor5/x/v3/ui/fileicons"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
)

type MediaBoxConfigKey int

const (
	MediaBoxConfig      MediaBoxConfigKey = iota
	I18nMediaLibraryKey i18n.ModuleKey    = "I18nMediaLibraryKey"

	ParamName               = "name"
	ParamParentID           = "parent_id"
	ParamSelectFolderID     = "select_folder_id"
	ParamSelectIDS          = "select_ids"
	ParamField              = "field"
	ParamCurrentDescription = "current_description"
	ParamCfg                = "cfg"
)

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(
		&media_library.MediaLibrary{},
	)
}

func configure(b *presets.Builder, mb *Builder, db *gorm.DB) {
	b.ExtraAsset("/cropper.js", "text/javascript", cropper.JSComponentsPack())
	b.ExtraAsset("/cropper.css", "text/css", cropper.CSSComponentsPack())

	b.FieldDefaults(presets.WRITE).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(MediaBoxComponentFunc(db, false)).
		SetterFunc(MediaBoxSetterFunc(db))

	b.FieldDefaults(presets.LIST).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(MediaBoxListFunc())

	b.FieldDefaults(presets.DETAIL).
		FieldType(media_library.MediaBox{}).
		ComponentFunc(MediaBoxComponentFunc(db, true))

	registerEventFuncs(b.GetWebBuilder(), mb)

	b.GetI18n().
		RegisterForModule(language.English, I18nMediaLibraryKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nMediaLibraryKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nMediaLibraryKey, Messages_ja_JP)

	configList(b, mb)
}

func MediaBoxComponentFunc(db *gorm.DB, readonly bool) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		cfg, ok := field.ContextValue(MediaBoxConfig).(*media_library.MediaBoxConfig)
		if !ok {
			cfg = &media_library.MediaBoxConfig{}
		}
		vErr, _ := ctx.Flash.(*web.ValidationErrors)
		if vErr == nil {
			vErr = &web.ValidationErrors{}
		}
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return QMediaBox(db).
			FieldName(field.FormKey).
			Value(&mediaBox).
			Label(field.Label).
			Config(cfg).
			Disabled(field.Disabled).
			Readonly(readonly).
			ErrorMessages(vErr.GetFieldErrors(fmt.Sprintf("%s.Values", field.FormKey))...)
	}
}

func MediaBoxSetterFunc(db *gorm.DB) presets.FieldSetterFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		jsonValuesField := fmt.Sprintf("%s.Values", field.FormKey)
		mediaBox := media_library.MediaBox{}
		err = mediaBox.Scan(ctx.Param(jsonValuesField))
		if err != nil {
			return
		}
		descriptionField := fmt.Sprintf("%s.Description", field.FormKey)
		mediaBox.Description = ctx.Param(descriptionField)
		err = reflectutils.Set(obj, field.Name, mediaBox)
		if err != nil {
			return
		}

		return
	}
}

type QMediaBoxBuilder struct {
	fieldName     string
	label         string
	value         *media_library.MediaBox
	config        *media_library.MediaBoxConfig
	db            *gorm.DB
	disabled      bool
	readonly      bool
	errorMessages []string
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

func (b *QMediaBoxBuilder) Disabled(v bool) (r *QMediaBoxBuilder) {
	b.disabled = v
	return b
}

func (b *QMediaBoxBuilder) Readonly(v bool) (r *QMediaBoxBuilder) {
	b.readonly = v
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

func (b *QMediaBoxBuilder) ErrorMessages(v ...string) (r *QMediaBoxBuilder) {
	b.errorMessages = v
	return b
}

func (b *QMediaBoxBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	if b.fieldName == "" {
		panic("FieldName required")
	}
	if b.value == nil {
		panic("Value required")
	}

	ctx := web.MustGetEventContext(c)
	errMessageFormKey := b.fieldName + ".Values"

	portalName := mainPortalName(b.fieldName)
	return h.Components(
		VSheet(
			h.If(b.label != "",
				h.Label(b.label).Class("v-label theme--light mb-2"),
			),
			web.Portal(
				mediaBoxThumbnails(ctx, b.value, b.fieldName, b.config, b.disabled, b.readonly),
			).Name(mediaBoxThumbnailsPortalName(b.fieldName)),
			web.Portal().Name(portalName),
			h.Div().Class("d-flex flex-column py-1 ga-1 text-caption").
				Attr(web.VAssign("dash.errorMessages", map[string]interface{}{errMessageFormKey: b.errorMessages})...).
				ClassIf("text-error", !b.disabled).
				ClassIf("text-grey", b.disabled).
				Children(
					h.Div(h.Text(fmt.Sprintf(`{{dash.errorMessages[%q][0]}}`, errMessageFormKey))).Attr("v-if", fmt.Sprintf(`dash.errorMessages[%q]`, errMessageFormKey)),
				),
		).
			Class("bg-transparent").
			Rounded(true),
	).MarshalHTML(c)
}

func mediaBoxThumb(msgr *Messages, cfg *media_library.MediaBoxConfig,
	f *media_library.MediaBox, field string, thumb string, disabled bool, option ...interface{},
) h.HTMLComponent {
	size := cfg.Sizes[thumb]
	fileSize := f.FileSizes[thumb]
	url := f.URLNoCached(thumb)
	if thumb == base.DefaultSizeKey {
		url = f.URLNoCached()
	}

	var ts interface{}
	if len(option) > 0 {
		ts = option[0]
	} else {
		ts = struct {
			Height interface{}
			Width  interface{}
		}{
			Height: 80,
			Width:  190,
		}
	}

	card := VCard(
		h.If(base.IsImageFormat(f.FileName),
			VImg().Src(url).Cover(true).Height(ts.(struct {
				Height interface{}
				Width  interface{}
			}).Height),
		).Else(
			h.Div(
				fileThumb(f.FileName),
				h.A().Text(f.FileName).Href(f.Url).Target("_blank"),
			).Style("text-align:center"),
		),
		h.If(base.IsImageFormat(f.FileName) && (size != nil || thumb == base.DefaultSizeKey) && !cfg.SimpleIMGURL,
			VCardActions(
				thumbName(thumb, size, fileSize, f),
			),
		),
	).Width(ts.(struct {
		Height interface{}
		Width  interface{}
	}).Width)

	if base.IsImageFormat(f.FileName) && (size != nil || thumb == base.DefaultSizeKey) && !disabled && !cfg.DisableCrop {
		card.Attr("@click", web.Plaid().
			EventFunc(loadImageCropperEvent).
			Query("field", field).
			Query(ParamMediaIDS, fmt.Sprint(f.ID)).
			Query("thumb", thumb).
			Query("cfg", h.JSONString(cfg)).
			Go())
	}
	return card
}

func fileThumb(filename string) h.HTMLComponent {
	return h.Div(
		fileicons.Icon(path.Ext(filename)[1:]).Attr("height", cardTitleHeight).Class("pt-4"),
	).Class("d-flex align-center justify-center")
}

func deleteConfirmation(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
			field = ctx.Param(ParamField)
			ids   = strings.Split(ctx.Param(ParamMediaIDS), ",")

			message string
		)
		if len(ids) == 0 {
			presets.ShowMessage(&r, "failed", ColorWarning)
			return
		} else if len(ids) == 1 {
			message = pMsgr.DeleteConfirmationText
		} else {
			message = msgr.DeleteObjects(len(ids))
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: deleteConfirmPortalName(field),
			Body: vx.VXDialog(
				h.Span(message),
			).
				Title(pMsgr.DialogTitleDefault).
				Attr("v-model", "vars.mediaLibrary_deleteConfirmation").
				Attr(web.VAssign("vars", `{mediaLibrary_deleteConfirmation: false}`)...).
				CancelText(pMsgr.Cancel).
				OkText(pMsgr.Delete).
				Attr("@click:ok", web.Plaid().
					EventFunc(DoDeleteEvent).
					BeforeScript("vars.mediaLibrary_deleteConfirmation =false").
					Queries(ctx.Queries()).
					Go()),
		})

		r.RunScript = "setTimeout(function(){ vars.mediaLibrary_deleteConfirmation = true }, 100)"
		return
	}
}

func doDelete(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			db              = mb.db
			ids             = strings.Split(ctx.Param(ParamMediaIDS), ",")
			objs            []media_library.MediaLibrary
			deleteIDs       []uint64
			deleteFolderIDS []uint
		)
		for _, idStr := range ids {
			id, innerErr := strconv.ParseUint(idStr, 10, 64)
			if innerErr != nil {
				continue
			}
			deleteIDs = append(deleteIDs, id)
		}
		defer web.AppendRunScripts(&r,
			"vars.mediaLibrary_deleteConfirmation = false",
			web.Plaid().EventFunc(ImageJumpPageEvent).MergeQuery(true).Queries(ctx.Queries()).Go(),
		)
		err = db.Where("id in ?", deleteIDs).Find(&objs).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return r, nil
			}
			panic(err)
		}
		for _, obj := range objs {
			if obj.Folder {
				deleteFolderIDS = append(deleteFolderIDS, obj.ID)
			}
			if err = mb.deleteIsAllowed(ctx.R, &obj); err != nil {
				return
			}
		}
		err = db.Transaction(func(tx *gorm.DB) (dbErr error) {
			if len(deleteFolderIDS) > 0 {
				if dbErr = tx.
					Model(&media_library.MediaLibrary{}).
					Where("parent_id in ? ", deleteFolderIDS).Update("parent_id", 0).Error; dbErr != nil {
					return
				}
			}

			if dbErr = tx.Delete(&media_library.MediaLibrary{}, "id  in ?", deleteIDs).Error; dbErr != nil {
				return
			}
			return
		})
		if err != nil {
			panic(err)
		}
		mb.onDelete(ctx, objs)
		return
	}
}

func ChooseFileButtonID(field string) string {
	return fmt.Sprintf("btnChooseFile_%s", field)
}

func mediaBoxThumbnails(ctx *web.EventContext, mediaBox *media_library.MediaBox, field string, cfg *media_library.MediaBoxConfig, disabled, readonly bool) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	c := VContainer().Class("media-box-wrap").Fluid(true)
	if cfg.BackgroundColor != "" {
		c.Attr("style", fmt.Sprintf("background-color: %s;", cfg.BackgroundColor))
	}
	vErr, _ := ctx.Flash.(*web.ValidationErrors)
	if vErr == nil {
		vErr = &web.ValidationErrors{}
	}
	var mediaboxID string

	if mediaBox != nil {
		mediaboxID = mediaBox.ID.String()
	}
	// button
	btnRow := VRow(
		VBtn(msgr.ChooseFile).
			Attr("id", ChooseFileButtonID(field)).
			Variant(VariantTonal).Color(ColorPrimary).Size(SizeSmall).PrependIcon("mdi-tray-arrow-up").
			Class("rounded img-upload-btn").
			Attr("style", "text-transform: none;").
			Attr("@click", web.Plaid().EventFunc(OpenFileChooserEvent).
				Query(ParamField, field).
				Query(ParamCfg, h.JSONString(cfg)).
				Query(ParamSelectIDS, mediaboxID).
				Go(),
			).Disabled(disabled),
	)
	if (mediaBox != nil && mediaBox.ID.String() != "" && mediaBox.ID.String() != "0") ||
		(cfg.SimpleIMGURL && mediaBox != nil && mediaBox.Url != "") {
		btnRow.AppendChildren(
			VBtn(msgr.Delete).
				Variant(VariantTonal).Color(ColorError).Size(SizeSmall).PrependIcon("mdi-delete-outline").
				Class("rounded ml-2 img-delete-btn").
				Attr("style", "text-transform: none").
				Attr("@click", web.Plaid().
					EventFunc(deleteFileEvent).
					Query(ParamField, field).
					Query(ParamCfg, h.JSONString(cfg)).
					Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
					Go(),
				).Disabled(disabled),
		)
	}
	if !readonly {
		c.AppendChildren(btnRow.Class())
	}
	if mediaBox.ID.String() != "" && mediaBox.ID.String() != "0" && !cfg.SimpleIMGURL {
		row := appendMediaBoxThumb(cfg, msgr, mediaBox, field, disabled)
		c.AppendChildren(row)

		fieldName := fmt.Sprintf("%s.Description", field)
		value := ctx.Param(fieldName)
		if value == "" {
			value = mediaBox.Description
		}
		if !(value == "" && readonly) {
			c.AppendChildren(
				VRow(
					VCol(
						h.If(
							readonly,
							h.Span(value),
						).Else(
							vx.VXField().
								Attr(presets.VFieldError(fieldName, value, vErr.GetFieldErrors(fmt.Sprintf("%s.Description", field)))...).
								Placeholder(msgr.DescriptionForAccessibility).
								Disabled(disabled),
						),
					).Cols(12).Class("pl-0 pt-0"),
				),
			)
		}
	} else if cfg.SimpleIMGURL {
		mediaBox.FileName = "simple.png"
		if mediaBox.Url != "" {
			row := appendMediaBoxThumb(cfg, msgr, mediaBox, field, disabled)
			c.AppendChildren(row)
		}
	}

	mediaBoxValue := ""
	if (mediaBox.ID.String() != "" && mediaBox.ID.String() != "0") ||
		cfg.SimpleIMGURL {
		mediaBoxValue = h.JSONString(mediaBox)
	}

	return h.Components(
		c,
		web.Portal().Name(cropperPortalName(field)),
		h.Input("").Type("hidden").
			Attr(web.VField(fmt.Sprintf("%s.Values", field), mediaBoxValue)...),
	)
}

func appendMediaBoxThumb(cfg *media_library.MediaBoxConfig, msgr *Messages, mediaBox *media_library.MediaBox, field string, disabled bool) h.HTMLComponent {
	row := VRow()
	if len(cfg.Sizes) == 0 {
		row.AppendChildren(
			VCol(
				mediaBoxThumb(msgr, cfg, mediaBox, field, base.DefaultSizeKey, disabled),
			).Cols(6).Sm(4).Class("pl-0 media-box-thumb"),
		)
	} else {
		var keys []string
		for k := range cfg.Sizes {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			sm := cfg.Sizes[k].Sm
			if sm == 0 {
				sm = 4
			}
			cols := cfg.Sizes[k].Cols
			if cols == 0 {
				cols = 6
			}
			row.AppendChildren(
				VCol(
					mediaBoxThumb(msgr, cfg, mediaBox, field, k, disabled),
				).Cols(cols).Sm(sm).Class("pl-0 media-box-thumb"),
			)
		}
	}
	return row
}

func SimpleMediaBox(url, fieldName string, readOnly bool, cfg *media_library.MediaBoxConfig, db *gorm.DB) *QMediaBoxBuilder {
	mdx := media_library.MediaBox{Url: url}
	if cfg == nil {
		cfg = &media_library.MediaBoxConfig{
			AllowType:    "image",
			DisableCrop:  true,
			SimpleIMGURL: true,
		}
	}
	return QMediaBox(db).
		FieldName(fieldName).
		Value(&mdx).
		Config(cfg).
		Disabled(readOnly).
		Readonly(readOnly)
}

func MediaBoxListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return h.Td(h.Img("").Src(mediaBox.URL(media_library.QorPreviewSizeName)).Style("height: 48px;"))
	}
}

func deleteFileField() web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.Param(ParamField)
		cfg := stringToCfg(ctx.Param(ParamCfg))
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, &media_library.MediaBox{}, field, cfg, false, false),
		})

		return
	}
}

func stringToCfg(v string) *media_library.MediaBoxConfig {
	var cfg media_library.MediaBoxConfig
	if v == "" {
		return &cfg
	}
	err := json.Unmarshal([]byte(v), &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}

func thumbName(name string, size *base.Size, fileSize int, f *media_library.MediaBox) h.HTMLComponent {
	div := h.Div().Class("pl-1")
	title := ""
	text := ""
	if name == base.DefaultSizeKey {
		title = name
		text = fmt.Sprintf("%d X %d", f.Width, f.Height)
	}
	if size != nil {
		title = name
		base.SaleUpDown(f.Width, f.Height, size)
		if size.Width != 0 && size.Height != 0 {
			text = fmt.Sprintf("%d X %d", size.Width, size.Height)
		}
	}
	// if fileSize != 0 {
	//	text = fmt.Sprintf("%s %s", text, media.ByteCountSI(fileSize))
	// }
	if title != "" {
		div.AppendChildren(h.Span(name))
	}
	if text != "" {
		div.AppendChildren(h.Br(), h.Span(text).Style("color:#757575;"))
	}
	return div
}

func updateDescription(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			db   = mb.db
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)

		obj := wrapFirst(mb, ctx, &r)
		if err = mb.updateDescIsAllowed(ctx.R, &obj); err != nil {
			return
		}
		old := wrapFirst(mb, ctx, &r)

		obj.File.Description = ctx.Param(ParamCurrentDescription)
		if err = db.Save(&obj).Error; err != nil {
			return
		}
		mb.onEdit(ctx, old, obj)
		presets.ShowMessage(&r, msgr.DescriptionUpdated, ColorSuccess)
		web.AppendRunScripts(&r,
			web.Plaid().EventFunc(ImageJumpPageEvent).
				MergeQuery(true).
				Queries(ctx.Queries()).
				Go())
		return
	}
}

func rename(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			db   = mb.db
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		obj := wrapFirst(mb, ctx, &r)
		if err = mb.updateNameIsAllowed(ctx.R, &obj); err != nil {
			return
		}
		old := wrapFirst(mb, ctx, &r)

		if obj.Folder {
			obj.File.FileName = ctx.Param(ParamName)
		} else {
			obj.File.FileName = ctx.Param(ParamName) + path.Ext(obj.File.FileName)
		}

		if err = db.Save(&obj).Error; err != nil {
			return
		}
		mb.onEdit(ctx, old, obj)
		presets.ShowMessage(&r, msgr.RenameUpdated, ColorSuccess)
		web.AppendRunScripts(&r,
			web.Plaid().EventFunc(ImageJumpPageEvent).
				MergeQuery(true).
				Queries(ctx.Queries()).
				Go())
		return
	}
}

func createFolder(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			dirName  = ctx.Param(ParamName)
			parentID = ctx.ParamAsInt(ParamParentID)
			m        = media_library.MediaLibrary{Folder: true, ParentId: uint(parentID)}
		)
		if err = mb.newFolderIsAllowed(ctx.R); err != nil {
			return
		}
		if dirName == "" {
			presets.ShowMessage(&r, "folder name can`t be empty", ColorWarning)
			return
		}
		m.File.FileName = dirName
		var uid uint
		if mb.currentUserID != nil {
			uid = mb.currentUserID(ctx)
		}
		m.UserID = uid
		m.ParentId = uint(parentID)
		if err = mb.saverFunc(mb.db, &m, "", ctx); err != nil {
			return
		}
		mb.onCreate(ctx, m)
		r.RunScript = web.Plaid().
			EventFunc(ImageJumpPageEvent).
			MergeQuery(true).
			Queries(ctx.Queries()).
			Go()
		return
	}
}

func wrapFirst(mb *Builder, ctx *web.EventContext, r *web.EventResponse) (obj media_library.MediaLibrary) {
	var (
		err   error
		db    = mb.db
		field = ctx.Param(ParamField)
		id    = ctx.Param(ParamMediaIDS)
		cfg   = ctx.Param(ParamCfg)
		pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
	)

	err = db.Where("id = ?", id).First(&obj).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			presets.ShowMessage(r, pMsgr.RecordNotFound, ColorError)
			renderFileChooserDialogContent(
				ctx,
				r,
				field,
				mb,
				stringToCfg(cfg),
			)
			return
		}
		panic(err)
	}
	return
}

func renameDialog(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			obj   = wrapFirst(mb, ctx, &r)
			pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		var fileName string
		if obj.Folder {
			fileName = obj.File.FileName
		} else {
			fileName = strings.TrimSuffix(obj.File.FileName, path.Ext(obj.File.FileName))
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: renameDialogPortalName,
			Body: web.Scope(
				vx.VXDialog(
					vx.VXField().Label(msgr.Name).Attr(web.VField(ParamName, fileName)...),
				).
					Attr("v-model", "dialogLocals.show").
					Title(msgr.Rename).
					CancelText(pMsgr.Cancel).
					Width(300).
					OkText(pMsgr.OK).
					Attr(":disable-ok", fmt.Sprintf("!form.%s", ParamName)).
					Attr("@click:ok", web.Plaid().BeforeScript("dialogLocals.show = false").EventFunc(RenameEvent).Queries(ctx.Queries()).Go()),
			).VSlot("{locals:dialogLocals}").Init("{show:true}"),
		})
		return
	}
}

func newFolderDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: newFolderDialogPortalName,
		Body: web.Scope(
			vx.VXDialog(
				vx.VXField().Label(msgr.Name).Attr(web.VField(ParamName, ctx.Param(ParamName))...),
			).
				Attr("v-model", "dialogLocals.show").
				Title(msgr.NewFolder).
				Width(300).
				CancelText(pMsgr.Cancel).
				OkText(pMsgr.OK).
				Attr("@click:ok", web.Plaid().BeforeScript("dialogLocals.show=false").EventFunc(CreateFolderEvent).Queries(ctx.Queries()).Go()),
		).VSlot("{locals:dialogLocals}").Init("{show:true}"),
	})
	return
}

func updateDescriptionDialog(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		obj := wrapFirst(mb, ctx, &r)
		var (
			pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: updateDescriptionDialogPortalName,
			Body: web.Scope(
				vx.VXDialog(
					vx.VXField().Label(msgr.UpdateDescriptionTextFieldPlaceholder).Attr(web.VField(ParamCurrentDescription, obj.File.Description)...),
				).
					Attr("v-model", "dialogLocals.show").
					Title(msgr.UpdateDescription).
					Width(300).
					CancelText(pMsgr.Cancel).
					OkText(pMsgr.OK).
					Attr("@click:ok", web.Plaid().BeforeScript("dialogLocals.show=false").EventFunc(UpdateDescriptionEvent).Queries(ctx.Queries()).Go()),
			).VSlot("{locals:dialogLocals}").Init("{show:true}"),
		})
		return
	}
}

func moveToFolderDialog(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			msgr  = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
		)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: moveToFolderDialogPortalName,
			Body: web.Scope(
				vx.VXDialog(
					VCardItem(
						VCard(
							VList(
								h.Components(folderGroupsComponents(db, ctx, -1)...),
							).ActiveColor(ColorPrimary).BgColor(ColorGreyLighten5),
						).Color(ColorGreyLighten5).Height(340).Class("overflow-auto"),
					),
				).
					Attr("v-model", "dialogLocals.show").
					Title(msgr.ChooseFolder).
					Size("large").
					Width(658).
					ContentHeight(571).
					CancelText(pMsgr.Cancel).
					OkText(pMsgr.OK).
					Attr("@click:ok", web.Plaid().
						EventFunc(MoveToFolderEvent).
						BeforeScript("dialogLocals.show = false").
						Queries(ctx.Queries()).
						Query(ParamParentID, web.Var(fmt.Sprintf("form.%s", ParamSelectFolderID))).
						Go()),
			).VSlot("{locals:dialogLocals,form}").Init("{show:true}").FormInit(fmt.Sprintf("{%s:0}", ParamSelectFolderID)),
		})
		return
	}
}

func moveToFolder(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			selectFolderID = ctx.ParamAsInt(ParamSelectFolderID)
			field          = ctx.Param(ParamField)
			selectIDs      = strings.Split(ctx.Param(ParamSelectIDS), ",")
			msgr           = i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
			inMediaLibrary = strings.Contains(ctx.R.RequestURI, "/"+mb.mb.Info().URIName())
		)
		if err = mb.moveToIsAllowed(ctx.R); err != nil {
			return
		}
		queries := ctx.Queries()
		delete(queries, searchKeywordName(inMediaLibrary, field))
		var ids []uint64

		for _, idStr := range selectIDs {
			selectID, innerErr := strconv.ParseUint(idStr, 10, 64)
			if innerErr != nil {
				continue
			}
			ids = append(ids, selectID)
		}
		presets.ShowMessage(&r, msgr.MovedFailed, ColorError)
		if len(ids) > 0 {
			for _, findID := range ids {
				var old, obj media_library.MediaLibrary
				db.First(&obj, findID)
				if obj.ID == 0 {
					continue
				}
				db.First(&old, findID)
				obj.ParentId = uint(selectFolderID)
				if err = db.Save(&obj).Error; err != nil {
					return
				}
				mb.onEdit(ctx, old, obj)
			}
			presets.ShowMessage(&r, msgr.MovedSuccess, ColorSuccess)
		}
		web.AppendRunScripts(
			&r,
			web.Plaid().PushState(true).Query(paramTab, tabFolders).Query(ParamParentID, selectFolderID).RunPushState(),
			web.Plaid().
				EventFunc(ImageJumpPageEvent).
				AfterScript(fmt.Sprintf(`vars.searchMsg="";"vars.media_parent_id=%v"`, selectFolderID)).
				Queries(queries).
				Query(paramTab, tabFolders).
				Query(ParamSelectIDS, "").
				Query(ParamParentID, selectFolderID).
				Go(),
		)
		return
	}
}

func nextFolder(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.ParamAsInt(ParamSelectFolderID)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: folderGroupPortalName(uint(id)),
			Body: h.Components(folderGroupsComponents(db, ctx, id)...),
		})
		return
	}
}

func folderGroupsComponents(db *gorm.DB, ctx *web.EventContext, parentID int) (items []h.HTMLComponent) {
	var (
		records   []*media_library.MediaLibrary
		count     int64
		selectIDs = strings.Split(ctx.Param(ParamSelectIDS), ",")
		idS       []uint64
	)

	for _, idStr := range selectIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		idS = append(idS, id)
	}

	if parentID == -1 {
		item := &media_library.MediaLibrary{
			Folder: true,
		}
		item.ID = 0
		item.File.FileName = "Root Directory"
		records = append(records, item)
	} else {
		db.Where("parent_id = ?  and folder = true", parentID).Find(&records)
	}
	for _, record := range records {
		if slices.Contains(selectIDs, fmt.Sprint(record.ID)) {
			continue
		}
		db.Model(media_library.MediaLibrary{}).Where("id not in ? and parent_id = ?  and folder = true", idS, record.ID).Count(&count)
		if count > 0 {
			items = append(items,
				VListGroup(
					web.Slot(
						VListItem(
							VListItemTitle(h.Text(record.File.FileName)),
						).Attr("v-bind", "props").
							PrependIcon("mdi-folder").
							Attr(":active", fmt.Sprintf(`form.%s==%v`, ParamSelectFolderID, record.ID)).
							Attr(web.VAssign("locals", fmt.Sprintf(`{folder%v:0}`, record.ID))...).
							Attr("@click", fmt.Sprintf("form.%s=%v;", ParamSelectFolderID, record.ID)+
								fmt.Sprintf(`if (!locals.folder%v){locals.folder%v=1;%s}`, record.ID, record.ID,
									web.Plaid().
										EventFunc(NextFolderEvent).
										Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
										Query(ParamSelectFolderID, record.ID).
										Go())),
					).Name("activator").Scope(" {  props }"),
					web.Portal().Name(folderGroupPortalName(record.ID)),
				).Value(record.ID),
			)
		} else {
			items = append(items, VListItem(h.Text(record.File.FileName)).
				Attr(":active", fmt.Sprintf(`form.%s==%v`, ParamSelectFolderID, record.ID)).
				Attr("@click", fmt.Sprintf("form.%s=%v;", ParamSelectFolderID, record.ID)).
				PrependIcon("mdi-folder"))
		}
	}
	return
}
