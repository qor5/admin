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
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'online', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'International', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}))

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
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/pages/10_2024-05-21-v01_International", nil))
		})
	})

	h := admin.TestHandler(TestDB, nil)

	cases := []TestCase{
		{
			Name:  "Check previewDevelopUrl",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/10_2024-05-21-v01_International", nil)
			},
			ExpectPageBodyContainsInOrder: []string{`<a href='example-publish.s3.ap-northeast-1.amazonaws.com/'>example-publish.s3.ap-northeast-1.amazonaws.com/</a>`},
		},
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
		{
			Name:  "PageBuilder Online Edit Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderOnlineData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_International").
					EventFunc("section_save_Page").
					Query(presets.ParamID, "10_2024-05-21-v01_International").
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
					PageURL("/pages/10_2024-05-21-v01_International").
					EventFunc("section_save_SEO").
					Query(presets.ParamID, "10_2024-05-21-v01_International").
					AddField("SEO.EnabledCustomize", "true").
					AddField("SEO.Title", "My seo title").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`vars.presetsMessage = { show: true, message: "The resource can not be modified", color: "warning"}`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
