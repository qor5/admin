package examples_presets

import (
	"net/http"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
)

func TestPresetsEditingTabController(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingTabController(pb, TestDB)

	cases := []TestCase{
		{
			Name:  "editing create",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return NewMultipartBuilder().
					PageURL("/customers").
					EventFunc(actions.Edit).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`tab:1`, `vx-tabs underline-border='full'`, "t1", "t2"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
