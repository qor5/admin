package integration_test

import (
	"golang.org/x/text/language"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/examples"
	. "github.com/qor5/web/v3/multipartestutils"
)

func TestLanguage(t *testing.T) {
	db := TestDB
	dbr, _ := db.DB()
	p := examples.Preset1(db)
	p.GetI18n().SupportLanguages(
		language.English,
		language.SimplifiedChinese,
		language.Japanese)
	var handler http.Handler = p

	cases := []TestCase{
		{
			Name: "English Icon",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/languages").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{presets.EnLanguageIcon},
		},
		{
			Name: "SimplifiedChinese Icon",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/languages").
					Query("lang", language.SimplifiedChinese.String()).
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{presets.ZhLanguageIcon},
		},
		{
			Name: "Japanese Icon",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/languages").
					Query("lang", language.Japanese.String()).
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{presets.JPIcon},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, handler)
		})
	}
}
