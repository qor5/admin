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
	AutoSyncExample(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "new",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/with-slug-products?__execute_event__=presets_New", nil)
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Title", "Title Slug", "Auto Sync"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
