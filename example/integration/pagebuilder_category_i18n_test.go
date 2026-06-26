package integration_test

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/presets/actions"
)

// TestPageCategoryI18n covers KGM-3903:
// In the Japanese UI, the Page Categories "New" form rendered its field labels
// (Name / Path / Description) in English, because the field-level translations
// (PageCategoriesName / PageCategoriesPath / PageCategoriesDescription) were not
// registered under presets.ModelsI18nModuleKey, so i18n.PT fell back to the
// English humanized field names.
func TestPageCategoryI18n(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Category New form in Japanese shows Japanese field labels",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.New).
					BuildEventFuncRequest()
				req.Header.Set("Accept-Language", "ja")
				return req
			},
			// Field labels must be the Japanese translations, in editing field order.
			ExpectPortalUpdate0ContainsInOrder: []string{"名前", "パス", "説明"},
		},
		{
			Name:  "Category New form in English shows English field labels",
			Debug: true,
			ReqFunc: func() *http.Request {
				pageBuilderData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/page_categories").
					EventFunc(actions.New).
					BuildEventFuncRequest()
				req.Header.Set("Accept-Language", "en")
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Path", "Description"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
