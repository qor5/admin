package examples_admin

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/pagebuilder"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
)

var duplicateAsNewPageData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '123', '/12', '', '');
SELECT setval('page_builder_categories_id_seq', 1, true);
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'Source Page', '/123', 1, 'online', '/12/123/index.html', '2030-01-01 00:00:00 +00:00', null, '2024-05-20 00:00:00 +00:00', null, '2024-05-20-v01', 'my version', '', '', '{"Title":"seo-title","EnabledCustomize":true,"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (2, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'Taken', '/123-copy', 1, 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01', '', '', '{}');
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (3, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, 'Taken Too', '/123-copy1', 1, 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01', '', '', '{}');
SELECT setval('page_builder_pages_id_seq', 3, true);
INSERT INTO public.my_contents (id,text) values (1,'first');
INSERT INTO public.my_contents (id,text) values (2,'second');
SELECT setval('my_contents_id_seq', 2, true);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'pages', 'MyContent', 1, 1, false, false, 'MyContent', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (2, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'pages', 'MyContent', 2, 2, false, true, 'MyContent', '', 0);
SELECT setval('page_builder_containers_id_seq', 2, true);
`, []string{"page_builder_pages", "page_builder_categories", "page_builder_containers", "my_contents"}))

var duplicateAsNewCampaignData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Campaign', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Campaign v2', 'draft', '', null, null, null, null, '2024-05-20-v02', '2024-05-20-v02','2024-05-20-v01');
SELECT setval('campaigns_id_seq', 1, true);
INSERT INTO public.my_contents (id,text) values (1,'shared-content');
INSERT INTO public.my_contents (id,text) values (2,'v2-content');
SELECT setval('my_contents_id_seq', 2, true);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'campaigns', 'MyContent', 1, 1, true, false, 'Shared One', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (2, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v02', 'campaigns', 'MyContent', 2, 1, false, false, 'MyContent', '', 0);
SELECT setval('page_builder_containers_id_seq', 2, true);
`, []string{"campaigns", "my_contents", "page_builder_containers"}))

func todayV01() string {
	return fmt.Sprintf("%s-v01", time.Now().Local().Format("2006-01-02"))
}

func TestPageBuilderDuplicateAsNew(t *testing.T) {
	pb := presets.New()
	b := PageBuilderExample(pb, TestDB)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Page duplicate as new creates an independent draft page",
			Debug: true,
			ReqFunc: func() *http.Request {
				duplicateAsNewPageData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/pages").
					EventFunc(pagebuilder.EventDuplicateAsNew).
					Query(presets.ParamID, "1_2024-05-20-v01").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var src pagebuilder.Page
				TestDB.First(&src, "id = 1")
				if src.Status.Status != publish.StatusOnline || src.Slug != "/123" {
					t.Fatalf("source page should be untouched, got %+v", src)
				}

				var pages []*pagebuilder.Page
				TestDB.Where("id > 3").Find(&pages)
				if len(pages) != 1 {
					t.Fatalf("expected exactly 1 new page, got %d", len(pages))
				}
				p := pages[0]
				if p.Title != "Source Page" {
					t.Errorf("title should be unchanged, got %q", p.Title)
				}
				if p.Slug != "/123-copy2" {
					t.Errorf("expected slug /123-copy2, got %q", p.Slug)
				}
				if p.CategoryID != 1 {
					t.Errorf("category should be kept, got %d", p.CategoryID)
				}
				if p.Status.Status != publish.StatusDraft || p.OnlineUrl != "" {
					t.Errorf("new page should be a draft without online url, got %+v", p.Status)
				}
				if p.ScheduledStartAt != nil || p.ActualStartAt != nil {
					t.Errorf("schedule should be cleared, got %+v", p.Schedule)
				}
				if p.Version.Version != todayV01() || p.VersionName != todayV01() || p.ParentVersion != "" {
					t.Errorf("unexpected version %+v", p.Version)
				}
				if p.SEO.Title != "seo-title" || !p.SEO.EnabledCustomize {
					t.Errorf("seo should be copied, got %+v", p.SEO)
				}

				var cons []*pagebuilder.Container
				TestDB.Order("display_order").Find(&cons, "page_id = ? AND page_version = ? AND page_model_name = ?", p.ID, p.Version.Version, "pages")
				if len(cons) != 2 {
					t.Fatalf("expected 2 copied containers, got %d", len(cons))
				}
				for _, c := range cons {
					if c.ModelID == 1 || c.ModelID == 2 {
						t.Errorf("container model should be a new row, got model_id %d", c.ModelID)
					}
				}
				if cons[0].Hidden || !cons[1].Hidden {
					t.Errorf("hidden flag should be copied, got %v %v", cons[0].Hidden, cons[1].Hidden)
				}
				var texts []string
				TestDB.Table("my_contents").Where("id IN ?", []uint{cons[0].ModelID, cons[1].ModelID}).Order("id").Pluck("text", &texts)
				if len(texts) != 2 || texts[0] != "first" || texts[1] != "second" {
					t.Errorf("container content should be copied, got %v", texts)
				}

				wantURL := fmt.Sprintf("/pages/%d_%s", p.ID, p.Version.Version)
				if er.PushState == nil || er.PushState.MyURL != wantURL {
					t.Errorf("expected push state to %s, got %+v", wantURL, er.PushState)
				}
			},
		},
		{
			Name:  "Campaign duplicate as new copies the opened version and detaches shared containers",
			Debug: true,
			ReqFunc: func() *http.Request {
				duplicateAsNewCampaignData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/campaigns/1_2024-05-20-v01").
					EventFunc(pagebuilder.EventDuplicateAsNew).
					Query(presets.ParamID, "1_2024-05-20-v01").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaigns []*Campaign
				TestDB.Where("id <> 1").Find(&campaigns)
				if len(campaigns) != 1 {
					t.Fatalf("expected 1 new campaign, got %d", len(campaigns))
				}
				c := campaigns[0]
				if c.Title != "Hello Campaign" {
					t.Errorf("should copy the opened version v01, got title %q", c.Title)
				}
				if c.Version.Version != todayV01() {
					t.Errorf("unexpected version %q", c.Version.Version)
				}

				var cons []*pagebuilder.Container
				TestDB.Find(&cons, "page_id = ? AND page_model_name = ?", c.ID, "campaigns")
				if len(cons) != 1 {
					t.Fatalf("expected 1 copied container, got %d", len(cons))
				}
				con := cons[0]
				if con.Shared {
					t.Errorf("copied shared container should become a normal container")
				}
				if con.ModelID == 1 {
					t.Errorf("copied shared container should point to a new model row")
				}
				var text string
				TestDB.Table("my_contents").Where("id = ?", con.ModelID).Pluck("text", &text)
				if text != "shared-content" {
					t.Errorf("content should be copied, got %q", text)
				}
				var srcShared pagebuilder.Container
				TestDB.First(&srcShared, "id = 1")
				if !srcShared.Shared || srcShared.ModelID != 1 {
					t.Errorf("source shared container should be untouched, got %+v", srcShared)
				}
			},
		},
		{
			Name:  "Listing row menu shows duplicate as new",
			Debug: true,
			ReqFunc: func() *http.Request {
				duplicateAsNewCampaignData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/campaigns").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{pagebuilder.EventDuplicateAsNew, "Duplicate as New"},
		},
		{
			Name:  "Detail page shows duplicate as new",
			Debug: true,
			ReqFunc: func() *http.Request {
				duplicateAsNewCampaignData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/campaigns/1_2024-05-20-v01").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{pagebuilder.EventDuplicateAsNew, "Duplicate as New"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
