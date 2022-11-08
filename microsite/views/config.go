package views

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/goplaid/ui/vuetify"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/qor/qor5/presets"
	"github.com/qor/oss"
	"github.com/qor/qor5/activity"
	"github.com/qor/qor5/microsite"
	"github.com/qor/qor5/microsite/utils"
	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const I18nMicrositeKey i18n.ModuleKey = "I18nMicrositeKey"

func Configure(b *presets.Builder, db *gorm.DB, ab *activity.ActivityBuilder, storage oss.StorageInterface, domain string, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	b.I18n().
		RegisterForModule(language.English, I18nMicrositeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nMicrositeKey, Messages_zh_CN)

	publish_view.Configure(b, db, ab, publisher, models...)
	for _, model := range models {
		model.Editing().Field("Package").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			this := obj.(microsite.MicroSiteInterface)

			if this.GetPackage().FileName == "" {
				return vuetify.VFileInput().Chips(true).ErrorMessages(field.Errors...).Label(field.Label).FieldName(field.Name).Attr("accept", ".rar,.zip,.7z,.tar").Clearable(false).
					On("change", web.Plaid().
						FieldValue("PackageChanged", "true").String())
			}
			return web.Scope(
				h.Div(
					h.Div(
						h.Div(
							h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, model.Info().Label(), "Current Package")).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
							h.A().Href(this.GetPackageUrl(domain)).Text(this.GetPackage().FileName),
						).Class("v-text-field__slot"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted"),

				vuetify.VFileInput().Chips(true).ErrorMessages(field.Errors...).Label(field.Label).FieldName(field.Name).Attr("accept", ".rar,.zip,.7z,.tar").Clearable(false).
					Attr("v-model", "locals.file").On("change", web.Plaid().
					FieldValue("PackageChanged", "true").String()),
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

				fileBytes, err := ioutil.ReadAll(f)
				if err != nil {
					return
				}

				filesList, err := this.GetFilesListAndPublishPreviewFiles(fileName, fileBytes, storage)
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
						content = append(content, h.A(h.Text(v)).Href(this.GetPublishedUrl(domain, v)))
					}
				} else {
					for k, v := range this.GetFileList() {
						if k != 0 {
							content = append(content, h.Br())
						}
						content = append(content, h.A(h.Text(v)).Href(this.GetPreviewUrl(domain, v)))
					}
				}

				return h.Div(
					h.Div(
						h.Div(
							content...,
						).Class("v-text-field__slot"),
					).Class("v-input__slot"),
				).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted")
			},
		).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			return nil
		})
	}

	return
}
