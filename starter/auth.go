package starter

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
	"github.com/pkg/errors"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/role"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/login"
	"github.com/qor5/x/v3/perm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	. "github.com/theplant/htmlgo"

	plogin "github.com/qor5/admin/v3/login"
)

// AuthConfig contains all authentication-related configuration
type AuthConfig struct {
	Secret                string `confx:"secret" validate:"required"`
	BaseURL               string `confx:"baseURL" validate:"required,url"`
	InitialUserEmail      string `confx:"initialUserEmail" validate:"required,email"`
	InitialUserPassword   string `confx:"initialUserPassword" validate:"required,min=12"`
	InitialUserRole       string `confx:"initialUserRole" validate:"required"`
	GoogleClientKey       string `confx:"googleClientKey"`
	GoogleClientSecret    string `confx:"googleClientSecret"`
	MicrosoftClientKey    string `confx:"microsoftClientKey"`
	MicrosoftClientSecret string `confx:"microsoftClientSecret"`
	GitHubClientKey       string `confx:"githubClientKey"`
	GitHubClientSecret    string `confx:"githubClientSecret"`
	EnableRecaptcha       bool   `confx:"enableRecaptcha"`
	RecaptchaSiteKey      string `confx:"recaptchaSiteKey"`
	RecaptchaSecretKey    string `confx:"recaptchaSecretKey"`
	MaxRetryCount         int    `confx:"maxRetryCount" validate:"min=1"`
	EnableTOTP            bool   `confx:"enableTOTP"`
}

// createRoleBuilder creates and configures the role builder
func (a *Handler) createRoleBuilder() *role.Builder {
	roleBuilder := role.New(a.DB).
		AfterInstall(func(_ *presets.Builder, mb *presets.ModelBuilder) error {
			mb.Listing().SearchFunc(func(ctx *web.EventContext, params *presets.SearchParams) (*presets.SearchResult, error) {
				u := GetCurrentUser(ctx.R)
				qdb := a.DB.WithContext(ctx.R.Context())
				// If the current user doesn't have 'admin' role, do not allow them to view admin and manager roles
				if currentRoles := u.GetRoles(); !slices.Contains(currentRoles, RoleAdmin) {
					qdb = qdb.Where("name NOT IN (?)", []string{RoleAdmin, RoleManager})
				}
				result, err := gorm2op.DataOperator(qdb).Search(ctx, params)
				if err != nil {
					return nil, errors.Wrap(err, "failed to search roles")
				}
				return result, nil
			})
			return nil
		})

	a.Use(roleBuilder)
	return roleBuilder
}

func (a *Handler) createLoginBuilder(presetsBuilder *presets.Builder) *login.Builder {
	loginBuilder := plogin.New(presetsBuilder).
		DB(a.DB).
		UserModel(&User{}).
		Secret(a.Auth.Secret).
		OAuthProviders(buildOAuthProviders(&a.Auth)...).
		HomeURLFunc(func(_ *http.Request, _ any) string {
			return strings.TrimSuffix(a.Prefix, "/") + "/"
		}).
		Recaptcha(a.Auth.EnableRecaptcha, login.RecaptchaConfig{
			SiteKey:   a.Auth.RecaptchaSiteKey,
			SecretKey: a.Auth.RecaptchaSecretKey,
		}).
		WrapBeforeSetPassword(func(in login.HookFunc) login.HookFunc {
			return func(r *http.Request, user any, extraVals ...any) error {
				if err := in(r, user, extraVals...); err != nil {
					return err
				}

				u := user.(*User)
				if u.GetAccountName() == a.Auth.InitialUserEmail {
					return errors.WithStack(
						&login.NoticeError{
							Level:   login.NoticeLevel_Error,
							Message: "Cannot change password for public user",
						},
					)
				}

				password := extraVals[0].(string)
				if len(password) < 12 {
					return errors.WithStack(
						&login.NoticeError{
							Level:   login.NoticeLevel_Error,
							Message: "Password cannot be less than 12 characters",
						},
					)
				}

				return nil
			}
		}).
		WrapAfterOAuthComplete(func(in login.HookFunc) login.HookFunc {
			return func(r *http.Request, user any, extraVals ...any) error {
				if err := in(r, user, extraVals...); err != nil {
					return err
				}

				gothUser := user.(goth.User)
				if gothUser.Email == "" {
					return nil
				}

				db := a.DB.WithContext(r.Context())

				// Check if user already exists
				if err := db.Where("o_auth_provider = ? AND o_auth_identifier = ?", gothUser.Provider, gothUser.Email).
					First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					// Create new user from OAuth data
					var name string
					if at := strings.LastIndex(gothUser.Email, "@"); at > 0 {
						name = gothUser.Email[:at]
					} else {
						name = gothUser.Email
					}

					newUser := &User{
						Name:             name,
						Status:           StatusActive,
						RegistrationDate: time.Now(),
						OAuthInfo: login.OAuthInfo{
							OAuthProvider:   gothUser.Provider,
							OAuthUserID:     gothUser.UserID,
							OAuthIdentifier: gothUser.Email,
							OAuthAvatar:     gothUser.AvatarURL,
						},
					}
					if err := db.Create(newUser).Error; err != nil {
						return errors.Wrap(err, "failed to create user from OAuth")
					}

					// Grant manager role to new OAuth user
					if err := grantUserRole(r.Context(), db, newUser.ID, RoleManager); err != nil {
						return err
					}
				}

				return nil
			}
		}).
		TOTP(a.Auth.EnableTOTP).
		MaxRetryCount(a.Auth.MaxRetryCount)

	loginBuilder.LoginPageFunc(plogin.NewAdvancedLoginPage(func(ctx *web.EventContext, config *plogin.AdvancedLoginPageConfig) (*plogin.AdvancedLoginPageConfig, error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nDemoKey, Messages_en_US).(*Messages)
		config.TitleLabel = msgr.SystemTitleLabel
		return config, nil
	})(loginBuilder.ViewHelper(), presetsBuilder))

	return loginBuilder
}

