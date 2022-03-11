package microsite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/oss"
	mediaoss "github.com/qor/qor5/media/oss"
	"github.com/qor/qor5/publish"
	publish_view "github.com/qor/qor5/publish/views"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var modelName = "s3_management_test"

func Configure(b *presets.Builder, db *gorm.DB, storage oss.StorageInterface, publisher *publish.Builder, siteStruct MicroSiteInterface) {

	mediaoss.Storage = storage
	var domain = ""

	db.AutoMigrate(siteStruct)
	m := b.Model(siteStruct)
	m.Listing("ID", "Name", "PrePath", "Status").
		SearchColumns("ID", "Name").
		PerPage(10)
	ed := m.Editing("Name", "Status", "Schedule", "PrePath", "FilesList", "Package")

	publish_view.Configure(b, db, publisher, m)

	ed.Field("Package").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		this := obj.(MicroSiteInterface)
		return vuetify.VFileInput().Chips(true).Label(field.Label).Value(this.GetPackage().FileName).FieldName(field.Name).Attr("accept", ".rar,.zip,.7z,.tar")
	}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			defer func() {
				if err != nil {
					fmt.Println(err)
				}
			}()
			this := obj.(MicroSiteInterface)
			fs := ctx.R.MultipartForm.File[field.Name]
			if len(fs) == 0 {
				//todo delete flag
				//this.SetFileList("")
				//this.SetPackage("", "")
				err = db.Select("files_list").Find(&this).Error
				if err != nil {
					return
				}

				return
			}
			var filePath string
			var fileName = fs[0].Filename
			var packagePath = this.GetPackagePath(modelName, fileName)
			var tempDir = "./temp"

			f, err := fs[0].Open()
			if err != nil {
				return
			}

			fileBytes, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			_, err = os.Stat(tempDir)
			if err != nil {
				err = os.Mkdir(tempDir, os.ModePerm)
				if err != nil {
					return err
				}
			}

			now := time.Now().Unix()
			//filePath = path.Join(tempDir, fmt.Sprintf("%d_%s_%d.zip", this.GetId(), this.VersionName, now))
			filePath = path.Join(tempDir, fmt.Sprintf("%s_%d_%d.zip", modelName, this.GetId(), now))
			err = ioutil.WriteFile(filePath, fileBytes, 0666)
			if err != nil {
				return err
			}
			defer os.Remove(filePath)

			list, err := this.PublishArchivePreviewFilesAndGetArchiveList(modelName, fileName, filePath, storage)
			if err != nil {
				return
			}

			filesList, err := json.Marshal(list)
			if err != nil {
				return
			}

			err = upload(storage, packagePath, bytes.NewReader(fileBytes))
			//_, err = storage.Put(packagePath, tempFile)
			if err != nil {
				return
			}

			this.SetFileList(string(filesList))
			this.SetPackage(fileName, packagePath)

			return
		})

	ed.Field("FilesList").ComponentFunc(
		func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (r h.HTMLComponent) {
			this := obj.(MicroSiteInterface)
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
					content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPreviewUrl(domain, modelName, v))))
				}
			}

			if len(otherFiles) > 0 {
				content = append(content, vuetify.VRow(h.Text("Assets")))
				for _, v := range otherFiles {
					if this.GetStatus() == publish.StatusOnline {
						content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPublishedUrl(domain, v))))
					} else {
						content = append(content, vuetify.VRow(h.A(h.Text(v)).Href(this.GetPreviewUrl(domain, modelName, v))))
					}
				}
			}

			return h.Components(
				h.Div(content...).Style("margin-top: 4px; padding-top: 12px"),
			)
		},
	)
	return
}
