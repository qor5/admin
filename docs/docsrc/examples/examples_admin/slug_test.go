package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestAutoSyncFrom(t *testing.T) {
	pb := presets.New()
	AutoSyncFromExample(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-title-slugs?__execute_event__=presets_New", nil)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Title", "Title Slug"},
		},
		{
			Name:  "sync",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/with-title-slugs?__execute_event__=slug_sync&field_name=Title&slug_label=Title%20Slug").
					AddField("Title", `A_Bc`).
					AddField("TitleWithSlug_Checkbox", `true`).
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"a_bc", "Title Slug"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
