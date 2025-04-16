package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
	"github.com/spf13/cast"
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
		panic(presets.ErrNotFound("wrong slug"))
	}

	_, err := cast.ToInt64E(segs[0])
	if err != nil {
		panic(presets.ErrNotFound(fmt.Sprintf("wrong slug %q: %v", slug, err)))
	}

	return map[string]string{
		"id":      segs[0],
		"version": segs[1],
	}
}

func (c *Category) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (c *Category) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (c *Category) PermissionRN() []string {
	return []string{"categories", strconv.Itoa(int(c.ID)), c.Version.Version}
}
