package admin

import (
	"net/http"
	"os"
	"strings"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/example/models"
	plogin "github.com/qor5/admin/login"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/role"
	"github.com/qor5/x/login"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

var (
	loginBuilder *login.Builder
	vh           *login.ViewHelper
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
				Goth: google.New(os.Getenv("LOGIN_GOOGLE_KEY"), os.Getenv("LOGIN_GOOGLE_SECRET"), os.Getenv("BASE_URL")+"/auth/callback?provider="+models.OAuthProviderGoogle),
				Key:  models.OAuthProviderGoogle,
				Text: "LoginProviderGoogleText",
				Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="16px" height="16px"><path fill="#fbc02d" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12	s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20	s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#e53935" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039	l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4caf50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36	c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1565c0" d="M43.611,20.083L43.595,20L42,20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571	c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg>`),
			},
			&login.Provider{
				Goth: microsoftonline.New(os.Getenv("LOGIN_MICROSOFTONLINE_KEY"), os.Getenv("LOGIN_MICROSOFTONLINE_SECRET"), os.Getenv("BASE_URL")+"/auth/callback"),
				Key:  models.OAuthProviderMicrosoftOnline,
				Text: "LoginProviderMicrosoftText",
				Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="16px" height="16px"><path fill="#f35325" d="M2 2h20v20H2z"/><path fill="#81bc06" d="M24 2h20v20H24z"/><path fill="#05a6f0" d="M2 24h20v20H2z"/><path fill="#ffba08" d="M24 24h20v20H24z"/></svg>`),
			},
			&login.Provider{
				Goth: github.New(os.Getenv("LOGIN_GITHUB_KEY"), os.Getenv("LOGIN_GITHUB_SECRET"), os.Getenv("BASE_URL")+"/auth/callback?provider="+models.OAuthProviderGithub),
				Key:  models.OAuthProviderGithub,
				Text: "LoginProviderGithubText",
				Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96" width="16px" height="16px"><path fill-rule="evenodd" clip-rule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z" fill="#24292f"/></svg>`),
			},
		).
		HomeURLFunc(func(r *http.Request, user interface{}) string {
			return "/"
		}).
		MaxRetryCount(5).
		Recaptcha(true, login.RecaptchaConfig{
			SiteKey:   os.Getenv("RECAPTCHA_SITE_KEY"),
			SecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),
		}).
		BeforeSetPassword(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
			u := user.(*models.User)
			if u.GetAccountName() == os.Getenv("LOGIN_INITIAL_USER_EMAIL") {
				return &login.NoticeError{
					Level:   login.NoticeLevel_Error,
					Message: "Cannot change password for public user",
				}
			}
			password := extraVals[0].(string)
			if len(password) < 12 {
				return &login.NoticeError{
					Level:   login.NoticeLevel_Error,
					Message: "Password cannot be less than 12 characters",
				}
			}
			return nil
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
		AfterOAuthComplete(func(r *http.Request, user interface{}, _ ...interface{}) error {
			u := user.(goth.User)
			if err := db.Where("o_auth_provider = ? and o_auth_identifier = ?", u.Provider, u.Email).First(&models.User{}).
				Error; err == gorm.ErrRecordNotFound {
				var name string
				at := strings.LastIndex(u.Email, "@")
				if at > 0 {
					name = u.Email[:at]
				} else {
					name = u.Email
				}

				user := &models.User{
					Name: name,
					OAuthInfo: login.OAuthInfo{
						OAuthProvider:   u.Provider,
						OAuthUserID:     u.UserID,
						OAuthIdentifier: u.Email,
						OAuthAvatar:     u.AvatarURL,
					},
				}

				if err := db.Create(user).Error; err != nil {
					panic(err)
				}

				if err := grantUserRole(user.ID, models.RoleManager); err != nil {
					panic(err)
				}
			}

			return nil
		}).
		AfterFailedToLogin(func(r *http.Request, user interface{}, _ ...interface{}) error {
			if user != nil {
				return ab.AddCustomizedRecord("login-failed", false, r.Context(), user)
			}
			return nil
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
		AfterConfirmSendResetPasswordLink(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
			resetLink := extraVals[0]
			_ = resetLink
			return ab.AddCustomizedRecord("send-reset-password-link", false, r.Context(), user)
		}).
		AfterResetPassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
			if err := expireAllSessionLogs(user.(*models.User).ID); err != nil {
				return err
			}
			return ab.AddCustomizedRecord("reset-password", false, r.Context(), user)
		}).
		AfterChangePassword(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return ab.AddCustomizedRecord("change-password", false, r.Context(), user)
		}).
		AfterExtendSession(func(r *http.Request, user interface{}, extraVals ...interface{}) error {
			oldToken := extraVals[0].(string)
			if err := updateCurrentSessionLog(r, user.(*models.User).ID, oldToken); err != nil {
				return err
			}

			return nil
		}).
		AfterTOTPCodeReused(func(r *http.Request, user interface{}, _ ...interface{}) error {
			return nil
		}).TOTP(false).MaxRetryCount(0)

	vh = loginBuilder.ViewHelper()
	loginBuilder.LoginPageFunc(loginPage(vh, pb))

	GenInitialUser()
}

func GenInitialUser() {
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

	if err := initDefaultRoles(); err != nil {
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
	if err := grantUserRole(user.ID, models.RoleManager); err != nil {
		panic(err)
	}
}

func grantUserRole(userID uint, roleName string) error {
	var roleID int
	if err := db.Table("roles").Where("name = ?", roleName).Pluck("id", &roleID).Error; err != nil {
		panic(err)
	}
	return db.Table("user_role_join").Create(
		&map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		}).Error
}

func initDefaultRoles() error {
	var cnt int64
	if err := db.Model(&role.Role{}).Count(&cnt).Error; err != nil {
		return err
	}

	if cnt == 0 {
		var roles []*role.Role
		for _, r := range models.DefaultRoles {
			roles = append(roles, &role.Role{
				Name: r,
			})
		}

		if err := db.Create(roles).Error; err != nil {
			return err
		}
	}

	return nil
}

func doOAuthCompleteInfo(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var isSubscribed bool
		position := r.FormValue("position")
		agree := r.FormValue("agree")
		if agree == "on" {
			isSubscribed = true
		}

		user := getCurrentUser(r)

		if err := db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"position":          position,
			"is_subscribed":     isSubscribed,
			"is_info_completed": true,
		}).Error; err != nil {
			http.Redirect(w, r, logoutURL, http.StatusFound)
		}

		http.Redirect(w, r, "/", http.StatusFound)

		return
	})
}
