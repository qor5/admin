package starter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/gormx"
	"github.com/theplant/inject"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/starter"
)

func TestUsers(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()
	cases := []TestCase{
		{
			Name:  "Index Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`qor@theplant.jp`},
		},
		{
			Name:  "Lock Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc("eventUnlockUser").
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`success`},
		},
		{
			Name:  "Send Reset Password Email Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc("eventSendResetPasswordEmail").
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`auth/reset-password`},
		},
		{
			Name:  "Revoke TOTP Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc("eventRevokeTOTP").
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`success`},
		},
		{
			Name:  "Edit User",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Type`, "Actions"},
		},
		{
			Name:  "Update User",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Type`, "Actions"},
		},
		{
			Name:  "User InValidate",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("Name", "test@theplant.jp").
					AddField("OAuthIdentifier", "test@theplant.jp").
					AddField("status", "active").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`test@theplant.jp`, `Email`, `Email is required`, `Company`, "Roles"},
		},
		{
			Name:  "User Update With Google Provider",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("Name", "viwer@theplant.jp").
					AddField("Account", "viwer@theplant.jp").
					AddField("OAuthProvider", "google").
					AddField("OAuthIdentifier", "viwer@theplant.jp").
					AddField("status", "active").
					AddField("Roles", "1").
					AddField("Roles", "2").
					AddField("Roles", "3").
					BuildEventFuncRequest()
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				user := &starter.User{}
				db.Preload("Roles").First(user, 1)
				if user.Account != "viwer@theplant.jp" {
					t.Fatalf(`expected "viwer@theplant.jp" but got "%s"`, user.Account)
					return
				}
				if len(user.Roles) != 3 {
					t.Fatalf(`expected 3 roles but got %d`, len(user.Roles))
					return
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
		})
	}
}
