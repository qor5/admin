package examples_presets

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var detailData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
'2024-03-28 05:52:28.497536 +00:00', null, null, '');

INSERT INTO public.credit_cards (id, customer_id, number, expire_year_month, name, type, phone, email) VALUES (2, 12,
'95550012', '', '', '', '', '');
`, []string{"customers", "credit_cards"}))

func TestPresetsDetailing(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailInlineEditDetails(pb, TestDB)

	pb1 := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailInlineEditFieldSections(pb1, TestDB)

	pb2 := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailPageCards(pb2, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail page show for completely customized",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb2
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Felix 1"},
		},

		{
			Name:  "page detail show for switchable",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer" +
						"&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Felix 1"},
		},

		{
			Name:  "page detail edit",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb1
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_edit_Details&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Email"},
		},
		{
			Name:  "page detail update",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_Details&id=12").
					AddField("Name", "123123").
					AddField("Email", "abc@example.com").
					AddField("Description", "hello description").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"123123", "abc@example.com", "hello description"},
		},
		{
			Name:  "page detail reload",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_reload_Details&id=12").
					AddField("Name", "123123").
					AddField("Email", "abc@example.com").
					AddField("Description", "hello description").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Edit", "Felix 1", "abc@example.com"},
			ExpectPortalUpdate0NotContains:     []string{"123123", "hello description", "Cancel", "Save"},
		},

		{
			Name:  "page detail show for field sections",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb1
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer" +
						"&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Felix 1", "<strong>abc@example.com</strong>"},
		},
		{
			Name:  "field section title i18n",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb1
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer" +
						"&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Field_sectionEN", "SectionEN"},
			ExpectPortalUpdate0NotContains:     []string{"Wrong"},
		},
		{
			Name:  "cancel edit section title i18n",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb1
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_section&id=12").
					Query("isCancel", "true").
					AddField("Name", "name").
					AddField("Email", "email").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"SectionEN"},
			ExpectPortalUpdate0NotContains:     []string{"Wrong"},
		},
		// PresetsDetailInlineEditDetails - Creating scenarios
		{
			Name:  "new form with section",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_New").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Details", "Name", "Email", "Description", "Avatar"},
		},
		{
			Name:  "create with section - save successfully",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update").
					AddField("Name", "New Customer").
					AddField("Email", "new@example.com").
					AddField("Description", "New description").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`"Name":"New Customer"`, `"Email":"new@example.com"`, `"Description":"New description"`, "Successfully Created"},
		},
		{
			Name:  "edit existing with section - open edit drawer",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Edit&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Details", "Felix 1", "abc@example.com"},
		},
		{
			Name:  "edit existing with section - update successfully",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return pb
			},
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Update&id=12").
					AddField("Name", "Updated Name").
					AddField("Email", "updated@example.com").
					AddField("Description", "Updated description").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`"Name":"Updated Name"`, `"Email":"updated@example.com"`, `"Description":"Updated description"`, "Successfully Updated"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

var detailNestedManyData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
'2024-03-28 05:52:28.497536 +00:00', null, null, '');

INSERT INTO public.credit_cards (id, customer_id, number, expire_year_month, name, type, phone, email) VALUES (2, 12,
'95550012', '', '', '', '', '');

INSERT INTO public.notes (id, source_type, source_id, content, created_at, updated_at) VALUES (1, 'Customer', 12, 
'This is my note 1', '2024-05-27 08:13:58.436186 +00:00', '2024-05-27 08:13:58.436186 +00:00');

`, []string{"customers", "credit_cards", "notes"}))

