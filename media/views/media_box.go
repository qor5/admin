package views

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/ui/v3/cropper"
	"github.com/qor5/ui/v3/fileicons"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type MediaBoxConfigKey int

var MediaLibraryPerPage int = 39

const MediaBoxConfig MediaBoxConfigKey = iota
const I18nMediaLibraryKey i18n.ModuleKey = "I18nMediaLibraryKey"

var permVerifier *perm.Verifier

func Configure(b *presets.Builder, db *gorm.DB) {
	err := db.AutoMigrate(&media_library.MediaLibrary{})
	if err != nil {
		panic(err)
	}

	b.ExtraAsset("/cropper.js", "text/javascript", cropper.JSComponentsPack())
	b.ExtraAsset("/cropper.css", "text/css", cropper.CSSComponentsPack())

	permVerifier = perm.NewVerifier("media_library", b.GetPermission())

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
		RegisterForModule(language.SimplifiedChinese, I18nMediaLibraryKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nMediaLibraryKey, Messages_ja_JP)

	configList(b, db)
}

func MediaBoxComponentFunc(db *gorm.DB) presets.FieldComponentFunc {
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
			Config(cfg).Disabled(field.Disabled)
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
				mediaBoxThumbnails(ctx, b.value, b.fieldName, b.config, b.disabled),
			).Name(mediaBoxThumbnailsPortalName(b.fieldName)),
			web.Portal().Name(portalName),
		).Class("pb-4").
			Rounded(true).
			Attr(web.ObjectAssign("vars", `{showFileChooser: false}`)...),
	).MarshalHTML(c)
}

func mediaBoxThumb(msgr *Messages, cfg *media_library.MediaBoxConfig,
	f *media_library.MediaBox, field string, thumb string, disabled bool) h.HTMLComponent {
	size := cfg.Sizes[thumb]
	fileSize := f.FileSizes[thumb]
	url := f.URL(thumb)
	if thumb == media.DefaultSizeKey {
		url = f.URL()
	}
	return VCard(
		h.If(media.IsImageFormat(f.FileName),
			VImg().Src(fmt.Sprintf("%s?%d", url, time.Now().UnixNano())).Height(150),
		).Else(
			h.Div(
				fileThumb(f.FileName),
				h.A().Text(f.FileName).Href(f.Url).Target("_blank"),
			).Style("text-align:center"),
		),
		h.If(media.IsImageFormat(f.FileName) && (size != nil || thumb == media.DefaultSizeKey),
			VCardActions(
				VChip(
					thumbName(thumb, size, fileSize, f),
				).Size(SizeSmall).Disabled(disabled).Attr("@click", web.Plaid().
					EventFunc(loadImageCropperEvent).
					Query("field", field).
					Query("id", fmt.Sprint(f.ID)).
					Query("thumb", thumb).
					FieldValue("cfg", h.JSONString(cfg)).
					Go()),
			),
		),
	)
}

func fileThumb(filename string) h.HTMLComponent {

	return h.Div(
		fileicons.Icon(path.Ext(filename)[1:]).Attr("height", "150").Class("pt-4"),
	).Class("d-flex align-center justify-center")
}

func deleteConfirmation(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue("id")
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
								EventFunc(doDeleteEvent).
								Query("field", field).
								Query("id", id).
								FieldValue("cfg", cfg).
								Go()),
					),
				),
			).MaxWidth("600px").
				Attr("v-model", "vars.mediaLibrary_deleteConfirmation").
				Attr(web.ObjectAssign("vars", `{mediaLibrary_deleteConfirmation: false}`)...),
		})

		r.RunScript = "setTimeout(function(){ vars.mediaLibrary_deleteConfirmation = true }, 100)"
		return
	}
}
func doDelete(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue("id")
		cfg := ctx.R.FormValue("cfg")

		var obj media_library.MediaLibrary
		err = db.Where("id = ?", id).First(&obj).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				renderFileChooserDialogContent(
					ctx,
					&r,
					field,
					db,
					stringToCfg(cfg),
				)
				r.RunScript = "vars.mediaLibrary_deleteConfirmation = false"
				return r, nil
			}
			panic(err)
		}
		if err = deleteIsAllowed(ctx.R, &obj); err != nil {
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
			db,
			stringToCfg(cfg),
		)
		r.RunScript = "vars.mediaLibrary_deleteConfirmation = false"
		return
	}
}

func mediaBoxThumbnails(ctx *web.EventContext, mediaBox *media_library.MediaBox, field string, cfg *media_library.MediaBoxConfig, disabled bool) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nMediaLibraryKey, Messages_en_US).(*Messages)
	c := VContainer().Fluid(true)

	if mediaBox.ID.String() != "" && mediaBox.ID.String() != "0" {
		row := VRow()
		if len(cfg.Sizes) == 0 {
			row.AppendChildren(
				VCol(
					mediaBoxThumb(msgr, cfg, mediaBox, field, media.DefaultSizeKey, disabled),
				).Cols(6).Sm(4).Class("pl-0"),
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
						mediaBoxThumb(msgr, cfg, mediaBox, field, k, disabled),
					).Cols(6).Sm(4).Class("pl-0"),
				)
			}
		}

		c.AppendChildren(row)

		fieldName := fmt.Sprintf("%s.Description", field)
		value := ctx.R.FormValue(fieldName)
		if len(value) == 0 {
			value = mediaBox.Description
		}
		c.AppendChildren(
			VRow(
				VCol(
					VTextField().
						Attr(web.VField(fieldName, value)...).
						Label(msgr.DescriptionForAccessibility).
						Density(DensityCompact).
						HideDetails(true).
						Variant(VariantOutlined).
						Disabled(disabled),
				).Cols(12).Class("pl-0 pt-0"),
			),
		)
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

		VBtn(msgr.ChooseFile).
			Variant(VariantFlat).
			Attr("@click", web.Plaid().EventFunc(openFileChooserEvent).
				Query("field", field).
				FieldValue("cfg", h.JSONString(cfg)).
				Go(),
			).Disabled(disabled),

		h.If(mediaBox != nil && mediaBox.ID.String() != "" && mediaBox.ID.String() != "0",
			VBtn(msgr.Delete).
				Variant(VariantFlat).
				Attr("@click", web.Plaid().EventFunc(deleteFileEvent).
					Query("field", field).
					FieldValue("cfg", h.JSONString(cfg)).
					Go(),
				).Disabled(disabled),
		),
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
			Body: mediaBoxThumbnails(ctx, &media_library.MediaBox{}, field, cfg, false),
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

func thumbName(name string, size *media.Size, fileSize int, f *media_library.MediaBox) h.HTMLComponent {
	text := name
	if size != nil {
		text = fmt.Sprintf("%s(%dx%d)", text, size.Width, size.Height)
	}
	if name == media.DefaultSizeKey {
		text = fmt.Sprintf("%s(%dx%d)", text, f.Width, f.Height)
	}
	if fileSize != 0 {
		text = fmt.Sprintf("%s %s", text, media.ByteCountSI(fileSize))
	}
	return h.Text(text)
}

func updateDescription(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		field := ctx.R.FormValue("field")
		id := ctx.R.FormValue("id")
		cfg := ctx.R.FormValue("cfg")

		var obj media_library.MediaLibrary
		err = db.Where("id = ?", id).First(&obj).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				renderFileChooserDialogContent(
					ctx,
					&r,
					field,
					db,
					stringToCfg(cfg),
				)
				// TODO: prompt that the record has been deleted?
				return r, nil
			}
			panic(err)
		}
		if err = updateDescIsAllowed(ctx.R, &obj); err != nil {
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
