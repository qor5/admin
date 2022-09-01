package login

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

var (
	errUserNotFound    = errors.New("user not found")
	errUserPassChanged = errors.New("password changed")
	errWrongPassword   = errors.New("wrong password")
	errUserLocked      = errors.New("user locked")
	errUserGetLocked   = errors.New("user get locked")
	errWrongTOTP       = errors.New("wrong totp")
)

const (
	totpDoURL       = "/auth/2fa/totp/do"
	totpSetupURL    = "/auth/2fa/totp/setup"
	totpValidateURL = "/auth/2fa/totp/validate"
)

type NotifyUserOfResetPasswordLinkFunc func(user interface{}, resetLink string) error
type PasswordValidationFunc func(password string) (message string, ok bool)
type HookFunc func(r *http.Request, user interface{}) error

type Provider struct {
	Goth goth.Provider
	Key  string
	Text string
	Logo HTMLComponent
}

type void struct{}

type Builder struct {
	secret                string
	providers             []*Provider
	authCookieName        string
	authSecureCookieName  string
	continueUrlCookieName string
	// seconds
	sessionMaxAge        int
	autoExtendSession    bool
	maxRetryCount        int
	noForgetPasswordLink bool

	homeURL             string
	loginURL            string
	logoutURL           string
	changePasswordURL   string
	doChangePasswordURL string

	authPrefixInterceptURLS map[string]void
	allowURLS               map[string]void

	loginPageFunc                 web.PageFunc
	forgetPasswordPageFunc        web.PageFunc
	resetPasswordLinkSentPageFunc web.PageFunc
	resetPasswordPageFunc         web.PageFunc
	changePasswordPageFunc        web.PageFunc
	totpSetupPageFunc             web.PageFunc
	totpValidatePageFunc          web.PageFunc

	notifyUserOfResetPasswordLinkFunc NotifyUserOfResetPasswordLinkFunc
	passwordValidationFunc            PasswordValidationFunc

	afterLoginHook                 HookFunc
	afterFailedToLoginHook         HookFunc
	afterUserLockedHook            HookFunc
	afterLogoutHook                HookFunc
	afterSendResetPasswordLinkHook HookFunc
	afterResetPasswordHook         HookFunc
	afterChangePasswordHook        HookFunc

	db                   *gorm.DB
	userModel            interface{}
	snakePrimaryField    string
	tUser                reflect.Type
	userPassEnabled      bool
	oauthEnabled         bool
	sessionSecureEnabled bool
	totpEnabled          bool
	totpIssuer           string

	i18nBuilder *i18n.Builder
}

func New() *Builder {
	r := &Builder{
		authCookieName:        "auth",
		authSecureCookieName:  "qor5_auth_secure",
		continueUrlCookieName: "qor5_continue_url",
		homeURL:               "/",
		loginURL:              "/auth/login",
		logoutURL:             "/auth/logout",
		changePasswordURL:     "/auth/change-password",
		doChangePasswordURL:   "/auth/do-change-password",
		allowURLS:             make(map[string]void),
		sessionMaxAge:         60 * 60,
		autoExtendSession:     true,
		maxRetryCount:         5,
		totpEnabled:           true,
		totpIssuer:            "qor5",
		i18nBuilder:           i18n.New(),
	}

	r.registerI18n()
	r.initAuthPrefixInterceptURLS()

	r.loginPageFunc = defaultLoginPage(r)
	r.forgetPasswordPageFunc = defaultForgetPasswordPage(r)
	r.resetPasswordLinkSentPageFunc = defaultResetPasswordLinkSentPage(r)
	r.resetPasswordPageFunc = defaultResetPasswordPage(r)
	r.changePasswordPageFunc = defaultChangePasswordPage(r)
	r.totpSetupPageFunc = defaultTOTPSetupPage(r)
	r.totpValidatePageFunc = defaultTOTPValidatePage(r)

	return r
}

func (b *Builder) initAuthPrefixInterceptURLS() {
	b.authPrefixInterceptURLS = map[string]void{
		// to redirect to login page
		b.loginURL: {},
		// below paths need logged-in status
		b.logoutURL:           {},
		b.changePasswordURL:   {},
		b.doChangePasswordURL: {},
		totpSetupURL:          {},
		totpValidateURL:       {},
	}
}

