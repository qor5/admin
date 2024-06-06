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

INSERT INTO public.notes (id, source_type, source_id, content, created_at, updated_at) VALUES (1, 'Customer', 12, 
'This is my note 1', '2024-05-27 08:13:58.436186 +00:00', '2024-05-27 08:13:58.436186 +00:00');

`, []string{"customers", "credit_cards", "notes"}))

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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Edit&detailField=Details&id=12").
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
					PageURL("/customers?__execute_event__=presets_Detailing_Field_Save&detailField=Details&id=12").
					AddField("Details.Name", "123123").
					AddField("Details.Email", "abc@example.com").
					AddField("Details.Description", "hello description").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"abc@example.com"},
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
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
