package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var redirectionData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.redirections (id, created_at, updated_at, deleted_at, source, target) VALUES (1, '2025-02-06 02:29:02.231371 +00:00', '2025-02-06 02:29:02.231371 +00:00', null, '/international/wen/test/index3.html', 'https://www.taobao.com');

`, []string{"redirections"}))

func TestRedirection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	RedirectionExample(pb, TestDB)

	dbr, _ := TestDB.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				redirectionData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/redirections", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"/international/wen/test/index3.html", "https://www.taobao.com"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
