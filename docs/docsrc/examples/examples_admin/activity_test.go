package examples_admin

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var activityData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.activity_logs (id, user_id, created_at, creator, action, model_keys, model_name, model_label, model_link, model_diffs) VALUES (1, 0, '2024-05-30 07:02:53.393836 +00:00', 'smile', 'Create', '1:xxx', 'WithActivityProduct', 'with-activity-products', '', '');

INSERT INTO public.with_activity_products (title, code, price, id, created_at, updated_at, deleted_at) VALUES ('P11111111111', 'code11111111', 0, 1, '2024-05-30 07:02:53.389781 +00:00', '2024-05-30 07:15:38.585837 +00:00', null);

`, []string{"with_activity_products", "activity_logs"}))

func mockHandleError(in http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				http.Error(w, fmt.Sprint(r), http.StatusInternalServerError)
			}
		}()
		in.ServeHTTP(w, r)
	})
}

func TestActivity(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	ActivityExample(pb, TestDB)
	h := mockHandleError(pb)
	dbr, _ := TestDB.DB()

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
			Name:  "create note",
			Debug: true,
			ReqFunc: func() *http.Request {
				activityData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-activity-products?__execute_event__=note_CreateNoteEvent").
					AddField("resource_id", "1").
					AddField("resource_type", "WithActivityProduct").
					AddField("Content", "Hello content, I am writing a content").
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
					AddField("resource_type", "WithActivityProduct").
					AddField("Content", "Hello content, I am writing a content").
					BuildEventFuncRequest()
				return req
			},
			PageMatch: func(t *testing.T, body *bytes.Buffer) {
				fmt.Println(body.String())
				if !strings.Contains(body.String(), "Missing required parameter") {
					t.Error("didn't check correct")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
