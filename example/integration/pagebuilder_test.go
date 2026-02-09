package integration_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qor5/web/v3"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/perm"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	m.Run()
}

var pageBuilderData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_123', '/12', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_123', '/12', '', 'China');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (2, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_456', '/45', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_34', '/34', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (3, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'category_34', '/34', '', 'China');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (4, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'test_change_category', '/category2', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (5, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'compare_change_category', '/', '', 'Japan');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (6, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'no_page', '/no_page', '', 'Japan');
SELECT setval('page_builder_categories_id_seq', 1, true);
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (2, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'category', '/', 4, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (3, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'category', '/category', 5, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'Japan', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
SELECT setval('page_builder_pages_id_seq', 3, true);
INSERT INTO public.page_builder_templates (id, created_at, updated_at, deleted_at, name, description, locale_code) VALUES (1, '2024-07-22 01:41:13.206348 +00:00', '2024-07-22 01:41:13.206348 +00:00', null, '123', '456', '');
`, []string{"page_builder_pages", "page_builder_categories", "page_builder_templates"}))

var pageBuilderContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '/12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (10,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'ListContent', 10, 1, false, false, 'ListContent', 'Japan', 0,'pages'),
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'Japan', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color, link, link_text, link_display_option) VALUES (10, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop');
SELECT setval('container_list_content_id_seq', 10, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_list_content", "container_headers"}))

var pageBuilderDemoContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);
INSERT INTO public.container_in_numbers (id, add_top_space, add_bottom_space, anchor_id, heading, items) VALUES (1, false, false, 'test1', '', 'null');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, locale_code) VALUES (1, '2024-06-25 02:21:41.014915 +00:00', '2024-06-25 02:21:41.014915 +00:00', null, 'InNumbers', 1, 'Japan');
INSERT INTO public.container_headings (id, add_top_space, add_bottom_space, anchor_id, heading, font_color, background_color, link, link_text, link_display_option, text) VALUES (1, false, false, '', '', '', '', '', '', '', '');
`, []string{"page_builder_pages", "page_builder_containers", "container_in_numbers", "page_builder_demo_containers", "container_headings"}))

var pageBuilderHiddenContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '/12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'Japan');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, true, 'Header', 'Japan', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_headers"}))

func TestPageBuilder(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"ID", "Title", "Live", "12312"},
		},
		{
			Name:  "New Page Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/pages").
					EventFunc(actions.New).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Template", "Title", `form["CategoryID"]`, `:clearable='true'`, `prefix='/'`},
		},

		{
			Name:  "Page Builder Detail Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				`Page`, "Category", `SEO`, `Activity`,
			},
		},
		{
			Name:  "Page Builder Detail Page with invalid slug",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_Japan_invalid", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				`Sorry, the requested page cannot be found. Please check the URL.`,
			},
		},
		{
			Name:  "Page Builder Detail Page with invalid id in slug",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/a_2024-05-18-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				`Sorry, the requested page cannot be found. Please check the URL.`,
			},
		},
		{
			Name:  "Page Builder Detail Editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					Query("containerDataID", "list-content_10_10Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{
				`eventFunc("page_builder_EditContainerEvent").mergeQuery(true).query("containerDataID", vars.containerDataID)`,
			},
		},
		{
			Name:  "Page Builder Detail editor(not found)",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("/page_builder/pages/10_2024-05-21-v01_JapanNotFound").
					Query("containerDataID", "list-content_10_10Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{
				`Sorry, the requested page cannot be found. Please check the URL.`,
			},
		},
		{
			Name:  "Page New Title Empty",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(actions.Update).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Title"},
		},
		{
			Name:  "Page Title Section Validate Empty",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(presets.ParamID, "1_2024-05-18-v01_Japan").
					AddField("Title", "").
					EventFunc("section_validate_Page").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Invalid Title"},
		},
		{
			Name:  "Page Title Section Empty",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					Query(presets.ParamID, "1_2024-05-18-v01_Japan").
					EventFunc("section_save_Page").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Title"},
		},
		{
			Name:  "Category New Title Empty",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Name"},
		},
		{
			Name:  "Category Validate Page Same Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "4_Japan").
					AddField("Name", "InvalidPath").
					AddField("LocaleCode", "Japan").
					AddField("Path", "category2").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully Updated"},
		},
		{
			Name:  "Category Validate No Page ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "6_Japan").
					AddField("Name", "InvalidPath").
					AddField("LocaleCode", "Japan").
					AddField("Path", "no_page2").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully Updated"},
		},
		{
			Name:  "Category Validate Page Invalid Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "4_Japan").
					AddField("Name", "InvalidPath").
					AddField("LocaleCode", "Japan").
					AddField("Path", "category").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"This path would cause URL conflicts with existing pages"},
		},
		{
			Name:  "Page Builder Editor Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/10_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Add Container", "Select an element and change the setting here."},
		},
		{
			Name:  "Page Builder Editor Page without perm.Update",
			Debug: true,
			HandlerMaker: func() http.Handler {
				mux, c := admin.TestHandlerComplex(TestDB, nil, false)
				c.GetPresetsBuilder().Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermUpdate).On("*:presets:pages:*"),
					),
				)
				return mux
			},
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/10_2024-05-21-v01_Japan", http.NoBody)
			},
			ExpectPageBodyNotContains: []string{"Add Container", "Select an element and change the setting here."},
		},
		{
			Name:  "Add a new page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(actions.Update).
					AddField("Title", "Hello 4").
					AddField("CategoryID", "1").
					AddField("Slug", "hello4").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var page pagebuilder.Page
				TestDB.First(&page, "slug = ?", "/hello4")
				if page.LocaleCode != "Japan" {
					t.Fatalf("wrong locale code, expected Japan, got %#+v", page.LocaleCode)
					return
				}
			},
		},
		{
			Name:  "Page Builder Editor Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					EventFunc(publish.EventDuplicateVersion).
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var pages []*pagebuilder.Page
				TestDB.Order("id DESC, version DESC").Find(&pages)
				if len(pages) != 2 {
					t.Fatalf("Page not duplicated %v", pages)
					return
				}
				if pages[0].Slug != pages[1].Slug {
					t.Fatalf("Page not duplicated %v", pages)
					return
				}
				var cons []*pagebuilder.Container
				TestDB.Find(&cons, "page_id = ? AND page_version = ?", pages[0].ID,
					pages[0].Version.Version)
				if len(cons) == 0 {
					t.Error("Container not duplicated", cons)
				}
			},
		},

		{
			Name:  "Page Builder Detail Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					EventFunc(publish.EventDuplicateVersion).
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var pages []*pagebuilder.Page
				TestDB.Order("id DESC, version DESC").Find(&pages)
				if len(pages) != 2 {
					t.Fatal("Page not duplicated", pages)
					return
				}
				if pages[0].Slug != pages[1].Slug {
					t.Fatalf("Page not duplicated %v", pages)
					return
				}
				var cons []*pagebuilder.Container
				TestDB.Find(&cons, "page_id = ? AND page_version = ?", pages[0].ID,
					pages[0].Version.Version)
				if len(cons) == 0 {
					t.Error("Container not duplicated", cons)
				}
			},
		},
		{
			Name:  "Page Builder ListContent add row",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/list-contents").
					EventFunc(actions.Edit).
					Query(presets.ParamOverlay, actions.Content).
					Query("portal_name", "pageBuilderRightContentPortal").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0NotContains: []string{">Update<"},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				if er.UpdatePortals[0].Name != "pageBuilderRightContentPortal" {
					t.Fatalf("error portalName %v", er.UpdatePortals[0].Name)
				}
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"@change-debounced"},
		},

		{
			Name:  "Page Builder add container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 3 {
					t.Error("containers not add", cons)
				}
				if cons[0].ModelName != "ListContent" || cons[1].ModelName != "Header" || cons[2].ModelName != "BrandGrid" {
					t.Error("containers not add under", cons)
				}
			},
		},
		{
			Name:  "Page Builder add container under",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("containerID", "10_Japan").
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 3 {
					t.Fatalf("cons not add %#+v", cons)
					return
				}
				if cons[0].ModelName != "ListContent" || cons[1].ModelName != "BrandGrid" || cons[2].ModelName != "Header" {
					t.Fatalf("containers not add under  %#+v", cons)
					return
				}
			},
		},
		{
			Name:  "Page Builder delete container dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.DeleteContainerConfirmationEvent).
					Query("containerID", "10_Japan").
					Query("containerName", "Header").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`page_builder_DeleteContainerEvent`, `query("containerID", "10_Japan")`},
		},
		{
			Name:  "Page Builder delete container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.DeleteContainerEvent).
					Query("containerID", "10_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 1 {
					t.Fatalf("containers not delete %#+v", cons)
					return
				}
			},
		},
		{
			Name:  "Page Builder toggle hidden ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ToggleContainerVisibilityEvent).
					Query("containerID", "10_Japan").
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container, 10).Error; err != nil {
					t.Error(err)
					return
				}
				if !container.Hidden {
					t.Fatalf("containers not hidden %#+v", container)
					return
				}
			},
		},
		{
			Name:  "Page Builder toggle visibility ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderHiddenContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ToggleContainerVisibilityEvent).
					Query("containerID", "11_Japan").
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container, 11).Error; err != nil {
					t.Error(err)
					return
				}
				if container.Hidden {
					t.Fatalf("containers is hidden %#+v", container)
					return
				}
			},
		},
		{
			Name:  "Page Builder Rename",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.RenameContainerEvent).
					Query("containerID", "10_Japan").
					Query("status", "draft").
					AddField("DisplayName", "hello").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container, 10).Error; err != nil {
					t.Error(err)
					return
				}
				if container.DisplayName != "hello" {
					t.Fatalf("containers not rename %#+v", container)
					return
				}
			},
		},
		{
			Name:  "Page Builder move down",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					Query("containerID", "10_Japan").
					Query("moveDirection", "down").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 2 {
					t.Error("containers not add", cons)
					return
				}
				if cons[0].ModelName != "Header" || cons[1].ModelName != "ListContent" {
					t.Error("container not move down", cons)
					return
				}
			},
		},
		{
			Name:  "Page Builder move up",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					Query("containerID", "11_Japan").
					Query("moveDirection", "up").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 2 {
					t.Error("containers not add", cons)
					return
				}
				if cons[0].ModelName != "Header" || cons[1].ModelName != "ListContent" {
					t.Error("container not move down", cons)
					return
				}
			},
		},
		{
			Name:  "Page Builder sorted move",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.MoveContainerEvent).
					Query("status", "draft").
					AddField("moveResult", `[{"index":0,"container_id":"11","locale":"Japan"},{"index":1,"container_id":"10","locale":"Japan"}]`).
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&cons)
				if len(cons) != 2 {
					t.Error("cons not add", cons)
					return
				}
				if cons[0].ModelName != "Header" || cons[1].ModelName != "ListContent" {
					t.Error("container not sort move", cons)
					return
				}
			},
		},
		{
			Name:  "Page Builder show sorted container left drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ShowSortedContainerDrawerEvent).
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"ListContent", "Header"},
		},
		{
			Name:  "Page Builder Preview  With SEO Title",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/preview").
					Query(presets.ParamID, "10_2024-05-21-v01_Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"1234567default", "list-contents", "headers"},
		},
		{
			Name:  "Demo Container List",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo_containers").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"All", "Filled", "Not Filled"},
		},
		{
			Name:  "Add New Demo Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					AddField("modelName", "InNumbers").
					AddField("id", "1").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []*pagebuilder.Container
				TestDB.Order("id desc").Find(&cons)
				if len(cons) != 1 {
					t.Fatalf("add container failed, expected 1 cons, got %d", len(cons))
					return
				}
				if cons[0].ModelName != "InNumbers" {
					t.Fatalf("add container failed, expected InNumbers, got %s", cons[0].ModelName)
					return
				}
				var mos []*containers.InNumbers
				TestDB.Order("id desc").Find(&mos)
				if len(mos) != 2 {
					t.Fatalf("add demo container model failed, expected 2 mos, got %d", len(mos))
					return
				}
				if mos[0].AnchorID != "test1" {
					t.Fatalf("add demo container model failed, expected test1, got %s", mos[0].AnchorID)
					return
				}
			},
		},
		{
			Name:  "InNumber Validate Items ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/in-numbers").
					EventFunc(actions.Validate).
					Query(presets.ParamID, "1").
					AddField("Items[0].Heading", "").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"Heading can`t Empty"},
		},
		{
			Name:  "Edit Demo Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/in-numbers").
					EventFunc(actions.Update).
					Query("demoContainer", "true").
					Query(presets.ParamID, "1").
					Query("status", "draft").
					AddField("AnchorID", "test_in_numbers").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var (
					mos  []*containers.InNumbers
					cons []*pagebuilder.DemoContainer
				)
				TestDB.Where("model_name = ? and locale_code = ? ", "InNumbers", "Japan").Find(&cons)
				if len(cons) != 1 {
					t.Fatalf("Expected 1  Demo Containers, got %v", len(cons))
					return
				}
				if !cons[0].Filled {
					t.Fatalf("Expected  Demo Container to be filled ")
					return
				}
				TestDB.Find(&mos)
				if len(mos) != 1 {
					t.Fatalf("Expected 1 model contianer, got %v", len(mos))
					return
				}
				if mos[0].AnchorID != "test_in_numbers" {
					t.Fatalf("Expected AnchorID 'test_in_numbers', got %v", mos[0].AnchorID)
					return
				}
			},
		},
		{
			Name:  "Page Detail Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					EventFunc("section_save_Page").
					Query(presets.ParamID, "10_2024-05-21-v01_Japan").
					AddField("Title", "123").
					AddField("Slug", "123").
					AddField("CategoryID", "0").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var cons []*pagebuilder.Page
				TestDB.Find(&cons)
				if len(cons) != 1 {
					t.Fatalf("Expected 1  Pages, got %v", len(cons))
					return
				}
				if cons[0].Title != "123" {
					t.Fatalf("Expected Page Title, got %s", cons[0].Title)
					return
				}
				if cons[0].Slug != "/123" {
					t.Fatalf("Expected Page Slug, got %s", cons[0].Slug)
					return
				}
			},
		},
		{
			Name:  "Page Builder preview demo container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ContainerPreviewEvent).
					Query("modelName", "InNumbers").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"test1"},
		},
		{
			Name:  "Template detail ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_templates/1").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"123", "456"},
		},
		{
			Name:  "Page Detail Editing No Category",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_Japan").
					Query(web.EventFuncIDName, "section_edit_Page").
					Query("id", "10_2024-05-21-v01_Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`"Title":"1234567"`, `"CategoryID":""`, `:clearable='true'`, `"Slug":"12313"`},
		},
		{
			Name:  "Page Detail Editing Has Category",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/1_2024-05-18-v01_Japan").
					Query(web.EventFuncIDName, "section_edit_Page").
					Query("id", "1_2024-05-18-v01_Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`"Title":"12312"`, `"CategoryID":1`, `:clearable='true'`, `"Slug":"123"`},
		},

		{
			Name:  "Page Builder add container Wrap SaveFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var (
					container pagebuilder.Container
					bd        containers.BrandGrid
				)
				TestDB.Order("id desc").First(&container)
				TestDB.Order("id desc").First(&bd)
				if container.ModelName != "BrandGrid" {
					t.Fatalf("containers not add")
					return
				}
				if bd.AnchorID == "" {
					t.Fatalf("wrap container creating error")
					return
				}
			},
		},
		{
			Name:  "Page Category Search",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					Query("keyword", "123").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"category_123"},
			ExpectPageBodyNotContains:     []string{"category_456"},
		},
		{
			Name:  "Page Category Validate Empty Name",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1_Japan").
					Query("Name", "").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Name"},
		},
		{
			Name:  "Page Category Validate Empty Name",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1_Japan").
					AddField("Name", "").
					AddField("LocaleCode", "Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Name"},
		},
		{
			Name:  "Page Category Validate InvalidPath",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1_Japan").
					AddField("Name", "category_123").
					AddField("Path", "/**)))=--").
					AddField("LocaleCode", "Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Invalid Path"},
		},
		{
			Name:  "Page Category Validate Existing Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					AddField("Name", "category_123").
					AddField("Path", "45").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Existing Path"},
		},
		{
			Name:  "Page Category Save Validate By ID Existing Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1_Japan").
					AddField("Name", "category_123").
					AddField("Path", "45").
					AddField("Description", "").
					AddField("LocaleCode", "Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Existing Path"},
		},
		{
			Name:  "Page Category Validate By ID Existing Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Validate).
					Query(presets.ParamID, "1_Japan").
					AddField("LocaleCode", "Japan").
					AddField("Path", "45").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Existing Path"},
		},
		{
			Name:  "Page Category Validate Event Existing Path",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.Validate).
					AddField("Path", "45").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Existing Path"},
		},
		{
			Name:  "Page Category Delete Related Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.DoDelete).
					Query(presets.ParamID, "1_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				TestDB.Model(&pagebuilder.Category{}).Where("id=1 and locale_code='Japan'").Count(&count)
				if count != 1 {
					t.Fatalf("category is Delete ")
					return
				}
			},
		},
		{
			Name:  "Page Category Delete no Related Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.DoDelete).
					Query(presets.ParamID, "2_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				TestDB.Model(&pagebuilder.Category{}).Where("id=2 and locale_code='Japan'").Count(&count)
				if count != 0 {
					t.Fatalf("category is  Not Deleted count: %d ", count)
					return
				}
			},
		},
		{
			Name:  "Demo Containers Listing",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/demo_containers").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{`=== "PageTitle"`},
		},
		{
			Name:  "Container Header Editing",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`black`},
		},
		{
			Name:  "Container Header Update",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headers").
					EventFunc(actions.Update).
					Query(presets.ParamID, "10").
					AddField("Color", "white").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				header := containers.WebHeader{}
				TestDB.First(&header, 10)
				if header.Color != "white" {
					t.Fatalf("container has not updated color")
					return
				}
			},
		},
		{
			Name:  "Container Heading Update Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headings").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("FontColor", "blue").
					AddField("AddTopSpace", "true").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`"AddTopSpace":true`, "blue", "LinkText 不能为空"},
		},
		{
			Name:  "Container Heading Update Reload Editing",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headings").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("LinkText", "ReplaceLinkText").
					AddField("Heading", "true").
					AddField("AddTopSpace", "true").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				heading := containers.Heading{}
				TestDB.First(&heading, 1)
				if heading.LinkText != "ReplaceLinkText" {
					t.Fatalf("container has not updated")
					return
				}
				if !heading.AddTopSpace {
					t.Fatalf("container has not updated")
				}
			},
		},

		{
			Name:  "PageBuilder Editor Replicate Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ReplicateContainerEvent).
					Query("containerID", "10_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var (
					container     pagebuilder.Container
					nextContainer pagebuilder.Container
					m             containers.ListContent
				)
				TestDB.Order("id desc").First(&container)
				if container.ID <= 11 || container.ModelID == 10 || container.ModelName != "ListContent" || container.DisplayOrder != 2 {
					t.Fatalf("Replicate Container Faield %#+v", container)
					return
				}
				TestDB.Order("id desc").First(&m, container.ModelID)
				if m.Link != "ijuhuheweq" {
					t.Fatalf("Replicate Container Model Faield %#+v", m)
					return
				}
				TestDB.First(&nextContainer, 11)
				if nextContainer.DisplayOrder != 3 {
					t.Fatalf("Replicate Container Faield %#+v", nextContainer)
					return
				}
			},
		},
		{
			Name:  "PageBuilder Wrap EditContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.EditContainerEvent).
					Query("containerDataID", "headers_10__10_Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`url("/headers").eventFunc("presets_Edit").query("id", "10").query("portal_name", "pageBuilderRightContentPortal").query("overlay", "content")`},
		},
		{
			Name:  "PageBuilder Wrap UpdateContainerEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.UpdateContainerEvent).
					Query("containerUri", "/page_builder/headers").
					Query("containerID", "10").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`plaid().vars(vars).locals(locals).form(form).dash(dash).url("/page_builder/headers").eventFunc("presets_Update").query("id", "10").query("portal_name", "pageBuilderRightContentPortal").query("overlay", "content").go()`},
		},
		{
			Name:  "Container Heading  Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headings").
					EventFunc(actions.Validate).
					Query(presets.ParamID, "1").
					AddField("LinkText", "").
					AddField("FontColor", "blue").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"LinkText 不能为空"},
		},
		{
			Name:  "Container In Number Delete First Row Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/in-numbers").
					EventFunc(actions.Update).
					AddField("__Deleted.Items", "0").
					AddField("Items[1].Heading", "").
					AddField("Items[1].Text", "blue").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Items[0].Heading"},
			ExpectPortalUpdate0NotContains:     []string{"Items[1].Heading"},
		},
		{
			Name:  "Category DoLocalize",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(l10n.DoLocalize).
					Query(presets.ParamID, "2_Japan").
					AddField("localize_from", "Japan").
					AddField("localize_to", "China").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully Localized"},
		},
		{
			Name:  "Category Delete Confirmation",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.DeleteConfirmation).
					Query(presets.ParamID, "1_Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"this will remove all the records in all localized languages"},
		},
		{
			Name:  "Category Delete",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.DoDelete).
					Query(presets.ParamID, "3_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var count int64
				TestDB.Model(pagebuilder.Category{}).Where("id=3").Count(&count)
				if count != 0 {
					t.Fatalf("Category Delete Failed %v", count)
				}
			},
		},
		{
			Name:  "PageBuilder Empty Edit",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.EditContainerEvent).
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Select an element and change the setting here`},
		},
		{
			Name:  "PageBuilder Reload Body",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ReloadRenderPageOrTemplateBodyEvent).
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{`PageBuilderNotifIframeBodyUpdatedPages`, "id='app'"},
		},
		{
			Name:  "PageBuilder Reload",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ReloadRenderPageOrTemplateEvent).
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"vx-scroll-iframe"},
		},
		{
			Name:  "Page DoLocalize",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(l10n.DoLocalize).
					Query(presets.ParamID, "10_2024-05-21-v01_Japan").
					AddField("localize_from", "Japan").
					AddField("localize_to", "China").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var (
					page  pagebuilder.Page
					now   = time.Now()
					count int64
					err   error
				)
				if err = utils.PrimarySluggerWhere(TestDB, &page, fmt.Sprintf("10_%v_China", page.GetNextVersion(&now))).First(&page).Error; err != nil {
					t.Fatalf("Copy Page Failed %v", err)
					return
				}
				if err = TestDB.Model(&pagebuilder.Container{}).Where("page_id=10 and locale_code='China'").Count(&count).Error; err != nil {
					t.Fatalf("Find Containers %v", err)
					return
				}
				if count == 0 {
					t.Fatalf("Containers not copied")
					return
				}
			},
		},
		{
			Name:  "Page Builder ReloadAddContainersListEvent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ReloadAddContainersListEvent).
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Header", "PageTitle", "Shared"},
		},
		{
			Name:  "Page Builder show hidden container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderHiddenContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_Japan").
					EventFunc(pagebuilder.ShowSortedContainerDrawerEvent).
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Header", "mdi-eye-off"},
		},
		{
			Name:  "Page Builder Edit Same Form Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderHiddenContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/headers").
					EventFunc(actions.Update).
					Query(presets.ParamID, "10").
					AddField("Color", "Black").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully Updated"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
