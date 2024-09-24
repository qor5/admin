package examples_admin

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/seo"
)

type SEOPost struct {
	gorm.Model
	Title string
	SEO   seo.Setting
}

func SEOExampleBasic(b *presets.Builder, db *gorm.DB) http.Handler {
	err := db.AutoMigrate(&SEOPost{})
	if err != nil {
		panic(err)
	}

	b.DataOperator(gorm2op.DataOperator(db))

	mb := b.Model(&SEOPost{})
	mb.Detailing("Title", seo.SeoDetailFieldName).Drawer(true)
	seob := seo.New(db).AutoMigrate()
	seob.RegisterSEO("Post", &SEOPost{}).
		RegisterContextVariable(
			"Title",
			func(object interface{}, _ *seo.Setting, _ *http.Request) string {
				if article, ok := object.(SEOPost); ok {
					return article.Title
				}
				return ""
			},
		).
		RegisterSettingVariables("Test")

	b.Use(seob)
	mb.Use(seob)
	return b
}
