package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestChangePassword(t *testing.T) {
	current := &User{
		Model:   gorm.Model{ID: 1},
		Name:    "admin",
		Address: "admin address",
		UserPass: login.UserPass{
			Account:  "qor@theplant.jp",
			Password: "$2a$10$N7gloPSgJtB23hYTO9Ww8uBqpAcLn7KOGFcYQFkg5IA92EG8LIZFu",
		},
	}
	h := changePasswordExample(presets.New(), TestDB, current)

	cases := []multipartestutils.TestCase{
		{
			Name:  "index",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Change Password"},
		},
		{
			Name:  "show dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/?__execute_event__=login_openChangePasswordDialog", http.NoBody)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Change your password"},
		},
		{
			Name:  "incorrect password",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/?__execute_event__=login_changePassword").
					AddField("old_password", `1234`).
					AddField("password", `12345`).
					AddField("confirm_password", `12345`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Unable to change password. Please check your inputs and try again."},
		},
		{
			Name:  "password not match",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/?__execute_event__=login_changePassword").
					AddField("old_password", `123`).
					AddField("password", `12345`).
					AddField("confirm_password", `123456`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"The new passwords do not match. Please try again."},
		},
		{
			Name:  "success",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/?__execute_event__=login_changePassword").
					AddField("old_password", `123`).
					AddField("password", `123456`).
					AddField("confirm_password", `123456`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Password successfully changed, please sign-in again", "/auth/logout"},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				require.True(t, current.IsPasswordCorrect("123456"))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}

func TestPasswordWithVisibleToggle(t *testing.T) {
	current := &User{
		Model:   gorm.Model{ID: 1},
		Name:    "admin",
		Address: "admin address",
		UserPass: login.UserPass{
			Account:  "qor@theplant.jp",
			Password: "$2a$10$N7gloPSgJtB23hYTO9Ww8uBqpAcLn7KOGFcYQFkg5IA92EG8LIZFu",
		},
	}

	h := changePasswordExample(presets.New(), TestDB, current)

	cases := []multipartestutils.TestCase{
		{
			Name:  "password field with visible toggle",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/auth/login", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`:password-visible-toggle='true'`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