func (b *Builder) AllowURL(v string) {
	b.allowURLS[v] = void{}
}

func (b *Builder) Secret(v string) (r *Builder) {
	b.secret = v
	return b
}

func (b *Builder) Providers(vs ...*Provider) (r *Builder) {
	if len(vs) == 0 {
		return b
	}
	b.oauthEnabled = true
	b.providers = vs
	var gothProviders []goth.Provider
	for _, v := range vs {
		gothProviders = append(gothProviders, v.Goth)
	}
	goth.UseProviders(gothProviders...)
	return b
}

func (b *Builder) AuthCookieName(v string) (r *Builder) {
	b.authCookieName = v
	return b
}

func (b *Builder) LoginURL(v string) (r *Builder) {
	b.loginURL = v
	return b
}

func (b *Builder) HomeURL(v string) (r *Builder) {
	b.homeURL = v
	return b
}

func (b *Builder) LoginPageFunc(v web.PageFunc) (r *Builder) {
	b.loginPageFunc = v
	return b
}

func (b *Builder) ForgetPasswordPageFunc(v web.PageFunc) (r *Builder) {
	b.forgetPasswordPageFunc = v
	return b
}

func (b *Builder) ResetPasswordLinkSentPageFunc(v web.PageFunc) (r *Builder) {
	b.resetPasswordLinkSentPageFunc = v
	return b
}

func (b *Builder) ResetPasswordPageFunc(v web.PageFunc) (r *Builder) {
	b.resetPasswordPageFunc = v
	return b
}

func (b *Builder) ChangePasswordPageFunc(v web.PageFunc) (r *Builder) {
	b.changePasswordPageFunc = v
	return b
}

func (b *Builder) NotifyUserOfResetPasswordLinkFunc(v NotifyUserOfResetPasswordLinkFunc) (r *Builder) {
	b.notifyUserOfResetPasswordLinkFunc = v
	return b
}

func (b *Builder) PasswordValidationFunc(v PasswordValidationFunc) (r *Builder) {
	b.passwordValidationFunc = v
	return b
}

func (b *Builder) wrapHook(v HookFunc) HookFunc {
	if v == nil {
		return nil
	}

	return func(r *http.Request, user interface{}) error {
		if GetCurrentUser(r) == nil {
			r = r.WithContext(context.WithValue(r.Context(), UserKey, user))
		}
		return v(r, user)
	}
}

