package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestPresetsBasicFilter(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsBasicFilter(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "index",
			Debug: true,
			ReqFunc: func() *http.Request {
				return multipartestutils.NewMultipartBuilder().
					PageURL("/posts").
					BuildEventFuncRequest()
			},
			ExpectPageBodyContainsInOrder: []string{"Status", "titleNoChoose", "warpBody"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
