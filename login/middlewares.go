package login

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
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
			if strings.HasPrefix(path, "/auth/") &&
				path != b.loginURL &&
				path != pathTOTPSetup &&
				path != pathTOTPValidate {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseUserClaimsFromCookie(r, b.authCookieName, b.secret)
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
						setFailCodeFlash(w, FailCodeUserNotFound)
					case errUserLocked:
						setFailCodeFlash(w, FailCodeUserLocked)
					case errUserPassChanged:
						setWarnCodeFlash(w, WarnCodePasswordHasBeenChanged)
					default:
						setFailCodeFlash(w, FailCodeSystemError)
					}
					http.Redirect(w, r, "/auth/logout", http.StatusFound)
					return
				}

				if b.sessionSecureEnabled {
					secureSalt = user.(SessionSecurer).GetSecure()
					_, err := parseBaseClaimsFromCookie(r, b.authSecureCookieName, b.secret+secureSalt)
					if err != nil {
						http.Redirect(w, r, "/auth/logout", http.StatusFound)
						return
					}
				}
			} else {
				user = claims
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

			if claims.Provider == "" && b.totpEnabled && !claims.TOTPValidated {
				if path == b.loginURL {
					next.ServeHTTP(w, r)
					return
				}
				if !user.(UserPasser).GetIsTOTPSetup() {
					if path == pathTOTPSetup {
						next.ServeHTTP(w, r)
						return
					}
					http.Redirect(w, r, pathTOTPSetup, http.StatusFound)
					return
				}
				if path == pathTOTPValidate {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, pathTOTPValidate, http.StatusFound)
				return
			}

			if claims.TOTPValidated || claims.Provider != "" {
				if path == pathTOTPSetup || path == pathTOTPValidate {
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
	return r.Context().Value(_userKey)
}
