package examples_vuetifyx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/web/v3/examples"
	"github.com/qor5/web/v3/multipartestutils"
)

func Handler() http.Handler {
	mux := http.NewServeMux()
	im := &examples.IndexMux{Mux: http.NewServeMux()}
	SamplesHandler(im)
	mux.Handle("/examples/", im.Mux)
	return mux
}

func TestFilterDemo(t *testing.T) {
	mux := Handler()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Index Page with keyword",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/examples/filter-demo", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{
				`vx-filter`,
				`DatetimeRangeItem`,
				`DateItem`,
				`DateRangeItem`,
				`DatetimeRangePickerItem`,
				`DateRangePickerItem`,
				`DatePickerItem`,
				`SelectItem`,
				`NumberItem`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, mux)
		})
	}
}
