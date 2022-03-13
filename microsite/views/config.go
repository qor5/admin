package views

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/oss"
	"github.com/qor/qor5/microsite"
	"github.com/qor/qor5/microsite/utils"
	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func Configure(b *presets.Builder, db *gorm.DB, storage oss.StorageInterface, domain string, publisher *publish.Builder, models ...*presets.ModelBuilder) {
	publish_view.Configure(b, db, publisher, models...)
	for _, model := range models {
		model.Editing().Field("Package").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			this := obj.(microsite.MicroSiteInterface)
			return vuetify.VFileInput().Chips(true).Label(field.Label).Value(this.GetPackage().FileName).FieldName(field.Name).Attr("accept", ".rar,.zip,.7z,.tar")
		}).
			SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
				defer func() {
					if err != nil {
						fmt.Println(err)
					}
				}()
				this := obj.(microsite.MicroSiteInterface)
				if this.GetUnixKey() == "" {
					this.SetUnixKey()
				}
				fs := ctx.R.MultipartForm.File[field.Name]
				if len(fs) == 0 {
					//todo delete flag
					err = db.Select("files_list").Find(&this).Error
					if err != nil {
						return
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
					return err
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
				if this.GetStatus() == publish.StatusOffline {
					return nil
				}
				htmlFiles := []string{}
				otherFiles := []string{}

				for _, v := range this.GetFileList() {
					if filepath.Ext(v) == ".html" {
						htmlFiles = append(htmlFiles, v)
					} else {
						otherFiles = append(otherFiles, v)
					}
				}

				var content []h.HTMLComponent
				content = append(content, vuetify.VRow(h.Label(field.Label)))

				// List all html files first
				for _, v := range htmlFiles {
					if this.GetStatus() == publish.StatusOnline {
						content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPublishedUrl(domain, v))))
					} else {
						content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPreviewUrl(domain, v))))
					}
				}

				if len(otherFiles) > 0 {
					content = append(content, vuetify.VRow(h.Text("Assets")))
					for _, v := range otherFiles {
						if this.GetStatus() == publish.StatusOnline {
							content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPublishedUrl(domain, v))))
						} else {
							content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPreviewUrl(domain, v))))
						}
					}
				}

				return h.Components(
					h.Div(content...).Style("margin-top: 4px; padding-top: 12px"),
				)
			},
		)
	}

	return
}
