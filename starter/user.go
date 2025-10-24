package starter

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"time"

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

	plogin "github.com/qor5/admin/v3/login"
	v "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
)

type UserModelBuilder *presets.ModelBuilder

// createUserModelBuilder creates and configures the user model builder
func (a *Handler) createUserModelBuilder(presetsBuilder *presets.Builder, activityBuilder *activity.Builder, loginSessionBuilder *plogin.SessionBuilder) UserModelBuilder {
	umb := presetsBuilder.Model(&User{})
	defer func() { activityBuilder.RegisterModel(umb) }()

	// ========== Search Configuration ==========
	umb.Listing().SearchFunc(func(ctx *web.EventContext, params *presets.SearchParams) (*presets.SearchResult, error) {
		u := GetCurrentUser(ctx.R)
		qdb := a.DB.WithContext(ctx.R.Context())

		// If the current user doesn't have 'admin' role, do not allow them to view admin and manager users
		// We didn't do this on permission because we are not supporting the permission on listing page
		if currentRoles := u.GetRoles(); !slices.Contains(currentRoles, RoleAdmin) {
			qdb = qdb.Joins("inner join user_role_join urj on users.id = urj.user_id inner join roles r on r.id = urj.role_id").
				Group("users.id").
				Having("COUNT(CASE WHEN r.name in (?) THEN 1 END) = 0", []string{RoleAdmin, RoleManager})
		}

		result, err := gorm2op.DataOperator(qdb).Search(ctx, params)
		if err != nil {
			return nil, errors.Wrap(err, "failed to search users")
		}
		return result, nil
	})

	// ========== Editing Configuration ==========
	editing := umb.Editing(
		"Type",
		"Actions",
		"Name",
		"OAuthProvider",
		"OAuthIdentifier",
		"Account",
		"Company",
		"Roles",
		"Status",
	)

	// Configure side panel for activity timeline
	editing.SidePanelFunc(func(obj any, ctx *web.EventContext) h.HTMLComponent {
		if ctx.R.FormValue(presets.ParamID) == "" {
			return nil
		}
		return activityBuilder.MustGetModelBuilder(umb).NewTimelineCompo(ctx, obj, "_side")
	})

	// Configure validation
	editing.ValidateFunc(func(obj any, _ *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*User)
		if u.OAuthProvider == "" && u.Account == "" {
			err.FieldError("Account", "Email is required")
		}
		return
	})

	// ========== Event Handlers ==========
	// Event: Unlock user
	umb.RegisterEventFunc("eventUnlockUser", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		var u User
		if err := a.DB.WithContext(ctx.R.Context()).Where("id = ?", uid).First(&u).Error; err != nil {
			return r, errors.Wrap(err, "failed to find user")
		}
		if err = u.UnlockUser(a.DB, &User{}); err != nil {
			return r, err
		}
		presets.ShowMessage(&r, "success", "")
		editing.UpdateOverlayContent(ctx, &r, &u, "", nil)
		return r, nil
	})

	// Event: Send reset password email
	umb.RegisterEventFunc("eventSendResetPasswordEmail", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		var u User
		if err := a.DB.WithContext(ctx.R.Context()).Where("id = ?", uid).First(&u).Error; err != nil {
			return r, errors.Wrap(err, "failed to find user")
		}
		token, err := u.GenerateResetPasswordToken(a.DB, &User{})
		if err != nil {
			return r, err
		}
		r.RunScript = fmt.Sprintf(`alert("http://localhost:9500/auth/reset-password?id=%s&token=%s")`, uid, token)
		return r, nil
	})

	// Event: Revoke TOTP
	umb.RegisterEventFunc("eventRevokeTOTP", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		var u *User
		if err := a.DB.WithContext(ctx.R.Context()).Where("id = ?", uid).First(&u).Error; err != nil {
			return r, errors.Wrap(err, "failed to find user")
		}
		err = login.RevokeTOTP(u, a.DB, &User{}, fmt.Sprint(u.ID))
		if err != nil {
			return r, errors.WithStack(err)
		}
		err = loginSessionBuilder.ExpireAllSessions(fmt.Sprint(u.ID))
		if err != nil {
			return r, errors.Wrap(err, "failed to expire all sessions")
		}
		presets.ShowMessage(&r, "success", "")
		editing.UpdateOverlayContent(ctx, &r, u, "", nil)
		return r, nil
	})

	// ========== Field Configurations ==========
	editing.Field("Type").ComponentFunc(func(obj any, _ *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		u := obj.(*User)
		if u.ID == 0 {
			return nil
		}

		var accountType string
		if u.IsOAuthUser() {
			accountType = "OAuth Account"
		} else {
			accountType = "Main Account"
		}

		return h.Div(
			v.VRow(
				v.VCol(
					h.Text(accountType),
				).Class("text-left deep-orange--text"),
			),
		).Class("mb-2")
	})

	editing.Field("Actions").ComponentFunc(func(obj any, _ *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		var actionBtns h.HTMLComponents
		u := obj.(*User)

		if !u.IsOAuthUser() && u.Account != "" {
			actionBtns = append(actionBtns,
				v.VBtn("Send Reset Password Email").
					Color("primary").
					Attr("@click", web.Plaid().EventFunc("eventSendResetPasswordEmail").
						Query("id", u.ID).Go()),
			)
		}

		if u.GetLocked() {
			actionBtns = append(actionBtns,
				v.VBtn("Unlock").Color("primary").
					Attr("@click", web.Plaid().EventFunc("eventUnlockUser").
						Query("id", u.ID).Go(),
					),
			)
		}

		if u.GetIsTOTPSetup() {
			actionBtns = append(actionBtns,
				v.VBtn("Revoke TOTP").
					Color("primary").
					Attr("@click", web.Plaid().EventFunc("eventRevokeTOTP").
						Query("id", u.ID).Go()),
			)
		}

		if len(actionBtns) == 0 {
			return nil
		}
		return h.Div(
			actionBtns...,
		).Class("mb-5 text-right")
	})

	editing.Field("Account").Label("Email").ComponentFunc(func(obj any, field *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		return vx.VXField().Attr(web.VField(field.Name, field.Value(obj))...).Label(field.Label).ErrorMessages(field.Errors...)
	}).SetterFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		u := obj.(*User)
		email := ctx.R.FormValue(field.Name)
		if email == "" {
			return
		}
		u.Account = email
		u.OAuthIdentifier = email
		return nil
	})

	editing.Field("OAuthProvider").Label("OAuth Provider").ComponentFunc(func(obj any, field *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		u := obj.(*User)
		if !u.IsOAuthUser() && u.ID != 0 {
			return nil
		} else {
			return v.VSelect().Attr(web.VField(field.Name, field.Value(obj))...).
				Label(field.Label).
				Items(OAuthProviders)
		}
	})

	editing.Field("OAuthIdentifier").Label("OAuth Identifier").ComponentFunc(func(obj any, field *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
		u := obj.(*User)
		if !u.IsOAuthUser() {
			return nil
		} else {
			return v.VTextField().Attr(web.VField(field.Name, field.Value(obj))...).Label(field.Label).ErrorMessages(field.Errors...).Disabled(true)
		}
	})

	editing.Field("Roles").
		ComponentFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var values []string
			u, ok := obj.(*User)
			if ok && u.ID != 0 {
				var userWithRoles User
				if err := a.DB.WithContext(ctx.R.Context()).Preload("Roles").Where("id = ?", u.ID).First(&userWithRoles).Error; err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						panic(err)
					}
				}
				for _, r := range userWithRoles.Roles {
					values = append(values, fmt.Sprint(r.ID))
				}
			}

			var roles []role.Role
			a.DB.Find(&roles)
			var allRoleItems []v.DefaultOptionItem
			for _, r := range roles {
				allRoleItems = append(allRoleItems, v.DefaultOptionItem{
					Text:  r.Name,
					Value: fmt.Sprint(r.ID),
				})
			}

			return vx.VXSelect().Label(field.Label).Chips(true).
				Items(allRoleItems).ItemTitle("text").ItemValue("value").
				Multiple(true).Attr(presets.VFieldError(field.Name, values, field.Errors)...).
				Disabled(field.Disabled)
		}).
		SetterFunc(func(obj any, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			u, ok := obj.(*User)
			if !ok {
				return
			}
			if u.GetAccountName() == a.Auth.InitialUserEmail {
				return perm.PermissionDenied //nolint:errhandle
			}
			rids := ctx.R.Form[field.Name]
			var roles []role.Role
			for _, id := range rids {
				uid, err1 := strconv.Atoi(id)
				if err1 != nil {
					continue
				}
				roles = append(roles, role.Role{
					Model: gorm.Model{ID: uint(uid)},
				})
			}
			u.Roles = roles
			return
		})

	// Field: Status
	editing.Field("Status").
		ComponentFunc(func(obj any, field *presets.FieldContext, _ *web.EventContext) h.HTMLComponent {
			return vx.VXSelect().Attr(presets.VFieldError(field.Name, field.Value(obj), field.Errors)...).
				Label(field.Label).
				Items([]string{"active", "inactive"})
		})

	// ========== Save Logic Configuration ==========
	editing.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj any, id string, ctx *web.EventContext) error {
			u := obj.(*User)
			if u.GetAccountName() == a.Auth.InitialUserEmail {
				return perm.PermissionDenied //nolint:errhandle
			}
			if u.RegistrationDate.IsZero() {
				u.RegistrationDate = time.Now()
			}
			if err := a.DB.Transaction(func(tx *gorm.DB) error {
				ctx.WithContextValue(gorm2op.CtxKeyDB{}, tx)
				defer ctx.WithContextValue(gorm2op.CtxKeyDB{}, nil)
				// First save the user to ensure we have a valid ID (covers both create and update)
				if err := in(obj, id, ctx); err != nil {
					return err
				}
				// Explicitly replace user roles in the join table
				var roleIDs []uint
				for _, r := range u.Roles {
					if r.ID != 0 {
						roleIDs = append(roleIDs, r.ID)
					}
				}
				if err := replaceUserRoles(tx, u.ID, roleIDs); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return errors.Wrap(err, "failed to save user")
			}
			return nil
		}
	})

	// ========== Listing Configuration ==========
	listing := umb.Listing("ID", "Name", "Account", "Status", activity.ListFieldNotes).PerPage(10)
	listing.Field("Account").Label("Email")
	listing.SearchColumns("users.Name", "Account")

	// Configure filter data
	listing.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		item, err := activityBuilder.MustGetModelBuilder(umb).NewHasUnreadNotesFilterItem(ctx.R.Context(), "users.")
		if err != nil {
			panic(err)
		}
		return []*vx.FilterItem{
			item,
			{
				Key:          "created",
				Label:        "Create Time",
				ItemType:     vx.ItemTypeDatetimeRange,
				SQLCondition: `users.created_at %s ?`,
			},
			{
				Key:          "name",
				Label:        "Name",
				ItemType:     vx.ItemTypeString,
				SQLCondition: `users.name %s ?`,
			},
			{
				Key:          "status",
				Label:        "Status",
				ItemType:     vx.ItemTypeSelect,
				SQLCondition: `users.status %s ?`,
				Options: []*vx.SelectItem{
					{Text: "Active", Value: "active"},
					{Text: "Inactive", Value: "inactive"},
				},
			},
			{
				Key:          "registration_date",
				Label:        "Registration Date",
				ItemType:     vx.ItemTypeDate,
				SQLCondition: `users.registration_date %s ?`,
				Folded:       true,
			},
			{
				Key:          "registration_date_range",
				Label:        "Registration Date Range",
				ItemType:     vx.ItemTypeDateRange,
				SQLCondition: `users.registration_date %s ?`,
				Folded:       true,
			},
		}
	})

	// Configure filter tabs
	listing.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nDemoKey, Messages_en_US).(*Messages)

		tab, err := activityBuilder.MustGetModelBuilder(umb).NewHasUnreadNotesFilterTab(ctx.R.Context())
		if err != nil {
			panic(err)
		}
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabsAll,
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: msgr.FilterTabsActive,
				Query: url.Values{"status": []string{"active"}},
			},
			tab,
		}
	})

	return umb
}

// replaceUserRoles replaces all roles of a user by role IDs via the join table explicitly.
// This does not rely on GORM association auto-save behavior.
func replaceUserRoles(db *gorm.DB, userID uint, roleIDs []uint) error {
	// Clear existing relations
	if err := db.Table("user_role_join").Where("user_id = ?", userID).Delete(nil).Error; err != nil {
		return errors.Wrap(err, "failed to delete user role")
	}
	if len(roleIDs) == 0 {
		return nil
	}
	// Bulk insert new relations
	rows := make([]map[string]any, 0, len(roleIDs))
	for _, rid := range roleIDs {
		rows = append(rows, map[string]any{"user_id": userID, "role_id": rid})
	}
	return errors.WithStack(db.Table("user_role_join").Create(&rows).Error)
}
