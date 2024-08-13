package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var activityData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."with_activity_products" ("id", "created_at", "updated_at", "deleted_at", "title", "code", "price") VALUES ('13', '2024-07-16 02:50:10.242888+00', '2024-07-16 02:50:10.242888+00', NULL, 'Air Jordan 4 RM Prowls', '10001', '200'),
('14', '2024-07-16 02:55:10.242888+00', '2024-07-16 02:55:10.242888+00', NULL, 'Nike Air Max 90', '10002', '150'),
('15', '2024-07-16 03:00:10.242888+00', '2024-07-16 03:00:10.242888+00', NULL, 'Adidas Yeezy Boost 350', '10003', '220'),
('16', '2024-07-16 03:05:10.242888+00', '2024-07-16 03:05:10.242888+00', NULL, 'Puma Suede Classic', '10004', '80'),
('17', '2024-07-16 03:10:10.242888+00', '2024-07-16 03:10:10.242888+00', NULL, 'Reebok Classic Leather', '10005', '90'),
('18', '2024-07-16 03:15:10.242888+00', '2024-07-16 03:15:10.242888+00', NULL, 'Asics Gel Lyte III', '10006', '120'),
('19', '2024-07-16 03:20:10.242888+00', '2024-07-16 03:20:10.242888+00', NULL, 'New Balance 574', '10007', '130'),
('20', '2024-07-16 03:25:10.242888+00', '2024-07-16 03:25:10.242888+00', NULL, 'Converse Chuck Taylor All Star', '10008', '70'),
('21', '2024-07-16 03:30:10.242888+00', '2024-07-16 03:30:10.242888+00', NULL, 'Vans Old Skool', '10009', '60'),
('22', '2024-07-16 03:35:10.242888+00', '2024-07-16 03:35:10.242888+00', NULL, 'Jordan 1 Retro High', '10010', '250'),
('23', '2024-07-16 03:40:10.242888+00', '2024-07-16 03:40:10.242888+00', NULL, 'Under Armour Curry 7', '10011', '140');

INSERT INTO "activity_users" ("id", "created_at", "updated_at", "deleted_at", "name", "avatar") VALUES ('1', '2024-07-26 08:03:52.17526+00', '2024-07-30 07:59:17.439158+00', NULL, 'John', 'https://i.pravatar.cc/300');

