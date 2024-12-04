package microsite

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/qor5/admin/v3/microsite/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
)

const I18nMicrositeKey i18n.ModuleKey = "I18nMicrositeKey"

func (mib *Builder) Install(b *presets.Builder) error {
	publisher := mib.publisher
	storage := mib.storage
	db := mib.db

	err := db.AutoMigrate(&MicroSite{})
	if err != nil {
		panic(err)
	}

	b.GetI18n().
		RegisterForModule(language.English, I18nMicrositeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nMicrositeKey, Messages_zh_CN)

	model := b.Model(&MicroSite{}).Use(publisher.ContextValueFuncs(mib.ContextValueProvider))

	model.Listing("ID", "Name", "PrePath", "Status").
		SearchColumns("ID::text", "Name").
		PerPage(10)
	model.Editing("StatusBar", "ScheduleBar", "Name", "Description", "PrePath", "FilesList", "Package")

	model.Editing().Field("Package").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		this := obj.(*MicroSite)

		if this.GetPackage().FileName == "" {
			return vuetify.VFileInput().Chips(true).ErrorMessages(field.Errors...).Label(field.Label).Attr("accept", ".rar,.zip,.7z,.tar").Clearable(false).
				On("change", fmt.Sprintf("form..%s = $event.target.files[0]", field.Name))
		}
		return web.Scope(
			h.Div(
				h.Div(
					h.Div(
						h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, model.Info().Label(), "Current Package")).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
						h.A().Href(this.GetPackageUrl(storage.GetEndpoint(ctx.R.Context()))).Text(this.GetPackage().FileName),
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
			this := obj.(*MicroSite)
			if this.GetUnixKey() == "" {
				this.SetUnixKey()
			}
			fs := ctx.R.MultipartForm.File[field.Name]
			if len(fs) == 0 {
				if this.GetID() != 0 {
					err = db.Where("id = ? AND version_name = ?", this.GetID(), this.VersionName).Select("files_list").Find(&this).Error
					if err != nil {
						return
					}
				}
				return
			}
			fileName := fs[0].Filename
			packagePath := getPackagePath(this, fileName, mib)

			f, err := fs[0].Open()
			if err != nil {
				return
			}
			defer f.Close()

			var buf bytes.Buffer
			tee := io.TeeReader(f, &buf)

			err = utils.Upload(storage, packagePath, tee)
			if err != nil {
				return
			}

			filesList, err := this.UnArchiveAndPublish(func(fn string) string {
				return getPreviewPath(this, fn,
					mib)
			},
				fileName,
				bytes.NewReader(buf.Bytes()),
				storage)
			if err != nil {
				return
			}

			this.SetFilesList(filesList)
			this.SetPackage(fileName, packagePath)

			return
		})

	model.Editing().Field("FilesList").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			this := obj.(*MicroSite)
			if this.Status.Status == publish.StatusOffline || len(this.GetFileList()) == 0 {
				return nil
			}

			var content []h.HTMLComponent

			content = append(content,
				h.Label(i18n.PT(ctx.R, presets.ModelsI18nModuleKey, model.Info().Label(), field.Label)).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
			)

			if this.Status.Status == publish.StatusOnline {
				for k, v := range this.GetFileList() {
					if k != 0 {
						content = append(content, h.Br())
					}
					content = append(content, h.A(h.Text(v)).Href(this.GetPublishedUrl(storage.GetEndpoint(ctx.R.Context()), v)))
				}
			} else {
				for k, v := range this.GetFileList() {
					if k != 0 {
						content = append(content, h.Br())
					}
					content = append(content, h.A(h.Text(v)).Href(getPreviewUrl(this, storage.GetEndpoint(ctx.R.Context()), v, mib)))
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

	return nil
}

func getPackagePath(m *MicroSite, fileName string, mib *Builder) string {
	return strings.TrimPrefix(path.Join(mib.packageAndPreviewPrepath, "__package__", m.GetUnixKey(), fileName), "/")
}

func getPreviewPath(m *MicroSite, fileName string, mib *Builder) string {
	return strings.TrimPrefix(path.Join(mib.packageAndPreviewPrepath, "__preview__", m.GetUnixKey(), fileName), "/")
}

func getPreviewUrl(m *MicroSite, domain, fileName string, mib *Builder) string {
	return strings.TrimSuffix(domain, "/") + "/" + getPreviewPath(m, fileName, mib)
}
