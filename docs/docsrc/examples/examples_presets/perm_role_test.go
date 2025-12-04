package examples_presets

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	"github.com/qor5/x/v3/ui/vuetify"
	"github.com/theplant/gofixtures"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/role"
)

type customer struct {
	CustomCode uint `gorm:"primarykey"`
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (c *customer) TableName() string { return "permtest_customers" }

func (c *customer) PrimarySlug() string {
	if c.CustomCode == 0 {
		return ""
	}
	return strconv.Itoa(int(c.CustomCode))
}

func (c *customer) PrimaryColumnValuesBySlug(slug string) map[string]string {
	return map[string]string{
		"custom_code": slug,
	}
}

func testPermHandler(db *gorm.DB, userRole string) http.Handler {
	mux := http.NewServeMux()
	if err := db.AutoMigrate(
		&models.User{},
		&role.Role{},
		&models.Order{},
		&customer{},
		&Group{},
	); err != nil {
		panic(err)
	}
	b := presets.New().RightDrawerWidth("700")
	defer b.Build()
	b.DataOperator(gorm2op.DataOperator(db))

	configPermOrder(b)
	configPermCustomer(b)
	mb := b.Model(&Group{})
	mb.Listing("ID", "Name")
	perm.Verbose = true
	b.Permission(
		perm.New().Policies(
			perm.PolicyFor(models.RoleEditor).WhoAre(perm.Allowed).ToDo(presets.PermActions, presets.PermBulkActions, presets.PermDoListingAction, presets.PermList).On("*:presets:orders:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Allowed).ToDo(presets.PermList).On("*:presets:orders:*"),
			perm.PolicyFor(models.RoleEditor).WhoAre(perm.Allowed).ToDo(presets.PermUpdate, presets.PermGet).On("*:presets:orders:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Allowed).ToDo(presets.PermGet).On("*:presets:orders:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Denied).ToDo(presets.PermCreate).On("*:presets:customers:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:presets:customers:*"),
			perm.PolicyFor(models.RoleEditor).WhoAre(perm.Allowed).ToDo(perm.Anything).On("*:presets:customers:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Denied).ToDo(presets.PermUpdate).On("*:presets:groups:*"),
			perm.PolicyFor(models.RoleViewer).WhoAre(perm.Allowed).ToDo(presets.PermRead...).On("*:presets:groups:*"),
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

func configPermOrder(b *presets.Builder) {
	// model order
	order := b.Model(&models.Order{})
	ol := order.Listing()
	ol.BulkAction("CustomBulkAction").ComponentFunc(func(selectedIds []string, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(vuetify.VBtn("CustomBulkAction"))
	})

	ol.Action("CustomAction").ComponentFunc(func(id string, ctx *web.EventContext) h.HTMLComponent {
		return h.Div(vuetify.VBtn("CustomAction"))
	})

	order.Detailing("source_section").Drawer(true)

	section := presets.NewSectionBuilder(order, "source_section").Editing("Source")
	order.Detailing("source_section").Section(section)
}

var permCustomerData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.permtest_customers (custom_code, created_at, updated_at, deleted_at, name) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'OldName');
`, []string{"permtest_customers"}))

func configPermCustomer(b *presets.Builder) {
	mb := b.Model(&customer{})
	de := mb.Detailing("name_section").Drawer(true)
	section := presets.NewSectionBuilder(mb, "name_section").
		Editing("Name").Viewing("Name")
	de.Section(section)
}

func TestPermWithoutID(t *testing.T) {
	dbr, _ := TestDB.DB()
	type Case struct {
		multipartestutils.TestCase
		Role string
	}
	cases := []Case{
		{
			TestCase: multipartestutils.TestCase{
				Name:  "deny create",
				Debug: true,
				ReqFunc: func() *http.Request {
					permCustomerData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/customers?__execute_event__=section_edit_name_section&id=1").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{`<vx-field label='Name' v-model='form["Name"]' :error-messages='dash.errorMessages["Name"]' v-assign:append='[dash.errorMessages, {"Name":null}]' v-assign='[form, {"Name":"OldName"}]' :disabled='false'></vx-field>`},
			},
			Role: models.RoleViewer,
		},
		{
			TestCase: multipartestutils.TestCase{
				Name:  "allow create",
				Debug: true,
				ReqFunc: func() *http.Request {
					permCustomerData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/customers?__execute_event__=section_edit_name_section&id=1").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{`<vx-field label='Name' v-model='form["Name"]' :error-messages='dash.errorMessages["Name"]' v-assign:append='[dash.errorMessages, {"Name":null}]' v-assign='[form, {"Name":"OldName"}]' :disabled='false'></vx-field>`},
			},
			Role: models.RoleEditor,
		},
	}

	for _, c := range cases {
		t.Run(c.TestCase.Name, func(t *testing.T) {
			h := testPermHandler(TestDB, c.Role)
			multipartestutils.RunCase(t, c.TestCase, h)
		})
	}
}

func TestBulkActionPerm(t *testing.T) {
	dbr, _ := TestDB.DB()
	type Case struct {
		multipartestutils.TestCase
		Role string
	}
	cases := []Case{
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Show order list",
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
				Name:  "Show order list",
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

func TestSectionEditPerm(t *testing.T) {
	dbr, _ := TestDB.DB()
	type Case struct {
		multipartestutils.TestCase
		Role string
	}
	cases := []Case{
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Show order detail with update perm",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=presets_DetailingDrawer&id=6").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{":prepend-icon='\"mdi-pencil-outline\"' v-show='true&&true'"},
			},
			Role: models.RoleEditor,
		},
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Show order detail without update perm",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=presets_DetailingDrawer&id=6").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{":prepend-icon='\"mdi-pencil-outline\"' v-show='true&&false'"},
			},
			Role: models.RoleViewer,
		},
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Save order section without update perm",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=section_save_source_section&id=6").
						AddField("Source", "newSource").
						BuildEventFuncRequest()
					return req
				},
				ExpectRunScriptContainsInOrder: []string{"permission denied"},
			},
			Role: models.RoleViewer,
		},
		{
			TestCase: multipartestutils.TestCase{
				Name:  "Save order section with update perm",
				Debug: true,
				ReqFunc: func() *http.Request {
					admin.OrdersExampleData.TruncatePut(dbr)
					req := multipartestutils.NewMultipartBuilder().
						PageURL("/orders?__execute_event__=section_save_source_section&id=6").
						AddField("Source", "newSource").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{"newSource"},
			},
			Role: models.RoleEditor,
		},
	}

	for _, c := range cases {
		t.Run(c.TestCase.Name, func(t *testing.T) {
			h := testPermHandler(TestDB, c.Role)
			multipartestutils.RunCase(t, c.TestCase, h)
		})
	}
}

var permGroupData = gofixtures.Data(gofixtures.Sql(`
insert into public.groups (id, name)
values  (1, 'Group1');
`, []string{"groups"}))

func TestGroupPerm(t *testing.T) {
	dbr, _ := TestDB.DB()
	t.Run("Show group list", func(t *testing.T) {
		h := testPermHandler(TestDB, models.RoleViewer)
		multipartestutils.RunCase(t, multipartestutils.TestCase{
			Name:  "Show group list",
			Debug: true,
			ReqFunc: func() *http.Request {
				permGroupData.TruncatePut(dbr)
				req := multipartestutils.NewMultipartBuilder().
					PageURL("/groups").
					BuildEventFuncRequest()
				return req
			},
			ExpectPageBodyContainsInOrder: []string{`Group1`},
			ExpectPageBodyNotContains:     []string{`eventFunc("presets_Edit").query("id", "1")`},
		}, h)
	})
}
