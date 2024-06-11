package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/pagebuilder"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var pageBuilderData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.campaigns (id, created_at, updated_at, deleted_at, title, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Campaign', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.campaign_products (id, created_at, updated_at, deleted_at, name, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version) VALUES (1, '2024-05-19 22:11:53.645941 +00:00', '2024-05-19 22:11:53.645941 +00:00', null, 'Hello Product', 'draft', '', null, null, null, null, '2024-05-20-v01', '2024-05-20-v01','');
INSERT INTO public.my_contents (id,text) values (1,'my-contents');
INSERT INTO public.campaign_contents (id,title,banner) values (1,'campaign-contents','banner');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-05 07:20:58.435363 +00:00', '2024-06-05 07:20:58.435363 +00:00', null, 1, '2024-05-20-v01', 'campaigns', 'MyContent', 1, 1, false, false, 'MyContent', '', 0);

`, []string{"campaigns", "campaign_products", "my_contents", "campaign_contents", "page_builder_containers"}))

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
			ExpectPageBodyContainsInOrder: []string{"publish_EventPublish", "iframe", "CampaignDetail"},
		},
		{
			Name:  "Product Detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/campaign-products/1_2024-05-20-v01", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"publish_EventPublish", "iframe", "ProductDetail"},
		},
		{
			Name:  "Campaign editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/page_builder/campaigns/editors/1_2024-05-20-v01", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"MyContent", "CampaignContent"},
			ExpectPageBodyNotContains:     []string{"ProductContent"},
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
					PageURL("/page_builder/campaigns/editors/1_2024-05-20-v01?__execute_event__=page_builder_AddContainerEvent&modelName=MyContent").
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
					PageURL("/page_builder/campaigns/editors/1_2024-05-20-v01?__execute_event__=page_builder_AddContainerEvent&modelName=CampaignContent").
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
			Name:  "Page Builder Editor Duplicate A Campaign",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_builder/campaigns/editors/1_2024-05-20-v01?__execute_event__=publish_EventDuplicateVersion").
					BuildEventFuncRequest()

				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var campaigns []*Campaign
				TestDB.Order("id DESC, version DESC").Find(&campaigns)
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
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
