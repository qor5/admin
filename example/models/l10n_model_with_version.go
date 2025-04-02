package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/oss"
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

func (lmv *L10nModelWithVersion) PrimarySlug() string {
	return fmt.Sprintf("%v_%v_%v", lmv.ID, lmv.Version.Version, lmv.LocaleCode)
}

func (lmv *L10nModelWithVersion) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 3 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"id":          segs[0],
		"version":     segs[1],
		"locale_code": segs[2],
	}
}

func (lmv *L10nModelWithVersion) GetPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (lmv *L10nModelWithVersion) GetUnPublishActions(ctx context.Context, db *gorm.DB, storage oss.StorageInterface) (actions []*publish.PublishAction, err error) {
	return
}

func (lmv *L10nModelWithVersion) PermissionRN() []string {
	return []string{"l10n_model_with_versions", strconv.Itoa(int(lmv.ID)), lmv.LocaleCode, lmv.Version.Version}
}
