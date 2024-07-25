package media

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/cropper"
	"github.com/qor5/x/v3/ui/fileicons"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type MediaBoxConfigKey int

const (
	MediaBoxConfig      MediaBoxConfigKey = iota
	I18nMediaLibraryKey i18n.ModuleKey    = "I18nMediaLibraryKey"

	ParamFolderName     = "folder_name"
	ParamParentID       = "parent_id"
	ParamSelectFolderID = "select_folder_id"
	ParamSelectIDS      = "select_ids"
)

func AutoMigrate(db *gorm.DB) (err error) {
	return db.AutoMigrate(
		&media_library.MediaLibrary{},
	)
}

func configure(b *presets.Builder, mb *Builder, db *gorm.DB) {
	mb.permVerifier = perm.NewVerifier("media_library", b.GetPermission())

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
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return QMediaBox(db).
			FieldName(field.FormKey).
			Value(&mediaBox).
			Label(field.Label).
			Config(cfg).
			Disabled(field.Disabled).
			Readonly(readonly)
	}
}

func MediaBoxSetterFunc(db *gorm.DB) presets.FieldSetterFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		jsonValuesField := fmt.Sprintf("%s.Values", field.FormKey)
		mediaBox := media_library.MediaBox{}
		err = mediaBox.Scan(ctx.R.FormValue(jsonValuesField))
		if err != nil {
			return
		}
		descriptionField := fmt.Sprintf("%s.Description", field.FormKey)
		mediaBox.Description = ctx.R.FormValue(descriptionField)
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
	disabled  bool
	readonly  bool
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

func (b *QMediaBoxBuilder) MarshalHTML(c context.Context) (r []byte, err error) {
	if len(b.fieldName) == 0 {
		panic("FieldName required")
	}
	if b.value == nil {
		panic("Value required")
	}

	ctx := web.MustGetEventContext(c)

	portalName := mainPortalName(b.fieldName)

	return h.Components(
		VSheet(
			h.If(len(b.label) > 0,
				h.Label(b.label).Class("v-label theme--light"),
			),
			web.Portal(
				mediaBoxThumbnails(ctx, b.value, b.fieldName, b.config, b.disabled, b.readonly),
			).Name(mediaBoxThumbnailsPortalName(b.fieldName)),
			web.Portal().Name(portalName),
		).Class("pb-4").
			Rounded(true).
			Attr(web.VAssign("vars", `{showFileChooser: false}`)...),
	).MarshalHTML(c)
}

func mediaBoxThumb(msgr *Messages, cfg *media_library.MediaBoxConfig,
	f *media_library.MediaBox, field string, thumb string, disabled bool,
) h.HTMLComponent {
	size := cfg.Sizes[thumb]
	fileSize := f.FileSizes[thumb]
	url := f.URL(thumb)
	if thumb == base.DefaultSizeKey {
		url = f.URL()
	}
	card := VCard(
		h.If(base.IsImageFormat(f.FileName),
			VImg().Src(fmt.Sprintf("%s?%d", url, time.Now().UnixNano())).Height(150),
		).Else(
			h.Div(
				fileThumb(f.FileName),
				h.A().Text(f.FileName).Href(f.Url).Target("_blank"),
			).Style("text-align:center"),
		),
		h.If(base.IsImageFormat(f.FileName) && (size != nil || thumb == base.DefaultSizeKey),
			VCardActions(
				thumbName(thumb, size, fileSize, f),
			),
		),
	)
	if base.IsImageFormat(f.FileName) && (size != nil || thumb == base.DefaultSizeKey) && !disabled && !cfg.DisableCrop {
		card.Attr("@click", web.Plaid().
			EventFunc(loadImageCropperEvent).
			Query("field", field).
			Query(mediaID, fmt.Sprint(f.ID)).
			Query("thumb", thumb).
			Query("cfg", h.JSONString(cfg)).
			Go())
	}
	return card
}

func fileThumb(filename string) h.HTMLComponent {
	return h.Div(
		fileicons.Icon(path.Ext(filename)[1:]).Attr("height", "150").Class("pt-4"),
	).Class("d-flex align-center justify-center")
}

