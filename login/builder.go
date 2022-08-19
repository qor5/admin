package login

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/goplaid/web"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	errUserNotFound    = errors.New("user not found")
	errUserPassChanged = errors.New("password changed")
	errWrongPassword   = errors.New("wrong password")
)

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
	sessionMaxAge     int
	autoExtendSession bool
	loginURL          string
	homeURL           string
	loginPageFunc     web.PageFunc

	ud           *userDao
	tUser        reflect.Type
	withUserPass bool
	withOAuth    bool
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
	b.withOAuth = true
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

func (b *Builder) UserModel(db *gorm.DB, m interface{}) (r *Builder) {
	b.tUser = underlyingReflectType(reflect.TypeOf(m))
	if _, ok := m.(UserPasser); ok {
		b.withUserPass = true
	}
	if _, ok := m.(OAuthUser); ok {
		b.withOAuth = true
	}

	b.ud = &userDao{
		db:    db,
		tUser: b.tUser,
	}
	return b
}

func (b *Builder) authUserPass(username string, password string) (user interface{}, err error) {
	user, err = b.ud.getUserByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errUserNotFound
		}
		return nil, err
	}
	if c := user.(UserPasser).IsPasswordCorrect(password); !c {
		return nil, errWrongPassword
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
	var claims UserClaims
	var secureSalt string
	if r.FormValue("login_type") == "1" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		user, err := b.authUserPass(username, password)
		if err != nil {
			setFailCodeFlash(w, FailCodeIncorrectUsernameOrPassword)
			setWrongLoginInputFlash(w, WrongLoginInputFlash{
				Iu: username,
				Ip: password,
			})
			return err
		}

		userID := fmt.Sprint(reflectutils.MustGet(user, "ID"))
		claims = UserClaims{
			UserID:           userID,
			PassUpdatedAt:    user.(UserPasser).getPassUpdatedAt(),
			RegisteredClaims: b.genBaseSessionClaim(userID),
		}
		secureSalt = user.(UserPasser).getPassLoginSalt()
	} else {
		ouser, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Println("completeUserAuthWithSetCookie", err)
			setFailCodeFlash(w, FailCodeCompleteUserAuthFailed)
			return err
		}
		if b.tUser != nil {
			user, err := b.ud.getUserByOAuthUserID(ouser.Provider, ouser.UserID)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					setFailCodeFlash(w, FailCodeSystemError)
					return err
				}
				// TODO: maybe the indentifier of some providers is not email
				indentifier := ouser.Email
				user, err := b.ud.getUserByOAuthIndentifier(ouser.Provider, indentifier)
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						setFailCodeFlash(w, FailCodeUserNotFound)
					} else {
						setFailCodeFlash(w, FailCodeSystemError)
					}
					return err
				}
				user, err = b.ud.updateOAuthUserIDAndLoginSalt(fmt.Sprint(reflectutils.MustGet(user, "ID")), ouser.UserID, genHashSalt())
				if err != nil {
					setFailCodeFlash(w, FailCodeSystemError)
					return err
				}
			}
			secureSalt = user.(OAuthUser).getOAuthLoginSalt()
		}

		claims = UserClaims{
			Provider:         ouser.Provider,
			Email:            ouser.Email,
			Name:             ouser.Name,
			UserID:           ouser.UserID,
			AvatarURL:        ouser.AvatarURL,
			RegisteredClaims: b.genBaseSessionClaim(ouser.UserID),
		}
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

func (b *Builder) Mount(mux *http.ServeMux) {
	if len(b.secret) == 0 {
		panic("secret is empty")
	}

	mux.HandleFunc("/auth/logout", b.logout)
	mux.HandleFunc("/auth/begin", b.beginAuth)
	mux.HandleFunc("/auth/callback", b.completeUserAuthCallback)
	mux.HandleFunc("/auth/userpass/login", b.completeUserAuthCallback)
	mux.Handle(b.loginURL, web.New().Page(b.loginPageFunc))
}
