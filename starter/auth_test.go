package starter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qor5/admin/v3/starter"
	"github.com/qor5/web/v3/multipartestutils"
	"github.com/qor5/x/v3/gormx"
	"github.com/stretchr/testify/require"
	"github.com/theplant/inject"
)

// Coverage focus: auth builder pages and role listing behavior for different roles
func TestAuthChangePasswordAndProfileRename(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	cases := []multipartestutils.TestCase{
		{
			Name:  "Change Password Page",
			Debug: false,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/auth/change-password", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Change your password", "Old password", "New password"},
		},
		{
			Name:  "Profile Rename",
			Debug: false,
			ReqFunc: func() *http.Request {
				// Send a stateful action to rename current user
				return multipartestutils.NewMultipartBuilder().
					PageURL("/?__execute_event__=__dispatch_stateful_action__").
					AddField("__action__", `
{
	"compo_type": "*login.ProfileCompo",
	"compo": {"id": ""},
	"injector": "__profile__",
	"sync_query": false,
	"method": "Rename",
	"request": {"name": "renamed@example.com"}
}
					`).
					BuildEventFuncRequest()
			},
			EventResponseMatch: func(t *testing.T, er *multipartestutils.TestEventResponse) {
				var u starter.User
				require.NoError(t, db.Where("account = ?", "test@example.com").First(&u).Error)
				require.Equal(t, "renamed@example.com", u.Name)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, env.handler)
		})
	}
}

func TestRolesIndex_AdminAndEditorVisibility(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()
	// Admin should see Admin/Manager roles
	{
		c := multipartestutils.TestCase{
			Name:  "Roles Visible For Admin",
			Debug: false,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/roles", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor", "Manager", ">Admin<"},
		}
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, env.handler)
		})
	}

	// Switch current user to Editor by modifying join table, withRoles middleware reloads from DB
	{
		var cur starter.User
		require.NoError(t, db.Where("account = ?", "test@example.com").First(&cur).Error)
		// Find Editor role id
		var editorRoleID uint
		require.NoError(t, db.Table("roles").Select("id").Where("name = ?", starter.RoleEditor).Scan(&editorRoleID).Error)
		require.NotZero(t, editorRoleID)
		// Reset joins
		require.NoError(t, db.Exec("DELETE FROM user_role_join WHERE user_id = ?", cur.ID).Error)
		require.NoError(t, db.Exec("INSERT INTO user_role_join(user_id, role_id) VALUES(?, ?)", cur.ID, editorRoleID).Error)

		c := multipartestutils.TestCase{
			Name:  "Roles Hidden For Editor",
			Debug: false,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/roles", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor"},
			ExpectPageBodyNotContains:     []string{"Manager", ">Admin<"},
		}
		t.Run(c.Name, func(t *testing.T) {
			multipartestutils.RunCase(t, c, env.handler)
		})
	}
}

func TestUsersListing_FilterAdminManagerForNonAdmin(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	// Prepare one Manager user to ensure listing filter can exclude it for non-admin
	mgr, err := starter.UpsertUser(context.Background(), db, &starter.UpsertUserOptions{
		Email:    "manager@example.com",
		Password: "1234567890abcd",
		Role:     []string{starter.RoleManager},
	})
	require.NoError(t, err)
	require.NotNil(t, mgr)

	// Switch current user to Editor
	var cur starter.User
	require.NoError(t, db.Where("account = ?", "test@example.com").First(&cur).Error)
	var editorRoleID uint
	require.NoError(t, db.Table("roles").Select("id").Where("name = ?", starter.RoleEditor).Scan(&editorRoleID).Error)
	require.NoError(t, db.Exec("DELETE FROM user_role_join WHERE user_id = ?", cur.ID).Error)
	require.NoError(t, db.Exec("INSERT INTO user_role_join(user_id, role_id) VALUES(?, ?)", cur.ID, editorRoleID).Error)

	// As Editor, Manager users should be filtered out in listing
	c := multipartestutils.TestCase{
		Name:  "Users listing hides Admin/Manager for non-admin",
		Debug: false,
		ReqFunc: func() *http.Request {
			return httptest.NewRequest("GET", "/users", http.NoBody)
		},
		ExpectPageBodyNotContains: []string{"manager@example.com"},
	}
	t.Run(c.Name, func(t *testing.T) {
		multipartestutils.RunCase(t, c, env.handler)
	})
}

type unloginKey struct{}

func TestAuthLoginPage(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler, func(handler *starter.Handler) *unloginKey {
		handler.WithHandlerHook(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("Cookie", "")
				next.ServeHTTP(w, r)
			})
		})
		return &unloginKey{}
	})
	c := multipartestutils.TestCase{
		Name:  "Login Page",
		Debug: true,
		ReqFunc: func() *http.Request {
			return httptest.NewRequest("GET", "/auth/login", http.NoBody)
		},
		ExpectPageBodyContainsInOrder: []string{"Sign in with Google", "Sign in with Microsoft", "Sign in with Github"},
	}
	t.Run(c.Name, func(t *testing.T) {
		multipartestutils.RunCase(t, c, env.handler)
	})

}
