package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
)

var commonContainerData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES (1, '2025-03-03 07:48:33.958831 +00:00', '2025-03-03 07:48:33.958831 +00:00', null, '123', '/123', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2025-03-03-v01', '2025-03-03-v01', '', '');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2025-03-03 07:48:38.445921 +00:00', '2025-03-03 07:48:38.445921 +00:00', null, 1, '2025-03-03-v01', 'pages', 'TailWindExampleFooter', 2, 1, false, false, 'TailWindExampleFooter', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (2, '2025-03-03 07:49:45.326457 +00:00', '2025-03-03 07:49:45.326457 +00:00', null, 1, '2025-03-03-v01', 'pages', 'TailWindExampleHeader', 2, 2, false, false, 'TailWindExampleHeader', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (3, '2025-03-03 07:49:46.922736 +00:00', '2025-03-03 07:49:46.922736 +00:00', null, 1, '2025-03-03-v01', 'pages', 'HeroImageHorizontal', 2, 3, false, false, 'HeroImageHorizontal', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (4, '2025-03-03 07:49:48.382729 +00:00', '2025-03-03 07:49:48.382729 +00:00', null, 1, '2025-03-03-v01', 'pages', 'TailWindHeroVertical', 2, 4, false, false, 'TailWindHeroVertical', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (5, '2025-03-03 07:49:50.127054 +00:00', '2025-03-03 07:49:50.127054 +00:00', null, 1, '2025-03-03-v01', 'pages', 'TailWindHeroList', 2, 5, false, false, 'TailWindHeroList', '', 0);
INSERT INTO public.container_tailwind_footer (id) VALUES (2);
INSERT INTO public.container_tailwind_header (id) VALUES (2);
INSERT INTO public.container_tailwind_hero_horizontal (id, content, style) VALUES (2, '{"Title":"This is a title","Body":"From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation.","Button":"Get Start","ButtonStyle":"unset","ImgInitial":false,"ImageUpload":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', '{"Layout":"left","TopSpace":0,"BottomSpace":0,"ImgInitial":false,"ImageBackground":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.container_tailwind_hero_list (id, content, style) VALUES (2, '{"Title":"This is a title"}', '{"TopSpace":0,"BottomSpace":0,"ImgInitial":false}');
INSERT INTO public.container_tailwind_hero_vertical (id, content, style) VALUES (2, '{"Title":"This is a title","Body":"From end-to-end solutions to consulting, we draw on decades of expertise to solve new challenges in e-commerce, content management, and digital innovation.","Button":"Get Start","ButtonStyle":"unset","ImgInitial":false,"ImageUpload":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', '{"TopSpace":0,"BottomSpace":0,"ImgInitial":false}');

`, []string{"page_builder_pages", "page_builder_containers", "container_tailwind_footer", "container_tailwind_header", "container_tailwind_hero_horizontal", "container_tailwind_hero_list", "container_tailwind_hero_vertical"}))

func TestCommonContainer(t *testing.T) {
	pb := presets.New()
	b := PageBuilderCommonContainerExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Common Container Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				commonContainerData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/pages/1_2025-03-03-v01", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"TailWindExampleFooter", "TailWindExampleHeader", "HeroImageHorizontal", "TailWindHeroVertical", "TailWindHeroList"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, b)
		})
	}
}
