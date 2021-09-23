package admin

import (
	"github.com/goplaid/x/presets"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/seo"
	qor_seo "github.com/qor/qor5/seo"
)

var SeoCollection *qor_seo.Collection

type GlobalVaribles struct {
	SiteName string
}

func ConfigureSeo(b *presets.Builder, db *gorm.DB) {
	SeoCollection = qor_seo.New("Site SEO")
	SeoCollection.RegisterGlobalVaribles(&GlobalVaribles{})
	SeoCollection.RegisterSettingModel(&seo.QorSEOSetting{})

	SeoCollection.RegisterSEO(&qor_seo.SEO{
		Name: "Not Found",
	})

	SeoCollection.RegisterSEO(&qor_seo.SEO{
		Name: "Internal Server Error",
	})

	SeoCollection.RegisterSEO(&qor_seo.SEO{
		Name:     "Post",
		Varibles: []string{"Title"},
		Context: func(objects ...interface{}) map[string]string {
			context := make(map[string]string)
			if len(objects) > 0 {
				if article, ok := objects[0].(models.Post); ok {
					context["Title"] = article.Title
				}
			}
			return context
		},
	})

	SeoCollection.Configure(b, db)
}
