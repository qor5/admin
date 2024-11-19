package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
	plogin "github.com/qor5/admin/v3/login"
)

func TestLogin(t *testing.T) {
	h, _ := admin.TestL18nHandler(TestDB)

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
		{
			Name:  "view by en (customized)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				mux := http.NewServeMux()
				c := admin.NewConfig(TestDB, false)
				loginSessionBuilder := c.GetLoginSessionBuilder()
				loginBuilder := c.GetLoginSessionBuilder().GetLoginBuilder()
				loginBuilder.LoginPageFunc(plogin.NewAdvancedLoginPage(func(ctx *web.EventContext, config *plogin.AdvancedLoginPageConfig) (*plogin.AdvancedLoginPageConfig, error) {
					config.WelcomeLabel = "Hello"
					return config, nil
				})(loginBuilder.ViewHelper(), c.GetPresetsBuilder()))
				loginSessionBuilder.Secret("test")
				loginSessionBuilder.Mount(mux)
				mux.Handle("/", c.GetPresetsBuilder())
				return mux
			},
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/auth/login").
					BuildEventFuncRequest()
				req.Header.Add("accept-language", "en")
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`Hello`, `Qor Admin System`, `Email`, `Password`, `Sign in`, `Forget your password?`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
