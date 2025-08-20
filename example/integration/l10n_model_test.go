package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
)

var l10nModelTestData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.l10n_models (id, created_at, updated_at, deleted_at, title, locale_code) VALUES (1, '2025-07-16 06:31:07.828871 +00:00', '2025-07-16 06:31:07.828871 +00:00', null, '123', 'Japan');

`, []string{"l10n_models"}))

func TestL10nModel(t *testing.T) {
	h := admin.TestHandler(TestDB, nil)
	dbr, _ := TestDB.DB()

	cases := []TestCase{
		{
			Name:  "Index L10n Model",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nModelTestData.TruncatePut(dbr)
				return httptest.NewRequest("GET", "/l-10-n-models", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"New", "123"},
		},
		{
			Name:  "L10n Edit",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nModelTestData.TruncatePut(dbr)
				req := NewMultipartBuilder().
					PageURL("/l-10-n-models").
					Query(presets.ParamID, "1_Japan").
					EventFunc(actions.Update).
					AddField("Title", "234").
					AddField("LocaleCode", "Japan").
					BuildEventFuncRequest()
				return req
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				m := models.L10nModel{}
				TestDB.First(&m, 1)
				if m.Title != "234" {
					t.Errorf("Expected title to be '234', got '%s'", m.Title)
				}
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, h)
		})
	}
}
