package microsite

import (
	"io"

	"github.com/qor/oss"
)

type MicroSiteInterface interface {
	GetID() uint
	GetVersionName() string
	GetUnixKey() string
	SetUnixKey()
	GetStatus() string

	GetPackagePath(fileName string) string
	GetPackageUrl(domain string) string
	GetPreviewPath(fileName string) string
	GetPreviewUrl(domain, fileName string) string
	GetPublishedPath(fileName string) string
	GetPublishedUrl(domain, fileName string) string
	GetFileList() (arr []string)
	SetFilesList(filesList []string)
	GetPackage() FileSystem
	SetPackage(fileName, url string)
	UnArchiveAndPublish(getPath func(string) string, fileName string, f io.Reader, storage oss.StorageInterface) (filesList []string, err error)
}
