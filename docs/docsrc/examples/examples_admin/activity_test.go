package examples_admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/perm"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
)

var activityData = gofixtures.Data(gofixtures.Sql(`

INSERT INTO "public"."activity_logs" ("id", "created_at", "updated_at", "deleted_at", "user_id", "action", "hidden", "model_name", "model_keys", "model_label", "model_link", "detail", "scope") VALUES ('29', '2024-07-16 02:50:10.250739+00', '2024-07-16 02:50:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', 'null', ',ownerID:1,'),
('30', '2024-07-16 02:55:10.250739+00', '2024-07-16 02:55:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '14', 'with-activity-products', '/examples/activity-example/with-activity-products/14', 'null', ',ownerID:1,'),
('31', '2024-07-16 03:00:10.250739+00', '2024-07-16 03:00:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '15', 'with-activity-products', '/examples/activity-example/with-activity-products/15', 'null', ',ownerID:1,'),
('32', '2024-07-16 03:05:10.250739+00', '2024-07-16 03:05:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '16', 'with-activity-products', '/examples/activity-example/with-activity-products/16', 'null', ',ownerID:1,'),
('33', '2024-07-16 03:10:10.250739+00', '2024-07-16 03:10:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', 'null', ',ownerID:1,'),
('34', '2024-07-16 03:15:10.250739+00', '2024-07-16 03:15:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '18', 'with-activity-products', '/examples/activity-example/with-activity-products/18', 'null', ',ownerID:1,'),
('35', '2024-07-16 03:20:10.250739+00', '2024-07-16 03:20:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '19', 'with-activity-products', '/examples/activity-example/with-activity-products/19', 'null', ',ownerID:1,'),
('36', '2024-07-16 03:25:10.250739+00', '2024-07-16 03:25:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '20', 'with-activity-products', '/examples/activity-example/with-activity-products/20', 'null', ',ownerID:1,'),
('37', '2024-07-16 03:30:10.250739+00', '2024-07-16 03:30:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '21', 'with-activity-products', '/examples/activity-example/with-activity-products/21', 'null', ',ownerID:1,'),
('38', '2024-07-16 03:35:10.250739+00', '2024-07-16 03:35:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '22', 'with-activity-products', '/examples/activity-example/with-activity-products/22', 'null', ',ownerID:1,'),
('39', '2024-07-16 03:40:10.250739+00', '2024-07-16 03:40:10.251259+00', NULL, '1', 'Create', 'f', 'WithActivityProduct', '23', 'with-activity-products', '/examples/activity-example/with-activity-products/23', 'null', ',ownerID:1,'),
('40', '2024-07-16 02:55:26.074417+00', '2024-07-16 02:55:26.075726+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '23', 'with-activity-products', '/examples/activity-example/with-activity-products/23', 'null', ',ownerID:1,'),
('41', '2024-09-10 09:19:30.248075+00', '2024-09-10 09:19:30.250049+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '22', 'with-activity-products', '/examples/activity-example/with-activity-products/22', 'null', ',ownerID:1,'),
('42', '2024-11-04 08:30:06.590733+00', '2024-11-04 08:30:06.594158+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '19', 'with-activity-products', '/examples/activity-example/with-activity-products/19', 'null', ',ownerID:1,'),
('43', '2024-07-16 02:56:16.18094+00', '2024-07-16 02:56:16.184033+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '15', 'with-activity-products', '/examples/activity-example/with-activity-products/15', 'null', ',ownerID:1,'),
('44', '2024-07-16 02:56:42.273117+00', '2024-07-16 02:56:42.275043+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', 'null', ',ownerID:1,'),
('45', '2024-07-16 02:56:45.176698+00', '2024-07-16 02:56:45.177268+00', NULL, '1', 'Note', 'f', 'WithActivityProduct', '13', 'with-activity-products', '/examples/activity-example/with-activity-products/13', '{"note":"The newest model of the off-legacy Air Jordans is ready to burst onto to the scene.","last_edited_at":"0001-01-01T00:00:00Z"}', ',ownerID:1,'),
('85', '2024-09-10 09:19:40.602161+00', '2024-09-10 09:19:40.60246+00', NULL, '1', 'Edit', 'f', 'WithActivityProduct', '22', 'with-activity-products', '/examples/activity-example/with-activity-products/22', '[{"Field":"Price","Old":"250","New":"251"}]', ',ownerID:1,'),
('86', '2024-09-10 09:19:52.521799+00', '2024-09-10 09:19:52.522825+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '20', 'with-activity-products', '/examples/activity-example/with-activity-products/20', 'null', ',ownerID:1,'),
('87', '2024-09-10 09:19:57.195723+00', '2024-09-10 09:19:57.195937+00', NULL, '1', 'Edit', 'f', 'WithActivityProduct', '20', 'with-activity-products', '/examples/activity-example/with-activity-products/20', '[{"Field":"Price","Old":"70","New":"72"}]', ',ownerID:1,'),
('119', '2024-11-04 08:30:12.193638+00', '2024-11-04 08:30:12.194166+00', NULL, '1', 'Delete', 'f', 'WithActivityProduct', '19', 'with-activity-products', '/examples/activity-example/with-activity-products/19', 'null', ',ownerID:1,'),
('120', '2024-11-04 08:30:48.658835+00', '2024-11-04 08:30:48.660793+00', NULL, '1', 'LastView', 't', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', 'null', ',ownerID:1,'),
('121', '2024-11-04 08:30:34.915831+00', '2024-11-04 08:30:34.916535+00', NULL, '1', 'Edit', 'f', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', '[{"Field":"AppovedAt","Old":"","New":"2024-11-06 00:00:00 +0800 CST"}]', ',ownerID:1,'),
('122', '2024-11-04 08:30:57.018058+00', '2024-11-04 08:30:57.018752+00', NULL, '1', 'Edit', 'f', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', '[{"Field":"Price","Old":"90","New":"0"}]', ',ownerID:1,'),
('123', '2024-11-04 08:31:13.297364+00', '2024-11-04 08:31:13.298046+00', NULL, '1', 'Edit', 'f', 'WithActivityProduct', '17', 'with-activity-products', '/examples/activity-example/with-activity-products/17', '[{"Field":"AppovedAt","Old":"2024-11-06 00:00:00 +0800 CST","New":""}]', ',ownerID:1,');

INSERT INTO "public"."activity_users" ("id", "created_at", "updated_at", "deleted_at", "name", "avatar") VALUES ('1', '2024-07-26 08:03:52.17526+00', '2024-11-04 08:31:13.292852+00', NULL, 'John', 'https://i.pravatar.cc/300');

INSERT INTO "public"."with_activity_products" ("id", "created_at", "updated_at", "deleted_at", "title", "code", "price", "approved", "edited", "stocked_at", "appoved_at") VALUES ('13', '2024-07-16 02:50:10.242888+00', '2024-07-16 02:50:10.242888+00', NULL, 'Air Jordan 4 RM Prowls', '10001', '200', NULL, NULL, NULL, NULL),
('14', '2024-07-16 02:55:10.242888+00', '2024-07-16 02:55:10.242888+00', NULL, 'Nike Air Max 90', '10002', '150', NULL, NULL, NULL, NULL),
('15', '2024-07-16 03:00:10.242888+00', '2024-07-16 03:00:10.242888+00', NULL, 'Adidas Yeezy Boost 350', '10003', '220', NULL, NULL, NULL, NULL),
('16', '2024-07-16 03:05:10.242888+00', '2024-07-16 03:05:10.242888+00', NULL, 'Puma Suede Classic', '10004', '80', NULL, NULL, NULL, NULL),
('17', '2024-07-16 03:10:10.242888+00', '2024-11-04 08:31:13.289941+00', NULL, 'Reebok Classic Leather', '10005', '0', 'f', 'f', '0001-01-01 00:00:00+00', NULL),
('18', '2024-07-16 03:15:10.242888+00', '2024-07-16 03:15:10.242888+00', NULL, 'Asics Gel Lyte III', '10006', '120', NULL, NULL, NULL, NULL),
('19', '2024-07-16 03:20:10.242888+00', '2024-07-16 03:20:10.242888+00', '2024-11-04 08:30:12.185633+00', 'New Balance 574', '10007', '130', NULL, NULL, NULL, NULL),
('20', '2024-07-16 03:25:10.242888+00', '2024-09-10 09:19:57.192539+00', NULL, 'Converse Chuck Taylor All Star', '10008', '72', NULL, NULL, NULL, NULL),
('21', '2024-07-16 03:30:10.242888+00', '2024-07-16 03:30:10.242888+00', NULL, 'Vans Old Skool', '10009', '60', NULL, NULL, NULL, NULL),
('22', '2024-07-16 03:35:10.242888+00', '2024-09-10 09:19:40.596903+00', NULL, 'Jordan 1 Retro High', '10010', '251', NULL, NULL, NULL, NULL),
('23', '2024-07-16 03:40:10.242888+00', '2024-07-16 03:40:10.242888+00', NULL, 'Under Armour Curry 7', '10011', '140', NULL, NULL, NULL, NULL);
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
				return httptest.NewRequest("GET", "/with-activity-products", http.NoBody)
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
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene.", "John", "Created"},
		},
		{
			Name:  "Activity Model details (without PermListNotes)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermListNotes).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0NotContains: []string{">Add Note</v-btn>", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
		},
		{
			Name:  "view all (with admin used)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					defer pb.Use(ab)
					ab.MaxCountShowInTimeline(1)
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene.", "View All"},
		},
		{
			Name:  "can not show more (without admin used)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					ab.MaxCountShowInTimeline(1)
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene.", "Reached the display limit, unable to load more."},
		},
		{
			Name:  "timeline without non-note logs",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					defer pb.Use(ab)
					ab.FindLogsForTimelineFunc(func(ctx context.Context, db *gorm.DB, modelName, modelKeys string) ([]*activity.ActivityLog, bool, error) {
						maxCount := 10
						var logs []*activity.ActivityLog
						err := db.Where("hidden = FALSE AND model_name = ? AND model_keys = ? AND action = ?", modelName, modelKeys, activity.ActionNote).
							Order("created_at DESC").Limit(maxCount + 1).Find(&logs).Error
						if err != nil {
							return nil, false, err
						}
						if len(logs) > maxCount {
							return logs[:maxCount], true, nil
						}
						return logs, false, nil
					})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
			ExpectPortalUpdate0NotContains:     []string{"Created"},
		},
		{
			Name:  "without PermAddNote for ModelBuilder",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermAddNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
			ExpectPortalUpdate0NotContains:     []string{">Add Note</v-btn>"},
		},
		{
			Name:  "without PermEditNote for ModelBuilder",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermEditNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
			ExpectPortalUpdate0NotContains:     []string{"mdi-square-edit-outline"},
		},
		{
			Name:  "without PermDeleteNote for ModelBuilder",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermDeleteNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=13").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 13", ">Add Note</v-btn>", "John", "Added a note", "The newest model of the off-legacy Air Jordans is ready to burst onto to the scene."},
			ExpectPortalUpdate0NotContains:     []string{"mdi-delete"},
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
			Name:  "Create note (without PermAddNote)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermAddNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
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
			ExpectRunScriptContainsInOrder: []string{"permission denied"},
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
			Name:  "Update note without model_keys",
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
				"model_keys": "",
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
			ExpectRunScriptContainsInOrder: []string{"permission denied"},
		},
		{
			Name:  "Update note (without PermEditNote)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermEditNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
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
			ExpectRunScriptContainsInOrder: []string{"permission denied"},
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
		{
			Name:  "Delete Note (without PermDeleteNote)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(activity.PermDeleteNote).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, nil)
				return pb
			},
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
			ExpectRunScriptContainsInOrder: []string{"permission denied"},
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
			Name:  "Page 1",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh&per_page=10", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"<div", "<v-btn", "mdi-chevron-left", ":disabled='true'", "<v-btn", "mdi-chevron-right", ":disabled='false'", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination"},
		},
		{
			Name:  "Page 2",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh&after=eyJJRCI6Mzd9&per_page=10", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"<div", "<v-btn", "mdi-chevron-left", ":disabled='false'", "<v-btn", "mdi-chevron-right", ":disabled='true'", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination"},
		},
		{
			Name:  "Goto Page 1 From Page 2",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh&before=eyJJRCI6MzZ9&per_page=10", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"<div", "<v-btn", "mdi-chevron-left", ":disabled='true'", "<v-btn", "mdi-chevron-right", ":disabled='false'", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination"},
		},
		{
			Name:  "Create note",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
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
			Name:  "Goto Previous Page From Current Page after new record inserted (using original before)",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh&before=eyJJRCI6MzZ9&per_page=10", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"<div", "<v-btn", "mdi-chevron-left", ":disabled='false'", "<v-btn", "mdi-chevron-right", ":disabled='false'", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination"},
		},
		{
			Name:  "Goto Previous Page Again (check records count displayed)",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh&before=eyJJRCI6MTIzfQ&per_page=10", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"</tr>", "</tr>", "</tr>", "</tr>", "</tr>", "</tr>", "</tr>", "</tr>", "</tr>", "</tr>",
				"<div", "<v-btn", "mdi-chevron-left", ":disabled='true'", "<v-btn", "mdi-chevron-right", ":disabled='false'", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination", "没有可显示的记录"},
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
			ExpectPortalUpdate0ContainsInOrder: []string{"Activity Log", "45", "A updated note", "edited at now"},
		},
		{
			Name:  "Activity log detail for edit action (changes)",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=87").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Activity Log", "87", `:items='[{"Field":"Price","Old":"70","New":"72"}]'`},
		},
		{
			Name:  "Activity log detail for edit action (new)",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=121").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Activity Log", "121", `:items='[{"Field":"AppovedAt","Old":"","New":"2024-11-06 00:00:00 +0800 CST"}]'`},
		},
		{
			Name:  "Activity log detail for edit action (delete)",
			Debug: true,
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=121").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Activity Log", "121", `:items='[{"Field":"AppovedAt","Old":"","New":"2024-11-06 00:00:00 +0800 CST"}]'`},
		},

		{
			Name:  "Activity logs without PermList",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermList).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					pb.Use(ab)
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"没有可显示的记录",
			},
		},

		{
			Name:  "Activity logs with skipResPermCheck without PermList",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermList).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					ab.SkipResPermCheck(true)
					pb.Use(ab)
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/activity-logs?lang=zh", http.NoBody)
			},
			ExpectPageBodyNotContains: []string{
				"没有可显示的记录",
			},
		},

		{
			Name:  "Activity log detail for edit action with skipResPermCheck without PermGet",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				pb.Permission(
					perm.New().Policies(
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
						perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermGet).On("*:presets:with_activity_products:*"),
					),
				)
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					ab.SkipResPermCheck(true)
					pb.Use(ab)
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				// activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=87").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Activity Log", "87"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}

	panicCase := multipartestutils.TestCase{
		Name:  "Activity log detail for edit action without PermGet",
		Debug: true,
		HandlerMaker: func() http.Handler {
			pb := presets.New()
			pb.Permission(
				perm.New().Policies(
					perm.PolicyFor(perm.Anybody).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),
					perm.PolicyFor(perm.Anybody).WhoAre(perm.Denied).ToDo(presets.PermGet).On("*:presets:with_activity_products:*"),
				),
			)
			activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
				pb.Use(ab)
			})
			return pb
		},
		ReqFunc: func() *http.Request {
			// activityData.TruncatePut(dbr)
			req := multipartestutils.NewMultipartBuilder().
				PageURL("/activity-logs?__execute_event__=presets_DetailingDrawer&id=87").
				BuildEventFuncRequest()
			return req
		},
	}
	t.Run(panicCase.Name, func(t *testing.T) {
		require.Panics(t, func() {
			multipartestutils.RunCase(t, panicCase, pb)
		})
	})
}

func TestActivityBeforeCreate(t *testing.T) {
	pb := presets.New()
	activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
		pb.Use(ab)

		ab.MustGetModelBuilder(mb).BeforeCreate(func(ctx context.Context, log *activity.ActivityLog) error {
			log.ModelName = log.ModelName + "_BeforeCreate"
			log.ModelLabel = log.ModelLabel + "_BeforeCreate"
			return nil
		})
	})

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Edit a item",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=section_save_Content&id=22").
					AddField("Content.Title", "Jordan 1 Retro High").
					AddField("Content.Code", "10011").
					AddField("Content.Approved", "false").
					AddField("Content.Edited", "false").
					AddField("Content.Price", "252").
					AddField("Content.StockedAt", "").
					AddField("Content.AppovedAt", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"PresetsNotifModelsUpdatedexamplesAdminWithActivityProduct"},
		},
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/activity-logs?lang=zh", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				"操作日志",
				"全部", "创建", "编辑", "删除", "备注",
				"<vx-filter", "操作类型", "操作时间", "操作者", "对象", "</vx-filter>",
				"日期时间", "操作者", "操作", "主键", "菜单名", "对象",
				"_BeforeCreate", "_BeforeCreate",
				"<div", "<v-btn", "mdi-chevron-left", "<v-btn", "mdi-chevron-right", "</div>",
			},
			ExpectPageBodyNotContains: []string{"v-pagination"},
		},
		{
			Name:  "Edit a item(before create error)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				pb := presets.New()
				activityExample(pb, TestDB, func(mb *presets.ModelBuilder, ab *activity.Builder) {
					pb.Use(ab)

					ab.MustGetModelBuilder(mb).BeforeCreate(func(ctx context.Context, log *activity.ActivityLog) error {
						return errors.New("before create error")
					})
				})
				return pb
			},
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=section_save_Content&id=22").
					AddField("Content.Title", "Jordan 1 Retro High").
					AddField("Content.Code", "10011").
					AddField("Content.Approved", "false").
					AddField("Content.Edited", "false").
					AddField("Content.Price", "252").
					AddField("Content.StockedAt", "").
					AddField("Content.AppovedAt", "").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"failed to run before create"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