func (b *Builder) AfterLogin(v HookFunc) (r *Builder) {
	b.afterLoginHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterFailedToLogin(v HookFunc) (r *Builder) {
	b.afterFailedToLoginHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterUserLocked(v HookFunc) (r *Builder) {
	b.afterUserLockedHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterLogout(v HookFunc) (r *Builder) {
	b.afterLogoutHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterSendResetPasswordLink(v HookFunc) (r *Builder) {
	b.afterSendResetPasswordLinkHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterResetPassword(v HookFunc) (r *Builder) {
	b.afterResetPasswordHook = b.wrapHook(v)
	return b
}

func (b *Builder) AfterChangePassword(v HookFunc) (r *Builder) {
	b.afterChangePasswordHook = b.wrapHook(v)
	return b
}

// seconds
// default 1h
func (b *Builder) SessionMaxAge(v int) (r *Builder) {
	b.sessionMaxAge = v
	return b
}

// extend the session if successfully authenticated
// default true
func (b *Builder) AutoExtendSession(v bool) (r *Builder) {
	b.autoExtendSession = v
	return b
}

// default 5
// MaxRetryCount <= 0 means no max retry count limit
func (b *Builder) MaxRetryCount(v int) (r *Builder) {
	b.maxRetryCount = v
	return b
}

func (b *Builder) TOTPEnabled(v bool) (r *Builder) {
	b.totpEnabled = v
	return b
}

func (b *Builder) TOTPIssuer(v string) (r *Builder) {
	b.totpIssuer = v
	return b
}

func (b *Builder) NoForgetPasswordLink(v bool) (r *Builder) {
	b.noForgetPasswordLink = v
	return b
}

func (b *Builder) DB(v *gorm.DB) (r *Builder) {
	b.db = v
	return b
}

func (b *Builder) I18n(v *i18n.Builder) (r *Builder) {
	b.i18nBuilder = v
	b.registerI18n()
	return b
}

func (b *Builder) registerI18n() {
	b.i18nBuilder.RegisterForModule(language.English, I18nLoginKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLoginKey, Messages_zh_CN)
}

func (b *Builder) UserModel(m interface{}) (r *Builder) {
	b.userModel = m
	b.tUser = underlyingReflectType(reflect.TypeOf(m))
	b.snakePrimaryField = snakePrimaryField(m)
	if _, ok := m.(UserPasser); ok {
		b.userPassEnabled = true
	}
	if _, ok := m.(OAuthUser); ok {
		b.oauthEnabled = true
	}
	if _, ok := m.(SessionSecurer); ok {
		b.sessionSecureEnabled = true
	}
	return b
}

func (b *Builder) newUserObject() interface{} {
	return reflect.New(b.tUser).Interface()
}

func (b *Builder) findUserByID(id string) (user interface{}, err error) {
	m := b.newUserObject()
	err = b.db.Where(fmt.Sprintf("%s = ?", b.snakePrimaryField), id).
		First(m).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errUserNotFound
		}
		return nil, err
	}
	return m, nil
}

// completeUserAuthCallback is for url "/auth/{provider}/callback"
func (b *Builder) completeUserAuthCallback(w http.ResponseWriter, r *http.Request) {
	var err error
	var user interface{}
	defer func() {
		if b.afterFailedToLoginHook != nil && err != nil && user != nil {
			b.afterFailedToLoginHook(r, user)
		}
	}()

	var ouser goth.User
	ouser, err = gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println("completeUserAuthWithSetCookie", err)
		setFailCodeFlash(w, FailCodeCompleteUserAuthFailed)
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	userID := ouser.UserID

	if b.userModel != nil {
		user, err = b.userModel.(OAuthUser).FindUserByOAuthUserID(b.db, b.newUserObject(), ouser.Provider, ouser.UserID)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				setFailCodeFlash(w, FailCodeSystemError)
				http.Redirect(w, r, b.logoutURL, http.StatusFound)
				return
			}
			// TODO: maybe the indentifier of some providers is not email
			indentifier := ouser.Email
			user, err = b.userModel.(OAuthUser).FindUserByOAuthIndentifier(b.db, b.newUserObject(), ouser.Provider, indentifier)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					setFailCodeFlash(w, FailCodeUserNotFound)
				} else {
					setFailCodeFlash(w, FailCodeSystemError)
				}
				http.Redirect(w, r, b.logoutURL, http.StatusFound)
				return
			}
			err = user.(OAuthUser).InitOAuthUserID(b.db, b.newUserObject(), ouser.Provider, indentifier, ouser.UserID)
			if err != nil {
				setFailCodeFlash(w, FailCodeSystemError)
				http.Redirect(w, r, b.logoutURL, http.StatusFound)
				return
			}
		}
		userID = objectID(user)
	}

	claims := UserClaims{
		Provider:         ouser.Provider,
		Email:            ouser.Email,
		Name:             ouser.Name,
		UserID:           userID,
		AvatarURL:        ouser.AvatarURL,
		RegisteredClaims: b.genBaseSessionClaim(userID),
	}
	if user == nil {
		user = &claims
	}

	if b.afterLoginHook != nil {
		if herr := b.afterLoginHook(r, user); herr != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.loginURL, http.StatusFound)
			return
		}
	}

	if err := b.setSecureCookiesByClaims(w, user, claims); err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	redirectURL := b.homeURL
	if v := b.getContinueURL(w, r); v != "" {
		redirectURL = v
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return
}

