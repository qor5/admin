package login

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type contextUserKey int

const _userKey contextUserKey = 1

var staticFileRe = regexp.MustCompile(`\.(css|js|gif|jpg|jpeg|png|ico|svg|ttf|eot|woff|woff2)$`)

func Authenticate(b *Builder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if staticFileRe.MatchString(strings.ToLower(r.URL.Path)) {
				next.ServeHTTP(w, r)
				return
			}

			path := strings.TrimRight(r.URL.Path, "/")
			if strings.HasPrefix(path, "/auth/") && path != b.loginURL {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseUserClaims(r, b.authParamName, b.secret)
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
				if path == b.loginURL {
					next.ServeHTTP(w, r)
				} else {
					http.Redirect(w, r, b.loginURL, http.StatusFound)
				}
				return
			}

			var user interface{}
			if b.tUser == nil {
				user = &claims
			} else {
				if claims.Provider == "" {
					user, err = b.ud.getUserByID(claims.UserID)
					if err == nil && user.(UserPasser).getPassUpdatedAt() != claims.PassUpdatedAt {
						err = errUserPassChanged
					}
				} else {
					user, err = b.ud.getUserByOAuthUserID(claims.Provider, claims.UserID)
					if err == nil {
						user.(OAuthUser).setAvatar(claims.AvatarURL)
					}
				}
			}
			if err != nil {
				log.Println(err)
				code := FailCodeSystemError
				if err == gorm.ErrRecordNotFound {
					code = FailCodeUserNotFound
				}
				if err == errUserPassChanged {
					code = 0
				}
				if code != 0 {
					setFailCodeFlash(w, code)
				}
				http.Redirect(w, r, "/auth/logout", http.StatusFound)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), _userKey, user))

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
					// FIXME: extend token lifetime
				}
			}

			if path == b.loginURL {
				http.Redirect(w, r, b.homeURL, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetCurrentUser(r *http.Request) (u interface{}) {
	return r.Context().Value(_userKey)
}