func (a *Handler) createLoginSessionBuilder(loginBuilder *login.Builder, activityBuilder *activity.Builder) *plogin.SessionBuilder {
	loginSessionBuilder := plogin.NewSessionBuilder(loginBuilder, a.DB).
		Activity(activityBuilder.RegisterModel(&User{})).
		IsPublicUser(func(u any) bool {
			user, ok := u.(*User)
			return ok && user.GetAccountName() == a.Auth.InitialUserEmail
		})
	a.Use(loginSessionBuilder)
	return loginSessionBuilder
}

// createProfileBuilder creates and configures the profile builder
func (a *Handler) createProfileBuilder(activityBuilder *activity.Builder, loginSessionBuilder *plogin.SessionBuilder) *plogin.ProfileBuilder {
	profileBuilder := plogin.NewProfileBuilder(
		func(ctx context.Context) (*plogin.Profile, error) {
			evCtx := web.MustGetEventContext(ctx)
			u := GetCurrentUser(evCtx.R)
			if u == nil {
				return nil, perm.PermissionDenied //nolint:errhandle
			}
			notifiCounts, err := activityBuilder.GetNotesCounts(ctx, "", nil)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get activity notes counts")
			}
			user := &plogin.Profile{
				ID:     fmt.Sprint(u.ID),
				Name:   u.Name,
				Roles:  u.GetRoles(),
				Status: strcase.ToCamel(u.Status),
				Fields: []*plogin.ProfileField{
					{Name: "Email", Value: u.Account},
					{Name: "Company", Value: u.Company},
				},
				NotifCounts: notifiCounts,
			}
			if u.OAuthAvatar != "" {
				user.Avatar = u.OAuthAvatar
			}
			return user, nil
		},
		func(ctx context.Context, newName string) error {
			evCtx := web.MustGetEventContext(ctx)
			u := GetCurrentUser(evCtx.R)
			if u == nil {
				return perm.PermissionDenied //nolint:errhandle
			}
			u.Name = newName
			if err := a.DB.Save(u).Error; err != nil {
				return errors.Wrap(err, "failed to update user name")
			}
			return nil
		},
	).SessionBuilder(loginSessionBuilder)

	a.Use(profileBuilder)
	return profileBuilder
}

