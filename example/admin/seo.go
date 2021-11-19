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
	)
	SeoCollection.RegisterSEO("Not Found")
	SeoCollection.RegisterSEO("Internal Server Error")

	seo.Messages_en_US.DynamicMessage = map[string]string{
		"SiteName": "SiteName",
		"Title":    "Title",
		"Post":     "Post",
	}

	seo.Messages_zh_CN.DynamicMessage = map[string]string{
		"SiteName": "站点名称",
		"Title":    "标题",
		"Post":     "帖子",
	}
	SeoCollection.Configure(b, db)
}
