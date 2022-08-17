package login

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"github.com/goplaid/web"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/sunfmin/reflectutils"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	errUserNotFound = errors.New("user not found")
)

type FetchUserToContextFunc func(claim *UserClaims, r *http.Request) (newR *http.Request, err error)
type ContentFunc func(ctx *web.EventContext, providers []*Provider, in HTMLComponent) (r HTMLComponent)

type Provider struct {
	Goth goth.Provider
	Key  string
	Text string
	Logo HTMLComponent
}

type Builder struct {
	db *gorm.DB

	secret                string
	loginURL              string
	authParamName         string
	homeURL               string
	continueUrlCookieName string
	extractors            []request.Extractor
	loginPageContentFunc  ContentFunc
	providers             []*Provider
	// seconds
	sessionMaxAge     int
	autoExtendSession bool

	tUser        reflect.Type
	withUserPass bool
	withOAuth    bool
}

func New(db *gorm.DB) *Builder {
	r := &Builder{
		db:                    db,
		authParamName:         "auth",
		loginURL:              "/auth/login",
		homeURL:               "/",
		continueUrlCookieName: "qor5_continue_url",
		sessionMaxAge:         60 * 60,
		autoExtendSession:     true,
	}
	return r
}

func (b *Builder) Secret(v string) (r *Builder) {
	b.secret = v
	return b
}

func (b *Builder) LoginURL(v string) (r *Builder) {
	b.loginURL = v
	return b
}

func (b *Builder) Providers(vs ...*Provider) (r *Builder) {
	b.providers = vs
	var gothProviders []goth.Provider
	for _, v := range vs {
		gothProviders = append(gothProviders, v.Goth)
	}
	goth.UseProviders(gothProviders...)
	return b
}

func (b *Builder) Extractors(vs ...request.Extractor) (r *Builder) {
	b.extractors = vs
	return b
}

func (b *Builder) HomeURL(v string) (r *Builder) {
	b.homeURL = v
	return b
}

func (b *Builder) LoginPageFunc(v ContentFunc) (r *Builder) {
	b.loginPageContentFunc = v
	return b
}

func (b *Builder) AuthParamName(v string) (r *Builder) {
	b.authParamName = v
	return b
}

// seconds
// default 1h
func (b *Builder) SessionMaxAge(v int) (r *Builder) {
	b.sessionMaxAge = v
	return b
}

// default true
func (b *Builder) AutoExtendSession(v bool) (r *Builder) {
	b.autoExtendSession = v
	return b
}

func underlyingReflectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return underlyingReflectType(t.Elem())
	}
	return t
}

func (b *Builder) UserModel(v interface{}) (r *Builder) {
	b.tUser = underlyingReflectType(reflect.TypeOf(v))
	if _, ok := v.(UserPasser); ok {
		b.withUserPass = true
	}
	if _, ok := v.(OAuthUser); ok {
		b.withOAuth = true
	}
	return b
}

func (b *Builder) authUserPass(username string, password string) (userID string, ok bool) {
	u := reflect.New(b.tUser).Interface()
	if err := b.db.Where("username = ?", username).First(u).Error; err != nil {
		return "", false
	}
	if c := u.(UserPasser).IsPasswordCorrect(password); !c {
		return "", false
	}
	return fmt.Sprint(reflectutils.MustGet(u, "ID")), true
}

type UserClaims struct {
	Provider  string
	Email     string
	Name      string
	UserID    string
	AvatarURL string
	Location  string
	IDToken   string
	jwt.RegisteredClaims
}

