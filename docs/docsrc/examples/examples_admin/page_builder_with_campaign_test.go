package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/publish"

	"github.com/qor5/admin/v3/pagebuilder"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var pageBuilderData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Campaign', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (2, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'UnPublish Campaign', 'online', 'campaigns/2/index.html', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaign_products (id, created_at, updated_at, deleted_at, name, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Product', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.my_contents (id,text) values (1,'my-contents');
INSERT INTO public.campaign_contents (id,title,banner) values (1,'campaign-contents','banner');
INSERT INTO public.product_contents (id,name) values (1,'demo-product-contents');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'campaigns', 'MyContent', 1, 1, false, false, 'MyContent', '', 0);

`, []string{"campaigns", "campaign_products", "my_contents", "campaign_contents", "product_contents", "page_builder_containers"}))

var pageBuilderDemoData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Campaign', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (2, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'UnPublish Campaign', 'online', 'campaigns/2/index.html', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaign_products (id, created_at, updated_at, deleted_at, name, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Product', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.my_contents (id,text) values (1,'my-contents');
INSERT INTO public.campaign_contents (id,title,banner) values (1,'campaign-contents','banner');
INSERT INTO public.product_contents (id,name) values (1,'demo-product-contents');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'campaigns', 'MyContent', 1, 1, false, false, 'MyContent', '', 0);
INSERT INTO page_builder_demo_containers (id, created_at, updated_at, deleted_at, model_name, model_id, locale_code) VALUES (1, '2024-06-25 02:21:41.014915 +00:00', '2024-06-25 02:21:41.014915 +00:00', null, 'ProductContent', 1, '');

`, []string{"campaigns", "campaign_products", "my_contents", "campaign_contents", "product_contents", "page_builder_containers", "page_builder_demo_containers"}))

var pageBuilderPageData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '123', '/12', '', '');
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (2, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '123', '/456', '', '');
SELECT setval('page_builder_categories_id_seq', 1, true);
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/123', 1, 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01', '', '', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
SELECT setval('page_builder_pages_id_seq', 1, true);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'pages', 'MyContent', 1, 1, false, false, 'MyContent', '', 0);
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (2, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'pages', 'MyContent', 2, 2, false, false, 'MyContent', '', 0);
INSERT INTO public.my_contents (id,text) values (1,'my-contents');
INSERT INTO public.my_contents (id,text) values (2,'my-contents2');