// return user if account exists even if there is an error returned
func (b *Builder) authUserPass(account string, password string) (user interface{}, err error) {
	user, err = b.userModel.(UserPasser).FindUser(b.db, b.newUserObject(), account)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errUserNotFound
		}
		return nil, err
	}

	u := user.(UserPasser)
	if u.GetLocked() {
		return user, errUserLocked
	}

	if !u.IsPasswordCorrect(password) {
		if b.maxRetryCount > 0 {
			if err = u.IncreaseRetryCount(b.db, b.newUserObject()); err != nil {
				return user, err
			}
			if u.GetLoginRetryCount() >= b.maxRetryCount {
				if err = u.LockUser(b.db, b.newUserObject()); err != nil {
					return user, err
				}
				return user, errUserGetLocked
			}
		}

		return user, errWrongPassword
	}

	if u.GetLoginRetryCount() != 0 {
		if err = u.UnlockUser(b.db, b.newUserObject()); err != nil {
			return user, err
		}
	}
	return user, nil
}

func (b *Builder) userpassLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	var user interface{}
	defer func() {
		if b.afterFailedToLoginHook != nil && err != nil && user != nil {
			b.afterFailedToLoginHook(r, user)
		}
	}()

	account := r.FormValue("account")
	password := r.FormValue("password")
	user, err = b.authUserPass(account, password)
	if err != nil {
		if err == errUserGetLocked && b.afterUserLockedHook != nil {
			if herr := b.afterUserLockedHook(r, user); herr != nil {
				setFailCodeFlash(w, FailCodeSystemError)
				http.Redirect(w, r, b.loginURL, http.StatusFound)
				return
			}
		}

		code := FailCodeSystemError
		switch err {
		case errWrongPassword, errUserNotFound:
			code = FailCodeIncorrectAccountNameOrPassword
		case errUserLocked, errUserGetLocked:
			code = FailCodeUserLocked
		}
		setFailCodeFlash(w, code)
		setWrongLoginInputFlash(w, WrongLoginInputFlash{
			Ia: account,
			Ip: password,
		})
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	u := user.(UserPasser)
	userID := objectID(user)
	claims := UserClaims{
		UserID:           userID,
		PassUpdatedAt:    u.GetPasswordUpdatedAt(),
		RegisteredClaims: b.genBaseSessionClaim(userID),
	}

	if !b.totpEnabled {
		if b.afterLoginHook != nil {
			if herr := b.afterLoginHook(r, user); herr != nil {
				setFailCodeFlash(w, FailCodeSystemError)
				http.Redirect(w, r, b.loginURL, http.StatusFound)
				return
			}
		}
	}

	if err = b.setSecureCookiesByClaims(w, user, claims); err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	if b.totpEnabled {
		if u.GetIsTOTPSetup() {
			http.Redirect(w, r, totpValidateURL, http.StatusFound)
			return
		}

		var key *otp.Key
		if key, err = totp.Generate(
			totp.GenerateOpts{
				Issuer:      b.totpIssuer,
				AccountName: u.GetAccountName(),
			},
		); err != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.logoutURL, http.StatusFound)
			return
		}

		if err = u.SetTOTPSecret(b.db, b.newUserObject(), key.Secret()); err != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.logoutURL, http.StatusFound)
			return
		}

		http.Redirect(w, r, totpSetupURL, http.StatusFound)
		return
	}

	redirectURL := b.homeURL
	if v := b.getContinueURL(w, r); v != "" {
		redirectURL = v
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return
}

func (b *Builder) genBaseSessionClaim(id string) jwt.RegisteredClaims {
	return genBaseClaims(id, b.sessionMaxAge)
}

func (b *Builder) setAuthCookiesFromUserClaims(w http.ResponseWriter, claims *UserClaims, secureSalt string) error {
	ss, err := signClaims(claims, b.secret)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     b.authCookieName,
		Value:    ss,
		Path:     "/",
		MaxAge:   b.sessionMaxAge,
		Expires:  time.Now().Add(time.Duration(b.sessionMaxAge) * time.Second),
		HttpOnly: true,
		Secure:   true,
	})

	if secureSalt != "" {
		ss, err = signClaims(&claims.RegisteredClaims, b.secret+secureSalt)
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{
			Name:     b.authSecureCookieName,
			Value:    ss,
			Path:     "/",
			MaxAge:   b.sessionMaxAge,
			Expires:  time.Now().Add(time.Duration(b.sessionMaxAge) * time.Second),
			HttpOnly: true,
			Secure:   true,
		})
	}

	return nil
}

