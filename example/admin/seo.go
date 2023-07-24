package admin

import (
	"net/http"

	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/seo"
	"gorm.io/gorm"
)

// @snippet_begin(SeoExample)
var SeoCollection *seo.Collection

func ConfigureSeo(b *presets.Builder, db *gorm.DB) {
	SeoCollection = seo.NewCollection()
	SeoCollection.RegisterSEO(&models.Post{}).RegisterContextVariables(
		"Title",
		func(object interface{}, _ *seo.Setting, _ *http.Request) string {
			if article, ok := object.(models.Post); ok {
				return article.Title
			}
			return ""
		},
	).RegisterSettingVaribles(struct{ Test string }{})
	SeoCollection.RegisterSEOByNames("Product", "Announcement")
	SeoCollection.Configure(b, db)
}

// @snippet_end
