package views

import (
	"bytes"
	"fmt"
	"io"

	"github.com/qor/oss"
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/microsite"
	"github.com/qor5/admin/microsite/utils"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/publish"
	publish_view "github.com/qor5/admin/publish/views"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nMicrositeKey i18n.ModuleKey = "I18nMicrositeKey"

func Configure(b *presets.Builder, db *gorm.DB, ab *activity.ActivityBuilder, storage oss.StorageInterface, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	b.I18n().
		RegisterForModule(language.English, I18nMicrositeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nMicrositeKey, Messages_zh_CN)

	publish_view.Configure(b, db, ab, publisher, models...)
	for _, model := range models {
		model.Editing().Field("Package").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			this := obj.(microsite.MicroSiteInterface)

			if this.GetPackage().FileName == "" {
				return vuetify.VFileInput().Chips(true).ErrorMessages(field.Errors...).Label(field.Label).Attr("accept", ".rar,.zip,.7z,.tar").Clearable(false).
					On("change", fmt.Sprintf("form..%s = $event.target.files[0]", field.Name))
			}
			return web.Scope(
				h.Div(
					h.Div(
						h.Div(
							h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, model.Info().Label(), "Current Package")).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
							h.A().Href(this.GetPackageUrl(storage.GetEndpoint())).Text(this.GetPackage().FileName),
						).Class("v-text-field__slot").Style("padding: 8px 0;"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted"),

				vuetify.VFileInput().Chips(true).ErrorMessages(field.Errors...).Label(field.Label).Attr("accept", ".rar,.zip,.7z,.tar").Clearable(false).
					Attr("v-model", "locals.file").On("change", fmt.Sprintf("form.%s = $event.target.files[0]", field.Name)),
			).Init(fmt.Sprintf(`{ file: new File([""], "%v", {
                  lastModified: 0,
                }) , change: false}`, this.GetPackage().FileName)).
				VSlot("{ locals }")
		}).
			SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
				if ctx.R.FormValue("PackageChanged") != "true" {
					return
				}
				this := obj.(microsite.MicroSiteInterface)
				if this.GetUnixKey() == "" {
					this.SetUnixKey()
				}
				fs := ctx.R.MultipartForm.File[field.Name]
				if len(fs) == 0 {
					if this.GetID() != 0 {
						err = db.Where("id = ? AND version_name = ?", this.GetID(), this.GetVersionName()).Select("files_list").Find(&this).Error
						if err != nil {
							return
						}
					}
					return
				}
				var fileName = fs[0].Filename
				var packagePath = this.GetPackagePath(fileName)

				f, err := fs[0].Open()
				if err != nil {
					return
				}

				fileBytes, err := io.ReadAll(f)
				if err != nil {
					return
				}

				filesList, err := this.UnArchiveAndPublish(this.GetPreviewPath, fileName, bytes.NewReader(fileBytes), storage)
				if err != nil {
					return
				}

				err = utils.Upload(storage, packagePath, bytes.NewReader(fileBytes))
				if err != nil {
					return
				}

				this.SetFilesList(filesList)
				this.SetPackage(fileName, packagePath)

				return
			})

		model.Editing().Field("FilesList").ComponentFunc(
			func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
				this := obj.(microsite.MicroSiteInterface)
				if this.GetStatus() == publish.StatusOffline || len(this.GetFileList()) == 0 {
					return nil
				}

				var content []h.HTMLComponent

				content = append(content,
					h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, model.Info().Label(), field.Label)).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
				)

				if this.GetStatus() == publish.StatusOnline {
					for k, v := range this.GetFileList() {
						if k != 0 {
							content = append(content, h.Br())
						}
						content = append(content, h.A(h.Text(v)).Href(this.GetPublishedUrl(storage.GetEndpoint(), v)))
					}
				} else {
					for k, v := range this.GetFileList() {
						if k != 0 {
							content = append(content, h.Br())
						}
						content = append(content, h.A(h.Text(v)).Href(this.GetPreviewUrl(storage.GetEndpoint(), v)))
					}
				}

				return h.Div(
					h.Div(
						h.Div(
							content...,
						).Class("v-text-field__slot").Style("padding: 8px 0;"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted")
			},
		).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			return nil
		})
	}

	return
}
