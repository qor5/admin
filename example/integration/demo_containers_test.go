package integration_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var demoContainerData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (1, '2024-08-26 03:00:44.699127 +00:00', '2024-08-26 03:00:44.699127 +00:00', null, 'Header', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (2, '2024-08-26 03:00:44.727313 +00:00', '2024-08-26 03:00:44.727313 +00:00', null, 'Heading', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (3, '2024-08-26 03:00:45.406616 +00:00', '2024-08-26 03:00:45.406616 +00:00', null, 'PageTitle', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (4, '2024-08-26 03:00:45.448964 +00:00', '2024-08-26 03:00:45.448964 +00:00', null, 'Video Banner', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (5, '2024-08-26 03:00:45.479946 +00:00', '2024-08-26 03:00:45.479946 +00:00', null, 'BrandGrid', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (6, '2024-08-26 03:00:46.095622 +00:00', '2024-08-26 03:00:46.095622 +00:00', null, 'Footer', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (7, '2024-08-26 03:00:46.309314 +00:00', '2024-08-26 03:00:46.309314 +00:00', null, 'Image', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (8, '2024-08-26 09:35:33.805101 +00:00', '2024-08-26 09:35:33.805101 +00:00', null, 'ListContent', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (9, '2024-08-26 09:35:33.812110 +00:00', '2024-08-26 09:35:33.812110 +00:00', null, 'InNumbers', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (10, '2024-08-26 09:35:33.815391 +00:00', '2024-08-26 09:35:33.815391 +00:00', null, 'ContactForm', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (11, '2024-08-26 09:35:33.820515 +00:00', '2024-08-26 09:35:33.820515 +00:00', null, 'ListContentLite', 1, false, 'International');
INSERT INTO public.page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, filled, locale_code) VALUES (12, '2024-08-26 09:35:33.823044 +00:00', '2024-08-26 09:35:33.823044 +00:00', null, 'ListContentWithImage', 1, false, 'International');
INSERT INTO public.container_list_content_with_image (id, add_top_space, add_bottom_space, anchor_id, items) VALUES (1, false, false, '', 'null');
INSERT INTO public.container_list_content_lite (id, add_top_space, add_bottom_space, anchor_id, items, background_color) VALUES (1, false, false, '', 'null', '');
INSERT INTO public.container_contact_form (id, add_top_space, add_bottom_space, anchor_id, heading, text, send_button_text, form_button_text, message_placeholder, name_placeholder, email_placeholder, thankyou_message, action_url, privacy_policy) VALUES (1, false, false, '', '', '', '', '', '', '', '', '', '', '');
INSERT INTO public.container_in_numbers (id, add_top_space, add_bottom_space, anchor_id, heading, items) VALUES (1, false, false, '', '', 'null');
INSERT INTO public.container_list_content (id, add_top_space, add_bottom_space, anchor_id, items, background_color, link, link_text, link_display_option) VALUES (1, false, false, '', 'null', '', '', '', '');
INSERT INTO public.container_images (id, add_top_space, add_bottom_space, anchor_id, image, background_color, transition_background_color) VALUES (1, false, false, '', null, '', '');
INSERT INTO public.container_footers (id, english_url, japanese_url) VALUES (1, '', '');
INSERT INTO public.container_brand_grids (id, add_top_space, add_bottom_space, anchor_id, brands) VALUES (1, false, false, '', 'null');
INSERT INTO public.container_video_banners (id, add_top_space, add_bottom_space, anchor_id, video, background_video, mobile_background_video, video_cover, mobile_video_cover, heading, popup_text, text, link_text, link) VALUES (1, false, false, '', null, null, null, null, null, '', '', '', '', '');
INSERT INTO public.container_page_title (id, add_top_space, add_bottom_space, anchor_id, hero_image, navigation_link, navigation_link_text, heading_icon, heading, text, tags) VALUES (1, false, false, '', null, '', '', '', '', '', 'null');
INSERT INTO public.container_headings (id, add_top_space, add_bottom_space, anchor_id, heading, font_color, background_color, link, link_text, link_display_option, text) VALUES (1, false, false, '', '', '', '', '', '', '', '');
INSERT INTO public.container_headers (id, color) VALUES (1, '');

`,

	[]string{
		`page_builder_demo_containers`, `container_list_content_with_image`, `container_brand_grids`, `container_list_content_lite`,
		`container_contact_form`, `container_in_numbers`, `container_list_content`, `container_images`, `container_footers`,
		`container_brand_grids`, `container_video_banners`, `container_page_title`, `container_headings`, `container_headers`,
	},
))

func TestDemoContainer(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "ListContentWithImage Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/list-content-with-images").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Items", "Add Row"},
		},
		{
			Name:  "ListContentLite Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/list-content-lites").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Items", "Add Row", "Background Color", "vx-select"},
		},
		{
			Name:  "ListContentLite Edit View Add Row",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/list-content-lites").
					EventFunc(actions.AddRowEvent).
					Query(presets.ParamID, "1").
					Query("listEditor_AddRowFormKey", "Items").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Heading", "vx-field", "vx-tiptap-editor"},
		},
		{
			Name:  "ContactForm Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/contact-forms").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Privacy Policy"},
		},
		{
			Name:  "InNumber Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/in-numbers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Heading", "Items", "Add Row"},
		},
		{
			Name:  "ListContent Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/list-contents").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Background Color", "vx-select", "Items", "Add Row", "Link Display Option"},
		},
		{
			Name:  "ImageContainer Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/images").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Background Color", "vx-select", "Choose File"},
		},
		{
			Name:  "WebFooter Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/footers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"English Url", "vx-field", "Japanese Url"},
		},
		{
			Name:  "BrandGrid Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/brand-grids").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Brands", "Add Row"},
		},
		{
			Name:  "VideoBanner Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/video-banners").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Video", "Choose File", "Link Text"},
		},
		{
			Name:  "PageTitle Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/page-titles").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Hero Image", "Choose File", "Navigation Link Text", "Tags", "Add Row"},
		},
		{
			Name:  "Heading Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/headings").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Top Space", "vx-checkbox", "Add Bottom Space", "Anchor ID", "vx-field", "Font Color", "vx-select", "vx-tiptap-editor"},
		},
		{
			Name:  "WebHeader Edit View",
			Debug: true,
			ReqFunc: func() *http.Request {
				demoContainerData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/headers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Color", "vx-select"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
