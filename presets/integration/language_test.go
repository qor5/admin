package integration_test

import (
	"net/http"
	"testing"

	"golang.org/x/text/language"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/presets/examples"
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
			Name: "languages listing",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/languages").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"Code", "Name"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, handler)
		})
	}
}
