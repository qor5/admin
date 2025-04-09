package examples_presets

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

var customerDataWithNumberRecord = gofixtures.Data(gofixtures.Sql(`
	INSERT INTO public.plain_nested_bodies (id, created_at, updated_at, deleted_at, name, items) VALUES (1, '2025-04-09 03:42:47.416003 +00:00', '2025-04-09 03:42:47.416003 +00:00', null, '123', null);
			`, []string{"plain_nested_bodies"}))

func TestPresetsPlainNestedField(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsPlainNestedField(pb, TestDB)

	cases := []TestCase{
		{
			Name:  "PlainNestedBody Update",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerDataWithNumberRecord.TruncatePut(SqlDB)
				return NewMultipartBuilder().
					PageURL("/plain-nested-bodies").
					EventFunc(actions.Update).
					Query(presets.ParamID, "1").
					AddField("Items[0].Number", "123").
					AddField("Items[0].Name", "234").
					BuildEventFuncRequest()
			},
			ResponseMatch: func(t *testing.T, w *httptest.ResponseRecorder) {
				m := PlainNestedBody{}
				TestDB.First(&m, 1)
				if m.Items == nil {
					t.Fatalf("Items is nil")
					return
				}
				if len(m.Items) == 0 {
					t.Fatalf("Number card is empty")
					return
				}
				if m.Items[0].Number != "123" {
					t.Fatalf("Number card is not 123")
					return
				}
				if m.Items[0].Name != "234" {
					t.Fatalf("Name card is not 234")
					return
				}
				return
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
