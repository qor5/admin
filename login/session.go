package login

import (
	"cmp"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/x/v3/login"
	"github.com/ua-parser/uap-go/uaparser"
	"gorm.io/gorm"
)

const (
	LoginTokenHashLen = 8 // The hash string length of the token stored in the DB.
)

type LoginSession struct {
	gorm.Model

	UserID    string    `gorm:"index;not null"`
	Device    string    `gorm:"not null"`
	IP        string    `gorm:"not null"`
	TokenHash string    `gorm:"index;not null"`
	ExpiredAt time.Time `gorm:"not null"`
}

func (sess *LoginSession) IsExpired() bool {
	return time.Now().After(sess.ExpiredAt)
}

type SessionBuilder struct {
	once sync.Once
	lb   *login.Builder
	db   *gorm.DB
	amb  *activity.ModelBuilder
}

func NewSessionBuilder(lb *login.Builder, db *gorm.DB) *SessionBuilder {
	return &SessionBuilder{lb: lb, db: db}
}

func (b *SessionBuilder) GetLoginBuilder() *login.Builder {
	return b.lb
}

func (b *SessionBuilder) ActivityModelBuilder(amb *activity.ModelBuilder) *SessionBuilder {
	b.amb = amb
	return b
}

func (b *SessionBuilder) CreateSession(r *http.Request, uid string) error {
	token := login.GetSessionToken(b.lb, r)
	client := uaparser.NewFromSaved().Parse(r.Header.Get("User-Agent"))

	if err := b.db.Create(&LoginSession{
		UserID:    uid,
		Device:    fmt.Sprintf("%v - %v", client.UserAgent.Family, client.Os.Family),
		IP:        ip(r),
		TokenHash: getStringHash(token, LoginTokenHashLen),
		ExpiredAt: time.Now().Add(time.Duration(b.lb.GetSessionMaxAge()) * time.Second),
	}).Error; err != nil {
		return err
	}

	return nil
}

func (b *SessionBuilder) ExtendSession(r *http.Request, uid string, oldToken string) (err error) {
	token := login.GetSessionToken(b.lb, r)
	tokenHash := getStringHash(token, LoginTokenHashLen)
	oldTokenHash := getStringHash(oldToken, LoginTokenHashLen)
	if err = b.db.Model(&LoginSession{}).
		Where("user_id = ? and token_hash = ?", uid, oldTokenHash).
		Updates(map[string]any{
			"token_hash": tokenHash,
			"expired_at": time.Now().Add(time.Duration(b.lb.GetSessionMaxAge()) * time.Second),
		}).Error; err != nil {
		return err
	}

	return nil
}

func (b *SessionBuilder) ExpireCurrentSession(r *http.Request, uid string) (err error) {
	token := login.GetSessionToken(b.lb, r)
	tokenHash := getStringHash(token, LoginTokenHashLen)
	if err = b.db.Model(&LoginSession{}).
		Where("user_id = ? and token_hash = ?", uid, tokenHash).
		Updates(map[string]any{
			"expired_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	return nil
}

func (b *SessionBuilder) ExpireAllSessions(uid string) (err error) {
	return b.db.Model(&LoginSession{}).
		Where("user_id = ?", uid).
		Updates(map[string]any{
			"expired_at": time.Now(),
		}).Error
}

func (b *SessionBuilder) ExpireOtherSessions(r *http.Request, uid string) (err error) {
	token := login.GetSessionToken(b.lb, r)

	return b.db.Model(&LoginSession{}).
		Where("user_id = ? AND token_hash != ?", uid, getStringHash(token, LoginTokenHashLen)).
		Updates(map[string]any{
			"expired_at": time.Now(),
		}).Error
}

func (b *SessionBuilder) IsSessionValid(r *http.Request, uid string) (valid bool, err error) {
	token := login.GetSessionToken(b.lb, r)
	if token == "" {
		return false, nil
	}
	sess := LoginSession{}
	if err = b.db.Where("user_id = ? and token_hash = ?", uid, getStringHash(token, LoginTokenHashLen)).
		First(&sess).
		Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return false, err
		}
		return false, nil
	}
	if sess.IsExpired() {
		return false, nil
	}
	// IP check
	if sess.IP != ip(r) {
		return false, nil
	}
	return true, nil
}

func (b *SessionBuilder) Middleware(cfgs ...login.MiddlewareConfig) func(next http.Handler) http.Handler {
	middleware := b.lb.Middleware(cfgs...)
	return func(next http.Handler) http.Handler {
		return middleware(b.validateSessionToken()(next))
	}
}

func (b *SessionBuilder) validateSessionToken() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := login.GetCurrentUser(r)
			if user == nil {
				next.ServeHTTP(w, r)
				return
			}
			if login.IsLoginWIP(r) {
				next.ServeHTTP(w, r)
				return
			}

			valid, err := b.IsSessionValid(r, activity.ObjectID(user))
			if err != nil || !valid {
				if r.URL.Path == b.lb.LogoutURL {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, b.lb.LogoutURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (b *SessionBuilder) AutoMigrate() (r *SessionBuilder) {
	if b.db == nil {
		panic("db is nil")
	}
	if err := b.db.AutoMigrate(&LoginSession{}); err != nil {
		panic(err)
	}
	return b
}

func (b *SessionBuilder) Setup() (r *SessionBuilder) {
	if b.db == nil {
		return b
	}
	b.once.Do(func() {
		logAction := func(r *http.Request, user any, action string) error {
			if b.amb != nil && user != nil {
				_, err := b.amb.Log(r.Context(), action, user, nil)
				return err
			}
			return nil
		}
		b.lb.AfterLogin(func(r *http.Request, user any, extraVals ...any) error {
			return cmp.Or(
				logAction(r, user, "login"),
				b.CreateSession(r, activity.ObjectID(user)),
			)
		}).
			AfterFailedToLogin(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return logAction(r, user, "login-failed")
			}).
			AfterUserLocked(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return logAction(r, user, "locked")
			}).
			AfterLogout(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return cmp.Or(
					logAction(r, user, "logout"),
					b.ExpireCurrentSession(r, activity.ObjectID(user)),
				)
			}).
			AfterConfirmSendResetPasswordLink(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
				return logAction(r, user, "send-reset-password-link")
			}).
			AfterResetPassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return cmp.Or(
					b.ExpireAllSessions(activity.ObjectID(user)),
					logAction(r, user, "reset-password"),
				)
			}).
			AfterChangePassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return cmp.Or(
					b.ExpireAllSessions(activity.ObjectID(user)),
					logAction(r, user, "change-password"),
				)
			}).
			AfterExtendSession(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
				oldToken := extraVals[0].(string)
				return cmp.Or(
					b.ExtendSession(r, activity.ObjectID(user), oldToken),
					logAction(r, user, "extend-session"),
				)
			}).
			AfterTOTPCodeReused(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return logAction(r, user, "totp-code-reused")
			})
	})
	return b
}
