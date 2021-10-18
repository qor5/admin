package admin

import (
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/seo"
	"gorm.io/gorm"
)

var SeoCollection *seo.Collection

type GlobalVaribles struct {
	SiteName string
}

func ConfigureSeo(b *presets.Builder, db *gorm.DB) {
	SeoCollection = seo.New("Site SEO")
	SeoCollection.RegisterGlobalVariables(&GlobalVaribles{})
	SeoCollection.RegisterSettingModel(&seo.QorSEOSetting{})

	SeoCollection.RegisterSEO(&seo.SEO{
		Name: "Not Found",
	})

	SeoCollection.RegisterSEO(&seo.SEO{
		Name: "Internal Server Error",
	})

	SeoCollection.RegisterSEO(&seo.SEO{
		Name:      "Post",
		Model:     &models.Post{},
		Variables: []string{"Title"},
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

	seo.Messages_en_US.DynamicMessage = map[string]string{
		"SiteName":              "SiteName",
		"Title":                 "Title",
		"Not Found":             "Not Found",
		"Internal Server Error": "Internal Server Error",
		"Post":                  "Post",
	}

	seo.Messages_zh_CN.DynamicMessage = map[string]string{
		"SiteName":              "站点名称",
		"Title":                 "标题",
		"Not Found":             "404页面",
		"Internal Server Error": "错误页面",
		"Post":                  "帖子",
	}
	SeoCollection.Configure(b, db)
}
