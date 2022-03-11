package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/qor/oss"
	"github.com/qor/qor5/media/media_library"
	"github.com/qor/qor5/publish"
	"github.com/qor/qor5/seo"
	"github.com/qor/qor5/slug"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model

	Title         string
	TitleWithSlug slug.Slug
	Seo           seo.Setting
	Body          string
	HeroImage     media_library.MediaBox `sql:"type:text;"`
	BodyImage     media_library.MediaBox `sql:"type:text;"`
	UpdatedAt     time.Time
	CreatedAt     time.Time

	publish.Status
	publish.Schedule
	publish.Version
}

func (p *Post) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Post) PrimaryColumnValuesBySlug(slug string) [][]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return [][]string{
		{"id", segs[0]},
		{"version", segs[1]},
	}
}

func (p *Post) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	return
}

func (p *Post) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction) {
	return
}
