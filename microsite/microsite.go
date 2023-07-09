package microsite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	archiver "github.com/mholt/archiver/v4"
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

func (this MicroSite) GetPackagePath(fileName string) string {
	return strings.TrimPrefix(path.Join(PackageAndPreviewPrepath, "__package__", this.GetUnixKey(), fileName), "/")
}

func (this MicroSite) GetPreviewPath(fileName string) string {
	return strings.TrimPrefix(path.Join(PackageAndPreviewPrepath, "__preview__", this.GetUnixKey(), fileName), "/")
}

func (this MicroSite) GetPreviewUrl(domain, fileName string) string {
	return strings.TrimSuffix(domain, "/") + "/" + this.GetPreviewPath(fileName)
}

func (this MicroSite) GetPublishedPath(fileName string) string {
	return path.Join(strings.TrimPrefix(strings.TrimSuffix(this.PrePath, "/"), "/"), fileName)
}

func (this MicroSite) GetPublishedUrl(domain, fileName string) string {
	return strings.TrimSuffix(domain, "/") + "/" + this.GetPublishedPath(fileName)
}

func (this MicroSite) GetPackageUrl(domain string) string {
	return strings.TrimSuffix(domain, "/") + "/" + strings.TrimPrefix(this.Package.Url, "/")
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
		var f *os.File
		f, err = storage.Get(this.Package.Url)
		if err != nil {
			return
		}

		_, err = this.UnArchiveAndPublish(this.GetPublishedPath, this.Package.FileName, f, storage)
		if err != nil {
			return
		}

		var previewPaths []string
		for _, v := range this.GetFileList() {
			previewPaths = append(previewPaths, this.GetPreviewPath(v))
		}
		err = utils.DeleteObjects(storage, previewPaths)
		if err != nil {
			return
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

func (this *MicroSite) UnArchiveAndPublish(getPath func(string) string, fileName string, f io.Reader, storage oss.StorageInterface) (filesList []string, err error) {
	format, reader, err := archiver.Identify(fileName, f)
	if err != nil {
		if err == archiver.ErrNoMatch {
			err = utils.Upload(storage, getPath(fileName), f)
			return
		}
		return
	}

	var wg = sync.WaitGroup{}
	var putError error
	var mutex sync.Mutex

	err = format.(archiver.Extractor).Extract(context.Background(), reader, nil, func(ctx context.Context, f archiver.File) (err error) {
		if f.IsDir() {
			return
		}

		rc, err := f.Open()
		if err != nil {
			return
		}
		defer rc.Close()

		filesList = append(filesList, f.NameInArchive)

		publishedPath := getPath(f.NameInArchive)
		wg.Add(1)
		putSemaphore <- struct{}{}
		go func() {
			defer func() {
				<-putSemaphore
				wg.Done()
			}()
			err2 := utils.Upload(storage, publishedPath, rc)
			if err2 != nil {
				mutex.Lock()
				putError = multierror.Append(putError, err2).ErrorOrNil()
				mutex.Unlock()
			}
		}()

		return
	})
	wg.Wait()
	err = multierror.Append(err, putError).ErrorOrNil()
	return
}
