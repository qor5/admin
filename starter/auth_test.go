package starter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	. "github.com/qor5/web/v3/multipartestutils"

	"github.com/qor5/admin/v3/starter"
	"github.com/qor5/x/v3/gormx"
	"github.com/stretchr/testify/require"
	"github.com/theplant/inject"
)

// Coverage focus: auth builder pages and role listing behavior for different roles
func TestAuthChangePasswordAndProfileRename(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	cases := []TestCase{
		{
			Name:  "Change Password Page",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/auth/change-password", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Change your password", "Old password", "New password"},
		},
		{
			Name:  "Profile Rename",
			Debug: true,
			ReqFunc: func() *http.Request {
				// Send a stateful action to rename current user
				return NewMultipartBuilder().
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
			EventResponseMatch: func(t *testing.T, er *TestEventResponse) {
				var u starter.User
				require.NoError(t, db.Where("account = ?", "test@example.com").First(&u).Error)
				require.Equal(t, "renamed@example.com", u.Name)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
		})
	}
}

func TestRolesIndex_AdminAndEditorVisibility(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()
	// Admin should see Admin/Manager roles
	{
		c := TestCase{
			Name:  "Roles Visible For Admin",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/roles", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor", "Manager", ">Admin<"},
		}
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
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

		c := TestCase{
			Name:  "Roles Hidden For Editor",
			Debug: true,
			ReqFunc: func() *http.Request {
				return httptest.NewRequest("GET", "/roles", http.NoBody)
			},
			ExpectPageBodyContainsInOrder: []string{"Viewer", "Editor"},
			ExpectPageBodyNotContains:     []string{"Manager", ">Admin<"},
		}
		t.Run(c.Name, func(t *testing.T) {
			RunCase(t, c, env.handler)
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
	c := TestCase{
		Name:  "Users listing hides Admin/Manager for non-admin",
		Debug: true,
		ReqFunc: func() *http.Request {
			return httptest.NewRequest("GET", "/users", http.NoBody)
		},
		ExpectPageBodyNotContains: []string{"manager@example.com"},
	}
	t.Run(c.Name, func(t *testing.T) {
		RunCase(t, c, env.handler)
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
	c := TestCase{
		Name:  "Login Page",
		Debug: true,
		ReqFunc: func() *http.Request {
			return httptest.NewRequest("GET", "/auth/login", http.NoBody)
		},
		ExpectPageBodyContainsInOrder: []string{"Sign in with Google", "Sign in with Microsoft", "Sign in with Github"},
	}
	t.Run(c.Name, func(t *testing.T) {
		RunCase(t, c, env.handler)
	})
}

func TestAuthLoigin(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler, func(handler *starter.Handler) *unloginKey {
		handler.WithHandlerHook(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("Cookie", "")
				next.ServeHTTP(w, r)
			})
		})
		return &unloginKey{}
	})
	form := url.Values{}
	form.Set("account", "qor@theplant.jp")
	form.Set("password", "admin123456789")
	req := httptest.NewRequest("POST", "/auth/userpass/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	env.handler.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusFound, res.StatusCode)
	require.Equal(t, "/", res.Header.Get("Location"))

	var hasAuthCookie bool
	for _, c := range res.Cookies() {
		if c.Name == "auth" && c.Value != "" {
			hasAuthCookie = true
			break
		}
	}
	require.True(t, hasAuthCookie)
}

func TestDoResetPassword_Success(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	// Prepare a dedicated user for reset flow to avoid affecting other tests
	ctx := context.Background()
	usr, err := starter.UpsertUser(ctx, db, &starter.UpsertUserOptions{
		Email:    "resetuser@example.com",
		Password: "origPassword1234",
		Role:     []string{starter.RoleViewer},
	})
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Generate reset token
	token, err := usr.GenerateResetPasswordToken(db, &starter.User{})
	require.NoError(t, err)
	userID := strconv.Itoa(int(usr.ID))

	// Call do-reset-password
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("token", token)
	form.Set("password", "newPassword1234")
	form.Set("confirm_password", "newPassword1234")
	req := httptest.NewRequest("POST", "/auth/do-reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	env.handler.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusFound, res.StatusCode)
	require.Equal(t, "/auth/login", res.Header.Get("Location"))

	// Verify new password works by logging in
	loginForm := url.Values{}
	loginForm.Set("account", "resetuser@example.com")
	loginForm.Set("password", "newPassword1234")
	loginReq := httptest.NewRequest("POST", "/auth/userpass/login", strings.NewReader(loginForm.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRR := httptest.NewRecorder()
	env.handler.ServeHTTP(loginRR, loginReq)
	loginRes := loginRR.Result()
	defer loginRes.Body.Close()

	require.Equal(t, http.StatusFound, loginRes.StatusCode)
	require.Equal(t, "/", loginRes.Header.Get("Location"))

	var hasAuthCookie bool
	for _, c := range loginRes.Cookies() {
		if c.Name == "auth" && c.Value != "" {
			hasAuthCookie = true
			break
		}
	}
	require.True(t, hasAuthCookie)

	// Follow redirect to home with cookies to ensure authenticated access works
	homeReq := httptest.NewRequest("GET", "/", http.NoBody)
	for _, c := range loginRes.Cookies() {
		homeReq.AddCookie(c)
	}
	homeRR := httptest.NewRecorder()
	env.handler.ServeHTTP(homeRR, homeReq)
	homeRes := homeRR.Result()
	defer homeRes.Body.Close()
	require.Equal(t, http.StatusOK, homeRes.StatusCode)
}

func TestDoResetPassword_FailedByInitialUser(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	ctx := context.Background()
	usr, err := starter.UpsertUser(ctx, db, &starter.UpsertUserOptions{
		Email:    "qor@theplant.jp",
		Password: "admin123456789",
		Role:     []string{starter.RoleViewer},
	})
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Generate reset token
	token, err := usr.GenerateResetPasswordToken(db, &starter.User{})
	require.NoError(t, err)
	userID := strconv.Itoa(int(usr.ID))

	// Call do-reset-password and expect redirect back to reset page
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("token", token)
	form.Set("password", "newPassword1234")
	form.Set("confirm_password", "newPassword1234")
	req := httptest.NewRequest("POST", "/auth/do-reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	env.handler.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusFound, res.StatusCode)
	loc := res.Header.Get("Location")
	require.True(t, strings.HasPrefix(loc, "/auth/reset-password?"))
	require.True(t, strings.Contains(loc, "id="))
	require.True(t, strings.Contains(loc, "token="))
}

func TestDoResetPassword_FailedByLessPassword(t *testing.T) {
	env := newTestEnv(t, starter.SetupPageBuilderForHandler)
	suite := inject.MustResolve[*gormx.TestSuite](env.lc)
	db := suite.DB()

	ctx := context.Background()
	usr, err := starter.UpsertUser(ctx, db, &starter.UpsertUserOptions{
		Email:    "resetuser@theplant.jp",
		Password: "admin123456789",
		Role:     []string{starter.RoleViewer},
	})
	require.NoError(t, err)
	require.NotNil(t, usr)

	// Generate reset token
	token, err := usr.GenerateResetPasswordToken(db, &starter.User{})
	require.NoError(t, err)
	userID := strconv.Itoa(int(usr.ID))

	// Call do-reset-password
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("token", token)
	form.Set("password", "newPassword")
	form.Set("confirm_password", "newPassword")
	req := httptest.NewRequest("POST", "/auth/do-reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	env.handler.ServeHTTP(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusFound, res.StatusCode)
	loc := res.Header.Get("Location")
	require.True(t, strings.HasPrefix(loc, "/auth/reset-password?"))
	require.True(t, strings.Contains(loc, "id="))
	require.True(t, strings.Contains(loc, "token="))
}
