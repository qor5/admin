package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var listModelData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.list_models (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, page_number, position, list_deleted, list_updated, version, version_name, parent_version) VALUES (1, '2024-10-18 07:23:43.969040 +00:00', '2024-10-18 07:23:43.969040 +00:00', null, '123', 'online', '/tmp/public', null, null, null, null, 1, 0, false, false, '2024-10-18-v01', '2024-10-18-v01', '');
INSERT INTO public.list_models (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, page_number, position, list_deleted, list_updated, version, version_name, parent_version) VALUES (2, '2024-10-18 07:23:43.969040 +00:00', '2024-10-18 07:23:43.969040 +00:00', null, '456', 'draft', '', null, null, null, null, 0, 0, false, false, '2024-10-18-v01', '2024-10-18-v01', '');
`, []string{`list_models`}))

func TestListModel(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index ListModel",
			Debug: true,
			ReqFunc: func() *http.Request {
				listModelData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/list-models", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"123", "Online"},
		},
		{
			Name:  "Detail ListModel",
			Debug: true,
			ReqFunc: func() *http.Request {
				listModelData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/list-models").
					EventFunc(actions.DetailingDrawer).
					Query(presets.ParamID, "1_2024-10-18-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Online", "Unpublish", "123", "Detail Path", "List Path"},
		},
		{
			Name:  "Detail ListModel Draft",
			Debug: true,
			ReqFunc: func() *http.Request {
				listModelData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/list-models").
					EventFunc(actions.DetailingDrawer).
					Query(presets.ParamID, "2_2024-10-18-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Draft", "Publish", "456"},
			ExpectPortalUpdate0NotContains:     []string{"Detail Path", "List Path"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
