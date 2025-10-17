package starter_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/starter"
)

func TestUsers(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)

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
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
		})
	}
}
