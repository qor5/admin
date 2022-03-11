package microsite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/go-unarr"
	"github.com/qor/oss"
	"github.com/qor/qor5/microsite/utils"
	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

type MicroSite struct {
	gorm.Model

	publish.Status
	publish.Schedule
	//publish.Version

	Name    string
	PrePath string

	Package   FileSystem `gorm:"type:text"`
	FilesList string     `gorm:"type:text"`
}

func (this MicroSite) GetId() uint {
	return this.ID
}

//func (this MicroSite) GetVersionName() string {
//	return this.VersionName
//}

//func (this *MicroSite) BeforeDelete(db *gorm.DB) (err error) {
//	if this.Status == Status_published {
//		err = Unpublish(db, this, false)
//		return
//	}
//
//	return
//}

func (this MicroSite) GetPackagePath(fileName string) string {
	return fmt.Sprintf("/%s/__package__/%d/%s", PackageAndPreviewPrepath, this.GetId(), fileName)
}

func (this MicroSite) GetPreviewPrePath() string {
	return "/" + path.Join(PackageAndPreviewPrepath, "__preview__", strconv.Itoa(int(this.GetId()))) + "/"
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

func (this *MicroSite) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	if len(this.GetFileList()) > 0 {
		file, err := storage.Get(this.Package.Url)
		if err != nil {
			panic(err)
		}

		now := time.Now().Unix()
		tempDir, err := GetTempFileDir()
		if err != nil {
			return
		}
		//filePath = path.Join(tempDir, fmt.Sprintf("%d_%s_%d.zip", this.GetId(), this.VersionName, now))
		filePath := path.Join(tempDir, fmt.Sprintf("%s_%d_%d.zip", PackageAndPreviewPrepath, this.GetId(), now))

		var dst *os.File
		if dst, err = os.Create(filePath); err != nil {
			panic(err)
		}
		if _, err = io.Copy(dst, file); err != nil {
			panic(err)
		}
		dst.Close()
		defer os.Remove(filePath)

		err = this.PublishArchiveFiles(this.Package.FileName, filePath, storage)
		if err != nil {
			panic(err)
		}

		var previewPaths []string
		for _, v := range this.GetFileList() {
			previewPaths = append(previewPaths, this.GetPreviewPath(v))
		}
		err = utils.DeleteObjects(storage, previewPaths)
		if err != nil {
			panic(err)
		}

	}
	return
}

func (this *MicroSite) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	var paths []string
	for _, v := range this.GetFileList() {
		paths = append(paths, this.GetPublishedPath(v))
	}
	err := utils.DeleteObjects(storage, paths)
	if err != nil {
		panic(err)
	}
	return
}

func (this *MicroSite) PublishArchiveFiles(fileName, filePath string, storage oss.StorageInterface) (err error) {
	a, err := unarr.NewArchive(filePath)
	if err != nil {
		if err.Error() == INVALID_ARCHIVER_ERROR.Error() {
			var f *os.File
			f, err = os.Open(filePath)
			if err != nil {
				return
			}
			var data []byte
			data, err = ioutil.ReadAll(f)
			if err != nil {
				return
			}
			err = f.Close()
			if err != nil {
				return
			}

			utils.Upload(storage, this.GetPublishedPath(fileName), bytes.NewReader(data))
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
	var putSemaphore = make(chan struct{}, MaxNumberOfFilesUploadedAtTheSameTime)

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

func (this *MicroSite) GetFilesListAndPublishPreviewFiles(fileName, filePath string, storage oss.StorageInterface) (filesList []string, err error) {
	a, err := unarr.NewArchive(filePath)
	if err != nil {
		if err.Error() == INVALID_ARCHIVER_ERROR.Error() {
			filesList = append(filesList, fileName)

			var f *os.File
			f, err = os.Open(filePath)
			if err != nil {
				return
			}
			var data []byte
			data, err = ioutil.ReadAll(f)
			if err != nil {
				return
			}
			err = f.Close()
			if err != nil {
				return
			}

			utils.Upload(storage, this.GetPreviewPath(fileName), bytes.NewReader(data))
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
	var putSemaphore = make(chan struct{}, MaxNumberOfFilesUploadedAtTheSameTime)

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
