package microsite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/go-unarr"
	"github.com/qor/oss"
	"github.com/qor5/admin/microsite/utils"
	"github.com/qor5/admin/publish"
	"gorm.io/gorm"
)

type MicroSite struct {
	gorm.Model
	publish.Status
	publish.Schedule
	publish.Version

	PrePath string

	Package   FileSystem `gorm:"type:text"`
	FilesList string     `gorm:"type:text"`

	UnixKey string
}

func (this *MicroSite) PermissionRN() []string {
	return []string{"microsite_models", strconv.Itoa(int(this.ID)), this.Version.Version}
}

func (this *MicroSite) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", this.ID, this.Version.Version)
}

func (this *MicroSite) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (this MicroSite) GetID() uint {
	return this.ID
}

func (this MicroSite) GetUnixKey() string {
	return this.UnixKey
}

func (this *MicroSite) SetUnixKey() {
	this.UnixKey = strconv.FormatInt(time.Now().UnixMilli(), 10)
	return
}

//func (this *MicroSite) BeforeDelete(db *gorm.DB) (err error) {
//	if this.Status == Status_published {
//		err = Unpublish(db, this, false)
//		return
//	}
//
//	return
//}

func (this MicroSite) GetPackagePath(fileName string) string {
	return fmt.Sprintf("/%s/__package__/%s/%s", PackageAndPreviewPrepath, this.GetUnixKey(), fileName)
}

func (this MicroSite) GetPreviewPrePath() string {
	return fmt.Sprintf("/%s/__preview__/%s", PackageAndPreviewPrepath, this.GetUnixKey())
}

func (this MicroSite) GetPreviewPath(fileName string) string {
	return path.Join(this.GetPreviewPrePath(), fileName)
}

func (this MicroSite) GetPreviewUrl(domain, fileName string) string {
	return strings.TrimSuffix(domain, "/") + this.GetPreviewPath(fileName)
}

func (this MicroSite) GetPublishedPath(fileName string) string {
	return "/" + path.Join(strings.TrimSuffix(this.PrePath, "/"), fileName)
}

func (this MicroSite) GetPublishedUrl(domain, fileName string) string {
	return strings.TrimSuffix(domain, "/") + this.GetPublishedPath(fileName)
}

func (this MicroSite) GetPackageUrl(domain string) string {
	return strings.TrimSuffix(domain, "/") + this.Package.Url
}

func (this MicroSite) GetFileList() (arr []string) {
	json.Unmarshal([]byte(this.FilesList), &arr)
	return
}

func (this *MicroSite) SetFilesList(filesList []string) {
	list, err := json.Marshal(filesList)
	if err != nil {
		return
	}
	this.FilesList = string(list)
	return
}

func (this *MicroSite) GetPackage() FileSystem {
	return this.Package
}

func (this *MicroSite) SetPackage(fileName, url string) {
	this.Package.FileName = fileName
	this.Package.Url = url
	return
}

func (this *MicroSite) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	if len(this.GetFileList()) > 0 {
		f, err := storage.Get(this.Package.Url)
		if err != nil {
			return nil, err
		}

		fileBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		err = this.PublishArchiveFiles(this.Package.FileName, fileBytes, storage)
		if err != nil {
			return nil, err
		}

		var previewPaths []string
		for _, v := range this.GetFileList() {
			previewPaths = append(previewPaths, this.GetPreviewPath(v))
		}
		err = utils.DeleteObjects(storage, previewPaths)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (this *MicroSite) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	var paths []string
	for _, v := range this.GetFileList() {
		paths = append(paths, this.GetPublishedPath(v))
	}
	err = utils.DeleteObjects(storage, paths)
	if err != nil {
		return
	}
	return
}

func (this *MicroSite) PublishArchiveFiles(fileName string, fileBytes []byte, storage oss.StorageInterface) (err error) {
	a, err := unarr.NewArchiveFromMemory(fileBytes)
	if err != nil {
		if err.Error() == INVALID_ARCHIVER_ERROR.Error() {
			utils.Upload(storage, this.GetPublishedPath(fileName), bytes.NewReader(fileBytes))
			return nil
		}
		return
	}
	defer a.Close()

	filesList, err := a.List()
	if err != nil {
		return
	}
	filesList = utils.RemoveUselessArchiveFiles(filesList)

	if len(filesList) > MaximumNumberOfFilesInArchive {
		err = TOO_MANY_FILE_ERROR
		return
	}

	var wg = sync.WaitGroup{}
	var putError error
	var putSemaphore = make(chan struct{}, MaximumNumberOfFilesUploadedAtTheSameTime)

	for _, v := range filesList {
		if putError != nil {
			err = putError
			break
		}
		e := a.EntryFor(v)
		if e != nil {
			if e == io.EOF {
				break
			}

			err = e
			return
		}
		data, e := a.ReadAll()
		if e != nil {
			err = e
			return
		}

		publishedPath := this.GetPublishedPath(a.Name())
		wg.Add(1)
		putSemaphore <- struct{}{}
		go func() {
			defer func() {
				<-putSemaphore
				wg.Done()
			}()
			err2 := utils.Upload(storage, publishedPath, bytes.NewReader(data))
			if err2 != nil {
				putError = err2
			}
		}()
	}

	if putError != nil {
		err = putError
		return
	}
	wg.Wait()

	return
}

func (this *MicroSite) GetFilesListAndPublishPreviewFiles(fileName string, fileBytes []byte, storage oss.StorageInterface) (filesList []string, err error) {
	a, err := unarr.NewArchiveFromMemory(fileBytes)
	if err != nil {
		if err.Error() == INVALID_ARCHIVER_ERROR.Error() {
			filesList = append(filesList, fileName)
			utils.Upload(storage, this.GetPreviewPath(fileName), bytes.NewReader(fileBytes))
			return filesList, nil
		}
		return
	}
	defer a.Close()

	filesList, err = a.List()
	if err != nil {
		return
	}
	filesList = utils.RemoveUselessArchiveFiles(filesList)

	if len(filesList) > MaximumNumberOfFilesInArchive {
		err = TOO_MANY_FILE_ERROR
		return
	}

	var wg = sync.WaitGroup{}
	var putError error
	var putSemaphore = make(chan struct{}, MaximumNumberOfFilesUploadedAtTheSameTime)

	for _, v := range filesList {
		if putError != nil {
			err = putError
			break
		}
		e := a.EntryFor(v)
		if e != nil {
			if e == io.EOF {
				break
			}

			err = e
			return
		}
		data, e := a.ReadAll()
		if e != nil {
			err = e
			return
		}

		previewPath := this.GetPreviewPath(a.Name())
		wg.Add(1)
		putSemaphore <- struct{}{}
		go func() {
			defer func() {
				<-putSemaphore
				wg.Done()
			}()
			err2 := utils.Upload(storage, previewPath, bytes.NewReader(data))
			if err2 != nil {
				putError = err2
			}
		}()
	}

	if putError != nil {
		err = putError
		return
	}
	wg.Wait()

	return
}
