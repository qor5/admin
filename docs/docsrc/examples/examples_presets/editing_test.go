package examples_presets

import (
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
)

var companyData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.companies (id, name) VALUES (12, 'terry_company');
`, []string{"companies"}))

func TestPresetsEditingValidate(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingValidate(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "editing create",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"timeout='2000'", "name must not be empty"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}

func TestPresetsEditingSetter(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsEditingSetter(pb, TestDB)

	cases := []multipartestutils.TestCase{
		{
			Name:  "default field setterFunc",
			Debug: true,
			ReqFunc: func() *http.Request {
				companyData.TruncatePut(SqlDB)
				return multipartestutils.NewMultipartBuilder().
					PageURL("/companies?__execute_event__=presets_Update").
					AddField("Name", "").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{"name must not be empty"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, pb)
		})
	}
}
