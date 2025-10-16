package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var seoData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.seo_posts (title, seo, id, created_at, updated_at, deleted_at) VALUES 
                                                                                      ('The seo post 1', 
'{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""}}', 1, '2024-05-31 10:02:13.114089 +00:00', '2024-05-31 10:02:13.114089 +00:00', null),
                                                                                       ('The seo post 2', 
'{"OpenGraphImageFromMediaLibrary":{"ID":0,"Url":"","VideoLink":"","FileName":"","Description":""},"Title":"123","EnabledCustomize":true,"OpenGraphImageURL":"http://www.text.jpg"}', 2, '2024-05-31 10:02:13.114089 +00:00', '2024-05-31 10:02:13.114089 +00:00', null)
                                                                                      ;

`, []string{"seo_posts"}))

func TestSEOExampleBasic(t *testing.T) {
	pb := presets.New()
	SEOExampleBasic(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				seoData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/seo-posts", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"The seo post 1"},
		},
		{
			Name:  "Edit SEO Title",
			Debug: true,
			ReqFunc: func() *http.Request {
				seoData.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/seo-posts?__execute_event__=section_save_SEO&id=1").
					AddField("SEO.EnabledCustomize", "true").
					AddField("SEO.Title", "My seo title").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`My seo title`},
		},
		{
			Name:  "SEO Detail With OpenGraphImageURL",
			Debug: true,
			ReqFunc: func() *http.Request {
				seoData.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/seo-posts").
					EventFunc(actions.DetailingDrawer).
					Query(presets.ParamID, "2").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`http://www.text.jpg`},
		},
		{
			Name:  "SEO Detail Without OpenGraphImageURL",
			Debug: true,
			ReqFunc: func() *http.Request {
				seoData.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/seo-posts").
					EventFunc(actions.DetailingDrawer).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0NotContains: []string{`http://www.text.jpg`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
