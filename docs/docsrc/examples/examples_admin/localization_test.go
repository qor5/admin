package examples_admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testingutils"
)

var l10nData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.l10n_models (id, created_at, updated_at, deleted_at, title, locale_code) VALUES (1, 
'2024-06-04 23:27:40.442281 +00:00', '2024-06-04 23:27:40.442281 +00:00', null, 'My model title', 'Japan');


`, []string{"l10n_models"}))

var l10nDataWithChina = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.l10n_models (id, created_at, updated_at, deleted_at, title, locale_code) VALUES (1, 
'2024-06-04 23:27:40.442281 +00:00', '2024-06-04 23:27:40.442281 +00:00', null, 'My model title', 'Japan');
INSERT INTO public.l10n_models (id, created_at, updated_at, deleted_at, title, locale_code) VALUES (1, '2024-06-04 23:50:28.847833 +00:00', '2024-06-04 23:50:28.844069 +00:00', null, '中文标题', 'China');


`, []string{"l10n_models"}))

func TestLocalization(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	LocalizationExample(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/l10n-models", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"My model title", "Japan"},
		},
		{
			Name:  "Index Page with locale code",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nDataWithChina.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/l10n-models?locale=China", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"中文标题"},
		},
		{
			Name:  "Show detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nDataWithChina.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/l10n-models?__execute_event__=presets_Edit&id=1_China").
					BuildEventFuncRequest()
				return req
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"中文标题"},
		},
		{
			Name:  "Update detail",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nDataWithChina.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/l10n-models?__execute_event__=presets_Update&id=1_China").
					AddField("Title", "Updated Title").
					AddField("LocaleCode", "China").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var m L10nModel
				TestDB.Find(&m, "id = ? AND locale_code = ?", 1, "China")
				if m.Title != "Updated Title" {
					t.Errorf("title is wrong %#+v", m)
				}
			},
		},
		{
			Name:  "Delete China locale",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nDataWithChina.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/l10n-models?__execute_event__=presets_DoDelete&id=1_China").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var m L10nModel
				TestDB.Unscoped().Find(&m, "id = ? AND deleted_at IS NOT NULL", 1)
				if !strings.Contains(m.LocaleCode, "del") {
					t.Errorf("delete is wrong %#+v", m)
				}
			},
		},
		{
			Name:  "Localize to China and Japan",
			Debug: true,
			ReqFunc: func() *http.Request {
				l10nData.TruncatePut(SqlDB)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/l10n-models?__execute_event__=l10n_DoLocalizeEvent&id=1_Japan&localize_from=Japan").
					AddField("localize_to", "China").
					BuildEventFuncRequest()
				return req
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var localeCodes []string
				TestDB.Raw("SELECT locale_code FROM l10n_models ORDER BY locale_code").Scan(&localeCodes)
				if diff := testingutils.PrettyJsonDiff(
					[]string{"China", "Japan"},
					localeCodes); diff != "" {
					t.Error(diff)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