func deleteConfirmation(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue(mediaID)
		cfg := ctx.R.FormValue("cfg")

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: deleteConfirmPortalName(field),
			Body: VDialog(
				VCard(
					VCardTitle(h.Text(msgr.DeleteConfirmationText(id))),
					VCardActions(
						VSpacer(),
						VBtn(msgr.Cancel).
							Variant(VariantFlat).
							Class("ml-2").
							On("click", "vars.mediaLibrary_deleteConfirmation = false"),

						VBtn(msgr.Delete).
							Color("primary").
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", web.Plaid().
								MergeQuery(true).
								EventFunc(doDeleteEvent).
								Query("field", field).
								Query(mediaID, id).
								FieldValue("cfg", cfg).
								Go()),
					),
				),
			).MaxWidth("600px").
				Attr("v-model", "vars.mediaLibrary_deleteConfirmation").
				Attr(web.VAssign("vars", `{mediaLibrary_deleteConfirmation: false}`)...),
		})

		r.RunScript = "setTimeout(function(){ vars.mediaLibrary_deleteConfirmation = true }, 100)"
		return
	}
}

func doDelete(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		db := mb.db
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue(mediaID)
		cfg := ctx.R.FormValue("cfg")

		var obj media_library.MediaLibrary
		err = db.Where("id = ?", id).First(&obj).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				renderFileChooserDialogContent(
					ctx,
					&r,
					field,
					mb,
					stringToCfg(cfg),
				)
				r.RunScript = "vars.mediaLibrary_deleteConfirmation = false"
				return r, nil
			}
			panic(err)
		}
		if err = mb.deleteIsAllowed(ctx.R, &obj); err != nil {
			return
		}

		err = db.Delete(&media_library.MediaLibrary{}, "id = ?", id).Error
		if err != nil {
			panic(err)
		}

		renderFileChooserDialogContent(
			ctx,
			&r,
			field,
			mb,
			stringToCfg(cfg),
		)
		r.RunScript = "vars.mediaLibrary_deleteConfirmation = false"
		return
	}
}

func mediaBoxThumbnails(ctx *web.EventContext, mediaBox *media_library.MediaBox, field string, cfg *media_library.MediaBoxConfig, disabled, readonly bool) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	c := VContainer().Fluid(true)
	if cfg.BackgroundColor != "" {
		c.Attr("style", fmt.Sprintf("background-color: %s;", cfg.BackgroundColor))
	}
	// button
	btnRow := VRow(
		VBtn(msgr.ChooseFile).
			Variant(VariantTonal).Color(ColorPrimary).Size(SizeXSmall).PrependIcon("mdi-upload-outline").
			Class("rounded-sm").
			Attr("style", "text-transform: none;").
			Attr("@click", web.Plaid().EventFunc(openFileChooserEvent).
				Query("field", field).
				FieldValue("cfg", h.JSONString(cfg)).
				Go(),
			).Disabled(disabled),
	)
	if mediaBox != nil && mediaBox.ID.String() != "" && mediaBox.ID.String() != "0" {
		btnRow.AppendChildren(
			VBtn(msgr.Delete).
				Variant(VariantTonal).Color(ColorError).Size(SizeXSmall).PrependIcon("mdi-delete-outline").
				Class("rounded-sm ml-2").
				Attr("style", "text-transform: none").
				Attr("@click", web.Plaid().EventFunc(deleteFileEvent).
					Query("field", field).
					FieldValue("cfg", h.JSONString(cfg)).
					Go(),
				).Disabled(disabled),
		)
	}
	if !readonly {
		c.AppendChildren(btnRow.Class())
	}
	if mediaBox.ID.String() != "" && mediaBox.ID.String() != "0" {
		row := VRow()
		if len(cfg.Sizes) == 0 {
			row.AppendChildren(
				VCol(
					mediaBoxThumb(msgr, cfg, mediaBox, field, base.DefaultSizeKey, disabled),
				).Cols(6).Sm(4).Class("pl-0"),
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
					).Cols(cols).Sm(sm).Class("pl-0"),
				)
			}
		}

		c.AppendChildren(row)

		fieldName := fmt.Sprintf("%s.Description", field)
		value := ctx.R.FormValue(fieldName)
		if len(value) == 0 {
			value = mediaBox.Description
		}
		if !(len(value) == 0 && readonly) {
			c.AppendChildren(
				VRow(
					VCol(
						h.If(
							readonly,
							h.Span(value),
						).Else(
							VTextField().
								Attr(web.VField(fieldName, value)...).
								Placeholder(msgr.DescriptionForAccessibility).
								Density(DensityCompact).
								HideDetails(true).
								Variant(VariantOutlined).
								Disabled(disabled),
						),
					).Cols(12).Class("pl-0 pt-0"),
				),
			)
		}
	}

	mediaBoxValue := ""
	if mediaBox.ID.String() != "" && mediaBox.ID.String() != "0" {
		mediaBoxValue = h.JSONString(mediaBox)
	}

	return h.Components(
		c,
		web.Portal().Name(cropperPortalName(field)),
		h.Input("").Type("hidden").
			Attr(web.VField(fmt.Sprintf("%s.Values", field), mediaBoxValue)...),
	)
}

