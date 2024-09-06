package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	plogin "github.com/qor5/admin/v3/login"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/login"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/osenv"
	"github.com/theplant/testingutils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	login.UserPass
	login.OAuthInfo
	login.SessionSecure
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("/tmp/login_example.db"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}

	user := &User{
		UserPass: login.UserPass{
			Account:  "qor@theplant.jp",
			Password: "123",
		},
	}
	user.EncryptPassword()
	db.Create(user)
}

var (
	loginGoogleKey    = osenv.Get("LOGIN_GOOGLE_KEY", "Google client key for Login with Google", "")
	loginGoogleSecret = osenv.Get("LOGIN_GOOGLE_SECRET", "Google client secret for Login with Google", "")
	loginGithubKey    = osenv.Get("LOGIN_GITHUB_KEY", "Github client key for Login with Github", "")
	loginGithubSecret = osenv.Get("LOGIN_GITHUB_SECRET", "Github client secret for Login with Github", "")
)

func main() {
	pb := presets.New()

	lb := plogin.New(pb).
		DB(db).
		UserModel(&User{}).
		Secret("123").
		OAuthProviders(
			&login.Provider{
				Goth: google.New(loginGoogleKey, loginGoogleSecret, "http://localhost:9500/auth/callback?provider=google"),
				Key:  "google",
				Text: "Login with Google",
				Logo: h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="16px" height="16px"><path fill="#fbc02d" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12	s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20	s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#e53935" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039	l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4caf50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36	c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1565c0" d="M43.611,20.083L43.595,20L42,20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571	c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg>`),
			},
			&login.Provider{
				Goth: github.New(loginGithubKey, loginGithubSecret, "http://localhost/auth/callback?provider=github"),
				Key:  "github",
				Text: "Login with Github",
				Logo: h.RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96" width="16px" height="16px"><path fill-rule="evenodd" clip-rule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z" fill="#24292f"/></svg>`),
			},
		).
		WrapAfterConfirmSendResetPasswordLink(func(in login.HookFunc) login.HookFunc {
			return func(r *http.Request, user interface{}, extraVals ...interface{}) error {
				if err := in(r, user, extraVals...); err != nil {
					return err
				}
				resetLink := extraVals[0]
				fmt.Println("#########################################start")
				testingutils.PrintlnJson(resetLink)
				fmt.Println("#########################################end")
				return nil
			}
		}).
		TOTP(false)

	pb.ProfileFunc(func(ctx *web.EventContext) h.HTMLComponent {
		return h.A(h.Text("logout")).Href(lb.LogoutURL)
	})

	r := http.NewServeMux()
	r.Handle("/", pb)
	lb.Mount(r)

	mux := http.NewServeMux()
	mux.Handle("/", lb.Middleware()(r))
	server := &http.Server{
		Addr:              ":9500",
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}
	log.Println("serving at http://localhost:9500")
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
