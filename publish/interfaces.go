package publish

import (
	"context"
	"time"

	"github.com/qor5/x/v3/oss"
	"gorm.io/gorm"
)

type PublishAction struct {
	Url      string
	Content  string
	IsDelete bool
}

// @snippet_begin(PublishList)
type List struct {
	PageNumber  int
	Position    int
	ListDeleted bool
	ListUpdated bool
}

// @snippet_end

// @snippet_begin(PublishSchedule)
type Schedule struct {
	ScheduledStartAt *time.Time `gorm:"index"`
	ScheduledEndAt   *time.Time `gorm:"index"`

	ActualStartAt *time.Time
	ActualEndAt   *time.Time
}

// @snippet_end

// @snippet_begin(PublishStatus)
var (
	StatusDraft   = "draft"
	StatusOnline  = "online"
	StatusOffline = "offline"
)

type Status struct {
	Status    string `gorm:"default:'draft'"`
	OnlineUrl string
}

// @snippet_end

// @snippet_begin(PublishVersion)
type Version struct {
	Version       string `gorm:"primaryKey;size:128;not null;"`
	VersionName   string
	ParentVersion string
}

// @snippet_end
type (
	PublishActionsFunc func(ctx context.Context, db *gorm.DB, storage oss.StorageInterface, obj any) (actions []*PublishAction, err error)
	PublishInterface   interface {
		GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*PublishAction, err error)
	}
)

type UnPublishInterface interface {
	GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*PublishAction, err error)
}

type WrapPublishInterface interface {
	WrapPublishActions(in PublishActionsFunc) PublishActionsFunc
}
type WrapUnPublishInterface interface {
	WrapUnPublishActions(in PublishActionsFunc) PublishActionsFunc
}

type AfterPublishInterface interface {
	AfterPublish(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) error
}

type AfterUnPublishInterface interface {
	AfterUnPublish(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) error
}

type StatusInterface interface {
	EmbedStatus() *Status
}

func (s *Status) EmbedStatus() *Status {
	return s
}

func EmbedStatus(v any) *Status {
	iface, ok := v.(StatusInterface)
	if !ok {
		return nil
	}
	return iface.EmbedStatus()
}

type VersionInterface interface {
	EmbedVersion() *Version
}

func (s *Version) EmbedVersion() *Version {
	return s
}

func EmbedVersion(v any) *Version {
	iface, ok := v.(VersionInterface)
	if !ok {
		return nil
	}
	return iface.EmbedVersion()
}

type ScheduleInterface interface {
	EmbedSchedule() *Schedule
}

func (s *Schedule) EmbedSchedule() *Schedule {
	return s
}

func EmbedSchedule(v any) *Schedule {
	iface, ok := v.(ScheduleInterface)
	if !ok {
		return nil
	}
	return iface.EmbedSchedule()
}

type ListInterface interface {
	EmbedList() *List
}

func (s *List) EmbedList() *List {
	return s
}

type (
	PreviewBuilderInterface interface {
		PreviewHTML(ctx context.Context, obj interface{}) string
		ExistedL10n() bool
	}
	PublishModelInterface interface {
		PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) string
	}
)
