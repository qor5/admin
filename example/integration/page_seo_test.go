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
	"github.com/qor5/admin/v3/seo"
)

var pageSEOData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"Title":"{{Title}}默认","Description":"{{Title}}默认","OpenGraphImageFromMediaLibrary":{"ID":16,"Url":"/system/media_libraries/16/file.jpeg","VideoLink":"","FileName":"1.jpeg","Description":"","FileSizes":{"@qor_preview":10918,"default":15749,"og":149917,"original":15749,"twitter-large":143797,"twitter-small":81181},"Width":474,"Height":255},"OpenGraphMetadata":[{"Property":"title","Content":"name"},{"Property":"1","Content":"2"},{"Property":"默认","Content":"默认"}],"EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
										(11, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"EnabledCustomize":false}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);
INSERT INTO public.qor_seo_settings (name, setting, variables, created_at, updated_at, deleted_at, locale_code) VALUES ('Page', '{"Title":"{{Title}}默认","Description":"{{Title}}默认","OpenGraphImageFromMediaLibrary":{"ID":12,"Url":"/system/media_libraries/17/file.jpeg","VideoLink":"","FileName":"1.jpeg","Description":"","FileSizes":{"@qor_preview":10918,"default":15749,"original":15749},"Width":474,"Height":255},"OpenGraphMetadata":[{"Property":"title","Content":"name"},{"Property":"1","Content":"2"},{"Property":"默认","Content":"默认"}]}', '{}', '2024-08-23 08:43:32.954799 +00:00', '2024-12-03 09:42:06.217590 +00:00', null, 'Japan');
INSERT INTO public.qor_seo_settings (name, setting, variables, created_at, updated_at, deleted_at, locale_code) VALUES ('Global SEO', '{"Title":"{{Title}}默认","Description":"{{Title}}默认","OpenGraphImageFromMediaLibrary":{"ID":12,"Url":"/system/media_libraries/18/file.jpeg","VideoLink":"","FileName":"1.jpeg","Description":"","FileSizes":{"@qor_preview":10918,"default":15749,"original":15749},"Width":474,"Height":255},"OpenGraphMetadata":[{"Property":"title","Content":"name"},{"Property":"1","Content":"2"},{"Property":"默认","Content":"默认"}]}', '{}', '2024-08-23 08:43:32.954799 +00:00', '2024-12-03 09:42:06.217590 +00:00', null, 'Japan');

`, []string{"page_builder_pages", "qor_seo_settings"}))

var pageGlobalSEOData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"EnabledCustomize":false}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);
INSERT INTO public.qor_seo_settings (name, setting, variables, created_at, updated_at, deleted_at, locale_code) VALUES ('Page', '{"EnabledCustomize":false}', '{}', '2024-08-23 08:43:32.954799 +00:00', '2024-12-03 09:42:06.217590 +00:00', null, 'Japan');
INSERT INTO public.qor_seo_settings (name, setting, variables, created_at, updated_at, deleted_at, locale_code) VALUES ('Global SEO', '{"Title":"{{Title}}默认","Description":"{{Title}}默认","OpenGraphImageFromMediaLibrary":{"ID":12,"Url":"/system/media_libraries/18/file.jpeg","VideoLink":"","FileName":"1.jpeg","Description":"","FileSizes":{"@qor_preview":10918,"default":15749,"original":15749},"Width":474,"Height":255},"OpenGraphMetadata":[{"Property":"title","Content":"name"},{"Property":"1","Content":"2"},{"Property":"默认","Content":"默认"}]}', '{}', '2024-08-23 08:43:32.954799 +00:00', '2024-12-03 09:42:06.217590 +00:00', null, 'Japan');

`, []string{"page_builder_pages", "qor_seo_settings"}))

func TestPageSeoTemplate(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Preview Not Found Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageSEOData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/preview?id=", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"404"},
		},
		{
			Name:  "Page Use Customize SEO OpenGraphImage",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageSEOData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/preview?id=10_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"/system/media_libraries/16/file.og.jpeg"},
		},
		{
			Name:  "Page Use Default SEO OpenGraphImage",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageSEOData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/preview?id=11_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"/system/media_libraries/17/file.og.jpeg"},
		},
		{
			Name:  "Page Use Global SEO OpenGraphImage",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageGlobalSEOData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/preview?id=10_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"/system/media_libraries/18/file.og.jpeg"},
		},
		{
			Name:  "Global SEO Delete OpenGraphImage",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageGlobalSEOData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/qor-seo-settings").
					EventFunc(actions.Update).
					Query(presets.ParamID, "Global SEO_Japan").
					AddField("Setting.OpenGraphImageFromMediaLibrary.Values", "").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, _ *TestEventResponse) {
				var m seo.QorSEOSetting
				TestDB.Where("name='Global SEO' and locale_code='Japan'").First(&m)
				if m.Setting.OpenGraphImageFromMediaLibrary.URL() != "" {
					t.Fatalf("delete OpenGraphImage failed :%v", m.Setting.OpenGraphImageFromMediaLibrary.URL())
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
