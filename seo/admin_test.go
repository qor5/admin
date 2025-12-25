package seo

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theplant/testingutils"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func TestUpdate(t *testing.T) {
	cases := []struct {
		name      string
		prepareDB func()
		builder   func() *Builder
		form      func() (*bytes.Buffer, *multipart.Writer)
		expected  *QorSEOSetting
		locale    string
	}{
		{
			name: "update_setting",
			prepareDB: func() {
				seoSetting := QorSEOSetting{
					Name:   "Product",
					Locale: l10n.Locale{LocaleCode: "en"},
					Setting: Setting{
						Title: "productA",
					},
				}
				if err := dbForTest.Save(&seoSetting).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product Detail")
				builder.RegisterSEO("Product")
				return builder
			},
			form: func() (*bytes.Buffer, *multipart.Writer) {
				form := &bytes.Buffer{}
				mwriter := multipart.NewWriter(form)
				must(mwriter.WriteField("Setting.Title", "productB"))
				must(mwriter.WriteField("id", fmt.Sprintf("Product_%s", "en")))
				must(mwriter.Close())
				return form, mwriter
			},
			expected: &QorSEOSetting{
				Name:   "Product",
				Locale: l10n.Locale{LocaleCode: "en"},
				Setting: Setting{
					Title: "productB",
				},
				Variables: map[string]string{},
			},
			locale: "en",
		},
		{
			name: "update_setting_without_locale",
			prepareDB: func() {
				seoSetting := QorSEOSetting{
					Name: "Product",
					Setting: Setting{
						Title: "productA",
					},
				}
				if err := dbForTest.Save(&seoSetting).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest).AutoMigrate()
				builder.RegisterSEO("Product Detail")
				builder.RegisterSEO("Product")
				return builder
			},
			form: func() (*bytes.Buffer, *multipart.Writer) {
				form := &bytes.Buffer{}
				mwriter := multipart.NewWriter(form)
				must(mwriter.WriteField("Setting.Title", "productB"))
				must(mwriter.WriteField("id", "Product_"))
				must(mwriter.Close())
				return form, mwriter
			},
			expected: &QorSEOSetting{
				Name: "Product",
				Setting: Setting{
					Title: "productB",
				},
				Variables: map[string]string{},
			},
			locale: "",
		},
		{
			name: "update_variables",
			prepareDB: func() {
				seoSetting := QorSEOSetting{
					Name:   "Product",
					Locale: l10n.Locale{LocaleCode: "en"},
					Setting: Setting{
						Title: "productA",
					},
					Variables: map[string]string{
						"varA": "A",
					},
				}
				if err := dbForTest.Save(&seoSetting).Error; err != nil {
					panic(err)
				}
			},
			builder: func() *Builder {
				builder := New(dbForTest, WithLocales("en")).AutoMigrate()
				builder.RegisterSEO("Product Detail")
				builder.RegisterSEO("Product")
				return builder
			},
			form: func() (*bytes.Buffer, *multipart.Writer) {
				form := &bytes.Buffer{}
				mwriter := multipart.NewWriter(form)
				must(mwriter.WriteField("Variables.varA", "B"))
				must(mwriter.WriteField("Setting.Title", "productA"))
				must(mwriter.WriteField("id", fmt.Sprintf("Product_%s", "en")))
				must(mwriter.Close())
				return form, mwriter
			},
			expected: &QorSEOSetting{
				Name:   "Product",
				Locale: l10n.Locale{LocaleCode: "en"},
				Setting: Setting{
					Title: "productA",
				},
				Variables: map[string]string{
					"varA": "B",
				},
			},
			locale: "en",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resetDB()
			if c.prepareDB != nil {
				c.prepareDB()
			}

			admin := presets.New().URIPrefix("/admin").DataOperator(gorm2op.DataOperator(dbForTest))
			server := httptest.NewServer(admin)

			l10nBuilder := l10n.New(dbForTest)
			l10nBuilder.RegisterLocales(c.locale, c.locale, c.locale, "")
			builder := c.builder()
			builder.Install(admin)

			form, mwriter := c.form()
			req, err := http.DefaultClient.Post(
				server.URL+"/admin/qor-seo-settings?__execute_event__=presets_Update",
				mwriter.FormDataContentType(),
				form)
			if err != nil {
				t.Fatal(err)
			}
			if req.StatusCode != 200 {
				t.Errorf("Update should be processed successfully, status code is %v", req.StatusCode)
			}

			seoSetting := &QorSEOSetting{}
			err = dbForTest.First(seoSetting, "name = ? and locale_code = ?", "Product", c.locale).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				t.Errorf("SEO Setting should be updated successfully")
			}
			var actualSetting QorSEOSetting
			actualSetting.Name = seoSetting.Name
			actualSetting.Setting = seoSetting.Setting
			actualSetting.LocaleCode = seoSetting.LocaleCode
			actualSetting.Variables = seoSetting.Variables
			r := testingutils.PrettyJsonDiff(c.expected, actualSetting)
			if r != "" {
				t.Error(r)
			}
		})
	}
}
