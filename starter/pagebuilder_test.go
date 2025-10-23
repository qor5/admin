package starter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/starter"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/gormx"
	"github.com/theplant/gofixtures"
	"github.com/theplant/inject"
)

var pageBuilderTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.page_builder_pages (id, created_at, updated_at, deleted_at, title, slug, category_id, seo, status, online_url, scheduled_start_at, scheduled_end_at, actual_start_at, actual_end_at, version, version_name, parent_version, locale_code) VALUES 
										(10, '2024-05-21 01:54:45.280106 +00:00', '2024-05-21 01:54:57.983233 +00:00', null, '1234567', '/12313', 0, '{"Title":"{{Title}}default","EnabledCustomize":true}', 'draft', '', null, null, null, null, '2024-05-21-v01', '2024-05-21-v01', '', 'China');
SELECT setval('page_builder_pages_id_seq', 10, true);
`, []string{"page_builder_pages"}))

func TestPageBuilder(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()
	dbr, _ := db.DB()
	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/pages", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"1234567"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, env.handler)
		})
	}
}