func (b *Builder) cleanAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     b.authCookieName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     b.authSecureCookieName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   true,
	})
}

func (b *Builder) setContinueURL(w http.ResponseWriter, r *http.Request) {
	continueURL := r.RequestURI
	if strings.Contains(continueURL, "?__execute_event__=") {
		continueURL = r.Referer()
	}
	if continueURL != b.homeURL && !strings.HasPrefix(continueURL, "/auth/") {
		http.SetCookie(w, &http.Cookie{
			Name:     b.continueUrlCookieName,
			Value:    continueURL,
			Path:     "/",
			HttpOnly: true,
		})
	}
}

func (b *Builder) getContinueURL(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(b.continueUrlCookieName)
	if err != nil || c.Value == "" {
		return ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:     b.continueUrlCookieName,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Path:     "/",
		HttpOnly: true,
	})

	return c.Value
}

func (b *Builder) setSecureCookiesByClaims(w http.ResponseWriter, user interface{}, claims UserClaims) (err error) {
	var secureSalt string
	if b.sessionSecureEnabled {
		if user.(SessionSecurer).GetSecure() == "" {
			err = user.(SessionSecurer).UpdateSecure(b.db, b.newUserObject(), objectID(user))
			if err != nil {
				return err
			}
		}
		secureSalt = user.(SessionSecurer).GetSecure()
	}
	if err = b.setAuthCookiesFromUserClaims(w, &claims, secureSalt); err != nil {
		return err
	}

	return nil
}

// logout is for url "/logout/{provider}"
func (b *Builder) logout(w http.ResponseWriter, r *http.Request) {
	err := gothic.Logout(w, r)
	if err != nil {
		//
	}

	b.cleanAuthCookies(w)

	if b.afterLogoutHook != nil {
		user := GetCurrentUser(r)
		if user != nil {
			if herr := b.afterLogoutHook(r, user); herr != nil {
				setFailCodeFlash(w, FailCodeSystemError)
				http.Redirect(w, r, b.loginURL, http.StatusFound)
				return
			}
		}
	}

	http.Redirect(w, r, b.loginURL, http.StatusFound)
}

