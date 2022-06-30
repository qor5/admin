package login

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"github.com/goplaid/web"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	. "github.com/theplant/htmlgo"
)

var (
	ErrUserNotFound = errors.New("user not found")
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
	secret                string
	loginURL              string
	fetchUserFunc         FetchUserToContextFunc
	authParamName         string
	homeURL               string
	continueUrlCookieName string
	extractors            []request.Extractor
	loginPageContentFunc  ContentFunc
	providers             []*Provider
	// seconds
	sessionMaxAge     int
	autoExtendSession bool
}

func New() *Builder {
	r := &Builder{
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

func (b *Builder) FetchUserToContextFunc(v FetchUserToContextFunc) (r *Builder) {
	b.fetchUserFunc = v
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
		http.Redirect(w, r, b.urlWithLoginFailCode(b.loginURL, code), http.StatusTemporaryRedirect)
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

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (b *Builder) completeUserAuthWithSetCookie(w http.ResponseWriter, r *http.Request) loginFailCode {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println("completeUserAuthWithSetCookie", err)
		return completeUserAuthFailed
	}

	claims := UserClaims{
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

	w.Header().Set("Location", b.loginURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// BeginAuth is for url "/auth/{provider}"
func (b *Builder) BeginAuth(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if b.completeUserAuthWithSetCookie(w, r) == 0 {
		http.Redirect(w, r, b.homeURL, http.StatusTemporaryRedirect)
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

func (b *Builder) Authenticate(in http.HandlerFunc) (r http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/auth/") && !strings.HasPrefix(r.URL.Path, "/auth/login") {
			in(w, r)
			return
		}

		if len(b.secret) == 0 {
			panic("secret is empty")
		}
		extractor := request.MultiExtractor(b.extractors)
		if len(b.extractors) == 0 {
			extractor = request.MultiExtractor{
				CookieExtractor(b.authParamName),
				AuthorizationHeaderExtractor{},
				request.ArgumentExtractor{b.authParamName},
				request.HeaderExtractor{b.authParamName},
			}
		}
		var claims UserClaims
		_, err := request.ParseFromRequest(r, extractor, b.keyFunc, request.WithClaims(&claims))

		if strings.HasPrefix(r.URL.Path, "/auth/login") {
			if err != nil || claims.Email == "" {
				in(w, r)
				return
			}
			if err == nil && claims.Email != "" {
				http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
				return
			}
		}

		if err != nil {
			log.Println(err)
			if b.homeURL != r.RequestURI {
				continueURL := r.RequestURI
				if strings.Contains(r.RequestURI, "?__execute_event__=") {
					continueURL = r.Referer()
				}
				http.SetCookie(w, &http.Cookie{
					Name:     b.continueUrlCookieName,
					Value:    continueURL,
					Path:     "/",
					HttpOnly: true,
				})
			}
			http.Redirect(w, r, b.loginURL, http.StatusTemporaryRedirect)
			return
		}

		newReq, err := b.fetchUserFunc(&claims, r)
		if err != nil {
			log.Println(err)
			code := systemError
			if err == ErrUserNotFound {
				code = userNotFound
			}
			http.Redirect(w, r, b.urlWithLoginFailCode(b.loginURL, code), http.StatusTemporaryRedirect)
			return
		}

		// extend the cookie if successfully authenticated
		if b.autoExtendSession {
			c, err := r.Cookie(b.authParamName)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     b.authParamName,
					Value:    c.Value,
					Path:     "/",
					MaxAge:   b.sessionMaxAge,
					Expires:  time.Now().Add(time.Duration(b.sessionMaxAge) * time.Second),
					HttpOnly: true,
				})
			}
		}

		in.ServeHTTP(w, newReq)
	}
}

type loginFailCode int

const (
	systemError loginFailCode = iota + 1
	completeUserAuthFailed
	userNotFound
)

var loginFailTexts = map[loginFailCode]string{
	systemError:            "System Error",
	completeUserAuthFailed: "Complete User Auth Failed",
	userNotFound:           "User Not Found",
}

var loginFailCodeQuery = "login_fc"

func (b *Builder) urlWithLoginFailCode(u string, code loginFailCode) string {
	pu, err := url.Parse(u)
	if err != nil {
		return u
	}
	q := pu.Query()
	q.Add(loginFailCodeQuery, fmt.Sprint(code))
	pu.RawQuery = q.Encode()
	return pu.String()
}

func (b *Builder) getLoginFailText(r *http.Request) string {
	sCode := r.URL.Query().Get(loginFailCodeQuery)
	if sCode == "" {
		return ""
	}
	code, err := strconv.Atoi(sCode)
	if err != nil {
		return ""
	}
	if code == 0 {
		return ""
	}
	text := loginFailTexts[loginFailCode(code)]
	if text == "" {
		text = loginFailTexts[systemError]
	}
	return text
}

func (b *Builder) defaultLoginPage(ctx *web.EventContext) (r web.PageResponse, err error) {

	ul := Div().Class("flex flex-col justify-center mt-8")
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

	loginFailText := b.getLoginFailText(ctx.R)
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
			Div(
				ul,
			).Class("max-w-xs sm:max-w-xl"),
		).Class("flex mt-4 justify-center h-screen"),
	)
	if b.loginPageContentFunc != nil {
		body = b.loginPageContentFunc(ctx, b.providers, body)
	}
	r.Body = body
	return
}

func (b *Builder) Mount(mux *http.ServeMux) {

	mux.HandleFunc("/auth/logout", b.Logout)
	mux.HandleFunc("/auth/begin", b.BeginAuth)
	mux.HandleFunc("/auth/callback", b.CompleteUserAuthCallback)
	mux.Handle("/auth/login", web.New().Page(b.defaultLoginPage))
}