INSERT INTO "public"."activity_logs" ("id", "created_at", "updated_at", "deleted_at", "user_id", "action", "hidden", "model_name", "model_keys", "model_label", "model_link", "detail") VALUES ('29', '2024-07-16 02:50:10.250739+00', '2024-07-16 02:50:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', 'null'),
('45', '2024-07-16 02:56:45.176698+00', '2024-07-16 02:56:45.177268+00', NULL, '1', 'Note', 'f', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', '{"note":"The newest model of the off-legacy Air Jordans is ready to burst onto to the scene.","last_edited_at":"0001-01-01T00:00:00Z"}'),
('44', '2024-07-16 02:56:42.273117+00', '2024-07-16 02:56:42.275043+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', 'null'),
('30', '2024-07-16 02:55:10.250739+00', '2024-07-16 02:55:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '14', 'with-activity-products', '/examples/activity-example/with-activity-products/14', 'null'),
('43', '2024-07-16 02:56:16.18094+00', '2024-07-16 02:56:16.184033+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '15', 'with-activity-products', '/examples/activity-example/with-activity-products/15', 'null'),
('31', '2024-07-16 03:00:10.250739+00', '2024-07-16 03:00:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '15', 'with-activity-products', '/examples/activity-example/with-activity-products/15', 'null'),
('32', '2024-07-16 03:05:10.250739+00', '2024-07-16 03:05:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '16', 'with-activity-products', '/examples/activity-example/with-activity-products/16', 'null'),
('33', '2024-07-16 03:10:10.250739+00', '2024-07-16 03:10:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', 'null'),
('34', '2024-07-16 03:15:10.250739+00', '2024-07-16 03:15:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '18', 'with-activity-products', '/examples/activity-example/with-activity-products/18', 'null'),
('35', '2024-07-16 03:20:10.250739+00', '2024-07-16 03:20:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '19', 'with-activity-products', '/examples/activity-example/with-activity-products/19', 'null'),
('42', '2024-07-16 02:56:11.064742+00', '2024-07-16 02:56:11.067334+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '19', 'with-activity-products', '/examples/activity-example/with-activity-products/19', 'null'),
('36', '2024-07-16 03:25:10.250739+00', '2024-07-16 03:25:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '20', 'with-activity-products', '/examples/activity-example/with-activity-products/20', 'null'),
('37', '2024-07-16 03:30:10.250739+00', '2024-07-16 03:30:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '21', 'with-activity-products', '/examples/activity-example/with-activity-products/21', 'null'),
('41', '2024-07-16 02:55:56.551122+00', '2024-07-16 02:55:56.55248+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '22', 'with-activity-products', '/examples/activity-example/with-activity-products/22', 'null'),
('38', '2024-07-16 03:35:10.250739+00', '2024-07-16 03:35:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '22', 'with-activity-products', '/examples/activity-example/with-activity-products/22', 'null'),
('40', '2024-07-16 02:55:26.074417+00', '2024-07-16 02:55:26.075726+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '23', 'with-activity-products', '/examples/activity-example/with-activity-products/23', 'null'),
('39', '2024-07-16 03:40:10.250739+00', '2024-07-16 03:40:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '23', 'with-activity-products', '/examples/activity-example/with-activity-products/23', 'null');
`, []string{"with_activity_products", "activity_logs", "activity_users"}))

func TestActivity(t *testing.T) {
	pb := presets.New()
	ActivityExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/with-activity-products", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"Under Armour Curry 7", "Asics Gel Lyte III", "Air Jordan 4 RM Prowls", "<v-badge", ">1</div>", "</v-badge>"},
		},
		{
			Name:  "Activity Model details should have timeline",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
		},
		{
			Name:  "Create note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*activity.TimelineCompo",
	"compo": {
		"id": "with-activity-products:13",
		"model_name": "WithActivityProduct",
		"model_keys": "13",
		"model_link": "/examples/activity-example/with-activity-products/13"
	},
	"injector": "__activity:with-activity-products__",
	"sync_query": false,
	"method": "CreateNote",
	"request": {
		"note": "The iconic all-black look."
	}
}	
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully created note", "The iconic all-black look."},
		},
		{
			Name:  "create note with invalid data",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*activity.TimelineCompo",
	"compo": {
		"id": "with-activity-products:13",
		"model_name": "WithActivityProduct",
		"model_keys": "13",
		"model_link": "/examples/activity-example/with-activity-products/13"
	},
	"injector": "__activity:with-activity-products__",
	"sync_query": false,
	"method": "CreateNote",
	"request": {
		"note": "     "
	}
}
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Note cannot be empty"},
		},
		{
			Name:  "Update note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*activity.TimelineCompo",
	"compo": {
		"id": "with-activity-products:13",
		"model_name": "WithActivityProduct",
		"model_keys": "13",
		"model_link": "/examples/activity-example/with-activity-products/13"
	},
	"injector": "__activity:with-activity-products__",
	"sync_query": false,
	"method": "UpdateNote",
	"request": {
		"log_id": 45,
		"note": "A updated note"
	}
}
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully updated note", "A updated note"},
		},
		{
			Name:  "Activity Model details should have timeline after note updated",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "A updated note", "edited at now"},
		},
		{
			Name:  "Delete Note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*activity.TimelineCompo",
	"compo": {
		"id": "with-activity-products:13",
		"model_name": "WithActivityProduct",
		"model_keys": "13",
		"model_link": "/examples/activity-example/with-activity-products/13"
	},
	"injector": "__activity:with-activity-products__",
	"sync_query": false,
	"method": "DeleteNote",
	"request": {
		"log_id": 45
	}
}
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully deleted note", "45"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestActivityAdmin(t *testing.T) {
	pb := presets.New()
	ActivityExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh", nil)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志列表",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作人", "操作对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "表的主键值", "菜单名", "表名",
			},
		},
		{
			Name:  "Update note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*activity.TimelineCompo",
	"compo": {
		"id": "with-activity-products:13",
		"model_name": "WithActivityProduct",
		"model_keys": "13",
		"model_link": "/examples/activity-example/with-activity-products/13"
	},
	"injector": "__activity:with-activity-products__",
	"sync_query": false,
	"method": "UpdateNote",
	"request": {
		"log_id": 45,
		"note": "A updated note"
	}
}
`).
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"Successfully updated note", "A updated note"},
		},
		{
			Name:  "Activity log detail after note updated",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=45").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct", "45", "A updated note", "edited at now"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