// buildOAuthProviders creates OAuth provider configurations
func buildOAuthProviders(authConfig *AuthConfig) []*login.Provider {
	var providers []*login.Provider

	// Google OAuth provider
	if authConfig.GoogleClientKey != "" && authConfig.GoogleClientSecret != "" {
		providers = append(providers, &login.Provider{
			Goth: google.New(
				authConfig.GoogleClientKey,
				authConfig.GoogleClientSecret,
				authConfig.BaseURL+"/auth/callback?provider="+OAuthProviderGoogle,
			),
			Key:  OAuthProviderGoogle,
			Text: "LoginProviderGoogleText",
			Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="24px" height="24px"><path fill="#fbc02d" d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12	s5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24s8.955,20,20,20	s20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"></path><path fill="#e53935" d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039	l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"></path><path fill="#4caf50" d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36	c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"></path><path fill="#1565c0" d="M43.611,20.083L43.595,20L42,20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571	c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"></path></svg>`),
		})
	}

	// Microsoft OAuth provider
	if authConfig.MicrosoftClientKey != "" && authConfig.MicrosoftClientSecret != "" {
		providers = append(providers, &login.Provider{
			Goth: microsoftonline.New(
				authConfig.MicrosoftClientKey,
				authConfig.MicrosoftClientSecret,
				// TODO: @molon 为什么这里偏偏不用给到 OAuthProviderMicrosoftOnline provider 参数呢？
				authConfig.BaseURL+"/auth/callback",
			),
			Key:  OAuthProviderMicrosoftOnline,
			Text: "LoginProviderMicrosoftText",
			Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="24px" height="24px"><path fill="#f35325" d="M2 2h20v20H2z"/><path fill="#81bc06" d="M24 2h20v20H24z"/><path fill="#05a6f0" d="M2 24h20v20H2z"/><path fill="#ffba08" d="M24 24h20v20H24z"/></svg>`),
		})
	}

	// GitHub OAuth provider
	if authConfig.GitHubClientKey != "" && authConfig.GitHubClientSecret != "" {
		providers = append(providers, &login.Provider{
			Goth: github.New(
				authConfig.GitHubClientKey,
				authConfig.GitHubClientSecret,
				authConfig.BaseURL+"/auth/callback?provider="+OAuthProviderGithub,
			),
			Key:  OAuthProviderGithub,
			Text: "LoginProviderGithubText",
			Logo: RawHTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 96" width="24px" height="24px"><path fill-rule="evenodd" clip-rule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a46.97 46.97 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0z" fill="#24292f"/></svg>`),
		})
	}

	return providers
}

func createDefaultRolesIfEmpty(ctx context.Context, db *gorm.DB) error {
	db = db.WithContext(ctx)

	var count int64
	if err := db.Model(&role.Role{}).Count(&count).Error; err != nil {
		return errors.Wrap(err, "failed to count roles")
	}

	if count > 0 {
		return nil
	}

	var roles []*role.Role
	for _, roleName := range DefaultRoles {
		roles = append(roles, &role.Role{
			Name: roleName,
		})
	}

	if err := db.Create(roles).Error; err != nil {
		return errors.Wrap(err, "failed to create default roles")
	}

	return nil
}

func createInitialUserIfEmpty(ctx context.Context, db *gorm.DB, opts *UpsertUserOptions) (*User, error) {
	db = db.WithContext(ctx)

	var count int64
	if err := db.Model(&User{}).Where("account = ?", opts.Email).Count(&count).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count users")
	}
	if count > 0 {
		return nil, nil
	}

	return UpsertUser(ctx, db, opts)
}

func grantUserRole(ctx context.Context, db *gorm.DB, userID uint, roleName string) error {
	db = db.WithContext(ctx)

	var roleID int
	if err := db.Model(&role.Role{}).Where("name = ?", roleName).Pluck("id", &roleID).Error; err != nil {
		return errors.Wrapf(err, "failed to get role id for role %s", roleName)
	}

	if err := db.Table("user_role_join").
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(map[string]any{"user_id": userID, "role_id": roleID}).Error; err != nil {
		return errors.Wrapf(err, "failed to grant role %s to user %d", roleName, userID)
	}

	return nil
}

type UpsertUserOptions struct {
	Email    string
	Password string
	Role     []string
}

func UpsertUser(ctx context.Context, db *gorm.DB, opts *UpsertUserOptions) (*User, error) {
	user := &User{
		Name:   opts.Email,
		Status: StatusActive,
		UserPass: login.UserPass{
			Account:  opts.Email,
			Password: opts.Password,
		},
	}
	user.EncryptPassword()

	db = db.WithContext(ctx)

	err := db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "account"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("account <> ''"),
				gorm.Expr("deleted_at IS NULL"),
			}},
			UpdateAll: true,
		},
		clause.Returning{Columns: []clause.Column{{Name: "id"}}},
	).Create(user).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert user")
	}

	for _, roleName := range opts.Role {
		if err := grantUserRole(ctx, db, user.ID, roleName); err != nil {
			return nil, errors.Wrapf(err, "failed to grant role %s to user", roleName)
		}
	}

	if err := db.Preload("Roles").Where("id = ?", user.ID).First(user).Error; err != nil {
		return nil, errors.Wrap(err, "failed to reload user with roles")
	}

	return user, nil
}

func GetCurrentUser(r *http.Request) *User {
	u, ok := login.GetCurrentUser(r).(*User)
	if !ok {
		return nil
	}
	return u
}
