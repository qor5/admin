package admin

import (
	"fmt"
	"net/http"
	"os"

	"github.com/markbates/goth/providers/google"
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/example/models"
	plogin "github.com/qor5/admin/login"
	"github.com/qor5/admin/presets"
	"github.com/qor5/x/login"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	loginBuilder *login.Builder
)

func getCurrentUser(r *http.Request) (u *models.User) {
	u, ok := login.GetCurrentUser(r).(*models.User)
	if !ok {
		return nil
	}

	return u
}

func initLoginBuilder(db *gorm.DB, pb *presets.Builder, ab *activity.ActivityBuilder) {
	ab.RegisterModel(&models.User{})
	loginBuilder = plogin.New(pb).
		DB(db).
		UserModel(&models.User{}).
		Secret(os.Getenv("LOGIN_SECRET")).
		OAuthProviders(
			&login.Provider{
				Goth: google.New(os.Getenv("LOGIN_GOOGLE_KEY"), os.Getenv("LOGIN_GOOGLE_SECRET"), os.Getenv("BASE_URL")+"/auth/callback?provider=google"),
				Key:  "google",
				Text: "Login with Google",
				Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="16px" height="16px"><path fill="#fbc02d" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12	s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20	s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#e53935" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039	l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4caf50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36	c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1565c0" d="M43.611,20.083L43.595,20L42,20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571	c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg>`),
			},
		).
		HomeURLFunc(func(r *http.Request, user interface{}) string {
			return "/"
		}).
		MaxRetryCount(5).
		NotifyUserOfResetPasswordLinkFunc(func(user interface{}, resetLink string) error {
			fmt.Println("#########################################start")
			fmt.Println("reset password link:", resetLink)
			fmt.Println("#########################################end")
			return nil
		}).
		PasswordValidationFunc(func(password string) (message string, ok bool) {
			if len(password) < 12 {
				return "Password cannot be less than 12 characters", false
			}
			return "", true
		}).
		RecaptchaConfig(login.RecaptchaConfig{
			SiteKey:   os.Getenv("RECAPTCHA_SITE_KEY"),
			SecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),
		}).
		AfterLogin(func(r *http.Request, user interface{}, _ ...interface{}) error {
			if err := ab.AddCustomizedRecord("log-in", false, r.Context(), user); err != nil {
				return err
			}

			if err := addSessionLogByUserID(r, user.(*models.User).ID); err != nil {
				return err
			}

			return nil
		}).
		AfterFailedToLogin(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("login-failed", false, r.Context(), user)
		}).
		AfterUserLocked(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("locked", false, r.Context(), user)
		}).
		AfterLogout(func(r *http.Request, user interface{}, _ ...interface{}) error {
			if err := ab.AddCustomizedRecord("log-out", false, r.Context(), user); err != nil {
				return err
			}

			if err := expireCurrentSessionLog(r, user.(*models.User).ID); err != nil {
				return err
			}

			return nil
		}).
		AfterSendResetPasswordLink(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("send-reset-password-link", false, r.Context(), user)
		}).
		AfterResetPassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("reset-password", false, r.Context(), user)
		}).
		AfterChangePassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("change-password", false, r.Context(), user)
		}).
		AfterExtendSession(func(r *http.Request, user interface{}, vals ...interface{}) error {
			if err := updateCurrentSessionLog(r, user.(*models.User).ID, vals[0].(string)); err != nil {
				return err
			}

			return nil
		}).
		AfterTOTPCodeReused(func(r *http.Request, user interface{}, _ ...interface{}) error {
			fmt.Println("#########################################start")
			fmt.Println("totp code is reused!")
			fmt.Println("#########################################end")
			return nil
		}).TOTPEnabled(false)

	genInitialPasswordUser()
}

func genInitialPasswordUser() {
	email := os.Getenv("LOGIN_INITIAL_USER_EMAIL")
	password := os.Getenv("LOGIN_INITIAL_USER_PASSWORD")
	if email == "" || password == "" {
		return
	}

	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		panic(err)
	}
	if count > 0 {
		return
	}

	user := &models.User{
		Name: email,
		UserPass: login.UserPass{
			Account:  email,
			Password: password,
		},
	}
	user.EncryptPassword()
	if err := db.Create(user).Error; err != nil {
		panic(err)
	}
}
