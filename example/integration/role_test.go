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

var roleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (2, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Manager');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (3, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Editor');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (4, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Viewer');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (1, '2024-08-23 08:43:32.969461 +00:00', '2024-09-12 06:25:17.533058 +00:00', null, 'Admin');
`, []string{`user_role_join`, `roles`}))

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
				roleData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/roles", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor"},
			ExpectPageBodyNotContains:     []string{"Manager", "Admin"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}

func TestRoleAdmin(t *testing.T) {
	h := admin.TestHandler(TestDB, &models.User{
		Model: gorm.Model{ID: 888},
		Name:  "admin@theplant.jp",
		Roles: []role.Role{
			{
				Name: models.RoleAdmin,
			},
		},
	})
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Role",
			Debug: true,
			ReqFunc: func() *http.Request {
				roleData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/roles", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor", "Manager", "Admin"},
		},
		{
			Name:  "Viewer Add Permission",
			Debug: true,
			ReqFunc: func() *http.Request {
				roleData.TruncatePut(dbr)
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
				roleData.TruncatePut(dbr)
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
				roleData.TruncatePut(dbr)
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
				roleData.TruncatePut(dbr)
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