// CompleteUserAuthCallback is for url "/auth/{provider}/callback"
func (b *Builder) CompleteUserAuthCallback(w http.ResponseWriter, r *http.Request) {
	if code := b.completeUserAuthWithSetCookie(w, r); code != 0 {
		b.setFailCodeFlash(w, code)
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

func (b *Builder) completeUserAuthWithSetCookie(w http.ResponseWriter, r *http.Request) loginFailCode {
	var claims UserClaims
	if r.FormValue("login_type") == "1" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		userID, ok := b.authUserPass(username, password)
		if !ok {
			b.setWrongInputFlash(w, wrongInputFlash{
				Iu: username,
				Ip: password,
			})
			return incorrectUsernameOrPassword
		}

		claims = UserClaims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				// Make the jwt 24 hour, don't care about the user.ExpireAt because it is the use refresh token to fetch
				// access token expire time
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Subject:   userID,
				ID:        userID,
			},
		}
	} else {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Println("completeUserAuthWithSetCookie", err)
			return completeUserAuthFailed
		}

		claims = UserClaims{
			Provider:  user.Provider,
			Email:     user.Email,
			Name:      user.Name,
			UserID:    user.UserID,
			AvatarURL: user.AvatarURL,
			RegisteredClaims: jwt.RegisteredClaims{
				// Make the jwt 24 hour, don't care about the user.ExpireAt because it is the use refresh token to fetch
				// access token expire time
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Subject:   user.Email,
				ID:        user.UserID,
			},
		}
	}
	ss, err := b.SignClaims(&claims)
	if err != nil {
		return systemError
	}
	http.SetCookie(w, &http.Cookie{
		Name:     b.authParamName,
		Value:    ss,
		Path:     "/",
		MaxAge:   b.sessionMaxAge,
		Expires:  time.Now().Add(time.Duration(b.sessionMaxAge) * time.Second),
		HttpOnly: true,
	})

	return 0
}

func (b *Builder) SignClaims(claims *UserClaims) (signed string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err = token.SignedString([]byte(b.secret))
	return
}

// Logout is for url "/logout/{provider}"
func (b *Builder) Logout(w http.ResponseWriter, r *http.Request) {
	err := gothic.Logout(w, r)
	if err != nil {
		//
	}
	http.SetCookie(w, &http.Cookie{
		Name:     b.authParamName,
		Value:    "",
		Path:     "/",
		Domain:   "",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
	})

	http.Redirect(w, r, b.loginURL, http.StatusFound)
}

// BeginAuth is for url "/auth/{provider}"
func (b *Builder) BeginAuth(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if b.completeUserAuthWithSetCookie(w, r) == 0 {
		http.Redirect(w, r, b.homeURL, http.StatusFound)
		return
	}
	gothic.BeginAuthHandler(w, r)
}

type CookieExtractor string

func (e CookieExtractor) ExtractToken(req *http.Request) (string, error) {
	ck, err := req.Cookie(string(e))
	if err != nil {
		return "", request.ErrNoTokenInRequest
	}

	if len(ck.Value) == 0 {
		return "", request.ErrNoTokenInRequest
	}

	return ck.Value, nil
}

type AuthorizationHeaderExtractor struct{}

func (e AuthorizationHeaderExtractor) ExtractToken(req *http.Request) (string, error) {
	if ah := req.Header.Get("Authorization"); ah != "" {
		// remove bearer
		segs := strings.Split(ah, " ")
		return segs[len(segs)-1], nil
	}
	return "", request.ErrNoTokenInRequest
}

func (b *Builder) keyFunc(t *jwt.Token) (interface{}, error) {
	return []byte(b.secret), nil
}

const failCodeFlashCookieName = "qor5_login_fc_flash"

type loginFailCode int

const (
	systemError loginFailCode = iota + 1
	completeUserAuthFailed
	userNotFound
	incorrectUsernameOrPassword
)

var loginFailTexts = map[loginFailCode]string{
	systemError:                 "System Error",
	completeUserAuthFailed:      "Complete User Auth Failed",
	userNotFound:                "User Not Found",
	incorrectUsernameOrPassword: "Incorrect username or password",
}

func (b *Builder) setFailCodeFlash(w http.ResponseWriter, c loginFailCode) {
	http.SetCookie(w, &http.Cookie{
		Name:  failCodeFlashCookieName,
		Value: fmt.Sprint(c),
		Path:  "/",
	})
}

const wrongInputFlashCookieName = "qor5_login_wi_flash"

type wrongInputFlash struct {
	Iu string // incorrect username
	Ip string // incorrect password
}

func (b *Builder) setWrongInputFlash(w http.ResponseWriter, f wrongInputFlash) {
	bf, _ := json.Marshal(&f)
	http.SetCookie(w, &http.Cookie{
		Name:  wrongInputFlashCookieName,
		Value: base64.StdEncoding.EncodeToString(bf),
		Path:  "/",
	})
}

type loginFlash struct {
	fc         loginFailCode
	wrongInput wrongInputFlash
}

