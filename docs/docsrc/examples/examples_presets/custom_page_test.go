package examples_presets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var customPageData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
'2024-03-28 05:52:28.497536 +00:00', null, null, '');
`, []string{"customers"}))

func TestPresetsCustomPage(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsCustomPage(pb, TestDB)

	cases := []TestCase{
		{
			Name:  "custom page basic",
			Debug: true,
			ReqFunc: func() *http.Request {
				customPageData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/custom", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`New Custom Page`},
		},
		{
			Name:  "custom page with id",
			Debug: true,
			ReqFunc: func() *http.Request {
				customPageData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/custom/12?name=vuetify", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`New Custom Page Param`, `12`, `vuetify`},
		},
		{
			Name:  "custom page with menu",
			Debug: true,
			ReqFunc: func() *http.Request {
				customPageData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/custom-menu", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`New Custom Page Show Menu`},
		},
		{
			Name:  "detail page with custom page buttons",
			Debug: true,
			ReqFunc: func() *http.Request {
				customPageData.TruncatePut(SqlDB)
				return NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				`NewCustomPageShowMenu`,
				`NewCustomPage`,
				`NewCustomPageByParam`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
