package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Edit&section=Details&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Details.Name", "Details.Email"},
		},
		{
			Name:  "page detail update",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&section=Details&id=12").
					AddField("Details.Name", "123123").
					AddField("Details.Email", "abc@example.com").
					AddField("Details.Description", "hello description").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"123123", "abc@example.com", "hello description"},
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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save"+
						"&id=12").
					Query("section", "section").
					Query("isCancel", "true").
					AddField("section.Name", "name").
					AddField("section.Email", "email").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"SectionEN"},
			ExpectPortalUpdate0NotContains:     []string{"Wrong"},
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
			ExpectPortalUpdate0ContainsInOrder: []string{"Felix 1"},
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
			Name:  "section validate globe err",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&section=name_section&id=12").
					AddField("name_section.Name", "").
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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&section=name_section&id=12").
					AddField("name_section.Name", "longlonglong name").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"customer name must no longer than 6"},
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
					Query("__execute_event__", "presets_Detailing_List_Field_Save").
					Query("section", "CreditCards").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "12").
					Query("isCancel", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"mdi-square-edit-outline"},
			ExpectPortalUpdate0NotContains:     []string{"Save"},
		},
		{
			Name:  "cancel section",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers").
					Query("__execute_event__", "presets_Detailing_Field_Save").
					Query("section", "section2").
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
INSERT INTO public.user_credit_cards (id, created_at, updated_at, deleted_at, name, credit_cards) VALUES (1, '2024-08-21 07:14:43.822238 +00:00', '2024-08-22 03:18:34.044182 +00:00', null, 'empty date', '[]');
INSERT INTO public.user_credit_cards (id, created_at, updated_at, deleted_at, name, credit_cards) VALUES (2, '2024-08-21 07:14:43.822238 +00:00', '2024-08-22 03:29:30.597570 +00:00', null, 'one card', '[{"ID":0,"CustomerID":0,"Number":"","ExpireYearMonth":"","Name":"terry","Type":"","Phone":"188","Email":""}]');
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
			ExpectPortalUpdate0ContainsInOrder: []string{"cards", "Add Row"},
		},
		{
			Name:  "click Add Row button",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Create").
					Query("section", "CreditCards").
					Query("id", "1").
					Query("sectionListUnsaved_CreditCards", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "Phone", "Cancel", "Save"},
			ExpectPortalUpdate0NotContains:     []string{"Add Row"},
		},
		{
			Name:  "save created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Save").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("section", "CreditCards").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Name", "terry", "Phone", "188", "Add Row"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		{
			Name:  "cancel created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Save").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("section", "CreditCards").
					Query("sectionListSaveBtn_CreditCards", "0").
					Query("id", "1").
					Query("isCancel", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Row"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save"},
		},
		{
			Name:  "delete created section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Delete").
					Query("sectionListUnsaved_CreditCards", "false").
					Query("section", "CreditCards").
					Query("sectionListDeleteBtn_CreditCards", "0").
					Query("id", "1").
					AddField("CreditCards[0].Name", "terry").
					AddField("CreditCards[0].Phone", "188").
					AddField("__Deleted_CreditCards[0].sectionListEditing", "true").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"Add Row"},
			ExpectPortalUpdate0NotContains:     []string{"Cancel", "Save", "terry", "188"},
		},
		{
			Name:  "delete section, have created but unsaved section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Delete").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("section", "CreditCards").
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
			ExpectPortalUpdate0NotContains:     []string{"terry", "188", "Add Row"},
		},
		{
			Name:  "cancel section, have created but unsaved section",
			Debug: true,
			ReqFunc: func() *http.Request {
				userCreditCardsData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/user-credit-cards").
					Query("__execute_event__", "presets_Detailing_List_Field_Save").
					Query("sectionListUnsaved_CreditCards", "true").
					Query("section", "CreditCards").
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
			ExpectPortalUpdate0NotContains:     []string{"joy", "177", "Add Row"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
