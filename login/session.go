package login

import (
	"cmp"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"github.com/ua-parser/uap-go/uaparser"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const (
	LoginTokenHashLen = 8 // The hash string length of the token stored in the DB.
)

func (b *SessionBuilder) Install(pb *presets.Builder) error {
	if b.pb != nil {
		return errors.Errorf("profile: already installed")
	}
	return b.installPreset(pb)
}

func (b *SessionBuilder) installPreset(pb *presets.Builder) error {
	if pb == nil {
		return errors.Errorf("profile: presets.Builder is nil")
	}

	b.pb = pb
	b.pb.GetI18n().
		RegisterForModule(language.English, I18nAdminLoginKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nAdminLoginKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nAdminLoginKey, Messages_ja_JP)

	type LoginSessionsDialog struct{}
	mb := b.pb.Model(&LoginSessionsDialog{}).URIName(uriNameLoginSessionsDialog).InMenu(false)
	mb.RegisterEventFunc(eventLoginSessionsDialog, b.handleEventLoginSessionsDialog)
	mb.RegisterEventFunc(eventExpireOtherSessions, b.handleEventExpireOtherSessions)
	return nil
}

type LoginSession struct {
	gorm.Model

	UserID    string    `gorm:"index;not null"`
	Device    string    `gorm:"not null"`
	IP        string    `gorm:"not null"`
	TokenHash string    `gorm:"index;not null"`
	ExpiredAt time.Time `gorm:"not null"`
}

type SessionBuilder struct {
	once              sync.Once
	calledAutoMigrate atomic.Bool // auto migrate flag

	lb           *login.Builder
	dbPrimitive  *gorm.DB // primitive db
	db           *gorm.DB // global db with table prefix scope
	tablePrefix  string
	amb          *activity.ModelBuilder
	pb           *presets.Builder
	isPublicUser func(user any) bool
}

func NewSessionBuilder(lb *login.Builder, db *gorm.DB) *SessionBuilder {
	return (&SessionBuilder{
		lb:          lb,
		db:          db,
		dbPrimitive: db,
	}).setup()
}

func (b *SessionBuilder) TablePrefix(prefix string) *SessionBuilder {
	if b.calledAutoMigrate.Load() {
		panic("please set table prefix before auto migrate")
	}
	b.tablePrefix = prefix
	if prefix == "" {
		b.db = b.dbPrimitive
	} else {
		b.db = b.dbPrimitive.Scopes(activity.ScopeWithTablePrefix(prefix)).Session(&gorm.Session{})
	}
	return b
}

func (b *SessionBuilder) AutoMigrate() (r *SessionBuilder) {
	if !b.calledAutoMigrate.CompareAndSwap(false, true) {
		panic("already migrated")
	}
	if err := AutoMigrateSession(b.dbPrimitive, b.tablePrefix); err != nil {
		panic(err)
	}
	return b
}

func AutoMigrateSession(db *gorm.DB, tablePrefix string) error {
	if tablePrefix != "" {
		db = db.Scopes(activity.ScopeWithTablePrefix(tablePrefix)).Session(&gorm.Session{})
	}
	dst := []any{&LoginSession{}}
	for _, v := range dst {
		err := db.Model(v).AutoMigrate(v)
		if err != nil {
			return errors.Wrap(err, "auto migrate")
		}
		if vv, ok := v.(interface {
			AfterMigrate(tx *gorm.DB, tablePrefix string) error
		}); ok {
			err := vv.AfterMigrate(db, tablePrefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *SessionBuilder) GetLoginBuilder() *login.Builder {
	return b.lb
}

func (b *SessionBuilder) Activity(amb *activity.ModelBuilder) *SessionBuilder {
	b.amb = amb
	return b
}

func (b *SessionBuilder) Secret(v string) *SessionBuilder {
	b.lb.Secret(v)
	return b
}

func (b *SessionBuilder) IsPublicUser(f func(user any) bool) *SessionBuilder {
	b.isPublicUser = f
	return b
}

func (b *SessionBuilder) Mount(mux *http.ServeMux) {
	b.lb.Mount(mux)
}

func (b *SessionBuilder) MountAPI(mux *http.ServeMux) {
	b.lb.MountAPI(mux)
}

func (b *SessionBuilder) CreateSession(r *http.Request, uid string) error {
	token := login.GetSessionToken(b.lb, r)
	client := uaparser.NewFromSaved().Parse(r.Header.Get("User-Agent"))
	if err := b.db.Create(&LoginSession{
		UserID:    uid,
		Device:    fmt.Sprintf("%v - %v", client.UserAgent.Family, client.Os.Family),
		IP:        ip(r),
		TokenHash: getStringHash(token, LoginTokenHashLen),
		ExpiredAt: b.db.NowFunc().Add(time.Duration(b.lb.GetSessionMaxAge()) * time.Second),
	}).Error; err != nil {
		return errors.Wrap(err, "login: failed to create session")
	}
	return nil
}

func (b *SessionBuilder) ExtendSession(r *http.Request, uid string, oldToken string) error {
	token := login.GetSessionToken(b.lb, r)
	tokenHash := getStringHash(token, LoginTokenHashLen)
	oldTokenHash := getStringHash(oldToken, LoginTokenHashLen)
	if err := b.db.Model(&LoginSession{}).
		Where("user_id = ? and token_hash = ?", uid, oldTokenHash).
		Updates(map[string]any{
			"token_hash": tokenHash,
			"expired_at": b.db.NowFunc().Add(time.Duration(b.lb.GetSessionMaxAge()) * time.Second),
		}).Error; err != nil {
		return errors.Wrap(err, "login: failed to extend session")
	}
	return nil
}

func (b *SessionBuilder) ExpireCurrentSession(r *http.Request, uid string) error {
	token := login.GetSessionToken(b.lb, r)
	tokenHash := getStringHash(token, LoginTokenHashLen)
	if err := b.db.Model(&LoginSession{}).
		Where("user_id = ? and token_hash = ?", uid, tokenHash).
		Updates(map[string]any{
			"expired_at": b.db.NowFunc(),
		}).Error; err != nil {
		return errors.Wrap(err, "login: failed to expire current session")
	}
	return nil
}

func (b *SessionBuilder) ExpireAllSessions(uid string) error {
	if err := b.db.Model(&LoginSession{}).
		Where("user_id = ?", uid).
		Updates(map[string]any{
			"expired_at": b.db.NowFunc(),
		}).Error; err != nil {
		return errors.Wrap(err, "login: failed to expire all sessions")
	}
	return nil
}

func (b *SessionBuilder) ExpireOtherSessions(r *http.Request, uid string) error {
	token := login.GetSessionToken(b.lb, r)
	if err := b.db.Model(&LoginSession{}).
		Where("user_id = ? AND token_hash != ?", uid, getStringHash(token, LoginTokenHashLen)).
		Updates(map[string]any{
			"expired_at": b.db.NowFunc(),
		}).Error; err != nil {
		return errors.Wrap(err, "login: failed to expire other sessions")
	}
	return nil
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
			return false, errors.Wrap(err, "login: failed to find session")
		}
		return false, nil
	}
	if sess.ExpiredAt.Before(b.db.NowFunc()) {
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

			valid, err := b.IsSessionValid(r, presets.MustObjectID(user))
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

func (b *SessionBuilder) setup() (r *SessionBuilder) {
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
				b.CreateSession(r, presets.MustObjectID(user)),
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
					b.ExpireCurrentSession(r, presets.MustObjectID(user)),
				)
			}).
			AfterConfirmSendResetPasswordLink(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
				return logAction(r, user, "send-reset-password-link")
			}).
			AfterResetPassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return cmp.Or(
					b.ExpireAllSessions(presets.MustObjectID(user)),
					logAction(r, user, "reset-password"),
				)
			}).
			AfterChangePassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return cmp.Or(
					b.ExpireAllSessions(presets.MustObjectID(user)),
					logAction(r, user, "change-password"),
				)
			}).
			AfterExtendSession(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
				oldToken := extraVals[0].(string)
				return cmp.Or(
					b.ExtendSession(r, presets.MustObjectID(user), oldToken),
					logAction(r, user, "extend-session"),
				)
			}).
			AfterTOTPCodeReused(func(r *http.Request, user interface{}, _ ...interface{}) error {
				return logAction(r, user, "totp-code-reused")
			})
	})
	return b
}

