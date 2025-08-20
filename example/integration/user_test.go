package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var userData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, company, status, favor_post_id, registration_date, account, password, pass_updated_at, login_retry_count, locked, locked_at, reset_password_token, reset_password_token_created_at, reset_password_token_expired_at, totp_secret, is_totp_setup, last_used_totp_code, last_totp_code_used_at, o_auth_provider, o_auth_user_id, o_auth_identifier, session_secure) VALUES (1, '2024-10-16 07:45:53.799531 +00:00', '2025-01-09 06:44:08.969119 +00:00', null, 'qor@theplant.jp', '', 'active', 0, '0001-01-01', 'qor@theplant.jp', '$2a$10$T8nXpBbtZTLI3u1dhgnyjOqL/ILI3lBKTsxNICVNVoNDWy8aWc0Eq', '1729064753799196000', 0, false, null, 'NzE0ZWRjYjUtMDY2Ni00ZDU0LTlhMGItZjU1MDI3Y2U0MGEy', '2025-01-09 06:44:08.967436 +00:00', '2025-01-09 06:54:08.967436 +00:00', '', false, '', null, '', '', '', 'fd79b2cc2e33eaa53fa28e04c6ce5f17');
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, company, status, favor_post_id, registration_date, account, password, pass_updated_at, login_retry_count, locked, locked_at, reset_password_token, reset_password_token_created_at, reset_password_token_expired_at, totp_secret, is_totp_setup, last_used_totp_code, last_totp_code_used_at, o_auth_provider, o_auth_user_id, o_auth_identifier, session_secure) VALUES (2, '2024-10-16 07:45:53.799531 +00:00', '2025-04-29 10:09:36.103221 +00:00', null, 'viwer@theplant.jp', '', 'active', 0, '2024-11-12', 'viwer@theplant.jp', '$2a$10$T8nXpBbtZTLI3u1dhgnyjOqL/ILI3lBKTsxNICVNVoNDWy8aWc0Eq', '1729064753799196000', 0, false, null, '', null, null, '', false, '', null, 'google', '', 'viwer@theplant.jp', '3569e78fec21684afd690cdef77658c0');
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, company, status, favor_post_id, registration_date, account, password, pass_updated_at, login_retry_count, locked, locked_at, reset_password_token, reset_password_token_created_at, reset_password_token_expired_at, totp_secret, is_totp_setup, last_used_totp_code, last_totp_code_used_at, o_auth_provider, o_auth_user_id, o_auth_identifier, session_secure) VALUES (3, '2025-01-09 06:46:36.217223 +00:00', '2025-04-29 09:34:01.517918 +00:00', null, 'test', '', 'inactive', 0, '2025-01-09', 'test@theplant.jp', '$2a$10$Xl2G1b89YovWo86AJfGuauO8yJ7IxXG.mi8IJV71JFGR5wWOyCHU2', '1736405573158607000', 0, false, null, 'ODkxMTZlMDctZTZjMS00ZjIxLTg4ZmQtYmFiYmE5NWE1MzFm', '2025-01-09 08:16:37.015593 +00:00', '2025-01-09 08:26:37.015593 +00:00', '', false, '', null, '', '', 'test@theplant.jp', '414e1ce054e6423c109516613c609dfa');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (2, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Manager');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (3, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Editor');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (4, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Viewer');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (1, '2024-08-23 08:43:32.969461 +00:00', '2024-09-12 06:25:17.533058 +00:00', null, 'Admin');
INSERT INTO public.user_role_join (user_id, role_id) VALUES (1, 1);
`, []string{`user_role_join`, `roles`, "users"}))

func TestUsers(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)

	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/users").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`viwer@theplant.jp`, `qor@theplant.jp`},
		},
		{
			Name:  "Edit User",
			Debug: true,
			ReqFunc: func() *http.Request {
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "2").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`viwer@theplant.jp`, `OAuth Provider`, `google`, `Email`, `Company`, "Roles"},
		},
		{
			Name:  "User InValidate",
			Debug: true,
			ReqFunc: func() *http.Request {
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					Query(presets.ParamID, "3").
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
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					Query(presets.ParamID, "2").
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
				user := &models.User{}
				TestDB.Preload("Roles").First(user, 2)
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
		{
			Name:  "User Create",
			Debug: true,
			ReqFunc: func() *http.Request {
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/users").
					EventFunc(actions.Update).
					AddField("Name", "viwer2@theplant.jp").
					AddField("Account", "viwer2@theplant.jp").
					AddField("OAuthProvider", "google").
					AddField("OAuthIdentifier", "viwer2@theplant.jp").
					AddField("status", "active").
					AddField("Roles", "1").
					AddField("Roles", "2").
					AddField("Roles", "3").
					BuildEventFuncRequest()
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				user := &models.User{}
				TestDB.Preload("Roles").Order("id desc").First(user)
				if user.Account != "viwer2@theplant.jp" {
					t.Fatalf(`expected "viwer2@theplant.jp" but got "%s"`, user.Account)
					return
				}
				if len(user.Roles) != 3 {
					t.Fatalf(`expected 3 roles but got %d`, len(user.Roles))
					return
				}
			},
		},
		{
			Name:  "User Open Dialog Post",
			Debug: true,
			ReqFunc: func() *http.Request {
				userData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/dialog-select-favor-posts").
					EventFunc(actions.OpenListingDialog).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Posts`, `Search`, `ID`, `Title`, "Body"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
