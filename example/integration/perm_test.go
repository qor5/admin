package integration_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

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
	"github.com/theplant/gofixtures"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
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
	); err != nil {
		panic(err)
	}
	b := presets.New().RightDrawerWidth("700")
	defer b.Build()
	b.DataOperator(gorm2op.DataOperator(db))

	configPermOrder(b)
	configPermCustomer(b)

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

	order.Detailing("source_section").Section("source_section").Editing("Source")
}

var permCustomerData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO public.permtest_customers (custom_code, created_at, updated_at, deleted_at, name) VALUES (1, '2024-05-17 15:25:31.134801 +00:00', '2024-05-17 15:25:31.134801 +00:00', null, 'OldName');
`, []string{"permtest_customers"}))

func configPermCustomer(b *presets.Builder) {
	mb := b.Model(&customer{})
	de := mb.Detailing("name_section").Drawer(true)
	se := de.Section("name_section")
	se.Editing("Name").Viewing("Name")
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
						PageURL("/customers?__execute_event__=presets_Detailing_Field_Edit&detailField=name_section&id=1").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{"<v-text-field type='text' :variant='\"underlined\"' v-model='form[\"name_section.Name\"]' v-assign='[form, {\"name_section.Name\":\"OldName\"}]' label='Name' :disabled='false'></v-text-field>"},
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
						PageURL("/customers?__execute_event__=presets_Detailing_Field_Edit&detailField=name_section&id=1").
						BuildEventFuncRequest()
					return req
				},
				ExpectPortalUpdate0ContainsInOrder: []string{"<v-text-field type='text' :variant='\"underlined\"' v-model='form[\"name_section.Name\"]' v-assign='[form, {\"name_section.Name\":\"OldName\"}]' label='Name' :disabled='false'></v-text-field>"},
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
			fmt.Println("1")
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
				ExpectPortalUpdate0ContainsInOrder: []string{":icon='\"mdi-square-edit-outline\"' v-show='isHovering&&true&&true'"},
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
				ExpectPortalUpdate0ContainsInOrder: []string{":icon='\"mdi-square-edit-outline\"' v-show='isHovering&&true&&false'"},
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
						PageURL("/orders?__execute_event__=presets_Detailing_Field_Save&id=6").
						Query("detailField", "source_section").
						AddField("source_section.Source", "newSource").
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
						PageURL("/orders?__execute_event__=presets_Detailing_Field_Save&id=6").
						Query("detailField", "source_section").
						AddField("source_section.Source", "newSource").
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