const (
	uriNameLoginSessionsDialog = "login-sessions-dialog"
	eventLoginSessionsDialog   = "loginSession_eventLoginSessionsDialog"
	eventExpireOtherSessions   = "loginSession_eventExpireOtherSessions"
)

func (b *SessionBuilder) OpenSessionsDialog() string {
	if b.pb == nil {
		panic("presets.Builder is nil")
	}
	return web.Plaid().URL("/" + uriNameLoginSessionsDialog).EventFunc(eventLoginSessionsDialog).Go()
}

type dataTableHeader struct {
	Title    string `json:"title"`
	Key      string `json:"key"`
	Width    string `json:"width"`
	Sortable bool   `json:"sortable"`
}

func (b *SessionBuilder) handleEventLoginSessionsDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nAdminLoginKey, Messages_en_US).(*Messages)
	pmsgr := presets.MustGetMessages(ctx.R)

	user := login.GetCurrentUser(ctx.R)
	if user == nil {
		return r, errors.New("login: user not found")
	}
	uid := presets.MustObjectID(user)
	currentTokenHash := getStringHash(login.GetSessionToken(b.lb, ctx.R), LoginTokenHashLen)

	var sessions []*LoginSession

	s, err := activity.ParseSchemaWithDB(b.db, &LoginSession{})
	if err != nil {
		return r, err
	}
	tableName := b.tablePrefix + s.Table

	// Only one record with the same `device+ip` is returned unless they are not expired.
	// Order by `expired_at` in descending order.
	// If the token is the current one, it should be the first one.
	// Max 100 records returned.
	raw := fmt.Sprintf(`
		WITH ranked_sessions AS (
		    SELECT *, ROW_NUMBER() OVER (PARTITION BY "device", "ip" ORDER BY "expired_at" DESC) AS rn
		    FROM %s
		    WHERE "user_id" = ? AND "deleted_at" IS NULL
		)
		SELECT *
		FROM ranked_sessions
		WHERE rn = 1 OR "expired_at" >= NOW()
		ORDER BY CASE WHEN "token_hash" = ? THEN 0 ELSE 1 END, "expired_at" DESC
		LIMIT 100;`, tableName)
	if err := b.db.Raw(raw, uid, currentTokenHash).Scan(&sessions).Error; err != nil {
		return r, errors.Wrap(err, "login: failed to find sessions")
	}

	isPublicUser := false
	if b.isPublicUser != nil {
		isPublicUser = b.isPublicUser(user)
	}

	if isPublicUser && len(sessions) > 10 {
		sessions = sessions[:10]
	}

	type sessionWrapper struct {
		*LoginSession
		Time   string
		Status string
	}
	now := b.db.NowFunc()
	wrappers := make([]*sessionWrapper, 0, len(sessions))
	activeCount := 0
	for _, v := range sessions {
		w := &sessionWrapper{
			LoginSession: v,
			Time:         pmsgr.HumanizeTime(v.CreatedAt),
			Status:       msgr.SessionStatusActive,
		}
		if isPublicUser {
			w.IP = msgr.HideIPAddressTips
		}
		if v.ExpiredAt.Before(now) {
			w.Status = msgr.SessionStatusExpired
		} else {
			activeCount++
		}
		if v.TokenHash == currentTokenHash {
			w.Status = msgr.SessionStatusCurrent
		}
		wrappers = append(wrappers, w)
	}
	tableHeaders := []dataTableHeader{
		{msgr.SessionTableHeaderTime, "Time", "25%", false},
		{msgr.SessionTableHeaderDevice, "Device", "25%", false},
		{msgr.SessionTableHeaderIPAddress, "IP", "25%", false},
		{msgr.SessionTableHeaderStatus, "Status", "25%", false},
	}
	table := v.VDataTable().Headers(tableHeaders).Items(wrappers).ItemsPerPage(-1).HideDefaultFooter(true)

	body := web.Scope().VSlot("{locals: xlocals}").Init("{dialog:true}").Children(
		v.VDialog().Attr("v-model", "xlocals.dialog").Width("60%").MaxWidth(828).Scrollable(true).Children(
			v.VCard().Children(
				v.VCardTitle().Class("d-flex align-center pa-6 ga-2").Children(
					h.Div().Class("text-h6").Text(msgr.SessionsDialogTitle),
					v.VSpacer(),
					v.VBtn("").Size(v.SizeXSmall).Icon("mdi-close").Variant(v.VariantText).Color(v.ColorGreyDarken1).Attr("@click", "xlocals.dialog=false"),
				),
				v.VCardText().Class("px-6 pt-0 pb-6").Attr("style", "max-height: 46vh;").ClassIf("mb-6", isPublicUser).Children(
					table,
				),

				h.Iff(!isPublicUser && activeCount > 1, func() h.HTMLComponent {
					return v.VCardActions().Class("px-6 pt-0 pb-6").Children(
						v.VSpacer(),
						v.VBtn(msgr.ExpireOtherSessions).Variant(v.VariantOutlined).Size(v.SizeSmall).Color(v.ColorWarning).PrependIcon("mdi-alert-circle-outline").
							Class("text-none font-weight-regular").
							Attr("@click", web.Plaid().URL("/"+uriNameLoginSessionsDialog).EventFunc(eventExpireOtherSessions).Go()),
					)
				}),

				// The old implementation doesn't make sense, so I removed it.
				// v.VCardActions().Class("px-6 pt-0 pb-6").Children(
				// 	v.VSpacer(),
				// 	v.VBtn(presetsMsgr.Cancel).Variant(v.VariantOutlined).Size(v.SizeSmall).Color(v.ColorSecondary).
				// 		Class("text-none text-caption font-weight-regular").
				// 		Attr("@click", "xlocals.dialog=false"),
				// 	v.VBtn(presetsMsgr.OK).Variant(v.VariantTonal).Size(v.SizeSmall).
				// 		Class("text-none text-caption font-weight-regular bg-secondary text-on-secondary").
				// 		Attr("@click", "xlocals.dialog=false"),
				// ),
			),
		),
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: presets.DialogPortalName, Body: body})
	return
}

func (b *SessionBuilder) handleEventExpireOtherSessions(ctx *web.EventContext) (r web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nAdminLoginKey, Messages_en_US).(*Messages)

	user := login.GetCurrentUser(ctx.R)
	if user == nil {
		return r, errors.New("login: user not found")
	}
	isPublicUser := false
	if b.isPublicUser != nil {
		isPublicUser = b.isPublicUser(user)
	}
	if isPublicUser {
		return r, perm.PermissionDenied
	}
	uid := presets.MustObjectID(user)

	if err = b.ExpireOtherSessions(ctx.R, uid); err != nil {
		return r, err
	}

	presets.ShowMessage(&r, msgr.SuccessfullyExpiredOtherSessions, "")
	web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).Go())
	return
}
