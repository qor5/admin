package integration_test

import (
	"gorm.io/gorm"
	"net/http"
	"testing"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

func testPermHandler(db *gorm.DB, userRole string) http.Handler {
	mux := http.NewServeMux()
	if err := db.AutoMigrate(
		&models.User{},
		&role.Role{},
		&models.Order{},
	); err != nil {
		panic(err)
	}
	b := presets.New().RightDrawerWidth("700")
	b.DataOperator(gorm2op.DataOperator(db))
	order := b.Model(&models.Order{})
	ol := order.Listing()
	ol.BulkAction("CustomBulkAction").ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(vuetify.VBtn("CustomBulkAction"))
	})

	ol.Action("CustomAction").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(vuetify.VBtn("CustomAction"))
	})

	defer b.Build()

	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(models.RoleEditor).WhoAre(perm.Allowed).ToDo(presets.PermActions, presets.PermBulkActions, presets.PermDoListingAction, presets.PermList).On(perm.Anything),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Allowed).ToDo(presets.PermList).On(perm.Anything),
		).SubjectsFunc(func(r *http.Request) []string {
			u, ok := login.GetCurrentUser(r).(*models.User)
			if !ok {
				return nil
			}
			return u.GetRoles()
		},
		),
	)
	u := &models.User{
		Model: gorm.Model{ID: 888},
		Roles: []role.Role{
			{
				Name: userRole,
			},
		},
	}
	m := login.MockCurrentUser(u)
	mux.Handle("/", m(b))
	return mux
}

func TestEditorPerm(t *testing.T) {
	dbr, _ := TestDB.DB()
	type Case struct {
		multipartestutils.TestCase
		Role string
	}
	cases := []Case{
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Show order detail",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=__reload__").
						BuildEventFuncRequest()
					return req
				},
				ExpectPageBodyContainsInOrder: []string{`CustomBulkAction`, "CustomAction"},
			},
			Role: models.RoleEditor,
		},
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Show order detail",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=__reload__").
						BuildEventFuncRequest()
					return req
				},
				ExpectPageBodyNotContains: []string{`CustomBulkAction`, "CustomAction"},
			},
			Role: models.RoleViewer,
		},
	}

	for _, c := range cases {
		t.Run(c.TestCase.Name, func(t *testing.T) {
			h := testPermHandler(TestDB, c.Role)
			multipartestutils.RunCase(t, c.TestCase, h)
		})
	}
}
