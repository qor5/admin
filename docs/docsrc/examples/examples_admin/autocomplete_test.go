package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/autocomplete"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/presets"
)

var autocompleteData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.auto_complete_posts (id, created_at, updated_at, deleted_at, title, body, status) VALUES (1, '2024-08-06 10:06:17.193170 +00:00', '2024-08-08 02:09:24.689204 +00:00', null, 'testname', 'test.body', '123131231');
INSERT INTO public.auto_complete_posts (id, created_at, updated_at, deleted_at, title, body, status) VALUES (2, '2024-08-06 10:06:17.193170 +00:00', '2024-08-08 02:09:24.689204 +00:00', null, 'testname', 'test.title', '123131231');

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
				return httptest.NewRequest("GET", "/auto-complete-posts", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"test.title", "test.body"},
		},
		{
			Name:  "Index AutoCompletePost Filter",
			Debug: true,
			ReqFunc: func() *http.Request {
				autocompleteData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/auto-complete-posts?f_title=test.title__2", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"test.title"},
			ExpectPageBodyNotContains:     []string{"test.body"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, b)
		})
	}
}
