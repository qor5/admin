package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"
)

var activityData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.with_activity_products (title, code, price, id, created_at, updated_at, deleted_at) VALUES ('P11111111111', 'code11111111', 0, 1, '2024-05-30 07:02:53.389781 +00:00', '2024-05-30 07:15:38.585837 +00:00', null);
INSERT INTO public.activity_logs (id, user_id, created_at, creator, action, model_keys, model_name, model_label, 
model_link, model_diffs, updated_at, deleted_at, comment) VALUES (1, 1, '2024-05-30 07:02:53.393836 +00:00', '{"ID":1,"Name":"John","Avatar":"https://i.pravatar.cc/300"}', 
'Create', '1:xxx', 'WithActivityProduct', 'with-activity-products', '', '', '2024-06-13 17:29:18.373000 +00:00', '2024-06-13 17:29:14.980000 +00:00', 'hello world');
`, []string{"with_activity_products", "activity_logs"}))

func TestActivity(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	ActivityExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	user := &User{Model: gorm.Model{ID: 1}, Name: "John"}

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/with-activity-products", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"P11111111111"},
		},
		{
			Name:  "Activity Model details should have timeline",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=presets_DetailingDrawer&id=1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"WithActivityProduct 1"},
		},
		{
			Name:  "Create note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=note_CreateNoteEvent").
					AddField(activity.ParamResourceKeys, "1").
					AddField(activity.ParamResourceComment, "Hello content, I am writing a content").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Hello content, I am writing a content"},
		},
		{
			Name:  "create note with invalid data",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=note_CreateNoteEvent").
					AddField(activity.ParamResourceComment, "   ").
					BuildEventFuncRequest()
				return req
			},
			ExpectRunScriptContainsInOrder: []string{"comment cannot be blank"},
		},
		{
			Name:  "Delete Note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=note_DeleteNoteEvent&id=1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0NotContains: []string{"hello world"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, login.MockCurrentUser(user)(pb))
		})
	}
}
