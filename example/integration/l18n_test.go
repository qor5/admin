package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestL18n(t *testing.T) {
	h, _ := admin.TestL18nHandler(TestDB)

	dbr, _ := TestDB.DB()
	profileData.TruncatePut(dbr)

	cases := []multipartestutils.TestCase{
		{
			Name:  "view by zh",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/auth/login").
					BuildEventFuncRequest()
				req.Header.Add("accept-language", "zh")
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`邮箱`},
		},
		{
			Name:  "view by ja",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/auth/login").
					BuildEventFuncRequest()
				req.Header.Add("accept-language", "ja")
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`メールアドレス`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
