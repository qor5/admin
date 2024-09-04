package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/autocomplete"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	. "github.com/qor5/web/v3/multipartestutils"
)

var autocompleteData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.auto_complete_posts (id, created_at, updated_at, deleted_at, title, body, status) VALUES (1, '2024-08-06 10:06:17.193170 +00:00', '2024-08-08 02:09:24.689204 +00:00', null, 'testname', 'test.body', '123131231');

`, []string{"auto_complete_posts"}))

func TestAutoComplete(t *testing.T) {
	pb := presets.New()

	ab := autocomplete.New().Prefix("/complete")
	b := AutoCompleteBasicFilterExample(pb, ab, TestDB)

	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index AutoCompletePost",
			Debug: true,
			ReqFunc: func() *http.Request {
				autocompleteData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/auto-complete-posts", nil)
			},
			ExpectPageBodyContainsInOrder: []string{"testname", "test.body"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
