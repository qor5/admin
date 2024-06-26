package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestOrders(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Show order detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.OrdersExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/orders?__execute_event__=presets_DetailingDrawer&id=11").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Basic Information`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
