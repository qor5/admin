package admin

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/seo"
)

// @snippet_begin(SeoExample)
var seoBuilder *seo.Builder

func configureSeo(pb *presets.Builder, db *gorm.DB, locales ...string) {
	seoBuilder = seo.New(db, seo.WithLocales(locales...)).AutoMigrate()
	seoBuilder.RegisterSEO("Post", &models.Post{}).RegisterContextVariable(
		"Title",
		func(object interface{}, _ *seo.Setting, _ *http.Request) string {
			if article, ok := object.(models.Post); ok {
				return article.Title
			}
			return ""
		},
	).RegisterSettingVariables("Test")
	seoBuilder.RegisterSEO("Product")
	seoBuilder.RegisterSEO("Announcement")
	pb.Use(seoBuilder)
}

// @snippet_end
