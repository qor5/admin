package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/pagebuilder/example/containers"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var pageBuilderContainerShareTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '12313', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'International');
SELECT setval('page_builder_pages_id_seq', 10, true);

INSERT INTO public.page_builder_containers (id,created_at, updated_at, deleted_at, page_id, page_version, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id,page_model_name) VALUES 
										   (9,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'BrandGrid', 10, 1, false, false, 'BrandGrid', 'International', 0,'pages'),
										   (10,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'ListContent', 10, 1, true, false, 'ListContent', 'International', 0,'pages'),
										   (11,'2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 10, '2024-05-21-v01', 'Header', 10, 2, true, false, 'Header', 'International', 0,'pages')  ;
SELECT setval('page_builder_containers_id_seq', 11, true);

INSERT INTO public.container_brand_grids (id, add_top_space, add_bottom_space, anchor_id, brands) VALUES (10, false, false, '', 'null');
SELECT setval('container_brand_grids_id_seq', 10, true);

INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color, link, link_text, link_display_option) VALUES (10, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop');
SELECT setval('container_list_content_id_seq', 10, true);

INSERT INTO public.container_headers (id, color) VALUES (10, 'black');
SELECT setval('container_headers_id_seq', 10, true);

`, []string{"page_builder_pages", "page_builder_containers", "container_list_content", "container_headers", "container_brand_grids"}))

func TestPageBuilderShareContainer(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Shared Containers",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerShareTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/shared_containers", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Header", "ListContent"},
		},

		{
			Name:  "Mark As Shared Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerShareTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.MarkAsSharedContainerEvent).
					Query("containerID", "9_International").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var con pagebuilder.Container
				TestDB.First(&con, 9)
				if con.ID != 9 || !con.Shared {
					t.Fatalf("Mark As Shared Container did not work %#+v", con)
					return
				}
			},
		},
		{
			Name:  "Shared Containers Editing",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerShareTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/shared_containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "10").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"ListContent"},
		},
		{
			Name:  "PageBuilder Editor Replicate  Shared Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerShareTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.ReplicateContainerEvent).
					Query("containerID", "10_International").
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
				if container.ID <= 11 || container.ModelID == 10 || container.ModelName != "ListContent" || container.Shared || container.DisplayOrder != 2 {
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
				return
			},
		},
		{
			Name:  "Page Builder Add Shared Container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderContainerShareTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/pages/10_2024-05-21-v01_International").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "ListContent").
					Query("sharedContainer", "true").
					Query("modelID", "10").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var (
					container pagebuilder.Container
					count     int64
				)
				TestDB.Order("id desc").First(&container)
				TestDB.Model(&containers.ListContent{}).Count(&count)
				if container.ModelName != "ListContent" {
					t.Fatalf("expected ListContent got %s", container.ModelName)
					return
				}
				if !container.Shared {
					t.Fatalf("expected Shared container got %t", container.Shared)
					return
				}
				if container.ModelID != 10 {
					t.Fatalf("expected 10 got %d", container.ModelID)
					return
				}
				if count != 1 {
					t.Fatalf("expected 1 got %d", count)
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
