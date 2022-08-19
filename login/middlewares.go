package login

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"

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

			claims, err := parseUserClaims(r, b.authCookieName, b.secret)
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
			var secureSalt string
			if b.tUser == nil {
				user = &claims
			} else {
				if claims.Provider == "" {
					user, err = b.ud.getUserByID(claims.UserID)
					if err == nil && user.(UserPasser).getPassUpdatedAt() != claims.PassUpdatedAt {
						err = errUserPassChanged
					}
					secureSalt = user.(UserPasser).getPassLoginSalt()
				} else {
					user, err = b.ud.getUserByOAuthUserID(claims.Provider, claims.UserID)
					if err == nil {
						user.(OAuthUser).setAvatar(claims.AvatarURL)
					}
					secureSalt = user.(OAuthUser).getOAuthLoginSalt()
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
				_, err := parseBaseClaims(r, b.authSecureCookieName, b.secret+secureSalt)
				if err != nil {
					http.Redirect(w, r, "/auth/logout", http.StatusFound)
					return
				}
			}

			r = r.WithContext(context.WithValue(r.Context(), _userKey, user))

			if b.autoExtendSession {
				claims.RegisteredClaims = b.genBaseSessionClaim(claims.UserID)
				if err := b.setAuthCookiesFromUserClaims(w, claims, secureSalt); err != nil {
					setFailCodeFlash(w, FailCodeSystemError)
					http.Redirect(w, r, "/auth/logout", http.StatusFound)
					return
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