// beginAuth is for url "/auth/{provider}"
func (b *Builder) beginAuth(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (b *Builder) sendResetPasswordLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	failRedirectURL := "/auth/forget-password"

	account := strings.TrimSpace(r.FormValue("account"))
	if account == "" {
		setFailCodeFlash(w, FailCodeAccountIsRequired)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	u, err := b.userModel.(UserPasser).FindUser(b.db, b.newUserObject(), account)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setFailCodeFlash(w, FailCodeUserNotFound)
		} else {
			setFailCodeFlash(w, FailCodeSystemError)
		}
		setWrongForgetPasswordInputFlash(w, WrongForgetPasswordInputFlash{
			Account: account,
		})
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	_, createdAt, _ := u.(UserPasser).GetResetPasswordToken()
	if createdAt != nil {
		v := 60 - int(time.Now().Sub(*createdAt).Seconds())
		if v > 0 {
			setSecondsToRedoFlash(w, v)
			setWrongForgetPasswordInputFlash(w, WrongForgetPasswordInputFlash{
				Account: account,
			})
			http.Redirect(w, r, failRedirectURL, http.StatusFound)
			return
		}
	}

	token, err := u.(UserPasser).GenerateResetPasswordToken(b.db, b.newUserObject())
	if err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		setWrongForgetPasswordInputFlash(w, WrongForgetPasswordInputFlash{
			Account: account,
		})
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	link := fmt.Sprintf("%s://%s/auth/reset-password?id=%s&token=%s", scheme, r.Host, objectID(u), token)
	if err = b.notifyUserOfResetPasswordLinkFunc(u, link); err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		setWrongForgetPasswordInputFlash(w, WrongForgetPasswordInputFlash{
			Account: account,
		})
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	if b.afterSendResetPasswordLinkHook != nil {
		if herr := b.afterSendResetPasswordLinkHook(r, u); herr != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, failRedirectURL, http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/auth/reset-password-link-sent?a=%s", account), http.StatusFound)
	return
}

func (b *Builder) doResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	userID := r.FormValue("user_id")
	token := r.FormValue("token")
	failRedirectURL := fmt.Sprintf("/auth/reset-password?id=%s&token=%s", userID, token)
	if userID == "" {
		setFailCodeFlash(w, FailCodeUserNotFound)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}
	if token == "" {
		setFailCodeFlash(w, FailCodeInvalidToken)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	if password == "" {
		setFailCodeFlash(w, FailCodePasswordCannotBeEmpty)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}
	if confirmPassword != password {
		setFailCodeFlash(w, FailCodePasswordNotMatch)
		setWrongResetPasswordInputFlash(w, WrongResetPasswordInputFlash{
			Password:        password,
			ConfirmPassword: confirmPassword,
		})
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}
	if b.passwordValidationFunc != nil {
		msg, ok := b.passwordValidationFunc(password)
		if !ok {
			setCustomErrorMessageFlash(w, msg)
			setWrongResetPasswordInputFlash(w, WrongResetPasswordInputFlash{
				Password:        password,
				ConfirmPassword: confirmPassword,
			})
			http.Redirect(w, r, failRedirectURL, http.StatusFound)
			return
		}
	}

	u, err := b.findUserByID(userID)
	if err != nil {
		if err == errUserNotFound {
			setFailCodeFlash(w, FailCodeUserNotFound)
		} else {
			setFailCodeFlash(w, FailCodeSystemError)
		}
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	storedToken, _, expired := u.(UserPasser).GetResetPasswordToken()
	if expired {
		setFailCodeFlash(w, FailCodeTokenExpired)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}
	if token != storedToken {
		setFailCodeFlash(w, FailCodeInvalidToken)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	err = u.(UserPasser).ConsumeResetPasswordToken(b.db, b.newUserObject())
	if err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	err = u.(UserPasser).SetPassword(b.db, b.newUserObject(), password)
	if err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, failRedirectURL, http.StatusFound)
		return
	}

	if b.afterResetPasswordHook != nil {
		if herr := b.afterResetPasswordHook(r, u); herr != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, failRedirectURL, http.StatusFound)
			return
		}
	}

	setInfoCodeFlash(w, InfoCodePasswordSuccessfullyReset)
	http.Redirect(w, r, b.loginURL, http.StatusFound)
	return
}

func (b *Builder) doChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	oldPassword := r.FormValue("old_password")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	user := GetCurrentUser(r).(UserPasser)
	if !user.IsPasswordCorrect(oldPassword) {
		setFailCodeFlash(w, FailCodeIncorrectPassword)
		setWrongChangePasswordInputFlash(w, WrongChangePasswordInputFlash{
			OldPassword:     oldPassword,
			NewPassword:     password,
			ConfirmPassword: confirmPassword,
		})
		http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
		return
	}

	if password == "" {
		setFailCodeFlash(w, FailCodePasswordCannotBeEmpty)
		setWrongChangePasswordInputFlash(w, WrongChangePasswordInputFlash{
			OldPassword:     oldPassword,
			NewPassword:     password,
			ConfirmPassword: confirmPassword,
		})
		http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
		return
	}
	if confirmPassword != password {
		setFailCodeFlash(w, FailCodePasswordNotMatch)
		setWrongChangePasswordInputFlash(w, WrongChangePasswordInputFlash{
			OldPassword:     oldPassword,
			NewPassword:     password,
			ConfirmPassword: confirmPassword,
		})
		http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
		return
	}
	if b.passwordValidationFunc != nil {
		msg, ok := b.passwordValidationFunc(password)
		if !ok {
			setCustomErrorMessageFlash(w, msg)
			setWrongChangePasswordInputFlash(w, WrongChangePasswordInputFlash{
				OldPassword:     oldPassword,
				NewPassword:     password,
				ConfirmPassword: confirmPassword,
			})
			http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
			return
		}
	}

	err := user.SetPassword(b.db, b.newUserObject(), password)
	if err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
		return
	}

	if b.afterChangePasswordHook != nil {
		if herr := b.afterChangePasswordHook(r, user); herr != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.changePasswordURL, http.StatusFound)
			return
		}
	}

	setInfoCodeFlash(w, InfoCodePasswordSuccessfullyChanged)
	http.Redirect(w, r, b.loginURL, http.StatusFound)
	return
}

