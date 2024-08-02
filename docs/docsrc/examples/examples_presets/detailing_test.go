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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Edit&detailField=field_section&id=12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"field_section.Name", "field_section.Email"},
		},
		{
			Name:  "page detail update",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&detailField=field_section&id=12").
					AddField("field_section.Name", "123123").
					AddField("field_section.Email", "abc@example.com").
					AddField("field_section.Description", "hello description").
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
			ExpectPortalUpdate0ContainsInOrder: []string{"Field_section_EN", "Section_EN"},
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
			Name:  "section validate",
			Debug: true,
			ReqFunc: func() *http.Request {
				detailData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&detailField=name_section&id=12").
					AddField("name_section.Name", "").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{"message: \"customer name must not be empty\""},
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
