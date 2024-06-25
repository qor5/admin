package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/example/models"

	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/web/v3/multipartestutils"
)

var profileData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, company, status, favor_post_id, registration_date, account, password, pass_updated_at, login_retry_count, locked, locked_at, reset_password_token, reset_password_token_created_at, reset_password_token_expired_at, totp_secret, is_totp_setup, last_used_totp_code, last_totp_code_used_at, o_auth_provider, o_auth_user_id, o_auth_identifier, session_secure) VALUES (1, '2024-06-18 03:24:28.001791 +00:00', '2024-06-19 07:07:18.502134 +00:00', null, 'qor@theplant.jp', '', 'Available', 0, '0001-01-01', 'qor@theplant.jp', '$2a$10$XKsTcchE1r1X5MyTD0k1keyUwub23DXsjSIQW73MtXfoiqrqbXAnu', '1718681068001453000', 0, false, null, '', null, null, '', false, '', null, '', '', '', 'cdedfb408d634c7240384e00203baf47');
`, []string{"users"}))

func TestProfile(t *testing.T) {
	h := admin.TestHandler(TestDB)
	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "login",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/auth/userpass/login").
					AddField("account", "qor@theplant.jp").
					AddField("password", "123").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyNotContains: []string{`/auth/userpass/login`},
		},
		{
			Name:  "rename",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/profile?__execute_event__=accountRenameEvent&id=1").
					AddField("name", "123@theplant.jp").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				u := models.User{}
				TestDB.First(&u, 1)
				if u.Name != "123@theplant.jp" {
					t.Fatalf("Rename failed %#+v", u)
				}
			},
		},
		{
			Name:  "login Sessions",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/profile?__execute_event__=loginSessionDialogEvent&id=1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Login Sessions`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
