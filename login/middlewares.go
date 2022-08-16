package login

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4/request"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

type contextUserKey int

const _userKey contextUserKey = 1

var staticFileRe = regexp.MustCompile(`\.(css|js|gif|jpg|jpeg|png|ico|svg|ttf|eot|woff|woff2)$`)

func fetchUserToContext(db *gorm.DB, tUser reflect.Type, claim *UserClaims, r *http.Request) (newR *http.Request, err error) {
	u := reflect.New(tUser).Interface()
	if claim.Provider == "" {
		err = db.Where("id = ?", claim.UserID).
			First(u).
			Error
	} else {
		err = db.Where("o_auth_provider = ? and o_auth_user_id = ?", claim.Provider, claim.UserID).
			First(u).
			Error
		if err == gorm.ErrRecordNotFound {
			err = db.Where("o_auth_provider = ? and o_auth_indentifier = ?", claim.Provider, claim.Email).
				First(u).
				Error
			if err == nil {
				err = db.Model(u).
					Where("id=?", fmt.Sprint(reflectutils.MustGet(u, "ID"))).
					Updates(map[string]interface{}{
						"o_auth_user_id": claim.UserID,
					}).
					Error
			}
		}
		if err == nil {
			u.(OAuthUser).SetAvatar(claim.AvatarURL)
		}
	}
	if err != nil {
		return r, err
	}

	return r.WithContext(context.WithValue(r.Context(), _userKey, u)), nil
}

func Authenticate(b *Builder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if staticFileRe.MatchString(strings.ToLower(r.URL.Path)) {
				next.ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "" || r.URL.Path == "/" {
				http.Redirect(w, r, b.homeURL, http.StatusFound)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/auth/") && !strings.HasPrefix(r.URL.Path, "/auth/login") {
				next.ServeHTTP(w, r)
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
					next.ServeHTTP(w, r)
					return
				}
				if err == nil && claims.Email != "" {
					http.Redirect(w, r, "/admin", http.StatusFound)
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
				http.Redirect(w, r, b.loginURL, http.StatusFound)
				return
			}

			newReq, err := fetchUserToContext(b.db, b.tUser, &claims, r)
			if err != nil {
				log.Println(err)
				code := systemError
				if err == ErrUserNotFound {
					code = userNotFound
				}
				http.Redirect(w, r, b.urlWithLoginFailCode(b.loginURL, code), http.StatusFound)
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

			next.ServeHTTP(w, newReq)
		})
	}
}

func GetCurrentUser(r *http.Request) (u interface{}) {
	return r.Context().Value(_userKey)
}
