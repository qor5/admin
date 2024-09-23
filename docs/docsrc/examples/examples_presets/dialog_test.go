package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	// "github.com/theplant/gofixtures"
)

func TestPresetsUtilsDialog(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsUtilsDialog(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "show vx dialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/customers?__execute_event__=presets_DetailingDrawer&id=1&var_current_active=__current_active_of_c512ed84__").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"vx-dialog", "type='error'", "title='Confirm'", "are you sure?"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