func MediaBoxListFunc() presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		mediaBox := field.Value(obj).(media_library.MediaBox)
		return h.Td(h.Img("").Src(mediaBox.URL(media_library.QorPreviewSizeName)).Style("height: 48px;"))
	}
}

func deleteFileField() web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		cfg := stringToCfg(ctx.R.FormValue("cfg"))
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: mediaBoxThumbnailsPortalName(field),
			Body: mediaBoxThumbnails(ctx, &media_library.MediaBox{}, field, cfg, false, false),
		})

		return
	}
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
		db := mb.db
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue(mediaID)
		cfg := ctx.R.FormValue("cfg")

		var obj media_library.MediaLibrary
		err = db.Where("id = ?", id).First(&obj).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				renderFileChooserDialogContent(
					ctx,
					&r,
					field,
					mb,
					stringToCfg(cfg),
				)
				// TODO: prompt that the record has been deleted?
				return r, nil
			}
			panic(err)
		}
		if err = mb.updateDescIsAllowed(ctx.R, &obj); err != nil {
			return
		}

		var media media_library.MediaLibrary
		if err = db.Find(&media, id).Error; err != nil {
			return
		}

		media.File.Description = ctx.R.FormValue("CurrentDescription")
		if err = db.Save(&media).Error; err != nil {
			return
		}

		r.RunScript = `vars.snackbarShow = true;`
		return
	}
}

func createFolder(mb *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			dirName  = ctx.Param(ParamFolderName)
			parentID = ctx.ParamAsInt(ParamParentID)
			m        = &media_library.MediaLibrary{Folder: true, ParentId: uint(parentID)}
		)
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
		if err = mb.db.Save(&m).Error; err != nil {
			return
		}
		r.RunScript = web.Plaid().MergeQuery(true).PushState(true).Go()
		return
	}
}

func newFolderDialog(_ *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: newFolderDialogPortalName,
		Body: web.Scope(
			VDialog(
				VCard(
					web.Slot(h.Text("New Folder")).Name("title"),
					web.Slot(
						VSpacer(),
						VBtn("").Icon("mdi-close").
							Variant(VariantText).Attr("@click", "dialogLocals.show=false"),
					).Name(VSlotAppend),
					VTextField().Variant(FieldVariantUnderlined).
						Class("px-6").
						Label("Folder Name").Attr(web.VField(ParamFolderName, "")...),
					VCardActions(
						VSpacer(),
						VBtn("Cancel").Color(ColorSecondary).Attr("@click", "dialogLocals.show=false"),
						VBtn("Ok").Color(ColorPrimary).Attr("@click",
							web.Plaid().EventFunc(CreateFolderEvent).MergeQuery(true).Go(),
						),
					),
				),
			).MaxWidth(300).Attr("v-model", "dialogLocals.show"),
		).VSlot("{locals:dialogLocals}").Init("{show:true}"),
	})
	return
}

