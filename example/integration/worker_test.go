package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
)

func TestWorker(t *testing.T) {
	h := admin.TestHandlerWorker(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Worker Schedule Job ",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/workers").
					EventFunc("worker_selectJob").
					Query("jobName", "scheduleJob").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`F1`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
