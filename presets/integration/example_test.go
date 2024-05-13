package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	examples2 "github.com/qor5/admin/v3/presets/examples"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/theplant/gofixtures"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

type reqCase struct {
	name               string
	reqFunc            func(db *sql.DB) *http.Request
	eventResponseMatch func(er *testEventResponse, db *gorm.DB, t *testing.T)
	pageMatch          func(body *bytes.Buffer, db *gorm.DB, t *testing.T)
}

var cases = []reqCase{
	{
		name: "Update",
		reqFunc: func(db *sql.DB) *http.Request {
			customerData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/my_customers").
				EventFunc(actions.Update).
				Query(presets.ParamID, "11").
				AddField("Bool1", "true").
				AddField("ID", "11").
				AddField("Int1", "42").
				AddField("Name", "Felix11").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			u := &examples2.Customer{}
			err := db.Find(u, 11).Error
			if err != nil {
				t.Error(err)
			}
			if u.Name != "Felix11" {
				t.Error(u)
			}
			return
		},
	},
	{
		name: "Create",
		reqFunc: func(db *sql.DB) *http.Request {
			emptyCustomerData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/my_customers").
				EventFunc(actions.Update).
				AddField("Bool1", "true").
				AddField("ID", "").
				AddField("Int1", "42").
				AddField("Name", "Felix").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			u := &examples2.Customer{}
			err := db.First(u).Error
			if err != nil {
				t.Error(err)
			}
			if u.Name != "Felix" {
				t.Error(u)
			}
			return
		},
	},

	{
		name: "New Form For Creating",
		reqFunc: func(db *sql.DB) *http.Request {
			emptyCustomerData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/credit-cards").
				EventFunc(actions.New).
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			partial := er.UpdatePortals[0].Body
			if strings.Index(partial, `v-model='form["Number"]' v-assign='[form, {"Number":""}]'`) < 0 {
				t.Error(`v-model='form["Number"]' v-assign='[form, {"Number":""}]'`, partial)
			}
			return
		},
	},

	{
		name: "Create CreditCard",
		reqFunc: func(db *sql.DB) *http.Request {
			creditCardData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/credit-cards").
				EventFunc(actions.Update).
				Query("customerID", "11").
				AddField("Number", "12345678").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			u := &examples2.CreditCard{}
			err := db.First(u).Error
			if err != nil {
				t.Error(err)
			}
			if u.Number != "12345678" {
				t.Error(u)
			}

			return
		},
	},

	{
		name: "Without Editing Config/Product Edit Form",
		reqFunc: func(db *sql.DB) *http.Request {
			productData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/products").
				EventFunc(actions.Edit).
				Query(presets.ParamID, "12").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			partial := er.UpdatePortals[0].Body
			if strings.Index(partial, `v-model='form["OwnerName"]' v-assign='[form, {"OwnerName":""}]'`) < 0 {
				t.Error(`can't find v-model='form["OwnerName"]' v-assign='[form, {"OwnerName":""}]'`, partial)
			}
			return
		},
	},

	{
		name: "Without Editing Config/Create Product",
		reqFunc: func(db *sql.DB) *http.Request {
			productData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/products").
				EventFunc(actions.Update).
				Query(presets.ParamID, "12").
				AddField("OwnerName", "owner1").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			u := &examples2.Product{}
			err := db.First(u).Error
			if err != nil {
				t.Error(err)
			}
			if u.OwnerName != "owner1" {
				t.Error(u)
			}

			return
		},
	},

	{
		name: "formDrawerAction AgreeTerms",
		reqFunc: func(db *sql.DB) *http.Request {
			customerData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/my_customers/11").
				EventFunc(actions.Action).
				Query(presets.ParamAction, "AgreeTerms").
				Query(presets.ParamID, "11").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			partial := er.UpdatePortals[0].Body
			if strings.Index(partial, `v-model='form["Agree"]' v-assign='[form, {"Agree":""}]'`) < 0 {
				t.Error(`can't find v-model='form["Agree"]' v-assign='[form, {"Agree":""}]'`, partial)
			}
			return
		},
	},

	{
		name: "doAction AgreeTerms",
		reqFunc: func(db *sql.DB) *http.Request {
			customerData.TruncatePut(db)
			return multipartestutils.NewMultipartBuilder().
				PageURL("/admin/my_customers/11").
				EventFunc(actions.DoAction).
				Query(presets.ParamAction, "AgreeTerms").
				Query(presets.ParamID, "11").
				AddField("Agree", "true").
				BuildEventFuncRequest()
		},
		eventResponseMatch: func(er *testEventResponse, db *gorm.DB, t *testing.T) {
			u := &examples2.Customer{}
			err := db.First(u).Error
			if err != nil {
				t.Error(err)
			}
			if u.TermAgreedAt == nil {
				t.Error(fmt.Sprintf("%#+v", u))
			}
			return
		},
	},
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(postgres.Open(os.Getenv("DBURL")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db.Debug()
}

type testPortalUpdate struct {
	Name        string `json:"name,omitempty"`
	Body        string `json:"body,omitempty"`
	AfterLoaded string `json:"afterLoaded,omitempty"`
}

type testEventResponse struct {
	PageTitle     string              `json:"pageTitle,omitempty"`
	Body          string              `json:"body,omitempty"`
	Reload        bool                `json:"reload,omitempty"`
	ReloadPortals []string            `json:"reloadPortals,omitempty"`
	UpdatePortals []*testPortalUpdate `json:"updatePortals,omitempty"`
	Data          interface{}         `json:"data,omitempty"`
}

func TestAll(t *testing.T) {
	db := ConnectDB()
	p := examples2.Preset1(db)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			dbraw, _ := db.DB()
			r := c.reqFunc(dbraw)
			p.ServeHTTP(w, r)

			if c.eventResponseMatch != nil {
				var er testEventResponse
				err := json.NewDecoder(w.Body).Decode(&er)
				if err != nil {
					t.Fatalf("%s for: %#+v", err, w.Body.String())
				}
				c.eventResponseMatch(&er, db, t)
			}

			if c.pageMatch != nil {
				c.pageMatch(w.Body, db, t)
			}
		})
	}
}
