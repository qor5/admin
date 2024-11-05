package integration_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/pagebuilder"
)

var pageBuilderOnlineData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'International', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}))

func TestPageBuilderOnline(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "PageBuilder Online Wrap EditContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.EditContainerEvent).
					Query("containerUri", "headers").
					Query("containerID", "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).reload().go()"},
		},
		{
			Name:  "PageBuilder Online Wrap UpdateContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.UpdateContainerEvent).
					Query("containerUri", "headers").
					Query("containerID", "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).reload().go()"},
		},
		{
			Name:  "PageBuilder Online Wrap AddContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).reload().go()"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
