package examples_presets

import (
	"net/http"
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
			Name:  "editing create",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerDataWithNumberRecord.TruncatePut(SqlDB)
				return NewMultipartBuilder().
					PageURL("/plain-nested-bodies").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
			},
			ExpectPageBodyNotContains:          []string{`v-card :variant='"outlined"' class='mx-0 mb-2 px-4 pb-0 pt-4'`},
			ExpectPortalUpdate0ContainsInOrder: []string{`v-btn :variant='"text"' color='primary' id='Items_1' :disabled='false' @click='plaid()`},
		},
		{
			Name:  "add row",
			Debug: true,
			ReqFunc: func() *http.Request {
				return NewMultipartBuilder().
					PageURL("/plain-nested-bodies").
					EventFunc("listEditor_addRowEvent").
					Query(presets.ParamID, "1").
					Query("listEditor_AddRowFormKey", "Items").
					Query("ItemsAddRowBtnID", "Items_1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`vx-field label='Name'`, `v-model='form["Items[0].Name"]'`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
