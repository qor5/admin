package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '123', '/12', '', 'International');
SELECT setval('page_builder_categories_id_seq', 1, true);
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'International', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
SELECT setval('page_builder_pages_id_seq', 1, true);
INSERT INTO public.page_builder_templates (id, created_at, updated_at, deleted_at, name, description, locale_code) VALUES (1, '2024-07-22 01:41:13.206348 +00:00', '2024-07-22 01:41:13.206348 +00:00', null, '123', '456', '');
`, []string{"page_builder_pages", "page_builder_categories", "page_builder_templates"}))

var pageBuilderContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (10,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'ListContent', 10, 1, false, false, 'ListContent', 'International', 0,'pages'),
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'International', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color, link, link_text, link_display_option) VALUES (10, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop');
SELECT setval('container_list_content_id_seq', 10, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_list_content", "container_headers"}))

var pageBuilderDemoContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);
INSERT INTO public.container_in_numbers (id, add_top_space, add_bottom_space, anchor_id, heading, items) VALUES (1, false, false, 'test1', '', 'null');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, locale_code) VALUES (1, '2024-06-25 02:21:41.014915 +00:00', '2024-06-25 02:21:41.014915 +00:00', null, 'InNumbers', 1, 'International');
`, []string{"page_builder_pages", "page_builder_containers", "container_in_numbers", "page_builder_demo_containers"}))

func TestPageBuilder(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"12312"},
		},
		{
			Name:  "New Page Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages?__execute_event__=presets_New", nil)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Template", "Title", `form["CategoryID"]`, `prefix='/'`},
		},

		{
			Name:  "Page Builder Detail Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_International", nil)
			},
			ExpectPageBodyContainsInOrder: []string{
				`Page`, `SEO`, `Activity`,
			},
		},
		{
			Name:  "Page Builder Detail editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages-editors/10_2024-05-21-v01_International?containerDataID=list-content_10", nil)
			},
			ExpectPageBodyContainsInOrder: []string{
				`presets_Edit`,
			},
		},
		{
			Name:  "Page New Title Empty",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages?__execute_event__=presets_Update").
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
					PageURL("/categories?__execute_event__=presets_Update").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name is required"},
		},
		{
			Name:  "Add a new page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages?__execute_event__=presets_Update").
					AddField("Title", "Hello 4").
					AddField("CategoryID", "1").
					AddField("Slug", "hello4").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var page pagebuilder.Page
				TestDB.First(&page, "slug = ?", "/hello4")
				if page.LocaleCode != "International" {
					t.Errorf("wrong locale code, expected International, got %#+v", page.LocaleCode)
				}
			},
		},
		{
			Name:  "Page Builder Editor Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=publish_EventDuplicateVersion").
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
				var containers []*pagebuilder.Container
				TestDB.Find(&containers, "page_id = ? AND page_version = ?", pages[0].ID,
					pages[0].Version.Version)
				if len(containers) == 0 {
					t.Error("Container not duplicated", containers)
				}
			},
		},

		{
			Name:  "Page Builder Detail Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/10_2024-05-21-v01_International?__execute_event__=publish_EventDuplicateVersion").
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
				var containers []*pagebuilder.Container
				TestDB.Find(&containers, "page_id = ? AND page_version = ?", pages[0].ID,
					pages[0].Version.Version)
				if len(containers) == 0 {
					t.Error("Container not duplicated", containers)
				}
			},
		},
		{
			Name:  "Page Builder ListContent add row",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/list-contents?__execute_event__=presets_Edit&id=10&overlay=content&portal_name=pageBuilderRightContentPortal").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0NotContains: []string{">Update<"},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				if er.UpdatePortals[0].Name != "pageBuilderRightContentPortal" {
					t.Errorf("error portalName %v", er.UpdatePortals[0].Name)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_AddContainerEvent&modelName=BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 3 {
					t.Error("containers not add", containers)
				}
				if containers[0].ModelName != "ListContent" || containers[1].ModelName != "Header" || containers[2].ModelName != "BrandGrid" {
					t.Error("containers not add under", containers)
				}
			},
		},
		{
			Name:  "Page Builder add container under",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_AddContainerEvent&containerID=10_International&modelName=BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 3 {
					t.Errorf("containers not add %#+v", containers)
					return
				}
				if containers[0].ModelName != "ListContent" || containers[1].ModelName != "BrandGrid" || containers[2].ModelName != "Header" {
					t.Errorf("containers not add under  %#+v", containers)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_DeleteContainerConfirmationEvent&containerID=10_International&containerName=Header").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`plaid().vars(vars).locals(locals).form(form).eventFunc("page_builder_DeleteContainerEvent").query("containerID", "10_International").go()`},
		},
		{
			Name:  "Page Builder delete container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_DeleteContainerEvent&containerID=10_International").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 1 {
					t.Errorf("containers not delete %#+v", containers)
					return
				}
			},
		},
		{
			Name:  "Page Builder toggle visibility ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_ToggleContainerVisibilityEvent&containerID=10_International&status=draft").
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
					t.Errorf("containers not hidden %#+v", container)
					return
				}
			},
		},
		{
			Name:  "Page Builder rename  ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_RenameContainerEvent&containerID=10_International&status=draft").
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
					t.Errorf("containers not rename %#+v", container)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_MoveUpDownContainerEvent&containerID=10_International&moveDirection=down").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 2 {
					t.Error("containers not add", containers)
					return
				}
				if containers[0].ModelName != "Header" || containers[1].ModelName != "ListContent" {
					t.Error("container not move down", containers)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_MoveUpDownContainerEvent&containerID=11_International&moveDirection=up").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 2 {
					t.Error("containers not add", containers)
					return
				}
				if containers[0].ModelName != "Header" || containers[1].ModelName != "ListContent" {
					t.Error("container not move down", containers)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_MoveContainerEvent&status=draft").
					AddField("moveResult", `[{"index":0,"container_id":"11","locale":"International"},{"index":1,"container_id":"10","locale":"International"}]`).
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 2 {
					t.Error("containers not add", containers)
					return
				}
				if containers[0].ModelName != "Header" || containers[1].ModelName != "ListContent" {
					t.Error("container not sort move", containers)
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
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_ShowSortedContainerDrawerEvent&status=draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"ListContent", "Header"},
		},
		{
			Name:  "Page Builder preview",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/preview?id=10_2024-05-21-v01_International").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"1234567", "list-contents", "headers"},
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
					PageURL("/page_builder/pages-editors/1_v1_International").
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
			Name:  "Edit Demo Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_builder/in-numbers?__execute_event__=presets_Update&id=1").
					AddField("AnchorID", "test_in_numbers").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var (
					mos  []*containers.InNumbers
					cons []*pagebuilder.DemoContainer
				)
				TestDB.Where("model_name = ? and locale_code = ? ", "InNumbers", "International").Find(&cons)
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
					PageURL("/pages/10_2024-05-21-v01_International?__execute_event__=presets_Detailing_Field_Save&detailField=Page&id=10_2024-05-21-v01_International").
					AddField("Page.Title", "123").
					AddField("Page.Slug", "/123").
					AddField("Page.CategoryID", "0").
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
			},
		},
		{
			Name:  "Page Builder preview demo container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages-editors/10_2024-05-21-v01_International?__execute_event__=page_builder_ContainerPreviewEvent&modelName=InNumbers").
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
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
