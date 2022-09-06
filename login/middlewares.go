package login

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ContextUserKey int

const UserKey ContextUserKey = 1

var staticFileRe = regexp.MustCompile(`\.(css|js|gif|jpg|jpeg|png|ico|svg|ttf|eot|woff|woff2)$`)

func Authenticate(b *Builder) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if staticFileRe.MatchString(strings.ToLower(r.URL.Path)) {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := b.allowURLS[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			path := strings.TrimRight(r.URL.Path, "/")
			if _, ok := b.authPrefixInterceptURLS[path]; strings.HasPrefix(path, "/auth/") && !ok {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseUserClaimsFromCookie(r, b.authCookieName, b.secret)
			if err != nil {
				log.Println(err)
				b.setContinueURL(w, r)
				if path == b.loginURL {
					next.ServeHTTP(w, r)
				} else {
					http.Redirect(w, r, b.loginURL, http.StatusFound)
				}
				return
			}

			var user interface{}
			var secureSalt string
			if b.userModel != nil {
				var err error
				user, err = b.findUserByID(claims.UserID)
				if err == nil {
					if claims.Provider == "" {
						if user.(UserPasser).GetPasswordUpdatedAt() != claims.PassUpdatedAt {
							err = errUserPassChanged
						}
						if user.(UserPasser).GetLocked() {
							err = errUserLocked
						}
					} else {
						user.(OAuthUser).SetAvatar(claims.AvatarURL)
					}
				}
				if err != nil {
					log.Println(err)
					switch err {
					case errUserNotFound:
						setFailCodeFlash(b.cookieConfig, w, FailCodeUserNotFound)
					case errUserLocked:
						setFailCodeFlash(b.cookieConfig, w, FailCodeUserLocked)
					case errUserPassChanged:
						setWarnCodeFlash(b.cookieConfig, w, WarnCodePasswordHasBeenChanged)
					default:
						setFailCodeFlash(b.cookieConfig, w, FailCodeSystemError)
					}
					if path == b.logoutURL {
						next.ServeHTTP(w, r)
					} else {
						http.Redirect(w, r, b.logoutURL, http.StatusFound)
					}
					return
				}

				if b.sessionSecureEnabled {
					secureSalt = user.(SessionSecurer).GetSecure()
					_, err := parseBaseClaimsFromCookie(r, b.authSecureCookieName, b.secret+secureSalt)
					if err != nil {
						if path == b.logoutURL {
							next.ServeHTTP(w, r)
						} else {
							http.Redirect(w, r, b.logoutURL, http.StatusFound)
						}
						return
					}
				}
			} else {
				user = claims
			}

			if b.autoExtendSession && time.Now().Sub(claims.IssuedAt.Time).Seconds() > float64(b.sessionMaxAge)/10 {
				claims.RegisteredClaims = b.genBaseSessionClaim(claims.UserID)
				if err := b.setAuthCookiesFromUserClaims(w, claims, secureSalt); err != nil {
					setFailCodeFlash(b.cookieConfig, w, FailCodeSystemError)
					if path == b.logoutURL {
						next.ServeHTTP(w, r)
					} else {
						http.Redirect(w, r, b.logoutURL, http.StatusFound)
					}
					return
				}
			}

			r = r.WithContext(context.WithValue(r.Context(), UserKey, user))

			if path == b.logoutURL {
				next.ServeHTTP(w, r)
				return
			}

			if claims.Provider == "" && b.totpEnabled && !claims.TOTPValidated {
				if path == b.loginURL {
					next.ServeHTTP(w, r)
					return
				}
				if !user.(UserPasser).GetIsTOTPSetup() {
					if path == totpSetupURL {
						next.ServeHTTP(w, r)
						return
					}
					http.Redirect(w, r, totpSetupURL, http.StatusFound)
					return
				}
				if path == totpValidateURL {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, totpValidateURL, http.StatusFound)
				return
			}

			if claims.TOTPValidated || claims.Provider != "" {
				if path == totpSetupURL || path == totpValidateURL {
					http.Redirect(w, r, b.homeURL, http.StatusFound)
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
	return r.Context().Value(UserKey)
}
