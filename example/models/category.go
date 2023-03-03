package models

import (
	"context"
	"fmt"
	"strconv"
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

func (c *Category) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", c.ID, c.Version.Version)
}

func (c *Category) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}
	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (c *Category) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}

func (c *Category) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}

func (c *Category) PermissionRN() []string {
	return []string{"categories", strconv.Itoa(int(c.ID)), c.Version.Version}
}
