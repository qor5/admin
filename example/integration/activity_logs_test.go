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

var activityLogsData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.cms_activity_logs (id, created_at, updated_at, deleted_at, user_id, action, hidden, model_name, model_keys, model_label, model_link, detail, scope) VALUES (1, '2024-10-16 08:34:41.503174 +00:00', '2024-10-16 08:34:41.503543 +00:00', null, '888', 'login', false, 'User', '888', '', '', 'null', '');
`, []string{`cms_activity_logs`}))

var eCDashboardData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.products (id, created_at, updated_at, deleted_at, code, name, price, image, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-10-18 09:11:31.659366 +00:00', '2024-10-18 09:11:31.659366 +00:00', null, '123', '123', 10, null, 'draft', '', null, null, null, null, '2024-10-18-v01', '2024-10-18-v01', '');
INSERT INTO public.orders (id, created_at, updated_at, deleted_at, source, status, delivery_method, payment_method, confirmed_at, order_items) VALUES (1, null, null, null, null, 'Pending', null, null, null, null);
`, []string{`products`, "orders"}))

func TestActivityLogs(t *testing.T) {
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
			Name:  "Index Activity Logs",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityLogsData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"login"},
		},

		{
			Name:  "Index ECDashboard",
			Debug: true,
			ReqFunc: func() *http.Request {
				eCDashboardData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/ec-dashboard", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Statistics", "Order Status", "Pending"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
