package admin

import (
	"net/http"

	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/seo"
	"gorm.io/gorm"
)

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
	SeoCollection.RegisterSEOByNames("Not Found", "Internal Server Error")
	SeoCollection.Configure(b, db)
}
