package models

import (
	"fmt"
	"strings"

	"github.com/qor5/admin/l10n"
	"gorm.io/gorm"
)

type L10nModel struct {
	gorm.Model
	Title string

	l10n.Locale
}

func (this *L10nModel) PrimarySlug() string {
	return fmt.Sprintf("%v_%v", this.ID, this.LocaleCode)
}

func (this *L10nModel) PrimaryColumnValuesBySlug(slug string) map[string]string {
	segs := strings.Split(slug, "_")
	if len(segs) != 2 {
		panic("wrong slug")
	}

	return map[string]string{
		"id":          segs[0],
		"locale_code": segs[1],
	}
}
