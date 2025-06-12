package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

var pageBuilderLocaleData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_123', '/12', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_123', '/12', '', 'China');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (2, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_456', '/45', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_34', '/34', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_34', '/34', '', 'China');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (4, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_45', '/45', '', 'China');
SELECT setval('page_builder_categories_id_seq', 1, true);
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'test001_en', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '22024-05-18-v01_en', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'test001_cn', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01_cn', '', 'China', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'test002_en', '/123', 1, 'draft', '', null, null, null, null, '2024-05-19-v01', '2024-05-19-v01_en', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'test002_cn', '/123', 1, 'draft', '', null, null, null, null, '2024-05-19-v01', '2024-05-19-v01_cn', '', 'China', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (2, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'test002_cn', '/123', 4, 'draft', '', null, null, null, null, '2024-05-19-v01', '2024-05-19-v01_cn', '', 'China', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
SELECT setval('page_builder_pages_id_seq', 1, true);
`, []string{"page_builder_pages", "page_builder_categories"}))

func TestPageBuilderVersion(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Page Builder Detail Open Version Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderLocaleData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages-version-list-dialog").
					EventFunc("presets_OpenListingDialog").
					Query("f_select_id", "1_2024-05-18-v01_Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"2024-05-18-v01_en", "2024-05-19-v01_en"},
			ExpectPortalUpdate0NotContains:     []string{"2024-05-18-v01_cn", "2024-05-19-v01_cn"},
		},
		{
			Name:  "Page Builder Detail Open Version Dialog With China Locale",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderLocaleData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages-version-list-dialog").
					EventFunc("presets_OpenListingDialog").
					Query("f_select_id", "1_2024-05-18-v01_China").
					Query("locale", "China").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"2024-05-18-v01_cn", "2024-05-19-v01_cn"},
			ExpectPortalUpdate0NotContains:     []string{"2024-05-18-v01_en", "2024-05-19-v01_en"},
		},

		{
			Name:  "Page DoLocalize Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderLocaleData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					EventFunc(l10n.Localize).
					Query("locale", "China").
					Query(presets.ParamID, "1_2024-05-19-v01_China").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Localize", "China"},
			ExpectPortalUpdate0NotContains:     []string{"Japan"},
		},
		{
			Name:  "Page DoLocalize With Category",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderLocaleData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					EventFunc(l10n.DoLocalize).
					Query(presets.ParamID, "2_2024-05-19-v01_China").
					AddField("localize_from", "China").
					AddField("localize_to", "Japan").
					BuildEventFuncRequest()
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var (
					pages      []pagebuilder.Page
					categories []pagebuilder.Category
				)
				TestDB.Where("id=2").Find(&pages)
				TestDB.Where("id=4").Find(&categories)
				if len(pages) != 2 {
					t.Errorf("Expected 2 pages, got %d", len(pages))
					return
				}
				if len(categories) != 2 {
					t.Errorf("Expected 2 categories, got %d", len(categories))
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
