package integration_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	m.Run()
}

var pageBuilderData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_categories (id, created_at, updated_at, deleted_at, name, path, description, locale_code) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, '123', '/12', '', 'International');

INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code, seo) VALUES (1, '2024-05-17 15:25:39.716658 +00:00', '2024-05-17 15:25:39.716658 +00:00', null, '12312', '/123', 1, 'draft', '', null, null, null, null, '2024-05-18-v01', '2024-05-18-v01', '', 'International', '{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}');
`, []string{"page_builder_pages", "page_builder_categories"}))

func TestPageBuilder(t *testing.T) {
	h := admin.TestHandler(TestDB)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name: "Index page",
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/pages", nil)
			},
			PageMatch: func(t *testing.T, body *bytes.Buffer) {
				t.Log(body.String())
			},
		},

		{
			Name: "Page Builder Detail page",
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages/1_2024-05-18-v01_International", nil)
			},
			PageMatch: func(t *testing.T, body *bytes.Buffer) {
				// t.Log(body.String())
				if !strings.Contains(body.String(), `eventFunc("createNoteDialogEvent").query("overlay", "dialog").query("id", "1_2024-05-18-v01_International").url("/pages").go()`) {
					t.Error(body.String())
				}
			},
		},

		{
			Name: "Page Builder Editor Notes add",
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().PageURL("http://localhost:9500/pages?__execute_event__=createNoteDialogEvent&id=1_2024-05-18-v01_International&overlay=dialog").
					BuildEventFuncRequest()
				bs, _ := httputil.DumpRequest(req, true)
				t.Log(string(bs))
				return req
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				t.Log(er)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