func (b *Builder) totpDo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var err error
	var user interface{}
	defer func() {
		if b.afterFailedToLoginHook != nil && err != nil && user != nil {
			b.afterFailedToLoginHook(r, user)
		}
	}()

	redirectURL := b.homeURL

	var claims *UserClaims
	claims, err = parseUserClaimsFromCookie(r, b.authCookieName, b.secret)
	if err != nil {
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	user, err = b.findUserByID(claims.UserID)
	if err != nil {
		if err == errUserNotFound {
			setFailCodeFlash(w, FailCodeUserNotFound)
		} else {
			setFailCodeFlash(w, FailCodeSystemError)
		}
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}
	u := user.(UserPasser)

	key := u.GetTOTPSecret()
	otp := r.FormValue("otp")
	isTOTPSetup := u.GetIsTOTPSetup()

	if !totp.Validate(otp, key) {
		err = errWrongTOTP
		setFailCodeFlash(w, FailCodeIncorrectTOTP)
		redirectURL := totpValidateURL
		if !isTOTPSetup {
			redirectURL = totpSetupURL
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	if !isTOTPSetup {
		if err = u.SetIsTOTPSetup(b.db, b.newUserObject(), true); err != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.logoutURL, http.StatusFound)
			return
		}
	}

	if b.afterLoginHook != nil {
		if herr := b.afterLoginHook(r, user); herr != nil {
			setFailCodeFlash(w, FailCodeSystemError)
			http.Redirect(w, r, b.loginURL, http.StatusFound)
			return
		}
	}

	claims.TOTPValidated = true
	err = b.setSecureCookiesByClaims(w, user, *claims)
	if err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
		http.Redirect(w, r, b.logoutURL, http.StatusFound)
		return
	}

	if v := b.getContinueURL(w, r); v != "" {
		redirectURL = v
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (b *Builder) Mount(mux *http.ServeMux) {
	if len(b.secret) == 0 {
		panic("secret is empty")
	}
	if b.userModel != nil {
		if b.db == nil {
			panic("db is required")
		}
	}

	wb := web.New()

	mux.HandleFunc(b.logoutURL, b.logout)
	mux.Handle(b.loginURL, b.i18nBuilder.EnsureLanguage(wb.Page(b.loginPageFunc)))

	if b.userPassEnabled {
		mux.HandleFunc("/auth/userpass/login", b.userpassLogin)
		mux.HandleFunc("/auth/do-reset-password", b.doResetPassword)
		mux.HandleFunc(b.doChangePasswordURL, b.doChangePassword)
		mux.Handle("/auth/reset-password", b.i18nBuilder.EnsureLanguage(wb.Page(b.resetPasswordPageFunc)))
		mux.Handle(b.changePasswordURL, b.i18nBuilder.EnsureLanguage(wb.Page(b.changePasswordPageFunc)))
		if !b.noForgetPasswordLink {
			mux.HandleFunc("/auth/send-reset-password-link", b.sendResetPasswordLink)
			mux.Handle("/auth/forget-password", b.i18nBuilder.EnsureLanguage(wb.Page(b.forgetPasswordPageFunc)))
			mux.Handle("/auth/reset-password-link-sent", b.i18nBuilder.EnsureLanguage(wb.Page(b.resetPasswordLinkSentPageFunc)))
		}
		if b.totpEnabled {
			mux.HandleFunc(totpDoURL, b.totpDo)
			mux.Handle(totpSetupURL, b.i18nBuilder.EnsureLanguage(wb.Page(b.totpSetupPageFunc)))
			mux.Handle(totpValidateURL, b.i18nBuilder.EnsureLanguage(wb.Page(b.totpValidatePageFunc)))
		}
	}
	if b.oauthEnabled {
		mux.HandleFunc("/auth/begin", b.beginAuth)
		mux.HandleFunc("/auth/callback", b.completeUserAuthCallback)
	}
}