func TestPresetsDetailNestedMany(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailNestedMany(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "page detail show",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer" +
						"&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Felix 1", ":hover='true'", "95550012", ":hover='false'", "95550012"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

var seedDetailActionsComponent = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.customers (id, name, email, description, company_id, created_at, updated_at, approved_at, 
term_agreed_at, approval_comment) VALUES (12, 'Felix 1', 'abc@example.com', '', 0, '2024-03-28 05:52:28.497536 +00:00', 
'2024-03-28 05:52:28.497536 +00:00', null, null, '');

INSERT INTO public.credit_cards (id, customer_id, number, expire_year_month, name, type, phone, email) VALUES (2, 12,
'95550012', '', '', '', '', '');

INSERT INTO public.notes (id, source_type, source_id, content, created_at, updated_at) VALUES (1, 'Customer', 12, 
'This is my note 1', '2024-05-27 08:13:58.436186 +00:00', '2024-05-27 08:13:58.436186 +00:00');

`, []string{"customers", "credit_cards", "notes"}))

func TestPresetsDetailActionsComponent(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailPageDetails(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "page detail show",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{">Agree Terms</v-btn>", ">Add Note</v-btn>"},
		},
		{
			Name:  "click agree terms action button",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12?__execute_event__=presets_Action&action=AgreeTerms&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`<v-checkbox`, `label='Agree the terms'></v-checkbox>`},
		},
		{
			Name:  "click action not exists",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12?__execute_event__=presets_Action&action=NotExists&id=12").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`cannot find requested action`},
		},
		{
			Name:  "do action not exists",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12?__execute_event__=presets_DoAction&action=NotExists&id=12").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`cannot find requested action`},
		},
		{
			Name:  "agree terms",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12?__execute_event__=presets_DoAction&action=AgreeTerms&id=12").
					AddField("Agree", "true").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{`emit("PresetsNotifModelsUpdatedexamplesPresetsCustomer"`, `["12"]`},
		},
		{
			Name:  "agree terms with false",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/12?__execute_event__=presets_DoAction&action=AgreeTerms&id=12").
					AddField("Agree", "false").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`You must agree the terms`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSectionValidate(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailInlineEditValidate(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "section validate name_section field error with VFieldError",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_name_section&id=12").
					AddField("Name", "long name exceeds 6 chars").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"v-text-field",
				`dash.errorMessages["Name"]`,
				"customer name must no longer than 6",
			},
		},
		{
			Name:  "section validate email_section field error with VFieldError",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_email_section&id=12").
					AddField("Email", "short").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"v-text-field",
				`dash.errorMessages["Email"]`,
				"customer email must longer than 6",
			},
		},
		{
			Name:  "section validate globe err",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_name_section&id=12").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"message: \"customer name must not be empty\""},
		},
		{
			Name:  "section validate field err",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_name_section&id=12").
					AddField("Name", "long name").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"customer name must no longer than 6"},
		},
		{
			Name:  "section validate field err with globe err",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_name_section&id=12").
					AddField("Name", "long long long long name").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"customer name must no longer than 6"},
			ExpectRunScriptContainsInOrder:     []string{"customer name must no longer than 20"},
		},

		{
			Name:  "section validate globe err with custom saveFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_email_section&id=12").
					AddField("Email", "").
					AddField("Name", "name").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"message: \"customer email must not be empty\""},
		},
		{
			Name:  "section validate field err with custom saveFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=section_save_email_section&id=12").
					AddField("Email", "short").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"customer email must longer than 6"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSectionLabel(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailSectionLabel(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "section label",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"section_with_label", "section_list_with_label"},
			ExpectPortalUpdate0NotContains:     []string{"section_without_label", "section_list_without_label"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSectionCancel(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailSectionLabel(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "cancel section list",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "12").
					Query("isCancel", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"mdi-pencil-outline"},
			ExpectPortalUpdate0NotContains:     []string{"Save"},
		},
		{
			Name:  "cancel section",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "section_save_section2").
					Query("id", "12").
					Query("isCancel", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Edit"},
			ExpectPortalUpdate0NotContains:     []string{"Save"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

var userCreditCardsData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.user_credit_cards (id, created_at, updated_at, deleted_at, name, credit_cards,credit_cards2) VALUES (1, '2024-08-21 07:14:43.822238 +00:00', '2024-08-22 03:18:34.044182 +00:00', null, 'empty date', '[]','[]');
INSERT INTO public.user_credit_cards (id, created_at, updated_at, deleted_at, name, credit_cards,credit_cards2) VALUES (2, '2024-08-21 07:14:43.822238 +00:00', '2024-08-22 03:29:30.597570 +00:00', null, 'one card', '[{"ID":0,"CustomerID":0,"Number":"","ExpireYearMonth":"","Name":"terry","Type":"","Phone":"188","Email":""}]','[]');
`, []string{"user_credit_cards"}))

func TestPresetsDetailListSection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailListSection(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "display section list",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_DetailingDrawer").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"cards", "Add Item", "cards2", "Add Item"},
		},
		{
			Name:  "click Add Row button",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards").
					Query("id", "1").
					Query("sectionListUnsaved_CreditCards", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Phone", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		{
			Name:  "save created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "terry", "Phone", "188", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		{
			Name:  "cancel created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					Query("isCancel", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		{
			Name:  "delete created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_delete_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListDeleteBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save", "terry", "188"},
		},
		{
			Name:  "delete section, have created but unsaved section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_delete_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListDeleteBtn_CreditCards", "0").
					Query("id", "2").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					AddField("CreditCards[1].Name", "tom").
					AddField("CreditCards[1].Phone", "199").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"tom", "199", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"terry", "188", "Add Item"},
		},
		{
			Name:  "reload section, have created but unsaved section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_reload_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "2").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					AddField("CreditCards[1].Name", "tom").
					AddField("CreditCards[1].Phone", "199").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry", "188", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},

		{
			Name:  "cancel section, have created but unsaved section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "2").
					Query("isCancel", "true").
					AddField("CreditCards[0].Name", "joy").
					AddField("CreditCards[0].Phone", "177").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					AddField("CreditCards[1].Name", "tom").
					AddField("CreditCards[1].Phone", "199").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry", "188", "tom", "199", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"joy", "177", "Add Item"},
		},
		{
			Name:  "click Add Row button 2",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards2").
					Query("id", "1").
					Query("sectionListUnsaved_CreditCards2", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Phone", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		{
			Name:  "save created section 2",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards2").
					Query("sectionListUnsaved_CreditCards2", "false").
					Query("sectionListSaveBtn_CreditCards2", "0").
					Query("id", "1").
					AddField("CreditCards2[0].Name", "terry").
					AddField("CreditCards2[0].Phone", "188").
					AddField("__Deleted_CreditCards2[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "terry", "Phone", "188", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		{
			Name:  "edit section 2",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_edit_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListEditBtn_CreditCards", "0").
					Query("id", "2").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Cancel", "Save", "This is hidden"},
		},
		{
			Name:  "save with empty name - validation error",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"card name must not be empty", "Cancel", "Save"},
		},
		{
			Name:  "save with name exceeds 10 chars - validation error",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "verylongname").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"card name must not exceed 10 characters", "Cancel", "Save"},
		},
		{
			Name:  "save CreditCards2 with empty phone - validation error",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards2").
					Query("sectionListUnsaved_CreditCards2", "false").
					Query("sectionListSaveBtn_CreditCards2", "0").
					Query("id", "1").
					AddField("CreditCards2[0].Name", "terry").
					AddField("CreditCards2[0].Phone", "").
					AddField("__Deleted_CreditCards2[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"card phone must not be empty", "Cancel", "Save"},
		},
		{
			Name:  "save with valid data - success",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "validname").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"validname", "188", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save", "must not be empty", "must not exceed"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailListSection_StatusxFieldViolations(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailListSectionStatusxFieldViolations(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "list section save returns validator field error and shows error",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			// Expect the field error text from BadRequest.FieldViolations to appear in the portal update
			ExpectPortalUpdate0ContainsInOrder: []string{"name is required"},
		},
		{
			Name:  "list section add should not re-init dash inside portal",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"dash.errorMessages"},
			ExpectPortalUpdate0NotContains:     []string{":dash-init"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailListSection_ItemStateIsolation(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailListSection_ItemStateIsolation(pb, TestDB)

	cases := []multipartestutils.TestCase{
		// 1) init data
		{
			Name:  "init data",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_DetailingDrawer").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"CreditCards", "Add Item"},
		},
		// 2) click add, expect empty component and no Add Item
		{
			Name:  "add item shows empty editor and hides add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"CreditCards[0].Name", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 3) save item with non-empty name, expect saved and Add Item reappears
		{
			Name:  "save created item with non-empty name",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		// 4) add again, expect empty editor and no Add Item
		{
			Name:  "add again shows empty editor and hides add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry", "CreditCards[1].Name", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 5) save second item with empty name -> expect validation error and no Add Item
		{
			Name:  "save second item empty name shows error and no add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "1").
					Query("id", "1").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name is required"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 6) click edit on first item; expect it enters edit mode
		{
			Name:  "edit first item while second has error keeps no add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_edit_CreditCards").
					Query("sectionListEditBtn_CreditCards", "0").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 7) cancel first item; expect no Add Item and second item still exists
		{
			Name:  "cancel first item keeps second item and no add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("isCancel", "true").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 8) edit first item again; expect same as step 6 (edit mode, no Add Item)
		{
			Name:  "edit first item again keeps no add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_edit_CreditCards").
					Query("sectionListEditBtn_CreditCards", "0").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 9) save first item; expect saved successfully, no Add Item, second item still exists
		{
			Name:  "save first item successfully keeps second item and no add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "first-saved").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"first-saved", "CreditCards[1].Name"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 10) cancel second item; expect Add Item shows
		{
			Name:  "cancel second item shows add button",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "1").
					Query("isCancel", "true").
					Query("id", "1").
					AddField("CreditCards[1].Name", "").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Item"},
		},
		// 11) add second item
		{
			Name:  "add second item",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_create_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"first-saved", "CreditCards[1].Name", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Item"},
		},
		// 12) save second item with name "second"
		{
			Name:  "save second item with name second",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("sectionListSaveBtn_CreditCards", "1").
					Query("id", "1").
					AddField("CreditCards[1].Name", "second").
					AddField("__Deleted_CreditCards[1].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"first-saved", "second", "Add Item"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		// 13) click edit on first item - editing existing item should NOT hide Add Item button
		{
			Name:  "edit first item",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_edit_CreditCards").
					Query("sectionListEditBtn_CreditCards", "0").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"CreditCards[0].Name", "Cancel", "Save", "second", "Add Item"},
		},
		// 14) save first item with empty name -> expect validation error and Add Item button still shows
		{
			Name:  "save first item empty name shows error and add button still shows",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "section_save_CreditCards").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name is required", "Add Item"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

var customerData = gofixtures.Data(gofixtures.Sql(`
				insert into customers (id, email,name) values (1, 'xxx@gmail.com','Terry');
			`, []string{"customers"}))

func TestPresetsDetailTabsField(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailTabsField(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail tabs field",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "presets_DetailingDrawer").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Email", "Name", "Terry", "Email", "xxx@gmail.com"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailTabsSection(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailTabsSection(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail tabs section display",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "presets_DetailingDrawer").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name", "email", `<v-tabs-window-item :value='"name"'>`, "Terry", `<v-tabs-window-item :value='"email"'>`, "xxx@gmail.com"},
		},
		{
			Name:  "detail tabs section save",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "section_save_name").
					Query("id", "1").
					AddField("Name", "terry1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry1"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailTabsSectionOrder(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailTabsSectionOrder(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail tabs section display",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "presets_DetailingDrawer").
					Query("id", "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"email", "name", `<v-tabs-window-item :value='"email"'>`, "xxx@gmail.com", `<v-tabs-window-item :value='"name"'>`, "Terry"},
		},
		{
			Name:  "detail tabs section save",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "section_save_name").
					Query("id", "1").
					AddField("Name", "terry1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"terry1"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailAfterTitle(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailAfterTitle(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail without drawer after title",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/1?__execute_event__=__reload__").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"After Title"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSidePanel(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailSidePanel(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail with drawer side panel",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/1?__execute_event__=presets_DetailingDrawer").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Side Panel 1", "Side Panel 2"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSectionView(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailSectionView(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail without drawer after title",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/1?__execute_event__=presets_DetailingDrawer").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"section-edit-area", "z-index:2"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailDisableSave(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailDisableSave(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "detail without drawer disable save",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers/1").
					EventFunc("section_edit_DisabledSection").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{fmt.Sprintf("dash.disabled.%s=true", presets.DisabledKeyButtonSave), fmt.Sprintf("dash.disabled.%s=false", presets.DisabledKeyButtonSave), "Savable"},
		},
		{
			Name:  "detail with drawer disable save",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies/1").
					EventFunc("section_edit_DisabledSectionCompany").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{fmt.Sprintf("dash.disabled.%s=true", presets.DisabledKeyButtonSave), fmt.Sprintf("dash.disabled.%s=false", presets.DisabledKeyButtonSave), "Savable"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsDetailSaverValidation(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsDetailSaverValidation(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name: "detail saver validation",
			ReqFunc: func() *http.Request {
				customPageData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					EventFunc("section_save_Customer").
					AddField("Name", "system").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"You can not use system as name"},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
