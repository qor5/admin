package admin_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/pagebuilder"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"github.com/theplant/testingutils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/qor5/admin/v3/cmd/qor5/website-template/admin"
)

var (
	TestDB *gorm.DB
	SqlDB  *sql.DB
)

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	TestDB.Logger = TestDB.Logger.LogMode(logger.Info)
	SqlDB, _ = TestDB.DB()
	m.Run()
}

var data = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES (2, '2024-06-08 02:14:07.024850 +00:00', '2024-06-08 02:14:07.024850 +00:00', null, 'My first page', 'my-first-page', 0, '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 'draft', '', null, null, null, null, '2024-06-08-v01', '2024-06-08-v01', '', '');
INSERT INTO public.page_builder_containers (id, created_at, updated_at, deleted_at, page_id, page_version, page_model_name, model_name, model_id, display_order, shared, hidden, display_name, locale_code, localize_from_model_id) VALUES (1, '2024-06-08 03:04:15.439286 +00:00', '2024-06-08 03:04:15.439286 +00:00', null, 2, '2024-06-08-v01', 'pages', 'MyHeader', 1, 1, false, false, 'MyHeader', '', 0);

INSERT INTO public.my_headers (menu_items, id, created_at, updated_at, deleted_at) VALUES ('null', 1, '2024-06-08 03:04:15.425264 +00:00', '2024-06-09 02:24:50.794541 +00:00', null);

`, []string{"page_builder_pages", "page_builder_containers", "my_headers"}))

func TestAll(t *testing.T) {
	mux := admin.Router(TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "index page",
			Debug: true,
			ReqFunc: func() *http.Request {
				data.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/admin/pages", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"My first page"},
		},
		{
			Name:  "add container to page",
			Debug: true,
			ReqFunc: func() *http.Request {
				data.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/admin/page_builder/pages/2_2024-06-08-v01?__execute_event__=page_builder_AddContainerEvent&containerName=MyHeader&modelName=MyHeader&tab=Elements").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var container pagebuilder.Container
				if err := TestDB.First(&container).Error; err != nil {
					t.Error("containers not add", er)
				}
				if container.ModelName != "MyHeader" {
					t.Error("containers not add", container.ModelName)
				}
				if container.PageModelName != "pages" {
					t.Error("containers not add for page model name", container.PageModelName)
				}
			},
		},
		{
			Name:  "add menu items to header",
			Debug: true,
			ReqFunc: func() *http.Request {
				data.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/admin/my-headers?__execute_event__=presets_Update&id=1").
					AddField("MenuItems[0].Text", "123").
					AddField("MenuItems[0].Link", "123").
					AddField("MenuItems[1].Text", "456").
					AddField("MenuItems[1].Link", "456").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var header admin.MyHeader
				TestDB.Find(&header)
				diff := testingutils.PrettyJsonDiff([]*admin.MenuItem{
					{"123", "123"},
					{"456", "456"},
				}, header.MenuItems)
				if diff != "" {
					t.Error(diff)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, mux)
		})
	}
}
