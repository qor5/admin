package microsite

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	"github.com/qor/qor5/publish"
	"gorm.io/gorm"
)

// MicroSiteInterface defined MicroSite itself's interface
type MicroSiteInterface interface {
	GetId() uint
	//GetVersionName() string
	//SetVersionPriority(string)
	GetStatus() string
	//TableName() string

	SetStatus(string)
	SetScheduledEndAt(*time.Time)

	GetPackagePath(modelName, fileName string) string
	GetPreviewPrePath(modelName string) string
	GetPreviewUrl(domain, modelName, fileName string) string
	GetPublishedPath(fileName string) string
	GetPublishedUrl(domain, fileName string) string
	GetFileList() (arr []string)
	SetFileList(fileList string)
	GetPackage() FileSystem
	SetPackage(fileName, url string)
	PublishArchivePreviewFilesAndGetArchiveList(modelName, fileName, filePath string, storage oss.StorageInterface) (list []string, err error)
	PublishArchiveFiles(fileName, filePath string, storage oss.StorageInterface) (err error)
}

// MicroSite default qor microsite setting struct
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

func (this *MicroSite) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	if len(this.GetFileList()) > 0 {
		file, err := storage.Get(this.Package.Url)
		if err != nil {
			panic(err)
		}

		var tempDir = "./temp"

		_, err = os.Stat(tempDir)
		if err != nil {
			err = os.Mkdir(tempDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		now := time.Now().Unix()
		//filePath = path.Join(tempDir, fmt.Sprintf("%d_%s_%d.zip", this.GetId(), this.VersionName, now))
		filePath := path.Join(tempDir, fmt.Sprintf("%s_%d_%d.zip", modelName, this.GetId(), now))

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
			previewPaths = append(previewPaths, this.GetPreviewPath(modelName, v))
		}
		err = deleteObjects(storage, previewPaths)
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
	err := deleteObjects(storage, paths)
	if err != nil {
		panic(err)
	}
	return
}

type FileSystem struct {
	FileName string
	Url      string
}

func (this FileSystem) Value() (driver.Value, error) {
	return json.Marshal(this)
}

func (this *FileSystem) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), this)
	case []byte:
		return json.Unmarshal(v, this)
	default:
		return errors.New("not supported")
	}
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

func (this MicroSite) GetPackagePath(modelName, fileName string) string {
	return fmt.Sprintf("/%s/__package__/%d/%s", modelName, this.GetId(), fileName)
}

func (this MicroSite) GetPreviewPrePath(modelName string) string {
	return "/" + path.Join(modelName, "__preview__", strconv.Itoa(int(this.GetId()))) + "/"
}

func (this MicroSite) GetPreviewPath(modelName, fileName string) string {
	return path.Join(this.GetPreviewPrePath(modelName), fileName)
}

func (this MicroSite) GetPreviewUrl(domain, modelName, fileName string) string {
	return strings.TrimSuffix(domain, "/") + this.GetPreviewPath(modelName, fileName)
}

func (this MicroSite) GetPublishedPath(fileName string) string {
	return "/" + path.Join(strings.TrimSuffix(this.PrePath, "/"), fileName)
}

func (this MicroSite) GetPublishedUrl(domain, fileName string) string {
	return strings.TrimSuffix(domain, "/") + this.GetPublishedPath(fileName)
}

// the archive file path
func (this MicroSite) GetFilePath() string {
	return this.Package.Url
}

// the archive file path in s3
func (this MicroSite) GetS3FilePath() string {
	if strings.HasPrefix(this.Package.Url, "http") || strings.HasPrefix(this.Package.Url, "//") {
		return this.Package.Url
	}
	//cdnURL := config.PublicS3Storage.GetEndpoint()
	cdnURL := "123"
	if !strings.HasPrefix(cdnURL, "http") {
		cdnURL = "//" + cdnURL
	}

	return cdnURL + this.Package.Url
}

// URL return file's url with given style
func (this MicroSite) URL(styles ...string) string {
	if this.GetS3FilePath() != "" && len(styles) > 0 {
		ext := path.Ext(this.GetS3FilePath())
		return fmt.Sprintf("%v.%v%v", strings.TrimSuffix(this.GetS3FilePath(), ext), styles[0], ext)
	}
	return this.GetS3FilePath()
}

func (this MicroSite) GetFileList() (arr []string) {
	json.Unmarshal([]byte(this.FilesList), &arr)
	return
}

func (this *MicroSite) SetFileList(fileList string) {
	this.FilesList = fileList
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

var INVALID_ARCHIVER_ERROR = errors.New("unarr: No valid RAR, ZIP, 7Z or TAR archive")
var TOO_MANY_FILE_ERROR = errors.New("Too many uploaded files, please contact the administrator")

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

			upload(storage, this.GetPublishedPath(fileName), bytes.NewReader(data))
			return nil
		}
		return
	}
	defer a.Close()

	list, err := a.List()
	if err != nil {
		return
	}
	list = RemoveUselessArchiveFiles(list)

	if len(list) > 200 {
		err = TOO_MANY_FILE_ERROR
		return
	}

	var wg = sync.WaitGroup{}
	var putError error
	var putSemaphore = make(chan struct{}, 10)

	for _, v := range list {
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

		s3Path := this.GetPublishedPath(a.Name())
		wg.Add(1)
		putSemaphore <- struct{}{}
		go func() {
			defer func() {
				<-putSemaphore
				wg.Done()
			}()
			err2 := upload(storage, s3Path, bytes.NewReader(data))
			//_, err2 := storage.Put(absolutePath, bytes.NewReader(data))
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

func (this *MicroSite) PublishArchivePreviewFilesAndGetArchiveList(modelName, fileName, filePath string, storage oss.StorageInterface) (list []string, err error) {
	a, err := unarr.NewArchive(filePath)
	//a, err := unarr.NewArchiveFromMemory()
	if err != nil {
		if err.Error() == INVALID_ARCHIVER_ERROR.Error() {
			list = append(list, fileName)

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

			upload(storage, this.GetPreviewPath(modelName, fileName), bytes.NewReader(data))
			//storage.Put(path.Join(s3Path, fileName), bytes.NewReader(data))
			return list, nil
		}
		return
	}
	defer a.Close()

	list, err = a.List()
	if err != nil {
		return
	}
	list = RemoveUselessArchiveFiles(list)

	if len(list) > 200 {
		err = TOO_MANY_FILE_ERROR
		return
	}

	var wg = sync.WaitGroup{}
	var putError error
	var putSemaphore = make(chan struct{}, 10)

	for _, v := range list {
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

		s3Path := this.GetPreviewPath(modelName, a.Name())
		wg.Add(1)
		putSemaphore <- struct{}{}
		go func() {
			defer func() {
				<-putSemaphore
				wg.Done()
			}()
			err2 := upload(storage, s3Path, bytes.NewReader(data))
			//_, err2 := storage.Put(absolutePath, bytes.NewReader(data))
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
