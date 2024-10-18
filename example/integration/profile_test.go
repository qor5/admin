package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/role"
)

var profileData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.users (id, created_at, updated_at, deleted_at, name, company, status, favor_post_id, registration_date, account, password, pass_updated_at, login_retry_count, locked, locked_at, reset_password_token, reset_password_token_created_at, reset_password_token_expired_at, totp_secret, is_totp_setup, last_used_totp_code, last_totp_code_used_at, o_auth_provider, o_auth_user_id, o_auth_identifier, session_secure) VALUES (1, '2024-06-18 03:24:28.001791 +00:00', '2024-06-19 07:07:18.502134 +00:00', null, 'qor@theplant.jp', '', 'Available', 0, '0001-01-01', 'qor@theplant.jp', '$2a$10$XKsTcchE1r1X5MyTD0k1keyUwub23DXsjSIQW73MtXfoiqrqbXAnu', '1718681068001453000', 0, false, null, '', null, null, '', false, '', null, '', '', '', 'cdedfb408d634c7240384e00203baf47');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (2, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Manager');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (3, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Editor');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (4, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Viewer');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (1, '2024-08-23 08:43:32.969461 +00:00', '2024-09-12 06:25:17.533058 +00:00', null, 'Admin');
INSERT INTO public.user_role_join (user_id, role_id) VALUES (1, 1);
`, []string{`user_role_join`, `roles`, "users"}))

func TestProfile(t *testing.T) {
	h := admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 1},
		Name:  "qor@theplant.jp",
		Roles: []role.Role{
			{
				Model: gorm.Model{ID: 1},
				Name:  models.RoleAdmin,
			},
		},
	})

	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "index",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`DEMO`, `portal-name='ProfileCompo:`, `<v-avatar`, `text='Q'`, `/v-avatar>`, `qor@theplant.jp`, `Admin`},
		},
		{
			Name:  "rename",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
		{
			"compo_type": "*login.ProfileCompo",
			"compo": {
				"id": ""
			},
			"injector": "__profile__",
			"sync_query": false,
			"method": "Rename",
			"request": {
				"name": "123@theplant.jp"
			}
		}
							`).
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
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
				req := NewMultipartBuilder().
					PageURL("/login-sessions-dialog?__execute_event__=loginSession_eventLoginSessionsDialog").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Login Sessions`, `Time`, `Device`, `IP`, `Status`},
		},
		{
			Name:  "Index Role",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/roles", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor", "Manager", ">Admin<"},
		},
		{
			Name:  "Viewer Add Permission",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/roles").
					EventFunc(actions.AddRowEvent).
					Query(presets.ParamID, "4").
					AddField("listEditor_AddRowFormKey", "Permissions").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Effect", "Actions", "Resources"},
		},
		{
			Name:  "Viewer Update Name",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/roles").
					EventFunc(actions.Update).
					Query(presets.ParamID, "4").
					AddField("Name", "Viewer1").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := role.Role{}
				TestDB.First(&m, 4)
				if m.Name != "Viewer1" {
					t.Fatalf("got %v want Viewer1", m.Name)
					return
				}
			},
		},
		{
			Name:  "Viewer Update Empty Name",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/roles").
					EventFunc(actions.Update).
					Query(presets.ParamID, "4").
					AddField("Name", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name is required"},
		},
		{
			Name:  "Viewer Delete",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/roles").
					EventFunc(actions.DoDelete).
					Query(presets.ParamID, "4").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				m := role.Role{}
				TestDB.First(&m, 4)
				if m.Name != "" {
					t.Fatalf("delete faield %#+v", m)
					return
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}

func TestRoleEditor(t *testing.T) {
	h := admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 888},
		Name:  "viwer@theplant.jp",
		Roles: []role.Role{
			{
				Name: models.RoleEditor,
			},
		},
	})
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Role",
			Debug: true,
			ReqFunc: func() *http.Request {
				profileData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/roles", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor"},
			ExpectPageBodyNotContains:     []string{"Manager", ">Admin<"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
	h = admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 888},
		Name:  "admin@theplant.jp",
		Roles: []role.Role{
			{
				Name: models.RoleAdmin,
			},
		},
	})
}