func (b *Builder) getFlash(w http.ResponseWriter, r *http.Request) (f loginFlash) {
	c, err := r.Cookie(failCodeFlashCookieName)
	if err == nil {
		v, _ := strconv.Atoi(c.Value)
		f.fc = loginFailCode(v)
		http.SetCookie(w, &http.Cookie{
			Name:   failCodeFlashCookieName,
			Path:   "/",
			MaxAge: -1,
		})
	}

	c, err = r.Cookie(wrongInputFlashCookieName)
	if err == nil {
		v, _ := base64.StdEncoding.DecodeString(c.Value)
		wi := wrongInputFlash{}
		json.Unmarshal([]byte(v), &wi)
		f.wrongInput = wi
		http.SetCookie(w, &http.Cookie{
			Name:   wrongInputFlashCookieName,
			Path:   "/",
			MaxAge: -1,
		})
	}

	return f
}

func (b *Builder) defaultLoginPage(ctx *web.EventContext) (r web.PageResponse, err error) {
	flash := b.getFlash(ctx.W, ctx.R)
	loginFailText := loginFailTexts[flash.fc]

	wrapperClass := "flex pt-8 h-screen flex-col max-w-md mx-auto"

	var oauthHTML HTMLComponent
	if b.withOAuth {
		ul := Div().Class("flex flex-col justify-center mt-8 text-center")
		for _, provider := range b.providers {
			ul.AppendChildren(
				A().
					Href("/auth/begin?provider="+provider.Key).
					Class("px-6 py-3 mt-4 font-semibold text-gray-900 bg-white border-2 border-gray-500 rounded-md shadow outline-none hover:bg-yellow-50 hover:border-yellow-400 focus:outline-none").
					Children(
						provider.Logo,
						Text(provider.Text),
					),
			)
		}

		oauthHTML = Div(
			ul,
		)
	}

	var userPassHTML HTMLComponent
	if b.withUserPass {
		wrapperClass += " pt-16"
		userPassHTML = Div(
			Form(
				Input("login_type").Type("hidden").Value("1"),
				Div(
					Label("Username").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("username"),
					Input("username").Placeholder("Username").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
						Value(flash.wrongInput.Iu),
				),
				Div(
					Label("Password").Class("block mb-2 text-sm text-gray-600 dark:text-gray-200").For("password"),
					Input("password").Placeholder("Password").Type("password").Class("block w-full px-4 py-2 mt-2 text-gray-700 placeholder-gray-400 bg-white border border-gray-200 rounded-md dark:placeholder-gray-600 dark:bg-gray-900 dark:text-gray-300 dark:border-gray-700 focus:border-blue-400 dark:focus:border-blue-400 focus:ring-blue-400 focus:outline-none focus:ring focus:ring-opacity-40").
						Value(flash.wrongInput.Ip),
				).Class("mt-6"),
				Div(
					Button("Sign In").Class("w-full px-6 py-3 tracking-wide text-white transition-colors duration-200 transform bg-blue-500 rounded-md hover:bg-blue-400 focus:outline-none focus:bg-blue-400 focus:ring focus:ring-blue-300 focus:ring-opacity-50"),
				).Class("mt-6"),
			).Method("post").Action("/auth/userpass/login"),
		)
	}
	var body HTMLComponent = Div(
		Style(StyleCSS),
		If(loginFailText != "",
			Div().Class("bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative text-center -mb-8").
				Role("alert").
				Children(
					Span(loginFailText).Class("block sm:inline"),
				),
		),
		Div(
			userPassHTML,
			oauthHTML,
		).Class(wrapperClass),
	)
	if b.loginPageContentFunc != nil {
		body = b.loginPageContentFunc(ctx, b.providers, body)
	}
	r.Body = body
	return
}

func (b *Builder) Mount(mux *http.ServeMux) {
	if len(b.secret) == 0 {
		panic("secret is empty")
	}

	mux.HandleFunc("/auth/logout", b.Logout)
	mux.HandleFunc("/auth/begin", b.BeginAuth)
	mux.HandleFunc("/auth/callback", b.CompleteUserAuthCallback)
	mux.HandleFunc("/auth/userpass/login", b.CompleteUserAuthCallback)
	mux.Handle(b.loginURL, web.New().Page(b.defaultLoginPage))
}
