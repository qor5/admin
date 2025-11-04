package examples_admin

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
)

// PermTestItem represents a single item in a list
type PermTestItem struct {
	Name string `json:"name"`
}

// PermTestItems is a slice type for testing permissions (must be pointer slice for ListEditor)
type PermTestItems []*PermTestItem

func (t PermTestItems) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *PermTestItems) Scan(value interface{}) error {
	if value == nil {
		*t = PermTestItems{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			*t = PermTestItems{}
			return nil
		}
		return json.Unmarshal([]byte(v), t)
	case []byte:
		if len(v) == 0 {
			*t = PermTestItems{}
			return nil
		}
		return json.Unmarshal(v, t)
	default:
		return errors.New("unsupported type for PermTestItems")
	}
}

// PermTestContainer represents a container with list of items for permission testing
type PermTestContainer struct {
	gorm.Model
	Title string        `json:"title"`
	Items PermTestItems `gorm:"type:jsonb" json:"items"`
}

// Simple user model for testing
type PermTestUser struct {
	gorm.Model
	Name  string
	Roles []role.Role `gorm:"many2many:perm_test_user_role_join"`
}

func (u *PermTestUser) GetRoles() []string {
	var roles []string
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return roles
}

const (
	RoleAdmin  = "Admin"
	RoleViewer = "Viewer"
)

func TestListEditorPermissions(t *testing.T) {
	if err := TestDB.AutoMigrate(&PermTestContainer{}, &PermTestUser{}, &role.Role{}); err != nil {
		t.Fatal(err)
	}

	cases := []multipartestutils.TestCase{
		{
			Name:  "Admin can see add button and sort button",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleAdmin)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 1},
					Title: "Test Container",
					Items: PermTestItems{
						&PermTestItem{Name: "Item 1"},
						&PermTestItem{Name: "Item 2"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "1").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"mdi-sort-variant", // Sort button visible (rendered first in header)
				"mdi-delete",       // Delete button visible (in each item card)
				"Add Item",         // Add button visible (rendered last, after items)
			},
		},
		{
			Name:  "Viewer cannot see add button, sort button, or delete button",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleViewer)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 2},
					Title: "Test Container 2",
					Items: PermTestItems{
						&PermTestItem{Name: "Item 1"},
						&PermTestItem{Name: "Item 2"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.Edit).
					Query(presets.ParamID, "2").
					BuildEventFuncRequest()
			},
			ExpectPortalUpdate0NotContains: []string{
				"Add Item",         // Add button hidden
				"mdi-sort-variant", // Sort button hidden
				"mdi-delete",       // Delete button hidden
			},
			ExpectPortalUpdate0ContainsInOrder: []string{
				"Item 1", // Content still visible (read-only)
				"Item 2",
			},
		},
		{
			Name:  "Admin can add items",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleAdmin)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 3},
					Title: "Test Container 3",
					Items: PermTestItems{
						&PermTestItem{Name: "Existing Item"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.AddRowEvent).
					Query(presets.ParamID, "3").
					Query(presets.ParamAddRowFormKey, "Items").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{
				"PresetsNotifRowUpdated", // Operation successful - row update notification fired
			},
		},
		{
			Name:  "Viewer cannot add items (shows error message)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleViewer)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 4},
					Title: "Test Container 4",
					Items: PermTestItems{
						&PermTestItem{Name: "Existing Item"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.AddRowEvent).
					Query(presets.ParamID, "4").
					Query(presets.ParamAddRowFormKey, "Items").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				// Permission check should show error message, no row update notification
				if strings.Contains(er.RunScript, "PresetsNotifRowUpdated") {
					t.Error("Expected operation to be blocked, but PresetsNotifRowUpdated was fired")
				}
				// Should show permission denied error
				if !strings.Contains(er.RunScript, "permission denied") {
					t.Error("Expected permission denied error message to be shown")
				}
			},
		},
		{
			Name:  "Admin can remove items",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleAdmin)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 5},
					Title: "Test Container 5",
					Items: PermTestItems{
						&PermTestItem{Name: "Item 1"},
						&PermTestItem{Name: "Item 2"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.RemoveRowEvent).
					Query(presets.ParamID, "5").
					AddField(presets.ParamRemoveRowFormKey, "Items[0]").
					BuildEventFuncRequest()
			},
			ExpectRunScriptContainsInOrder: []string{
				"PresetsNotifRowUpdated", // Operation successful - row update notification fired
			},
		},
		{
			Name:  "Viewer cannot remove items (shows error message)",
			Debug: true,
			HandlerMaker: func() http.Handler {
				return createPermTestHandler(TestDB, RoleViewer)
			},
			ReqFunc: func() *http.Request {
				// Seed data
				container := &PermTestContainer{
					Model: gorm.Model{ID: 6},
					Title: "Test Container 6",
					Items: PermTestItems{
						&PermTestItem{Name: "Item 1"},
						&PermTestItem{Name: "Item 2"},
					},
				}
				TestDB.Create(container)

				return multipartestutils.NewMultipartBuilder().
					PageURL("/perm-test-containers").
					EventFunc(actions.RemoveRowEvent).
					Query(presets.ParamID, "6").
					AddField(presets.ParamRemoveRowFormKey, "Items[0]").
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				// Permission check should show error message, no row update notification
				if strings.Contains(er.RunScript, "PresetsNotifRowUpdated") {
					t.Error("Expected operation to be blocked, but PresetsNotifRowUpdated was fired")
				}
				// Should show permission denied error
				if !strings.Contains(er.RunScript, "permission denied") {
					t.Error("Expected permission denied error message to be shown")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, nil)
		})
	}
}

func createPermTestHandler(db *gorm.DB, userRole string) http.Handler {
	mux := http.NewServeMux()

	if err := db.AutoMigrate(&PermTestContainer{}, &PermTestUser{}, &role.Role{}); err != nil {
		panic(err)
	}

	b := presets.New()
	b.DataOperator(gorm2op.DataOperator(db))

	// Configure permissions
	b.Permission(
		perm.New().Policies(
			// Admin has full access
			perm.PolicyFor(RoleAdmin).WhoAre(perm.Allowed).ToDo(perm.Anything).On(perm.Anything),

			// Viewer has read-only access
			perm.PolicyFor(RoleViewer).WhoAre(perm.Allowed).ToDo(presets.PermList, presets.PermGet).On(perm.Anything),
			perm.PolicyFor(RoleViewer).WhoAre(perm.Denied).ToDo(presets.PermCreate, presets.PermUpdate, presets.PermDelete).On(perm.Anything),
		).SubjectsFunc(func(r *http.Request) []string {
			u, ok := login.GetCurrentUser(r).(*PermTestUser)
			if !ok {
				return nil
			}
			return u.GetRoles()
		}),
	)

	// Register model with nested field
	mb := b.Model(&PermTestContainer{})
	mb.Listing("ID", "Title")
	mb.Editing("Title", "Items")

	// Configure nested field builder for Items
	fb := b.NewFieldsBuilder(presets.WRITE).Model(&PermTestItem{}).Only("Name")
	mb.Editing().Field("Items").Nested(fb)

	// Create mock user with specified role
	var r role.Role
	if err := db.FirstOrCreate(&r, role.Role{Name: userRole}).Error; err != nil {
		panic(err)
	}

	u := &PermTestUser{
		Model: gorm.Model{ID: 999},
		Name:  "Test User",
		Roles: []role.Role{r},
	}

	m := login.MockCurrentUser(u)
	mux.Handle("/", m(b))
	return mux
}
