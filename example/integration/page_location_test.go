package integration_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
)

var pageLocationData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (2, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/345', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'China', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (3, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/456', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'China', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
SELECT setval('page_builder_pages_id_seq', 1, true);
`, []string{"page_builder_pages"}))

func TestPageLocation(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Default Location",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageLocationData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"/123"},
			ExpectPageBodyNotContains:     []string{"/345", "/456"},
		},
		{
			Name:  "China Location",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageLocationData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					Query("locale", "China").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"/456", "/345"},
			ExpectPageBodyNotContains:     []string{"/123"},
		},
		{
			Name:  "Japan Location",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageLocationData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					Query("locale", "Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"/123"},
			ExpectPageBodyNotContains:     []string{"/345", "/456"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