`, []string{"page_builder_pages", "page_builder_categories", "page_builder_containers", "my_contents"}))

func forUnpublishCreateFile(filePath string, content string) {
	var (
		err  error
		file *os.File
	)

	// Create all necessary directories
	dir := filepath.Dir(filePath)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(err)
	}

	// Create or truncate the file
	file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write content to the file
	if _, err = file.WriteString(content); err != nil {
		panic(err)
	}
}

func forUnpublishFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func TestPageBuilderCampaign(t *testing.T) {
	pb := presets.New()
	b := PageBuilderExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index Campaign",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/campaigns", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Hello Campaign"},
		},
		{
			Name:  "Campaign Detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/campaigns/1_2024-05-20-v01", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"publish_EventPublish", "CampaignDetail"},
		},
		{
			Name:  "Product Detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/campaign-products/1_2024-05-20-v01", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"publish_EventPublish", "ProductDetail"},
		},
		{
			Name:  "Campaign My Contents",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/my-contents?__execute_event__=presets_Edit&id=1&overlay=content&portal_name=pageBuilderRightContentPortal", nil)
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				if er.UpdatePortals[0].Name != "pageBuilderRightContentPortal" {
					t.Errorf("error portalName %v", er.UpdatePortals[0].Name)
				}
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Editing MyContent 1", "my-contents"},
		},
		{
			Name:  "CampaignContents edit",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/campaign-contents?__execute_event__=presets_Edit&id=1&overlay=content&portal_name=pageBuilderRightContentPortal", nil)
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				if er.UpdatePortals[0].Name != "pageBuilderRightContentPortal" {
					t.Errorf("error portalName %v", er.UpdatePortals[0].Name)
				}
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Editing CampaignContent 1", "campaign-contents"},
		},
		{
			Name:  "Campaign add container MyContent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaigns-editors/1_2024-05-20-v01?__execute_event__=page_builder_AddContainerEvent&modelName=MyContent").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container).Error; err != nil {
					t.Error("containers not add", er)
				}
				if container.ModelName != "MyContent" {
					t.Error("containers not add", container.ModelName)
				}
				if container.PageModelName != "campaigns" {
					t.Error("containers not add for page model name", container.PageModelName)
				}
			},
		},

		{
			Name:  "Campaign add CampaignContent",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaigns-editors/1_2024-05-20-v01?__execute_event__=page_builder_AddContainerEvent&modelName=CampaignContent").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.Order("id desc").First(&container).Error; err != nil {
					t.Error("containers not add", er)
				}
				if container.ModelName != "CampaignContent" {
					t.Error("containers not add", container.ModelName)
				}
				if container.PageModelName != "campaigns" {
					t.Error("containers not add for page model name", container.PageModelName)
				}
			},
		},

		{
			Name:  "Add a new campaigns",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/campaigns?__execute_event__=presets_Update").
					AddField("Title", "Hello 4").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaign Campaign
				if err := TestDB.First(&campaign, "title = ?", "Hello 4").Error; err != nil {
					t.Error(err)
				}
			},
		},

		{
			Name:  "Page Builder Detail Duplicate A Campaign",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/campaigns/1_2024-05-20-v01?__execute_event__=publish_EventDuplicateVersion").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaigns []*Campaign
				TestDB.Where("id=1").Order("id DESC, version DESC").Find(&campaigns)
				if len(campaigns) != 2 {
					t.Fatal("Campaign not duplicated", campaigns)
				}
				var containers []*pagebuilder.Container
				TestDB.Find(&containers, "page_id = ? AND page_version = ?", campaigns[0].ID,
					campaigns[0].Version.Version)
				if len(containers) == 0 {
					t.Error("Container not duplicated", containers)
				}
			},
		},

		{
			Name:  "Page Builder Editor Duplicate A Campaign",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaigns-editors/1_2024-05-20-v01?__execute_event__=publish_EventDuplicateVersion").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaigns []*Campaign
				TestDB.Where("id=1").Order("id DESC, version DESC").Find(&campaigns)
				if len(campaigns) != 2 {
					t.Fatal("Campaign not duplicated", campaigns)
				}
				var containers []*pagebuilder.Container
				TestDB.Find(&containers, "page_id = ? AND page_version = ?", campaigns[0].ID,
					campaigns[0].Version.Version)
				if len(containers) == 0 {
					t.Error("Container not duplicated", containers)
				}
			},
		},
		{
			Name:  "Page Builder Campaign Detail Publish",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/campaigns/1_2024-05-20-v01?__execute_event__=publish_EventPublish&id=1_2024-05-20-v01").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var (
					body []byte
					err  error
				)
				body, err = os.ReadFile("/tmp/publish/campaigns/1/index.html")
				if err != nil {
					t.Fatalf("publish failed!")
					return
				}
				if !strings.Contains(string(body), "Hello Campaign") {
					t.Fatalf("publish failed! %s", string(body))
					return
				}
				var campaigns []*Campaign
				TestDB.Where("id=? and version=? and status=?", "1", "2024-05-20-v01", publish.StatusOnline).Find(&campaigns)
				if len(campaigns) != 1 {
					t.Fatalf("campaign not published, %#+v", campaigns)
					return
				}
				if campaigns[0].OnlineUrl == "" {
					t.Fatalf("campaign not published, %#+v", campaigns)
					return
				}
				body, err = os.ReadFile("/tmp/publish/campaigns/index.html")
				if err != nil {
					t.Fatalf("publish warp content failed! %v", err)
					return
				}
				if !strings.Contains(string(body), "Campaign List") {
					t.Fatalf("publish warp content failed! %v", string(body))
					return
				}
			},
		},
		{
			Name:  "Page Builder Campaign Detail UnPublish",
			Debug: true,
			ReqFunc: func() *http.Request {
				forUnpublishCreateFile("/tmp/publish/campaigns/2/index.html", "UnPublish Campaign")
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/campaigns/2_2024-05-20-v01?__execute_event__=publish_EventUnpublish&id=2_2024-05-20-v01").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaigns []*Campaign
				TestDB.Where("id=? and version=? and status=?", "2", "2024-05-20-v01", publish.StatusOffline).Find(&campaigns)
				if len(campaigns) != 1 {
					t.Fatalf("campaign not unpublished, %#+v", campaigns)
					return
				}
				if forUnpublishFileExists("/tmp/publish/campaigns/2/index.html") {
					t.Fatalf("campaign not unpublished, %#+v", campaigns)
					return
				}
			},
		},
		{
			Name:  "Page Builder Campaign Products Editor Publish",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaign-products-editors/1_2024-05-20-v01?__execute_event__=publish_EventPublish&id=1_2024-05-20-v01").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				body, err := os.ReadFile("/tmp/publish/campaign-products/1/index.html")
				if err != nil {
					t.Error("publish failed!")
					return
				}
				if !strings.Contains(string(body), "Hello Product") {
					t.Errorf("publish failed! %s", string(body))
					return
				}
				var campaignProducts []*CampaignProduct
				TestDB.Where("id=? and version=? and status=?", "1", "2024-05-20-v01", publish.StatusOnline).Find(&campaignProducts)
				if len(campaignProducts) != 1 {
					t.Fatalf("Product not published, %#+v", campaignProducts)
					return
				}
				if campaignProducts[0].OnlineUrl == "" {
					t.Fatalf("Product not published, %#+v", campaignProducts)
					return
				}
			},
		},
		{
			Name:  "CampaignProduct Add New Demo Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaign-products-editors/1_2024-05-20-v01").
					EventFunc(pagebuilder.AddContainerEvent).
					AddField("modelName", "ProductContent").
					AddField("id", "1").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var cons []*pagebuilder.Container
				TestDB.Order("id desc").Find(&cons)
				if len(cons) != 2 {
					t.Fatalf("add container failed, expected 2 cons, got %d", len(cons))
					return
				}
				if cons[0].ModelName != "ProductContent" {
					t.Fatalf("add container failed, expected ProductContent, got %s", cons[0].ModelName)
					return
				}
				var mos []*ProductContent
				TestDB.Order("id desc").Find(&mos)
				if len(mos) != 2 {
					t.Fatalf("add demo container model failed, expected 2 mos, got %d", len(mos))
					return
				}
				if mos[0].Name != "demo-product-contents" {
					t.Fatalf("add demo container model failed, expected demo-product-contents, got %s", mos[0].Name)
					return
				}
			},
		},
		{
			Name:  "Edit Demo Container",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderDemoData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_builder/product-contents?__execute_event__=presets_Update&id=1").
					AddField("Name", "demo-product-contents2").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var (
					mos  []*ProductContent
					cons []*pagebuilder.DemoContainer
				)
				TestDB.Where("model_name = ? and locale_code = ? ", "ProductContent", "").Find(&cons)
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
				if mos[0].Name != "demo-product-contents2" {
					t.Fatalf("Expected name 'demo-product-contents2', got %v", mos[0].Name)
					return
				}
			},
		},
		{
			Name:  "Edit Demo Container Zero Value",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/page_builder/product-contents?__execute_event__=presets_Update&id=1").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				var (
					mos  []*ProductContent
					cons []*pagebuilder.DemoContainer
				)
				TestDB.Where("model_name = ? and locale_code = ? ", "ProductContent", "").Find(&cons)
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
				if mos[0].Name != "" {
					t.Fatalf("Expected name '', got %v", mos[0].Name)
					return
				}
			},
		},
		{
			Name:  "Edit section validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/campaigns/2_2024-05-20-v01?__execute_event__=presets_Detailing_Field_Save&id=2_2024-05-20-v01").
					Query("detailField", "CampaignDetail").
					AddField("CampaignDetail.Title", "").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"title could not be empty"},
		},
		{
			Name:  "Pages Detail Save",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderPageData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/pages/1_2024-05-20-v01?__execute_event__=presets_Detailing_Field_Save&detailField=Page&id=1_2024-05-20-v01").
					AddField("Page.Title", "123").
					AddField("Page.Slug", "/123").
					AddField("Page.CategoryID", "2").
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
			Name:  "Page Builder Editor Duplicate A Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderPageData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/pages/1_2024-05-20-v01?__execute_event__=publish_EventDuplicateVersion").
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
			Name:  "Update Category",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderPageData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories?__execute_event__=presets_Update&id=1").
					AddField("Name", "123").
					AddField("Path", "").
					AddField("Description", "321").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var m pagebuilder.Category
				if err := TestDB.First(&m, 1).Error; err != nil {
					t.Fatalf("update cateogry failed %v", er)
					return
				}
				if m.Name != "123" || m.Description != "321" {
					t.Fatalf("update cateogry failed %#+v", m)
					return
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
