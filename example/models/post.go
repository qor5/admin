package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
	"github.com/qor5/x/v3/oss"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model

	Title         string
	TitleWithSlug string
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

func (p *Post) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (p *Post) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (p *Post) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (p *Post) PermissionRN() []string {
	return []string{"posts", strconv.Itoa(int(p.ID)), p.Version.Version}
}
