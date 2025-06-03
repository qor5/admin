package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"gorm.io/gorm"
)

type L10nModel struct {
	gorm.Model
	Title string

	l10n.Locale
}

func (lm *L10nModel) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", lm.ID, lm.LocaleCode)
}

func (lm *L10nModel) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic(presets.ErrNotFound("wrong slug"))
	}

	return map[string]string{
		"id":          segs[0],
		"locale_code": segs[1],
	}
}

func (lm *L10nModel) PermissionRN() []string {
	return []string{"l10n_models", strconv.Itoa(int(lm.ID)), lm.LocaleCode}
}
