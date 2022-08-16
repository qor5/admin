package admin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/markbates/goth/providers/google"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/login"
	"github.com/qor/qor5/note"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func getCurrentUser(r *http.Request) (u *models.User) {
	u, ok := r.Context().Value(_userKey).(*models.User)
	if !ok {
		return nil
	}

	return u
}

type contextUserKey int

const _userKey contextUserKey = 1

func newLoginBuilder(db *gorm.DB) *login.Builder {
	return login.New(db).
		UserModel(&models.User{}).
		Secret(os.Getenv("LOGIN_SECRET")).
		Providers(
			&login.Provider{
				Goth: google.New(os.Getenv("LOGIN_GOOGLE_KEY"), os.Getenv("LOGIN_GOOGLE_SECRET"), os.Getenv("BASE_URL")+"/auth/callback?provider=google"),
				Key:  "google",
				Text: "Login with Google",
				Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" class="inline w-4 h-4 mr-3 text-gray-900 fill-current" viewBox="0 0 48 48" width="48px" height="48px"><path fill="#fbc02d" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12	s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20	s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#e53935" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039	l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4caf50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36	c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1565c0" d="M43.611,20.083L43.595,20L42,20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571	c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg>`),
			},
		).
		FetchUserToContextFunc(func(claim *login.UserClaims, r *http.Request) (newR *http.Request, err error) {
			u := &models.User{}
			if claim.Provider == "" {
				err = db.Where("id = ?", claim.UserID).
					Preload("Roles").
					First(u).
					Error
			} else {
				err = db.Where("o_auth_provider = ? and o_auth_user_id = ?", claim.Provider, claim.UserID).
					Preload("Roles").
					First(u).
					Error
				if err == gorm.ErrRecordNotFound {
					u = &models.User{
						UserPass: login.UserPass{
							Username: claim.Email,
						},
						Name:          claim.Name,
						AvatarURL:     claim.AvatarURL,
						OAuthProvider: claim.Provider,
						OAuthUserID:   claim.UserID,
					}
					err = db.Create(u).Error
					if err != nil {
						panic(err)
					}
				}
			}
			if err != nil {
				panic(err)
			}
			// TODO: update user info if claim info changed?

			newR = r.WithContext(context.WithValue(r.Context(), _userKey, u))
			newR = newR.WithContext(context.WithValue(newR.Context(), note.UserIDKey, u.ID))
			newR = newR.WithContext(context.WithValue(newR.Context(), note.UserKey, fmt.Sprintf("%v (%v)", u.Name, u.Username)))
			return
		}).
		HomeURL("/admin")
}

func authenticate(loginBuilder *login.Builder) func(next http.Handler) http.Handler {
	re := regexp.MustCompile(`\.(css|js|gif|jpg|jpeg|png|ico|svg|ttf|eot|woff|woff2)$`)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			if strings.HasSuffix(p, "/") {
				p = p[:len(p)-1]
				if p == "" {
					p = "/admin"
				}
				r.URL.Path = p
				http.Redirect(w, r, r.URL.String(), http.StatusFound)
				return
			}

			if re.MatchString(strings.ToLower(r.URL.Path)) {
				next.ServeHTTP(w, r)
				return
			}

			loginBuilder.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			}).ServeHTTP(w, r)
		})
	}
}
