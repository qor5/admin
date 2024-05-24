package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/note"
	"github.com/qor5/admin/v3/pagebuilder"
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
`, []string{"page_builder_pages", "page_builder_categories"}))

var pageBuilderContainerTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES 
										   (10,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'ListContent', 10, 1, false, false, 'ListContent', 'International', 0),
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, false, false, 'Header', 'International', 0)  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color, link, link_text, link_display_option) VALUES (10, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop');
SELECT setval('container_list_content_id_seq', 10, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);


`, []string{"page_builder_pages", "page_builder_containers", "container_list_content", "container_headers"}))

func TestPageBuilder(t *testing.T) {
	h := admin.TestHandler(TestDB)
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
			Name:  "Page Builder Detail Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_International", nil)
			},
			ExpectPageBodyContainsInOrder: []string{
				`.eventFunc("createNoteDialogEvent").query("overlay", "dialog").query("id", "1_2024-05-18-v01_International").url("/pages")`,
			},
		},

		{
			Name:  "Page Builder Editor Show Add Notes Dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages?__execute_event__=createNoteDialogEvent&id=1_2024-05-18-v01_International&overlay=dialog").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`eventFunc("createNoteEvent")`},
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
					t.Errorf("wrong locale code, expected International, got %#+v", page)
				}
			},
		},

		{
			Name:  "Page Builder Editor Add a Note",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages?__execute_event__=createNoteEvent&overlay=dialog&resource_id=1_2024-05-18-v01_International&resource_type=Pages").
					AddField("Content", "Hello").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				n := note.QorNote{}
				TestDB.Find(&n)
				if n.Content != "Hello" {
					t.Error("Note not created", n)
				}
			},
		},
		{
			Name:  "Page Builder Editor Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/1_2024-05-18-v01_International?__execute_event__=publish_EventDuplicateVersion").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var pages []*pagebuilder.Page
				TestDB.Order("id DESC, version DESC").Find(&pages)
				if len(pages) != 2 {
					t.Fatal("Page not duplicated", pages)
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
					PageURL("/page_builder/list-contents?__execute_event__=presets_AutoSave_Edit&id=10&overlay=content").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0NotContains:     []string{"Update"},
			ExpectPortalUpdate0ContainsInOrder: []string{"@change-debounced"},
		},
		// TODO run with under under containerDataID will to be headers_2
		// {
		//	Name:  "Page Builder add container",
		//	Debug: true,
		//	ReqFunc: func() *http.Request {
		//		pageBuilderContainerTestData.TruncatePut(dbr)
		//		req := NewMultipartBuilder().
		//			PageURL("/page_builder/editors/10_2024-05-21-v01_International?__execute_event__=page_builder_AddContainerEvent&modelName=Header").
		//			BuildEventFuncRequest()
		//
		//		return req
		//	},
		//	ExpectRunScriptContainsInOrder: []string{"page_builder_ReloadRenderPageOrTemplateEvent", "containerDataID", "headers_1", "/page_builder/headers", "presets_AutoSave_Edit"},
		// },
		{
			Name:  "Page Builder add container under",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/editors/10_2024-05-21-v01_International?__execute_event__=page_builder_AddContainerEvent&containerID=10_International&modelName=Header").
					BuildEventFuncRequest()

				return req
			},
			ExpectRunScriptContainsInOrder: []string{"page_builder_ReloadRenderPageOrTemplateEvent", "containerDataID", "headers_1", "/page_builder/headers", "presets_AutoSave_Edit"},
		},
		{
			Name:  "Page Builder delete container dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/editors/10_2024-05-21-v01_International?__execute_event__=page_builder_DeleteContainerConfirmationEvent&containerID=10_International&containerName=Header").
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
					PageURL("/page_builder/editors/10_2024-05-21-v01_International?__execute_event__=page_builder_DeleteContainerEvent&containerID=10_International").
					BuildEventFuncRequest()

				return req
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
