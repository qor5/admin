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
const (
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
	Version       string `gorm:"primary_key;size:128;not null;"`
	VersionName   string
	ParentVersion string
}

// @snippet_end
type (
	PublishActionsFunc func(db *gorm.DB, ctx context.Context, storage oss.StorageInterface, obj any) (actions []*PublishAction, err error)
	PublishInterface   interface {
		GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (actions []*PublishAction, err error)
	}
)

type UnPublishInterface interface {
	GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (actions []*PublishAction, err error)
}

type WrapPublishInterface interface {
	WrapPublishActions(in PublishActionsFunc) PublishActionsFunc
}
type WrapUnPublishInterface interface {
	WrapUnPublishActions(in PublishActionsFunc) PublishActionsFunc
}

type AfterPublishInterface interface {
	AfterPublish(db *gorm.DB, storage oss.StorageInterface, ctx context.Context) error
}

type AfterUnPublishInterface interface {
	AfterUnPublish(db *gorm.DB, storage oss.StorageInterface, ctx context.Context) error
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
		PreviewHTML(obj interface{}) string
		ExistedL10n() bool
	}
	PublishModelInterface interface {
		PublishUrl(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) string
	}
)
