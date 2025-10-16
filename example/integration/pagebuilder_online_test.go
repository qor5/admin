package integration_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/oss"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
)

var pageBuilderOnlineData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'Japan', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}))

var editLastestVersion = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
(1, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'all_draft', 'draft-page-1', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(1, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, 'all_draft', 'draft-page-1', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(2, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'draft_after_online', 'draft-online-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(2, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, 'draft_after_online', 'draft-online-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(3, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, 'online_after_draft', 'draft-online-page-reverse', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(3, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'online_after_draft', 'draft-online-page-reverse', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(4, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'draft_after_offline', 'draft-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(4, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, 'draft_after_offline', 'draft-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'offline', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(5, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'all_offline', 'offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'offline', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(5, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, 'all_offline', 'offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'offline', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(6, '2024-05-21 01:54:45.280106 +00:00', '2024-05-23 03:54:57.983233 +00:00', null, 'online_after_offline', 'online-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(6, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'online_after_offline', 'online-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'offline', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan'),
(7, '2024-05-21 01:54:45.280106 +00:00', '2024-05-23 03:54:57.983233 +00:00', null, 'offline_after_online', 'online-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'offline', '/index.html', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan'),
(7, '2024-05-22 02:54:45.280106 +00:00', '2024-05-22 02:54:57.983233 +00:00', null, 'offline_after_online', 'online-offline-page', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '/index.html', null, null, null, null, '2024-05-22-v01', '2024-05-22-v01', '', 'Japan')
;`,
	[]string{"page_builder_pages"}))

type StorageWithError struct {
	oss.StorageInterface
	ErrGetURL error
}

func (s *StorageWithError) GetURL(ctx context.Context, path string) (string, error) {
	if s.ErrGetURL != nil {
		return "", s.ErrGetURL
	}
	return s.StorageInterface.GetURL(ctx, path)
}

func TestPageBuilderOnline(t *testing.T) {
	dbr, _ := TestDB.DB()

	t.Run("Check previewDevelopUrl(panic)", func(t *testing.T) {
		h, _ := admin.TestHandlerComplex(TestDB, nil, false, admin.WithStorageWrapper(func(si oss.StorageInterface) oss.StorageInterface {
			return &StorageWithError{StorageInterface: si, ErrGetURL: errors.New("get url error")}
		}))
		pageBuilderOnlineData.TruncatePut(dbr)
		require.Panics(t, func() {
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/pages/10_2024-05-21-v01_Japan", http.NoBody))
		})
	})

	h := admin.TestHandler(TestDB, nil)

	cases := []TestCase{
		{
			Name:  "Check previewDevelopUrl",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/10_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`<a href='example-publish.s3.ap-northeast-1.amazonaws.com/' target='_blank'>example-publish.s3.ap-northeast-1.amazonaws.com/</a>`},
		},
		{
			Name:  "PageBuilder Online Wrap EditContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.EditContainerEvent).
					Query("containerUri", "headers").
					Query("containerID", "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).dash(dash).reload().go()"},
		},
		{
			Name:  "PageBuilder Online Wrap UpdateContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.UpdateContainerEvent).
					Query("containerUri", "headers").
					Query("containerID", "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).dash(dash).reload().go()"},
		},
		{
			Name:  "PageBuilder Online Wrap AddContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"vars.presetsMessage = { show: true, message:", "plaid().vars(vars).locals(locals).form(form).dash(dash).reload().go()"},
		},
		{
			Name:  "PageBuilder Online Edit Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					EventFunc("section_save_Page").
					Query(presets.ParamID, "10_2024-05-21-v01_Japan").
					AddField("Title", "123").
					AddField("Slug", "123").
					AddField("CategoryID", "0").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`vars.presetsMessage = { show: true, message: "The resource can not be modified", color: "warning"}`},
		},
		{
			Name:  "PageBuilder Online Edit Seo",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					EventFunc("section_save_SEO").
					Query(presets.ParamID, "10_2024-05-21-v01_Japan").
					AddField("SEO.EnabledCustomize", "true").
					AddField("SEO.Title", "My seo title").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`vars.presetsMessage = { show: true, message: "The resource can not be modified", color: "warning"}`},
		},
		{
			Name:  "PageBuilder Online Edit Draft All Draft",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "all_draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/1_2024-05-22-v01_Japan").go()'>`, "Edit Last Draft"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/1_2024-05-21-v01_Japan").go()'>`},
		},
		{
			Name:  "Pages Listing After Online",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "draft_after_online").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/2_2024-05-22-v01_Japan").go()'>`, "Edit Last Draft"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/2_2024-05-21-v01_Japan").go()'>`},
		},
		{
			Name:  "Pages Listing  Online After Draft",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "online_after_draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/3_2024-05-21-v01_Japan").go()'>`, "Edit Last Draft"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/3_2024-05-22-v01_Japan").go()'>`},
		},
		{
			Name:  "Pages Listing  Draft After Offline",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "draft_after_offline").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/4_2024-05-22-v01_Japan").go()'>`, "Edit Last Draft"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-pencil"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/pages/4_2024-05-21-v01_Japan").go()'>`},
		},
		{
			Name:  "Pages Listing  All Offline",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "all_offline").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`<v-list-item :prepend-icon='"mdi-eye"' href='/page_builder/pages/preview?id=5_2024-05-22-v01_Japan' target='_blank'>`, "Preview"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-eye"' href='/page_builder/pages/preview?id=5_2024-05-21-v01_Japan'>`},
		},
		{
			Name:  "Pages Listing  Online After Offline",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "online_after_offline").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`index.html`, "Preview"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-eye"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/preview?id=6_2024-05-22-v01_Japan").go()'>`},
		},
		{
			Name:  "Pages Listing  Offline After Online",
			Debug: true,
			ReqFunc: func() *http.Request {
				editLastestVersion.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query("keyword", "offline_after_online").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`index.html`, "Preview"},
			ExpectPageBodyNotContains:     []string{`<v-list-item :prepend-icon='"mdi-eye"' @click='plaid().vars(vars).locals(locals).form(form).dash(dash).pushState(true).url("/preview?id=7_2024-05-22-v01_Japan").go()'>`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
