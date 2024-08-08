package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var publishData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.with_publish_products (id, created_at, updated_at, deleted_at, name, price, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Product', 200, 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01', '');


`, []string{"with_publish_products"}))

var emptyData = gofixtures.Data(gofixtures.Sql(``, []string{"with_publish_products"}))

func TestPublish(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PublishExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/with-publish-products", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Hello Product"},
		},
		{
			Name:  "Not Found Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/samples/publish/products", nil)
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
					t.Errorf("Expected text/html; charset=utf-8, got %v", w.Header().Get("Content-Type"))
				}
			},
		},
		{
			Name:  "Publish Model New should not have publish bar",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-publish-products?__execute_event__=presets_New").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Price"},
			ExpectPortalUpdate0NotContains:     []string{`"/with-publish-products-version-list-dialog"`},
		},
		{
			Name:  "Create should have first version",
			Debug: true,
			ReqFunc: func() *http.Request {
				emptyData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-publish-products?__execute_event__=presets_Update").
					AddField("Name", "123321").
					AddField("Price", "200").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var p WithPublishProduct
				TestDB.Find(&p)
				if len(p.Version.Version) == 0 {
					t.Errorf("version not updated for publish product %#+v", p)
				}
			},
		},
		{
			Name:  "Default Right Drawer Width should be 600",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-publish-products?__execute_event__=presets_Edit&id=1_2024-05-20-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`:width='"600"'`},
		},
		{
			Name:  "Detailing drawer control bar should be on top",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-publish-products?__execute_event__=presets_DetailingDrawer&id=1_2024-05-20-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"publish_EventDuplicateVersion", "Price"},
		},
		{
			Name:  "should allow to open version dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-publish-products-version-list-dialog?__execute_event__=presets_OpenListingDialog&f_select_id=1_2024-05-20-v01").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Versions List"},
		},
		{
			Name:  "view rename dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				publishData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/with-publish-products-version-list-dialog?__execute_event__=publish_eventRenameVersionDialog&id=1_2024-05-20-v01&overlay=dialog&version_name=2024-05-20-v01", nil)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Rename Version", "2024-05-20-v01", "Cancel", "OK"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
