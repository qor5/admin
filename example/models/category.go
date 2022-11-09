package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/qor/oss"
	"github.com/qor5/admin/publish"
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model

	Name     string
	Products pq.StringArray `gorm:"type:text[]"`
	publish.Status
	publish.Schedule
	publish.Version
}

func (p *Category) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Category) PrimaryColumnValuesBySlug(slug string) [][]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return [][]string{
		{"id", segs[0]},
		{"version", segs[1]},
	}
}

func (p *Category) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}

func (p *Category) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}
