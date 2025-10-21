package examples_presets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func TestPresetsListingKeywordSearchOff(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsKeywordSearchOff(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page with keyword",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers?keyword=thisismykeyword", http.NoBody)
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
				return httptest.NewRequest("GET", "/customers?__execute_event__=__reload__", http.NoBody)
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
				return httptest.NewRequest("GET", "/customers", http.NoBody)
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

func TestPresetsListingCustomizationFilters(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsListingCustomizationFilters(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "DateOptions",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`StartAt`, "EndAt", "Approved_Start_At", "Cancel1", "Approved_End_At"},
		},
		{
			Name:  "DateOptions Filter Validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/customers?f_created.gte=2024-09-13%2000%3A00&f_created.lt=2024-09-12%2000%3A00", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{`CreatedAt Error`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

var listingDatatableData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
'2024-03-28 05:52:28.497536 +00:00', null, null, '');
`, []string{"customers"}))

func TestPresetsListingDatatableFunc(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsCustomPage(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "custom page basic",
			Debug: true,
			ReqFunc: func() *http.Request {
				listingDatatableData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`v-card`, `Felix 1`, `abc@example.com`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsListingFilterNotificationFunc(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsListingFilterNotificationFunc(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "Filter Notification",
			Debug: true,
			ReqFunc: func() *http.Request {
				listingDatatableData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`Filter Notification`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDataOperatorWithGRPC(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDataOperatorWithGRPC(pb, TestDB)
	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Customer",
			Debug: true,
			ReqFunc: func() *http.Request {
				listingDatatableData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{`v-card`, `Felix 1`, `abc@example.com`},
		},
		{
			Name:  "Update Customer",
			Debug: true,
			ReqFunc: func() *http.Request {
				listingDatatableData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query(presets.ParamID, "12").
					EventFunc(actions.Update).
					AddField("Name", "system").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"invalid argument"},
		},
		{
			Name:  "Delete Customer",
			Debug: true,
			ReqFunc: func() *http.Request {
				listingDatatableData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query(presets.ParamID, "12").
					EventFunc(actions.DoDelete).
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"PresetsNotifModelsDeletedexamplesPresetsCustomer"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
