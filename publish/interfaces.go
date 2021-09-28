package publish

import (
	"github.com/qor/oss"
	"gorm.io/gorm"
)

type PublishAction struct {
	Url      string
	Content  string
	IsDelete bool
}

type PublishInterface interface {
	GetPublishActions(db *gorm.DB) (actions []*PublishAction)
}
type UnPublishInterface interface {
	GetUnPublishActions(db *gorm.DB) (actions []*PublishAction)
}

type AfterPublishInterface interface {
	AfterPublish(db *gorm.DB, storage oss.StorageInterface) error
}

type AfterUnPublishInterface interface {
	AfterUnPublish(db *gorm.DB, storage oss.StorageInterface) error
}

type StatusInterface interface {
	GeStatus() string
	SetStatus(s string)
	GetOnlineUrl() string
	SetOnlineUrl(s string)
}

type VersionInterface interface {
	GetVersionName() string
	SetVersionName(v string)
}
