package examples_presets

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3/multipartestutils"
)

func TestPresetsConfirmDialog(t *testing.T) {
	pb := presets.New()
	PresetsConfirmDialog(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/confirm-dialog", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Delete File"},
		},
		{
			Name:  "OpenConfirmDialog",
			Debug: true,
			ReqFunc: func() *http.Request {
				values := url.Values{
					"__execute_event__":                  {"presets_ConfirmDialog"},
					"presets_ConfirmDialog_ConfirmEvent": {"alert(\"file deleted\")"},
					"presets_ConfirmDialog_PromptText":   {"Are you sure you want to delete this file?"},
				}
				return multipartestutils.NewMultipartBuilder().
					PageURL("/confirm-dialog?" + values.Encode()).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`Confirm`, `Are you sure you want to delete this file?`, `alert("file deleted")`},
		},
		{
			Name:  "OpenConfirmDialogWithCustomizeTitle",
			Debug: true,
			ReqFunc: func() *http.Request {
				values := url.Values{
					"__execute_event__":                  {"presets_ConfirmDialog"},
					"presets_ConfirmDialog_ConfirmEvent": {"alert(\"file deleted\")"},
					"presets_ConfirmDialog_PromptText":   {"Are you sure you want to delete this file?"},
					"presets_ConfirmDialog_TitleText":    {"customed title"},
				}
				return multipartestutils.NewMultipartBuilder().
					PageURL("/confirm-dialog?" + values.Encode()).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`customed title`, `Are you sure you want to delete this file?`, `alert("file deleted")`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
