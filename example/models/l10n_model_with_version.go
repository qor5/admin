package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/qor/oss"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/publish"
	"gorm.io/gorm"
)

type L10nModelWithVersion struct {
	gorm.Model
	Title string

	publish.Status
	publish.Version
	publish.Schedule
	l10n.Locale
}

func (this *L10nModelWithVersion) PrimarySlug() string {
	return fmt.Sprintf("%v_%v_%v", this.ID, this.Version.Version, this.LocaleCode)
}

func (this *L10nModelWithVersion) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 3 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"version":     segs[1],
		"locale_code": segs[2],
	}
}

func (this *L10nModelWithVersion) GetPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}
func (this *L10nModelWithVersion) GetUnPublishActions(db *gorm.DB, ctx context.Context, storage oss.StorageInterface) (objs []*publish.PublishAction, err error) {
	return
}
