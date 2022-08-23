package login

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/goplaid/web"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	errUserNotFound    = errors.New("user not found")
	errUserPassChanged = errors.New("password changed")
	errWrongPassword   = errors.New("wrong password")
	errUserLocked      = errors.New("user locked")
)

type NotifyUserOfResetPasswordLinkFunc func(user interface{}, resetLink string) error
type PasswordValidationFunc func(password string) (message string, ok bool)

type Provider struct {
	Goth goth.Provider
	Key  string
	Text string
	Logo HTMLComponent
}

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

	loginURL string
	homeURL  string

	loginPageFunc                 web.PageFunc
	forgetPasswordPageFunc        web.PageFunc
	resetPasswordLinkSentPageFunc web.PageFunc
	resetPasswordPageFunc         web.PageFunc

	notifyUserOfResetPasswordLinkFunc NotifyUserOfResetPasswordLinkFunc
	passwordValidationFunc            PasswordValidationFunc

	db                   *gorm.DB
	userModel            interface{}
	snakePrimaryField    string
	tUser                reflect.Type
	userPassEnabled      bool
	oauthEnabled         bool
	sessionSecureEnabled bool
}

func New() *Builder {
	r := &Builder{
		authCookieName:        "auth",
		authSecureCookieName:  "qor5_auth_secure",
		continueUrlCookieName: "qor5_continue_url",
		loginURL:              "/auth/login",
		homeURL:               "/",
		sessionMaxAge:         60 * 60,
		autoExtendSession:     true,
	}
	r.loginPageFunc = defaultLoginPage(r)
	r.forgetPasswordPageFunc = defaultForgetPasswordPage(r)
	r.resetPasswordLinkSentPageFunc = defaultResetPasswordLinkSentPage(r)
	r.resetPasswordPageFunc = defaultResetPasswordPage(r)
	return r
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

func (b *Builder) NotifyUserOfResetPasswordLinkFunc(v NotifyUserOfResetPasswordLinkFunc) (r *Builder) {
	b.notifyUserOfResetPasswordLinkFunc = v
	return b
}

func (b *Builder) PasswordValidationFunc(v PasswordValidationFunc) (r *Builder) {
	b.passwordValidationFunc = v
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

// MaxRetryCount <= 0 means no max retry count limit
// default 0
func (b *Builder) MaxRetryCount(v int) (r *Builder) {
	b.maxRetryCount = v
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
		return nil, errUserLocked
	}

	if !u.IsPasswordCorrect(password) {
		if b.maxRetryCount > 0 {
			if err = u.IncreaseRetryCount(b.db, b.newUserObject()); err != nil {
				return nil, err
			}
			if u.GetLoginRetryCount() >= b.maxRetryCount {
				if err = u.LockUser(b.db, b.newUserObject()); err != nil {
					return nil, err
				}
				return nil, errUserLocked
			}
		}

		return nil, errWrongPassword
	}

	if u.GetLoginRetryCount() != 0 {
		if err = u.UnlockUser(b.db, b.newUserObject()); err != nil {
			return nil, err
		}
	}
	return user, nil
}

// completeUserAuthCallback is for url "/auth/{provider}/callback"
func (b *Builder) completeUserAuthCallback(w http.ResponseWriter, r *http.Request) {
	if err := b.completeUserAuthWithSetCookie(w, r); err != nil {
		http.Redirect(w, r, b.loginURL, http.StatusFound)
		return
	}
	redirectURL := b.homeURL
	c, _ := r.Cookie(b.continueUrlCookieName)
	if c != nil && c.Value != "" {
		redirectURL = c.Value
		http.SetCookie(w, &http.Cookie{
			Name:    b.continueUrlCookieName,
			Value:   "",
			MaxAge:  -1,
			Expires: time.Unix(1, 0),
			Path:    "/",
		})
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
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
	})
	http.SetCookie(w, &http.Cookie{
		Name:     b.authSecureCookieName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
	})
}

