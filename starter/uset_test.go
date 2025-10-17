package starter_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/starter"
)

func TestUsers(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)

	cases := []TestCase{
		{
			Name:  "Index Users",
			Debug: true,
			ReqFunc: func() *http.Request {
				req := NewMultipartBuilder().
					PageURL("/users").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{ `qor@theplant.jp`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
		})
	}
}
