package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
)

var activityLogsData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.cms_activity_logs (id, created_at, updated_at, deleted_at, user_id, action, hidden, model_name, model_keys, model_label, model_link, detail, scope) VALUES (1, '2024-10-16 08:34:41.503174 +00:00', '2024-10-16 08:34:41.503543 +00:00', null, '1', 'login', false, 'User', '1', '', '', 'null', '');
`, []string{`cms_activity_logs`}))

func TestActivityLogs(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Activity Logs",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityLogsData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"login"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
