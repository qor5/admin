package admin

import (
	"net/http"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/seo"
	"gorm.io/gorm"
)

// @snippet_begin(SeoExample)
var seoBuilder *seo.Builder

func ConfigureSeo(pb *presets.Builder, db *gorm.DB, locales ...string) {
	seoBuilder = seo.New(db, seo.WithLocales(locales...))
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
	seoBuilder.Install(pb)
}

// @snippet_end
