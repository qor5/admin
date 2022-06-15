package publish

import (
	"context"
	"time"

	"github.com/qor/oss"
	"gorm.io/gorm"
)

type PublishAction struct {
	Url      string
	Content  string
	IsDelete bool
}

type PublishInterface interface {
	GetPublishActions(db *gorm.DB, ctx context.Context) (actions []*PublishAction)
}

type UnPublishInterface interface {
	GetUnPublishActions(db *gorm.DB, ctx context.Context) (actions []*PublishAction)
}

type AfterPublishInterface interface {
	AfterPublish(db *gorm.DB, storage oss.StorageInterface, ctx context.Context) error
}

type AfterUnPublishInterface interface {
	AfterUnPublish(db *gorm.DB, storage oss.StorageInterface, ctx context.Context) error
}

type AfterSaveNewVersionInterface interface {
	AfterSaveNewVersion(db *gorm.DB, ctx context.Context) error
}

type StatusInterface interface {
	GetStatus() string
	SetStatus(s string)
	GetOnlineUrl() string
	SetOnlineUrl(s string)
}

type VersionInterface interface {
	GetVersion() string
	SetVersion(v string)
	GetVersionName() string
	SetVersionName(v string)
}

type ScheduleInterface interface {
	GetStatus() string

	GetScheduledStartAt() *time.Time
	GetScheduledEndAt() *time.Time
	SetScheduledStartAt(v *time.Time)
	SetScheduledEndAt(v *time.Time)

	GetPublishedAt() *time.Time
	GetUnPublishedAt() *time.Time
	SetPublishedAt(v *time.Time)
	SetUnPublishedAt(v *time.Time)
}

type ListInterface interface {
	GetPageNumber() int
	SetPageNumber(pageNumber int)
	GetPosition() int
	SetPosition(position int)
	GetListDeleted() bool
	SetListDeleted(listDeleted bool)
	GetListUpdated() bool
	SetListUpdated(listUpdated bool)
}
