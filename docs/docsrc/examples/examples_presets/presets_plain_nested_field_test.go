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
				insert into customers (id, email,name) values (1, 'xxx@gmail.com','Terry');
				insert into credit_cards (id, customer_id, number) values (1, 1, '1234567890');
			`, []string{"customers", "credit_cards"}))

func TestPresetsPlainNestedField(t *testing.T) {
	pb := presets.New().DataOperator(gorm2op.DataOperator(TestDB))
	PresetsPlainNestedField(pb, TestDB)

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
			ExpectPageBodyNotContains:          []string{"Notes"},
			ExpectPortalUpdate0ContainsInOrder: []string{`<label class="v-label theme--light text-caption">Notes</label>`},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, pb)
		})
	}
}
