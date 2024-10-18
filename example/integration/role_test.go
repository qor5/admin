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
	"github.com/qor5/admin/v3/role"
)

var roleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (2, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Manager');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (3, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Editor');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (4, '2024-08-23 08:43:32.969461 +00:00', '2024-08-23 08:43:32.969461 +00:00', null, 'Viewer');
INSERT INTO public.roles (id, created_at, updated_at, deleted_at, name) VALUES (1, '2024-08-23 08:43:32.969461 +00:00', '2024-09-12 06:25:17.533058 +00:00', null, 'Admin');
`, []string{`user_role_join`, `roles`}))

func TestRole(t *testing.T) {
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
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
