package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"github.com/theplant/testenv"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/examples"
)

var TestDB *gorm.DB

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	TestDB = env.DB
	m.Run()
}

var customerData = gofixtures.Data(gofixtures.Sql(`
				insert into customers (id, name) values (11, 'Felix1');
			`, []string{"customers"}))

var productData = gofixtures.Data(gofixtures.Sql(`
				insert into preset_products (id, name) values (12, 'Product 1');
			`, []string{"preset_products"}))

var (
	emptyCustomerData = gofixtures.Data(gofixtures.Sql(``, []string{"customers"}))
	creditCardData    = gofixtures.Data(customerData, gofixtures.Sql(``, []string{"credit_cards"}))
)

func TestExample(t *testing.T) {
	db := TestDB
	dbr, _ := db.DB()
	p := examples.Preset1(db)
	var handler http.Handler = p

	cases := []TestCase{
		{
			Name: "Update",
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/my_customers").
					EventFunc(actions.Update).
					Query(presets.ParamID, "11").
					AddField("Bool1", "true").
					AddField("ID", "11").
					AddField("Int1", "42").
					AddField("Name", "Felix11").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				u := &examples.Customer{}
				err := db.Find(u, 11).Error
				if err != nil {
					t.Error(err)
				}
				if u.Name != "Felix11" {
					t.Error(u)
				}
			},
		},
		{
			Name: "Create",
			ReqFunc: func() *http.Request {
				emptyCustomerData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/my_customers").
					EventFunc(actions.Update).
					AddField("Bool1", "true").
					AddField("ID", "").
					AddField("Int1", "42").
					AddField("Name", "Felix").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				u := &examples.Customer{}
				err := db.First(u).Error
				if err != nil {
					t.Error(err)
				}
				if u.Name != "Felix" {
					t.Error(u)
				}
			},
		},

		{
			Name: "New Form For Creating",
			ReqFunc: func() *http.Request {
				emptyCustomerData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/credit-cards").
					EventFunc(actions.New).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`v-model='form["Number"]`, `v-assign='[form, {"Number":""`},
		},

		{
			Name: "Create CreditCard",
			ReqFunc: func() *http.Request {
				creditCardData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/credit-cards").
					EventFunc(actions.Update).
					Query("customerID", "11").
					AddField("Number", "12345678").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				u := &examples.CreditCard{}
				err := db.First(u).Error
				if err != nil {
					t.Error(err)
				}
				if u.Number != "12345678" {
					t.Error(u)
				}
			},
		},

		{
			Name: "Without Editing Config/Product Edit Form",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/products").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{`v-model='form["OwnerName"]`, `v-assign='[form, {"OwnerName":""`},
		},

		{
			Name: "Without Editing Config/Create Product",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/products").
					EventFunc(actions.Update).
					Query(presets.ParamID, "12").
					AddField("OwnerName", "owner1").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				u := &examples.Product{}
				err := db.First(u).Error
				if err != nil {
					t.Error(err)
				}
				if u.OwnerName != "owner1" {
					t.Error(u)
				}
			},
		},

		{
			Name: "formDrawerAction AgreeTerms",
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/my_customers/11").
					EventFunc(actions.Action).
					Query(presets.ParamAction, "AgreeTerms").
					Query(presets.ParamID, "11").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				partial := er.UpdatePortals[0].Body
				if !strings.Contains(partial, `v-model='form["Agree"]' v-assign='[form, {"Agree":""}]'`) {
					t.Error(`can't find v-model='form["Agree"]' v-assign='[form, {"Agree":""}]'`, partial)
				}
			},
		},

		{
			Name: "doAction AgreeTerms",
			ReqFunc: func() *http.Request {
				customerData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/my_customers/11").
					EventFunc(actions.DoAction).
					Query(presets.ParamAction, "AgreeTerms").
					Query(presets.ParamID, "11").
					AddField("Agree", "true").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				u := &examples.Customer{}
				err := db.First(u).Error
				if err != nil {
					t.Error(err)
				}
				if u.TermAgreedAt == nil {
					t.Error(fmt.Sprintf("%#+v", u))
				}
			},
		},

		{
			Name: "Create Products Observe Change",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/products").
					EventFunc(actions.New).
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"Object.values(vars.presetsDataChanged)",
				"vx-dialog", "If you leave before submitting the form, you will lose all the unsaved input.",
			},
		},
		{
			Name: "Edit Products  Observe Change",
			ReqFunc: func() *http.Request {
				productData.TruncatePut(dbr)
				return NewMultipartBuilder().
					PageURL("/admin/products").
					EventFunc(actions.Edit).
					Query("id", "12").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"Object.values(vars.presetsDataChanged)",
				"vx-dialog", "If you leave before submitting the form, you will lose all the unsaved input.",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, handler)
		})
	}
}
