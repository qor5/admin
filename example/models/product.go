package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/qor/oss"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/publish"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model

	Code  string
	Name  string
	Price int
	Image media_library.MediaBox `sql:"type:text;"`
	publish.Status
	publish.Schedule
	publish.Version
}

func (p *Product) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", p.ID, p.Version.Version)
}

func (p *Product) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (p *Product) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (p *Product) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (p *Product) PermissionRN() []string {
	return []string{"products", strconv.Itoa(int(p.ID)), p.Code, p.Version.Version}
}
