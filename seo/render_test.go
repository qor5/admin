package seo

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/qor5/admin/v3/l10n"
)

type CustomerUser struct {
	Name string
	SEO  Setting
}
type AdminUser struct {
	Name string
	SEO  Setting
}

func TestRenderSameType(t *testing.T) {
	u, _ := url.Parse("http://dev.qor5.com/product/1")
	defaultRequest := &http.Request{
		Method: "GET",
		URL:    u,
	}
	globalSeoSetting := QorSEOSetting{
		Name: "Global SEO",
		Setting: Setting{
			Title: "global | {{SiteName}}",
		},
		Variables: map[string]string{"SiteName": "Qor5 dev"},
		Locale:    l10n.Locale{LocaleCode: "en"},
	}
	cases := []struct {
		name      string
		prepareDB func()
		builder   *Builder
		obj       interface{}
		want      string
	}{
		{
			name: "render_seo_in_the_case_of_types_with_the_same_name_in_different_packages",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				seoSettings := []*QorSEOSetting{
					{
						Name: "Customer",
						Setting: Setting{
							Description: "customer user description",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Admin User",
						Setting: Setting{
							Description: "admin user description",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
				}
				if err := dbForTest.Save(&seoSettings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Customer", CustomerUser{Name: "CustomerA"}).
					RegisterContextVariable("UserName",
						func(obj interface{}, _ *Setting, _ *http.Request) string {
							return obj.(*CustomerUser).Name
						},
					)
				builder.RegisterSEO("Admin", AdminUser{Name: "Administrator"}).
					RegisterContextVariable("UserName",
						func(obj interface{}, _ *Setting, _ *http.Request) string {
							return obj.(*AdminUser).Name
						},
					)
				return builder
			}(),
			obj: &CustomerUser{
				Name: "CustomerA",
				SEO: Setting{
					Description:      "{{UserName}}",
					EnabledCustomize: true,
				},
			},
			want: `
			<title>global | Qor5 dev</title>
			<meta name='description' content='CustomerA'>
			`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dbForTest.Exec("truncate qor_seo_settings;")
			c.prepareDB()
			if got, _ := c.builder.Render(c.obj, defaultRequest).MarshalHTML(context.TODO()); !metaEqual(string(got), c.want) {
				t.Errorf("Render = %v\nExpected = %v", string(got), c.want)
			}
		})
	}
}
