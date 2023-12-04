package seo

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/l10n"
	l10n_view "github.com/qor5/admin/l10n/views"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/theplant/testingutils"
	"gorm.io/gorm"
)

func TestUpdate(t *testing.T) {
	cases := []struct {
		name      string
		prepareDB func()
		form      func() (*bytes.Buffer, *multipart.Writer)
		expected  *QorSEOSetting
		locale    string
	}{
		{
			name: "update_setting",
			prepareDB: func() {
				resetDB()
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
			form: func() (*bytes.Buffer, *multipart.Writer) {
				form := &bytes.Buffer{}
				mwriter := multipart.NewWriter(form)
				mwriter.WriteField("Setting.Title", "productB")
				mwriter.Close()
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
			name: "update_variables",
			prepareDB: func() {
				resetDB()
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
			form: func() (*bytes.Buffer, *multipart.Writer) {
				form := &bytes.Buffer{}
				mwriter := multipart.NewWriter(form)
				mwriter.WriteField("Variables.varA", "B")
				mwriter.Close()
				return form, mwriter
			},
			expected: &QorSEOSetting{
				Name:   "Product",
				Locale: l10n.Locale{LocaleCode: "en"},
				Variables: map[string]string{
					"varA": "B",
				},
			},
			locale: "en",
		},
	}

	admin := presets.New().URIPrefix("/admin").DataOperator(gorm2op.DataOperator(dbForTest))
	server := httptest.NewServer(admin)
	builder := NewBuilder(dbForTest, WithLocales("en"))
	builder.RegisterMultipleSEO("Product Detail", "Product")
	builder.Configure(admin)

	l10nBuilder := l10n.New().RegisterLocales("en", "en", "English")
	l10n_view.Configure(admin, dbForTest, l10nBuilder, nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			form, mwriter := c.form()
			req, err := http.DefaultClient.Post(
				server.URL+"/admin/qor-seo-settings?__execute_event__=presets_Update&id=Product_"+c.locale,
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
				t.Errorf(r)
			}
		})
	}

}
