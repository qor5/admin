package seo

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	_ "github.com/lib/pq"
)

func TestRender(t *testing.T) {
	u, _ := url.Parse("http://dev.qor5.com/product/1")
	defaultRequest := &http.Request{
		Method: "GET",
		URL:    u,
	}

	globlalSeoSetting := TestQorSEOSetting{
		QorSEOSetting: QorSEOSetting{
			Name: GlobalSEO,
			Setting: Setting{
				Title: "global | {{SiteName}}",
			},
			Variables: map[string]string{"SiteName": "Qor5 dev"},
		},
	}

	tests := []struct {
		name       string
		prepareDB  func()
		collection *Collection
		obj        interface{}
		want       string
	}{
		{
			name:       "Render Golabl SEO with setting variables and default context variables",
			prepareDB:  func() { GlobalDB.Save(&globlalSeoSetting) },
			collection: NewCollection().RegisterSettingModel(&TestQorSEOSetting{}),
			obj:        GlobalSEO,
			want: `
			<title>global | Qor5 dev</title>
			<meta property='og:url' name='og:url' content='http://dev.qor5.com/product/1'>
			`,
		},

		{
			name: "Render seo setting with global setting variables",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name: "Product",
						Setting: Setting{
							Title: "product | {{SiteName}}",
						},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: NewCollection().RegisterSettingModel(&TestQorSEOSetting{}).RegisterSEOByNames("Product"),
			obj:        "Product",
			want:       `<title>product | Qor5 dev</title>`,
		},

		{
			name: "Render seo setting with setting and context variables",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name: "Product",
						Setting: Setting{
							Title: "product {{ProductTag}} | {{SiteName}}",
						},
						Variables: map[string]string{"ProductTag": "Men"},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: func() *Collection {
				collection := NewCollection().RegisterSettingModel(&TestQorSEOSetting{})
				collection.RegisterSEO("Product").
					RegisterSettingVaribles(struct{ ProductTag string }{}).
					RegisterContextVariables("og:image", func(_ interface{}, _ *Setting, _ *http.Request) string {
						return "http://dev.qor5.com/images/logo.png"
					})
				return collection
			}(),
			obj: "Product",
			want: `
			<title>product Men | Qor5 dev</title>
			<meta property='og:image' name='og:image' content='http://dev.qor5.com/images/logo.png'>`,
		},

		{
			name: "Render model setting with global and seo setting variables",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name:      "Product",
						Variables: map[string]string{"ProductTag": "Men"},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: func() *Collection {
				collection := NewCollection().RegisterSettingModel(&TestQorSEOSetting{})
				collection.RegisterSEO(&Product{})
				return collection
			}(),
			obj: Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1 | {{ProductTag}} | {{SiteName}}",
					EnabledCustomize: true,
				},
			},
			want: `<title>product1 | Men | Qor5 dev</title>`,
		},

		{
			name: "Render model setting with default seo setting",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name: "Product",
						Setting: Setting{
							Title: "product | Qor5 dev",
						},
						Variables: map[string]string{"ProductTag": "Men"},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: func() *Collection {
				collection := NewCollection().RegisterSettingModel(&TestQorSEOSetting{})
				collection.RegisterSEO(&Product{})
				return collection
			}(),
			obj: Product{
				Name: "product 1",
				SEO: Setting{
					Title:            "product1 | {{ProductTag}} | {{SiteName}}",
					EnabledCustomize: false,
				},
			},
			want: `<title>product | Qor5 dev</title>`,
		},

		{
			name: "Render model setting with inherite gloabl and seo setting",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name: "Product",
						Setting: Setting{
							Description: "product description",
						},
						Variables: map[string]string{"ProductTag": "Men"},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: func() *Collection {
				collection := NewCollection().RegisterSettingModel(&TestQorSEOSetting{})
				collection.RegisterSEO(&Product{})
				return collection
			}(),
			obj: Product{
				Name: "product 1",
				SEO: Setting{
					Keywords:         "shoes, {{ProductTag}}",
					EnabledCustomize: true,
				},
			},
			want: `
			<title>global | Qor5 dev</title>
			<meta name='description' content='product description'>
			<meta name='keywords' content='shoes, Men'>
			`,
		},

		{
			name: "Render model setting without inherite gloabl and seo setting",
			prepareDB: func() {
				GlobalDB.Save(&globlalSeoSetting)
				product := TestQorSEOSetting{
					QorSEOSetting: QorSEOSetting{
						Name: "Product",
						Setting: Setting{
							Description: "product description",
						},
						Variables: map[string]string{"ProductTag": "Men"},
					},
				}
				GlobalDB.Save(&product)
			},
			collection: func() *Collection {
				collection := NewCollection().SetInherited(false).RegisterSettingModel(&TestQorSEOSetting{})
				collection.RegisterSEO(&Product{})
				return collection
			}(),
			obj: Product{
				Name: "product 1",
				SEO: Setting{
					Keywords:         "shoes, {{ProductTag}}",
					EnabledCustomize: true,
				},
			},
			want: `
			<title></title>
			<meta name='description'>
			<meta name='keywords' content='shoes, Men'>
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetDB()
			tt.prepareDB()
			if got, _ := tt.collection.Render(tt.obj, defaultRequest).MarshalHTML(context.TODO()); !metaEqual(string(got), tt.want) {
				t.Errorf("Render = %v, want %v", string(got), tt.want)
			}
		})
	}

}
