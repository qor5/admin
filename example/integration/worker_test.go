package integration_test

import (
	"net/http"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets/actions"
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
		{
			Name:  "Worker Arg Job ",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/workers").
					EventFunc("worker_selectJob").
					Query("jobName", "argJob").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`F1`, "F2", "F3"},
		},
		{
			Name:  "Worker Arg Job Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				admin.PostsExampleData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/workers").
					EventFunc(actions.Update).
					AddField("Job", "argJob").
					AddField("F1", "aaa").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`cannot be aaa`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, h)
		})
	}
}
