package integration_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/theplant/gofixtures"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/example/admin"
	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/role"
)

var sessionData = gofixtures.Data(gofixtures.Sql(`
INSERT INTO "public"."cms_login_sessions" ("id", "created_at", "updated_at", "deleted_at", "user_id", "device", "ip", "token_hash", "expired_at", "extended_at", "last_token_hash") VALUES 
('2', '2024-11-26 16:23:40.940667+00', '2024-11-26 16:32:18.264294+00', NULL, '1', 'Chrome - Mac OS X', '127.0.0.1', '37e8f8b8', '2024-11-26 17:32:18.26309+00', '2024-11-26 16:32:18.263089+00', '95606eff'),
('3', '2024-11-27 06:01:32.667847+00', '2024-11-27 06:01:32.667847+00', NULL, '1', 'Chrome - Mac OS X', '127.0.0.1', '2d528b6d', '2024-11-27 07:01:32.665893+00', '2024-11-27 06:01:32.665892+00', '');
`, []string{"cms_login_sessions"}))

func TestSession(t *testing.T) {
	_, conf := admin.TestHandlerComplex(TestDB, &models.User{
		Model: gorm.Model{ID: 1},
		Name:  "qor@theplant.jp",
		Roles: []role.Role{
			{
				Model: gorm.Model{ID: 1},
				Name:  models.RoleAdmin,
			},
		},
	}, false)

	db := TestDB.Scopes(activity.ScopeWithTablePrefix("cms_")).Session(&gorm.Session{})
	dbr, _ := db.DB()

	sb := conf.GetLoginSessionBuilder()

	const uid = "1"

	{
		sessionData.TruncatePut(dbr)

		var sessions []*login.LoginSession
		err := db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 2)

		// active last session
		err = db.Model(&login.LoginSession{}).Where("id = ?", sessions[0].ID).
			Update("expired_at", db.NowFunc().Add(10*time.Minute)).Error
		require.NoError(t, err)
	}

	var currentTokenHash string

	{
		before := db.NowFunc()

		// create session
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token0"})
		err = sb.CreateSession(r, uid)
		require.NoError(t, err)

		// check session info after create
		var sessions []*login.LoginSession
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)

		session := sessions[0]
		require.True(t, before.Before(session.ExtendedAt))
		require.Empty(t, session.LastTokenHash)
		currentTokenHash = session.TokenHash
	}

	{
		before := db.NowFunc()

		// extend session
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token0"})
		err = sb.ExtendSession(r, uid, "token0", "token1")
		require.NoError(t, err)

		// check session info after extend
		var sessions []*login.LoginSession
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)

		session := sessions[0]
		require.True(t, before.Before(session.ExtendedAt))
		require.NotEmpty(t, session.LastTokenHash)
		require.Equal(t, currentTokenHash, session.LastTokenHash)
		require.NotEqual(t, currentTokenHash, session.TokenHash)
		currentTokenHash = session.TokenHash
	}

	{
		// valid for new token hash
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token1"})
		err = sb.IsSessionValid(r, uid)
		require.NoError(t, err)

		simulateValidateError := true
		sb.WithValidateSessionHook(func(next login.ValidateSessionFunc) login.ValidateSessionFunc {
			return func(ctx context.Context, input *login.ValidateSessionInput) (*login.ValidateSessionOutput, error) {
				if simulateValidateError {
					return nil, errors.New("additional validate error")
				}
				return next(ctx, input)
			}
		})
		err = sb.IsSessionValid(r, uid)
		require.ErrorContains(t, err, "additional validate error")
		simulateValidateError = false
	}

	{
		// valid for last token hash
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token0"})
		err = sb.IsSessionValid(r, uid)
		require.NoError(t, err)

		// force change the extended_at
		err = db.Model(&login.LoginSession{}).Where("token_hash = ?", currentTokenHash).
			Update("extended_at", db.NowFunc().Add(-10*time.Minute)).Error
		require.NoError(t, err)

		// invalid for last token hash
		r, err = http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token0"})
		err = sb.IsSessionValid(r, uid)
		require.ErrorContains(t, err, "session not found")
	}

	{
		// change ip
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token1"})
		r.Header.Set("X-Forwarded-For", "192.168.1.1")
		err = sb.IsSessionValid(r, uid)
		require.ErrorContains(t, err, "IP mismatch")

		// disable ip check
		sb.DisableIPCheck(true)
		err = sb.IsSessionValid(r, uid)
		require.NoError(t, err)
	}

	{
		before := db.NowFunc()

		var sessions []*login.LoginSession
		err := db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validCount := lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Equal(t, 2, validCount)
		// expired before before
		require.Equal(t, 1, lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.Before(before) }))

		// expire other sessions
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token1"})
		err = sb.ExpireOtherSessions(r, uid)
		require.NoError(t, err)

		// check session info after expire
		sessions = nil
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validCount = lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Equal(t, 1, validCount)
		// expired before before
		require.Equal(t, 1, lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.Before(before) }))
	}

	{
		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token1"})

		// update last actived at
		before := db.NowFunc()
		err = sb.UpdateSessionLastActivedAt(r, uid)
		require.NoError(t, err)

		var session *login.LoginSession
		err = db.Where("user_id = ? AND last_actived_at > ?", uid, before).First(&session).Error
		require.NoError(t, err)

		// expire current session
		err = sb.ExpireCurrentSession(r, uid)
		require.NoError(t, err)

		// check session info after expire
		var sessions []*login.LoginSession
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validCount := lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Equal(t, 0, validCount)
	}

	{
		// active all session
		err := db.Model(&login.LoginSession{}).Where("user_id = ?", uid).
			Update("expired_at", db.NowFunc().Add(10*time.Minute)).Error
		require.NoError(t, err)

		var sessions []*login.LoginSession
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validCount := lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Equal(t, 3, validCount)

		// expire one session
		tokenHashToExpire := sessions[0].TokenHash
		err = sb.ExpireSession(uid, tokenHashToExpire)
		require.NoError(t, err)
		sessions = nil
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validSessions := lo.Filter(sessions, func(session *login.LoginSession, _ int) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Len(t, validSessions, 2)
		_, exists := lo.Find(validSessions, func(session *login.LoginSession) bool {
			return session.TokenHash == tokenHashToExpire
		})
		require.False(t, exists)

		// expire all sessions
		err = sb.ExpireAllSessions(uid)
		require.NoError(t, err)

		sessions = nil
		err = db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sessions).Error
		require.NoError(t, err)
		require.Len(t, sessions, 3)
		validCount = lo.CountBy(sessions, func(session *login.LoginSession) bool { return session.ExpiredAt.After(db.NowFunc()) })
		require.Equal(t, 0, validCount)
	}

	{
		// add validate max lifetime hook
		sb.WithValidateSessionHook(login.ValidateMaxLifetimeHook(5 * time.Minute))

		// active current session
		err := db.Model(&login.LoginSession{}).Where("token_hash = ?", currentTokenHash).
			Update("expired_at", db.NowFunc().Add(10*time.Minute)).Error
		require.NoError(t, err)

		r, err := http.NewRequest("GET", "/", http.NoBody)
		require.NoError(t, err)
		r.AddCookie(&http.Cookie{Name: "auth", Value: "token1"})

		err = sb.IsSessionValid(r, uid)
		require.NoError(t, err) // current session is valid

		// change its created_at
		err = db.Model(&login.LoginSession{}).Where("token_hash = ?", currentTokenHash).
			Update("created_at", db.NowFunc().Add(-6*time.Minute)).Error
		require.NoError(t, err)

		err = sb.IsSessionValid(r, uid)
		require.ErrorContains(t, err, "session lifetime exceeded")
	}
}
