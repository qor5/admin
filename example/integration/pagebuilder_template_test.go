package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var pageBuilderTemplateData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_templates (id, created_at, updated_at, deleted_at, name, description, locale_code)
VALUES (1, '2024-07-22 01:41:13.206348 +00:00', '2024-07-22 01:41:13.206348 +00:00', null, '123', '456',
        'Japan');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, model_name,
                                            model_id, display_order, shared, hidden, display_name, locale_code,
                                            localize_from_model_id, page_model_name)
VALUES (1, '2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 1, '', 'ListContent', 1, 1,
        false, false, 'ListContent', 'Japan', 0, 'templates'),
       (2, '2024-05-21 01:55:06.952248 +00:00', '2024-05-21 01:55:06.952248 +00:00', null, 1, '', 'Header', 2, 2, false,
        false, 'Header', 'Japan', 0, 'templates');
SELECT setval('page_builder_containers_id_seq', 1, true);
INSERT INTO public.container_headers (id, color)
VALUES (1, 'black'),(2, 'black');
SELECT setval('container_headers_id_seq', 1, true);
INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color,
                                           link, link_text, link_display_option)
VALUES (1, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop'),(2, true, true, '', null, 'grey', 'ijuhuheweq', '', 'desktop');
SELECT setval('container_list_content_id_seq', 1, true);
`, []string{"page_builder_templates", "page_builder_containers", "container_list_content", "container_headers"}))

func TestPageBuilderTemplate(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_templates", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Add Page Template", "123", "456"},
		},
		{
			Name:  "New Template Drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.New).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"New Template", "Name", "Description", "Create"},
		},
		{
			Name:  "Edit Template Drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Editing Template", "Name", "123", "Description", "456", "Update"},
		},
		{
			Name:  "Update Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("Name", "template_name").
					AddField("Description", "template_description").
					AddField("LocaleCode", "Japan").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Template
				TestDB.First(&m, 1)
				if m.Name != "template_name" {
					t.Fatalf("wrong Name, expected template_name, got %#+v", m.Name)
				}
				if m.Description != "template_description" {
					t.Fatalf("wrong Description, expected template_description, got %#+v", m.Description)
				}
				if m.LocaleCode != "Japan" {
					t.Fatalf("wrong LocaleCode, expected Japan, got %#+v", m.LocaleCode)
				}
			},
		},
		{
			Name:  "Add a new Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.Update).
					AddField("Name", "template_name").
					AddField("Description", "template_description").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Template
				TestDB.Order("id desc").First(&m)
				if m.Name != "template_name" {
					t.Fatalf("wrong Name, expected template_name, got %#+v", m.Name)
				}
				if m.Description != "template_description" {
					t.Fatalf("wrong Description, expected template_description, got %#+v", m.Description)
				}
				if m.LocaleCode != "Japan" {
					t.Fatalf("wrong LocaleCode, expected Japan, got %#+v", m.LocaleCode)
				}
			},
		},

		{
			Name:  "Delete Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.DoDelete).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Template
				TestDB.First(&m, 1)
				if m.ID == 1 {
					t.Fatalf("delete error got %#+v", m)
					return
				}
			},
		},
		{
			Name:  "New Page With Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(actions.Update).
					AddField("Title", "HelloPage").
					AddField("Slug", "hello").
					AddField(pagebuilder.ParamTemplateSelectedID, "1_Japan").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Page
				TestDB.First(&m)
				if m.ID == 0 {
					t.Fatalf("new page error")
					return
				}
				var containers []pagebuilder.Container
				TestDB.Where("page_model_name = 'pages'").Find(&containers)
				if len(containers) != 2 {
					t.Fatalf("wrong number of containers, expected 2, got %d", len(containers))
					return
				}
			},
		},
		{
			Name:  "New Page Without Template",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(actions.Update).
					AddField("Title", "HelloPage").
					AddField("Slug", "hello").
					AddField(pagebuilder.ParamTemplateSelectedID, "").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Page
				TestDB.First(&m)
				if m.ID == 0 {
					t.Fatalf("new page error")
					return
				}
				var containers []pagebuilder.Container
				TestDB.Where("page_model_name = 'pages'").Find(&containers)
				if len(containers) != 0 {
					t.Fatalf("wrong number of containers, expected 0, got %d", len(containers))
					return
				}
			},
		},
		{
			Name:  "Template Editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"ListContent", "Header"},
		},

		{
			Name:  "Template Editor with invalid slug",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan_invalid").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"Sorry, the requested page cannot be found. Please check the URL."},
		},

		{
			Name:  "Template Editor with invalid id in slug",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/a_Japan").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{"Sorry, the requested page cannot be found. Please check the URL."},
		},

		{
			Name:  "Template Editor add container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("modelName", "BrandGrid").
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
			Name:  "Template add container under",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.AddContainerEvent).
					Query("containerID", "1_Japan").
					Query("modelName", "BrandGrid").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 3 {
					t.Fatalf("containers not add %#+v", containers)
					return
				}
				if containers[0].ModelName != "ListContent" || containers[1].ModelName != "BrandGrid" || containers[2].ModelName != "Header" {
					t.Fatalf("containers not add under  %#+v", containers)
					return
				}
			},
		},
		{
			Name:  "Template delete container dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.DeleteContainerConfirmationEvent).
					Query("containerID", "2_Japan").
					Query("containerName", "Header").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`page_builder_DeleteContainerEvent`, `query("containerID", "2_Japan")`},
		},
		{
			Name:  "Template delete container ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.DeleteContainerEvent).
					Query("containerID", "2_Japan").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var containers []pagebuilder.Container
				TestDB.Order("display_order asc").Find(&containers)
				if len(containers) != 1 {
					t.Fatalf("containers not delete %#+v", containers)
					return
				}
			},
		},
		{
			Name:  "Template toggle visibility ",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.ToggleContainerVisibilityEvent).
					Query("containerID", "1_Japan").
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container, 1).Error; err != nil {
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
			Name:  "Template Editor Rename Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.RenameContainerEvent).
					Query("containerID", "1_Japan").
					Query("status", "draft").
					AddField("DisplayName", "hello").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container, 1).Error; err != nil {
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
			Name:  "Template move down",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					Query("containerID", "1_Japan").
					Query("moveDirection", "down").
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
			Name:  "Page Builder move up",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.MoveUpDownContainerEvent).
					Query("containerID", "2_Japan").
					Query("moveDirection", "up").
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
			Name:  "Template sorted move",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.MoveContainerEvent).
					Query("status", "draft").
					AddField("moveResult", `[{"index":0,"container_id":"2","locale":"Japan"},{"index":1,"container_id":"1","locale":"Japan"}]`).
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
			Name:  "Template show sorted container left drawer",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/1_Japan").
					EventFunc(pagebuilder.ShowSortedContainerDrawerEvent).
					Query("status", "draft").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"ListContent", "Header"},
		},
		{
			Name:  "Template preview",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page_templates/preview").
					Query(presets.ParamID, "1_Japan").
					BuildEventFuncRequest()

				return req
			},
			ExpectPageBodyContainsInOrder: []string{"list-contents", "headers"},
		},
		{
			Name:  "Template OpenTemplateDialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(actions.OpenListingDialog).
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{presets.CloseListingDialogVarScript},
		},
		{
			Name:  "Template ReloadSelectedTemplate",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(pagebuilder.ReloadSelectedTemplateEvent).
					Query(pagebuilder.ParamTemplateSelectedID, "1").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"123"},
		},
		{
			Name:  "Template Keyword",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTemplateData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_templates").
					EventFunc(actions.OpenListingDialog).
					Query("keyword", "012").
					BuildEventFuncRequest()

				return req
			},
			ExpectPortalUpdate0NotContains: []string{"123"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
