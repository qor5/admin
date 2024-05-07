package seo_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/seo"
	adminUser "github.com/qor5/admin/v3/seo/testdata/admin"
	customerUser "github.com/qor5/admin/v3/seo/testdata/customer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbForTest *gorm.DB

func init() {
	if db, err := gorm.Open(postgres.Open(os.Getenv("DBURL")), &gorm.Config{}); err != nil {
		panic(err)
	} else {
		dbForTest = db
	}
}

func metaEqual(got, want string) bool {
	for _, s := range strings.Split(want, "\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !strings.Contains(got, s) {
			return false
		}
	}
	return true
}

func TestRenderSameType(t *testing.T) {
	u, _ := url.Parse("http://dev.qor5.com/product/1")
	defaultRequest := &http.Request{
		Method: "GET",
		URL:    u,
	}
	globalSeoSetting := seo.QorSEOSetting{
		Name: "Global SEO",
		Setting: seo.Setting{
			Title: "global | {{SiteName}}",
		},
		Variables: map[string]string{"SiteName": "Qor5 dev"},
		Locale:    l10n.Locale{LocaleCode: "en"},
	}
	cases := []struct {
		name      string
		prepareDB func()
		builder   *seo.Builder
		obj       interface{}
		want      string
	}{
		{
			name: "render_seo_in_the_case_of_types_with_the_same_name_in_different_packages",
			prepareDB: func() {
				dbForTest.Save(&globalSeoSetting)
				seoSettings := []*seo.QorSEOSetting{
					{
						Name: "Customer",
						Setting: seo.Setting{
							Description: "customer user description",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
					{
						Name: "Admin User",
						Setting: seo.Setting{
							Description: "admin user description",
						},
						Locale: l10n.Locale{LocaleCode: "en"},
					},
				}
				if err := dbForTest.Save(&seoSettings).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *seo.Builder {
				builder := seo.New(dbForTest, seo.WithLocales("en"))
				builder.RegisterSEO("Customer", customerUser.User{Name: "CustomerA"}).
					RegisterContextVariable("UserName",
						func(obj interface{}, _ *seo.Setting, _ *http.Request) string {
							return obj.(*customerUser.User).Name
						},
					)
				builder.RegisterSEO("Admin", adminUser.User{Name: "Administrator"}).
					RegisterContextVariable("UserName",
						func(obj interface{}, _ *seo.Setting, _ *http.Request) string {
							return obj.(*adminUser.User).Name
						},
					)
				return builder
			}(),
			obj: &customerUser.User{
				Name: "CustomerA",
				SEO: seo.Setting{
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
