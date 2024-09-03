package examples_presets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestPresetsListingKeywordSearchOff(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsKeywordSearchOff(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page with keyword",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers?keyword=thisismykeyword", nil)
			},
			ExpectPageBodyNotContains: []string{`model-value='"thisismykeyword"'`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsRowMenuIcon(t *testing.T) {

	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsRowMenuAction(pb, TestDB)
	TestDB.AutoMigrate(&CreditCard{})
	cases := []multipartestutils.TestCase{
		{
			Name:  "row menu with no icon",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return httptest.NewRequest("GET", "/customers?__execute_event__=__reload__", nil)
			},
			ExpectPageBodyContainsInOrder: []string{`:icon='\"mdi-close\"'\u003e\u003c/v-icon\u003e\n\u003c/template\u003e\n\n\u003cv-list-item-title\u003ewith-icon\u003c/v-list-item-title\u003e`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsListingCustomizationFields(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsListingCustomizationFields(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "WrapColumns",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers", nil)
			},
			ExpectPageBodyContainsInOrder: []string{`min-width: 123px; color: red;`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsListingCustomizationBulkActionsLabelI18n(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsListingCustomizationBulkActions(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "CN button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=__reload__").
					Query("lang", "zh-Hans").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`审批`},
		},
		{
			Name:  "EN button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=__reload__").
					Query("lang", "en").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`Approve`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