func moveToFolderDialog(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: moveToFolderDialogPortalName,
			Body: web.Scope(
				VDialog(
					VCard(
						web.Slot(h.Text("Choose Folder")).Name("title"),
						web.Slot(
							VSpacer(),
							VBtn("").Icon("mdi-close").
								Variant(VariantText).Attr("@click", "dialogLocals.show=false"),
						).Name(VSlotAppend),
						VCardItem(
							VCard(
								VList(
									VListGroup(
										web.Slot(
											VListItem(
												VListItemTitle(h.Text("Root Directory")),
											).Attr("v-bind", "props").
												PrependIcon("mdi-folder").
												Attr(":active", fmt.Sprintf(`form.%s==0`, ParamSelectFolderID)).
												Attr("@click", fmt.Sprintf("form.%s=0;", ParamSelectFolderID)),
										).Name("activator").Scope(" {  props }"),
										h.Components(folderGroupsComponents(db, ctx, 0)...),
									).Value(0)).ActiveColor(ColorPrimary).BgColor(ColorGreyLighten5),
							).Color(ColorGreyLighten5).Height(340).Class("overflow-auto"),
						),

						VCardActions(
							VSpacer(),
							VBtn("Cancel").Color(ColorSecondary).Attr("@click", "dialogLocals.show=false"),
							VBtn("Save").Color(ColorPrimary).
								Attr("@click", web.Plaid().
									EventFunc(MoveToFolderEvent).Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).Go()),
						),
					).Height(571).Width(658).Class("pa-6"),
				).MaxWidth(658).Attr("v-model", "dialogLocals.show"),
			).VSlot("{locals:dialogLocals,form}").Init("{show:true}").FormInit(fmt.Sprintf("{%s:0}", ParamSelectFolderID)),
		})
		return
	}
}

func moveToFolder(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			selectFolderID = ctx.Param(ParamSelectFolderID)
			selectIDs      = strings.Split(ctx.Param(ParamSelectIDS), ",")
		)
		var ids []uint

		for _, idStr := range selectIDs {
			selectID, err1 := strconv.Atoi(idStr)
			if err1 != nil {
				continue
			}
			ids = append(ids, uint(selectID))
		}
		presets.ShowMessage(&r, "move failed", ColorWarning)
		if len(ids) > 0 {
			if err = db.Model(media_library.MediaLibrary{}).Where("id in  ?", ids).Update("parent_id", selectFolderID).Error; err != nil {
				return
			}
			presets.ShowMessage(&r, "move success", ColorSuccess)
		}
		r.RunScript = web.Plaid().MergeQuery(true).PushState(true).Go()
		return
	}
}

func nextFolder(mb *Builder) web.EventFunc {
	db := mb.db
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := uint(ctx.ParamAsInt(presets.ParamID))
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: folderGroupPortalName(id),
			Body: h.Components(folderGroupsComponents(db, ctx, id)...),
		})
		return
	}
}

func folderGroupsComponents(db *gorm.DB, ctx *web.EventContext, parentID uint) (items []h.HTMLComponent) {
	var (
		records   []*media_library.MediaLibrary
		count     int64
		selectIDs = strings.Split(ctx.Param(ParamSelectIDS), ",")
	)
	db.Where("parent_id = ?  and folder = true", parentID).Find(&records)
	for _, record := range records {
		if slices.Contains(selectIDs, fmt.Sprint(record.ID)) {
			continue
		}
		db.Model(media_library.MediaLibrary{}).Where("parent_id = ?  and folder = true", record.ID).Count(&count)
		if count > 0 {
			items = append(items,
				VListGroup(
					web.Slot(
						VListItem(
							VListItemTitle(h.Text(record.File.FileName)),
						).Attr("v-bind", "props").
							PrependIcon("mdi-folder").
							Attr(":active", fmt.Sprintf(`form.%s==%v`, ParamSelectFolderID, record.ID)).
							Attr("@click", fmt.Sprintf("form.%s=%v;", ParamSelectFolderID, record.ID)+
								web.Plaid().
									EventFunc(NextFolderEvent).
									Query(ParamSelectIDS, ctx.Param(ParamSelectIDS)).
									Query(presets.ParamID, record.ID).Go()),
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
