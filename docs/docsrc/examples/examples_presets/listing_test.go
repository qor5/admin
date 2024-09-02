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

// var detailData = gofixtures.Data(gofixtures.Sql(`
// INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
// term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
// '2024-03-28 05:52:28.497536 +00:00', null, null, '');

// `, []string{"customers", "credit_cards"}))

func TestPresetsRowMenuIcon(t *testing.T) {

	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsRowMenuAction(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page with keyword",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers?keyword=thisismykeyword", nil)
			},
			ExpectPageBodyNotContains: []string{`<div class="v-list-item__prepend"><i class="v-icon notranslate v-theme--light v-icon--size-default" aria-hidden="true"><!----></i><div class="v-list-item__spacer"></div></div><div class="v-list-item__content" data-no-activator=""><!----><!----><div class="v-list-item-title">without icon</div></div>`},
			ExpectPageBodyContainsInOrder: []string{`<div class="v-list-item__prepend"><i class="mdi-delete mdi v-icon notranslate v-theme--light v-icon--size-default" aria-hidden="true"></i><div class="v-list-item__spacer"></div></div><div class="v-list-item__content" data-no-activator=""><!----><!----><div class="v-list-item-title">with-icon</div></div>`},
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