func (b *Builder) completeUserAuthWithSetCookie(w http.ResponseWriter, r *http.Request) error {
	var err error
	var claims UserClaims
	var user interface{}
	if r.FormValue("login_type") == "1" {
		account := r.FormValue("account")
		password := r.FormValue("password")
		user, err = b.authUserPass(account, password)
		if err != nil {
			code := FailCodeSystemError
			switch err {
			case errWrongPassword, errUserNotFound:
				code = FailCodeIncorrectAccountNameOrPassword
			case errUserLocked:
				code = FailCodeUserLocked
			}

			setFailCodeFlash(w, code)

			if code == FailCodeIncorrectAccountNameOrPassword {
				setWrongLoginInputFlash(w, WrongLoginInputFlash{
					Ia: account,
					Ip: password,
				})
			}

			return err
		}

		userID := objectID(user)
		claims = UserClaims{
			UserID:           userID,
			PassUpdatedAt:    user.(UserPasser).GetPasswordUpdatedAt(),
			RegisteredClaims: b.genBaseSessionClaim(userID),
		}
	} else {
		ouser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Println("completeUserAuthWithSetCookie", err)
			setFailCodeFlash(w, FailCodeCompleteUserAuthFailed)
			return err
		}
		userID := ouser.UserID
		if b.userModel != nil {
			user, err = b.userModel.(OAuthUser).FindUserByOAuthUserID(b.db, b.newUserObject(), ouser.Provider, ouser.UserID)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					setFailCodeFlash(w, FailCodeSystemError)
					return err
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
					return err
				}
				err = user.(OAuthUser).InitOAuthUserID(b.db, b.newUserObject(), ouser.Provider, indentifier, ouser.UserID)
				if err != nil {
					setFailCodeFlash(w, FailCodeSystemError)
					return err
				}
			}
			userID = objectID(user)
		}

		claims = UserClaims{
			Provider:         ouser.Provider,
			Email:            ouser.Email,
			Name:             ouser.Name,
			UserID:           userID,
			AvatarURL:        ouser.AvatarURL,
			RegisteredClaims: b.genBaseSessionClaim(userID),
		}
	}

	var secureSalt string
	if b.sessionSecureEnabled {
		if user.(SessionSecurer).GetSecure() == "" {
			err = user.(SessionSecurer).UpdateSecure(b.db, b.newUserObject(), objectID(user))
			if err != nil {
				setFailCodeFlash(w, FailCodeSystemError)
				return err
			}
		}
		secureSalt = user.(SessionSecurer).GetSecure()
	}
	if err := b.setAuthCookiesFromUserClaims(w, &claims, secureSalt); err != nil {
		setFailCodeFlash(w, FailCodeSystemError)
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

	storedToken, expired := u.(UserPasser).GetResetPasswordToken()
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

	setNoticeCodeFlash(w, NoticeCodePasswordSuccessfullyReset)
	http.Redirect(w, r, "/auth/login", http.StatusFound)
	return
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

	mux.HandleFunc("/auth/logout", b.logout)
	mux.Handle(b.loginURL, wb.Page(b.loginPageFunc))

	if b.userPassEnabled {
		mux.HandleFunc("/auth/userpass/login", b.completeUserAuthCallback)
		mux.HandleFunc("/auth/do-reset-password", b.doResetPassword)
		mux.Handle("/auth/reset-password", wb.Page(b.resetPasswordPageFunc))
		if !b.noForgetPasswordLink {
			mux.HandleFunc("/auth/send-reset-password-link", b.sendResetPasswordLink)
			mux.Handle("/auth/forget-password", wb.Page(b.forgetPasswordPageFunc))
			mux.Handle("/auth/reset-password-link-sent", wb.Page(b.resetPasswordLinkSentPageFunc))
		}
	}
	if b.oauthEnabled {
		mux.HandleFunc("/auth/begin", b.beginAuth)
		mux.HandleFunc("/auth/callback", b.completeUserAuthCallback)
	}
}
