package microsite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/mholt/archiver/v4"
	"github.com/qor5/admin/v3/microsite/utils"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type MicroSite struct {
	gorm.Model
	publish.Status
	publish.Schedule
	publish.Version
	Name        string
	Description string
	PrePath     string
	Package     FileSystem `gorm:"type:text"`
	FilesList   string     `gorm:"type:text"`
	UnixKey     string
}

func (this *MicroSite) PermissionRN() []string {
	return []string{"microsite_models", strconv.Itoa(int(this.ID)), this.Version.Version}
}

func (this *MicroSite) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", this.ID, this.Version.Version)
}

func (*MicroSite) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
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
}

func (this *MicroSite) GetPackage() FileSystem {
	return this.Package
}

func (this *MicroSite) SetPackage(fileName, url string) {
	this.Package.FileName = fileName
	this.Package.Url = url
}

type contextKeyType int

const contextKey contextKeyType = iota

func (b *Builder) ContextValueProvider(in context.Context) context.Context {
	return context.WithValue(in, contextKey, b)
}

func builderFromContext(c context.Context) (b *Builder, ok bool) {
	b, ok = c.Value(contextKey).(*Builder)
	return
}

func (this *MicroSite) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	if len(this.GetFileList()) == 0 {
		return
	}
	mib, ok := builderFromContext(ctx)
	if !ok {
		panic("use publisher.ContextValueFuncs(micrositeBuilder.ContextValueFunc) to set up microsite builder into context")
	}

	var previewPaths []string

	wg := sync.WaitGroup{}
	var copyError error
	var mutex sync.Mutex
	for _, v := range this.GetFileList() {
		wg.Add(1)
		copySemaphore <- struct{}{}
		go func(v string) {
			defer func() {
				wg.Done()
				<-copySemaphore
			}()
			err = utils.Copy(storage, getPreviewPath(this, v, mib), this.GetPublishedPath(v))
			if err != nil {
				mutex.Lock()
				copyError = multierror.Append(copyError, err).ErrorOrNil()
				mutex.Unlock()
				return
			}
			mutex.Lock()
			previewPaths = append(previewPaths, getPreviewPath(this, v, mib))
			mutex.Unlock()
		}(v)
	}

	wg.Wait()

	if len(previewPaths) > 0 {
		err = utils.DeleteObjects(storage, previewPaths)
	}
	err = multierror.Append(err, copyError).ErrorOrNil()

	return
}

func (this *MicroSite) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
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

func (*MicroSite) UnArchiveAndPublish(getPath func(string) string, fileName string, f io.Reader, storage oss.StorageInterface) (filesList []string, err error) {
	format, reader, err := archiver.Identify(context.Background(), fileName, f)
	if err != nil {
		if errors.Is(err, archiver.NoMatch) {
			err = utils.Upload(storage, getPath(fileName), f)
			return
		}
		return
	}

	wg := sync.WaitGroup{}
	var putError error
	var mutex sync.Mutex

	err = format.(archiver.Extractor).Extract(context.Background(), reader, func(ctx context.Context, info archiver.FileInfo) (err error) {
		if info.IsDir() || strings.Contains(info.NameInArchive, "__MACOSX") || strings.Contains(info.NameInArchive, "DS_Store") {
			return
		}

		rc, err := info.Open()
		if err != nil {
			return
		}

		mutex.Lock()
		filesList = append(filesList, info.NameInArchive)
		mutex.Unlock()

		publishedPath := getPath(info.NameInArchive)
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

		return rc.Close()
	})
	wg.Wait()
	err = multierror.Append(err, putError).ErrorOrNil()
	return
}
