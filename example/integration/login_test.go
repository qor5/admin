package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
)

func TestLogin(t *testing.T) {
	h := admin.TestL18nHandler(TestDB)

	dbr, _ := TestDB.DB()
	profileData.TruncatePut(dbr)

	cases := []multipartestutils.TestCase{
		{
			Name:  "view by en",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/auth/login").
					BuildEventFuncRequest()
				req.Header.Add("accept-language", "en")
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`Welcome`, `Qor Admin System`, `Email`, `Password`, `Sign in`, `Forget your password?`},
		},
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
			ExpectPageBodyContainsInOrder: []string{`欢迎`, `Qor 管理系统`, `邮箱`, `密码`, `登录`, `忘记密码？`},
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
			ExpectPageBodyContainsInOrder: []string{`ようこそ`, `Qor 管理システム`, `メールアドレス`, `パスワード`, `サインイン`, `パスワードをお忘れですか？`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
