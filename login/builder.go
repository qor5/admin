package login

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"github.com/goplaid/web"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	. "github.com/theplant/htmlgo"
)

type FetchUserFunc func(claim *UserClaims, r *http.Request) (newR *http.Request, err error)

type Provider struct {
	Goth    goth.Provider
	Key     string
	Text    string
	LogoURL string
}

type Builder struct {
	secret        string
	loginURL      string
	fetchUserFunc FetchUserFunc
	authParamName string
	homeURL       string
	extractors    []request.Extractor
	loginPageFunc web.PageFunc
	providers     []*Provider
}

func New() *Builder {
	r := &Builder{
		authParamName: "auth",
		homeURL:       "/",
	}
	r.loginPageFunc = r.defaultLoginPage
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

func (b *Builder) LoginPageFunc(v web.PageFunc) (r *Builder) {
	b.loginPageFunc = v
	return b
}

func (b *Builder) AuthParamName(v string) (r *Builder) {
	b.authParamName = v
	return b
}

func (b *Builder) FetchUserFunc(v FetchUserFunc) (r *Builder) {
	b.fetchUserFunc = v
	return b
}

type UserClaims struct {
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

	if b.completeUserAuthWithSetCookie(w, r) != nil {
		http.Redirect(w, r, b.loginURL, http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, b.homeURL, http.StatusTemporaryRedirect)
}

func (b *Builder) completeUserAuthWithSetCookie(w http.ResponseWriter, r *http.Request) error {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		return err
	}

	claims := UserClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(user.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
			ID:        user.UserID,
		},
	}
	ss, err := b.SignClaims(&claims)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     b.authParamName,
		Value:    ss,
		Path:     "/",
		Expires:  user.ExpiresAt,
		Secure:   true,
		HttpOnly: true,
	})

	return nil
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
	w.Header().Set("Location", b.loginURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// BeginAuth is for url "/auth/{provider}"
func (b *Builder) BeginAuth(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	if b.completeUserAuthWithSetCookie(w, r) == nil {
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
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, b.loginURL, http.StatusTemporaryRedirect)
			return
		}

		newReq, err := b.fetchUserFunc(&claims, r)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, b.loginURL, http.StatusTemporaryRedirect)
			return
		}

		in.ServeHTTP(w, newReq)
	}
}

func (b *Builder) defaultLoginPage(ctx *web.EventContext) (r web.PageResponse, err error) {
	ul := Ul()
	for _, provider := range b.providers {
		ul.AppendChildren(
			Li(
				A().Text(provider.Text).Href("/auth/begin?provider=" + provider.Key),
			),
		)
	}
	r.Body = Div(ul)
	return
}

func (b *Builder) Mount(mux *http.ServeMux) {

	mux.HandleFunc("/auth/logout", b.Logout)
	mux.HandleFunc("/auth/begin", b.BeginAuth)
	mux.HandleFunc("/auth/callback", b.CompleteUserAuthCallback)
	mux.Handle("/auth/login", web.New().Page(b.loginPageFunc))
}
