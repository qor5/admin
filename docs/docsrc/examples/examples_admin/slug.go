package examples_admin

import (
	"net/http"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/slug"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func AutoSyncFromExample(b *presets.Builder, db *gorm.DB) http.Handler {
	return autoSyncFromExample(b, db, nil)
}

func autoSyncFromExample(b *presets.Builder, db *gorm.DB, customize func(mb *presets.ModelBuilder, sb *slug.Builder)) http.Handler {
	b.GetI18n().SupportLanguages(language.English, language.SimplifiedChinese, language.Japanese)

	b.DataOperator(gorm2op.DataOperator(db))
	sb := slug.New()
	defer func() {
		b.Use(sb)
	}()

	type WithTitleSlug struct {
		ID            uint
		Title         string
		TitleWithSlug slug.Slug
	}

	err := db.AutoMigrate(&WithTitleSlug{})
	if err != nil {
		panic(err)
	}

	mb := b.Model(&WithTitleSlug{})
	defer func() { mb.Use(sb) }()

	if customize != nil {
		customize(mb, sb)
	}
	return b
}
