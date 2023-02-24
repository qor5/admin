package admin

import (
	"fmt"

	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/admin/example/models"
	"github.com/qor5/admin/note"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/presets/gorm2op"
	"github.com/qor5/admin/publish"
	publish_view "github.com/qor5/admin/publish/views"
	"github.com/qor5/admin/role"
	"github.com/qor5/admin/utils"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/qor5/x/login"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configUser(b *presets.Builder, db *gorm.DB) {
	user := b.Model(&models.User{})
	// MenuIcon("people")
	note.Configure(db, b, user)

	user.Listing().Searcher = func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
		u := getCurrentUser(ctx.R)
		qdb := db

		// If the current user doesn't has 'admin' role, do not allow them to view admin and manager users
		// We didn't do this on permission because of we are not supporting the permission on listing page
		if currentRoles := u.GetRoles(); !utils.Contains(currentRoles, models.RoleAdmin) {
			qdb = db.Joins("INNER JOIN user_role_join urj on users.id = urj.user_id inner join roles r on r.id = urj.role_id").Where("r.name NOT IN (?)", []string{models.RoleAdmin, models.RoleManager})
		}

		return gorm2op.DataOperator(qdb).Search(model, params, ctx)
	}

	ed := user.Editing(
		"Type",
		"Actions",
		"Name",
		"OAuthProvider",
		"Account",
		"Company",
		"Roles",
		"Status",
		"FavorPostID",
	)

	ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.User)
		if u.Account == "" {
			err.FieldError("Account", "Email is required")
		}
		return
	})
	user.RegisterEventFunc("roles_selector", rolesSelector(db))
	user.RegisterEventFunc("eventUnlockUser", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		u := models.User{}
		if err = db.Where("id = ?", uid).First(&u).Error; err != nil {
			return r, err
		}
		if err = u.UnlockUser(db, &models.User{}); err != nil {
			return r, err
		}
		presets.ShowMessage(&r, "success", "")
		ed.UpdateOverlayContent(ctx, &r, &u, "", nil)
		return r, nil
	})

	user.RegisterEventFunc("eventSendResetPasswordEmail", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		u := models.User{}
		if err = db.Where("id = ?", uid).First(&u).Error; err != nil {
			return r, err
		}
		token, err := u.GenerateResetPasswordToken(db, &models.User{})
		if err != nil {
			return r, err
		}
		r.VarsScript = fmt.Sprintf(`alert("http://localhost:9500/auth/reset-password?id=%s&token=%s")`, uid, token)
		return r, nil
	})

	user.RegisterEventFunc("eventRevokeTOTP", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		uid := ctx.R.FormValue("id")
		u := &models.User{}
		if err = db.Where("id = ?", uid).First(u).Error; err != nil {
			return r, err
		}
		err = login.RevokeTOTP(u, db, &models.User{}, fmt.Sprint(u.ID))
		if err != nil {
			return r, err
		}
		err = expireAllSessionLogs(u.ID)
		if err != nil {
			return r, err
		}
		presets.ShowMessage(&r, "success", "")
		ed.UpdateOverlayContent(ctx, &r, u, "", nil)
		return r, nil
	})

	ed.Field("Type").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*models.User)
		if u.ID == 0 {
			return nil
		}

		var accountType string
		if u.OAuthProvider == "" && u.Account != "" {
			accountType = "Main Account"
		} else {
			accountType = "OAuth Account"
		}

		return h.Div(
			VRow(
				VCol(
					h.Text(accountType),
				).Class("text-left deep-orange--text"),
			),
		).Class("mb-2")
	})

	ed.Field("Actions").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var actionBtns h.HTMLComponents
		u := obj.(*models.User)

		if u.OAuthProvider == "" && u.Account != "" {
			actionBtns = append(actionBtns,
				VBtn("Send Reset Password Email").
					Color("primary").
					Attr("@click", web.Plaid().EventFunc("eventSendResetPasswordEmail").
						Query("id", u.ID).Go()),
			)
		}

		if u.GetLocked() {
			actionBtns = append(actionBtns,
				VBtn("Unlock").Color("primary").
					Attr("@click", web.Plaid().EventFunc("eventUnlockUser").
						Query("id", u.ID).Go(),
					),
			)
		}

		if u.GetIsTOTPSetup() {
			actionBtns = append(actionBtns,
				VBtn("Revoke TOTP").
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

	ed.Field("Account").Label("Email").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		return presets.InputWithDefaults(VTextField(), obj, field).ErrorMessages(field.Errors...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		u := obj.(*models.User)
		email := ctx.R.FormValue(field.Name)
		u.Account = email
		u.OAuthIndentifier = email
		return nil
	})

	ed.Field("OAuthProvider").Label("OAuthProvider").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		u := obj.(*models.User)
		if p := field.Value(obj); p == "" && u.ID != 0 {
			return nil
		} else {
			return VSelect().FieldName(field.Name).
				Label(field.Label).Value(p).
				Items(models.OAuthProviders)
		}
	})

	var roles []role.Role
	db.Find(&roles)
	var allRoleItems = []DefaultOptionItem{}
	for _, r := range roles {
		allRoleItems = append(allRoleItems, DefaultOptionItem{
			Text:  r.Name,
			Value: fmt.Sprint(r.ID),
		})
	}

	ed.Field("Roles").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var selectedItems = []DefaultOptionItem{}
			var values = []string{}
			u, ok := obj.(*models.User)
			if ok {
				var roles []role.Role
				db.Model(u).Association("Roles").Find(&roles)
				for _, r := range roles {
					values = append(values, fmt.Sprint(r.ID))
					selectedItems = append(selectedItems, DefaultOptionItem{
						Text:  r.Name,
						Value: fmt.Sprint(r.ID),
					})
				}
			}

			return vx.VXAutocomplete().Label(field.Label).
				// ItemText("text").ItemValue("value").
				FieldName(field.Name).
				Multiple(true).Chips(true).Clearable(true).DeletableChips(true).
				Value(values).
				SelectedItems(selectedItems).
				Items(allRoleItems).
				CacheItems(true)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			u, ok := obj.(*models.User)
			if !ok {
				return
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

			if u.ID == 0 {
				err = reflectutils.Set(obj, field.Name, roles)
			} else {
				err = db.Model(u).Association(field.Name).Replace(roles)
			}
			if err != nil {
				return
			}
			return
		})

	ed.Field("Status").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSelect().FieldName(field.Name).
				Label(field.Label).Value(field.Value(obj)).
				Items([]string{"active", "inactive"})
		})

	configureFavorPostSelectDialog(b)
	ed.Field("FavorPostID").Label("Favorite Post").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		id := field.Value(obj).(uint)
		return web.Portal(favorPostSelector(id)).Name("favorPostSelector")
	})

	oldSaver := ed.Saver
	ed.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		u := obj.(*models.User)
		u.RegistrationDate = time.Now()
		return oldSaver(obj, id, ctx)
	})

	cl := user.Listing("ID", "Name", "Account", "Status", "Notes").PerPage(10)
	cl.Field("Account").Label("Email")
	cl.SearchColumns("Name", "Account")

	cl.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vx.FilterItem{
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
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "users", "Users", u.ID, "Users"),
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

	cl.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsActive"),
				Query: url.Values{"status": []string{"active"}},
			},
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsAll"),
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsHasUnreadNotes"),
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})
}

func rolesSelector(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var roles []role.Role
		var items []DefaultOptionItem
		searchKey := ctx.R.FormValue("keyword")
		sql := db.Order("name").Limit(3)
		if searchKey != "" {
			sql = sql.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", searchKey))
		}
		sql.Find(&roles)
		for _, r := range roles {
			items = append(items, DefaultOptionItem{
				Text:  r.Name,
				Value: fmt.Sprint(r.ID),
			})
		}
		r.Data = items
		return
	}
}

func favorPostSelector(id uint) h.HTMLComponent {
	var items []*models.Post
	if id > 0 {
		p := &models.Post{}
		if err := db.Where("id = ?", id).Order("version desc").First(p).Error; err == nil {
			items = append(items, p)
		}
	}
	return h.Div(
		VAutocomplete().
			Label("Favorite Post").
			FieldName("FavorPostID").
			Items(items).
			ItemText("Title").
			ItemValue("ID").
			Value(id).
			Readonly(true).
			Clearable(true),
	).Attr("@click", web.Plaid().EventFunc(actions.OpenListingDialog).
		URL("/dialog-select-favor-posts").
		Go())
}

func configureFavorPostSelectDialog(pb *presets.Builder) {
	b := pb.Model(&models.Post{}).
		URIName("dialog-select-favor-posts").
		InMenu(false)
	publish_view.Configure(pb, db, nil, nil, b)
	lb := b.Listing("ID", "Title", "TitleWithSlug", "HeroImage", "Body").
		SearchColumns("title", "body").
		PerPage(10).
		OrderableFields([]*presets.OrderableField{
			{
				FieldName: "ID",
				DBColumn:  "id",
			},
			{
				FieldName: "Title",
				DBColumn:  "title",
			},
		}).
		SelectableColumns(true)
	lb.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent { return nil })
	lb.RowMenu().Empty()
	registerSelectFavorPostEvent(pb)
	lb.CellWrapperFunc(func(cell h.MutableAttrHTMLComponent, id string, obj interface{}, dataTableID string) h.HTMLComponent {
		cell.SetAttr("@click.self", web.Plaid().
			Query("id", strings.Split(id, "_")[0]).
			EventFunc("selectFavorPost").
			Go(),
		)
		return cell
	})
	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		u := getCurrentUser(ctx.R)

		return []*vx.FilterItem{
			{
				Key:          "hasUnreadNotes",
				Invisible:    true,
				SQLCondition: fmt.Sprintf(hasUnreadNotesQuery, "posts", "Posts", u.ID, "Posts"),
			},
			{
				Key:          "created",
				Label:        "Create Time",
				ItemType:     vx.ItemTypeDatetimeRange,
				SQLCondition: `created_at %s ?`,
			},
			{
				Key:          "title",
				Label:        "Title",
				ItemType:     vx.ItemTypeString,
				SQLCondition: `title %s ?`,
			},
			{
				Key:      "status",
				Label:    "Status",
				ItemType: vx.ItemTypeSelect,
				Options: []*vx.SelectItem{
					{Text: publish.StatusDraft, Value: publish.StatusDraft},
					{Text: publish.StatusOnline, Value: publish.StatusOnline},
					{Text: publish.StatusOffline, Value: publish.StatusOffline},
				},
				SQLCondition: `status %s ?`,
			},
			{
				Key:      "multi_statuses",
				Label:    "Multiple Statuses",
				ItemType: vx.ItemTypeMultipleSelect,
				Options: []*vx.SelectItem{
					{Text: publish.StatusDraft, Value: publish.StatusDraft},
					{Text: publish.StatusOnline, Value: publish.StatusOnline},
					{Text: publish.StatusOffline, Value: publish.StatusOffline},
				},
				SQLCondition: `status %s ?`,
				Folded:       true,
			},
			{
				Key:          "id",
				Label:        "ID",
				ItemType:     vx.ItemTypeNumber,
				SQLCondition: `id %s ?`,
				Folded:       true,
			},
		}
	})

	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsAll"),
				ID:    "all",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: i18n.T(ctx.R, I18nExampleKey, "FilterTabsHasUnreadNotes"),
				ID:    "hasUnreadNotes",
				Query: url.Values{"hasUnreadNotes": []string{"1"}},
			},
		}
	})

	// select many
	// lb.BulkAction("Confirm").ButtonCompFunc(func(ctx *web.EventContext) h.HTMLComponent {
	// 	return VBtn("Confirm").
	// 		Color("primary").
	// 		Attr("@click", web.Plaid().
	// 			EventFunc("selectFavorPost").
	// 			Query("ids", ctx.R.URL.Query().Get(presets.ParamSelectedIds)).
	// 			MergeQuery(true).
	// 			Go())
	// })
}

func registerSelectFavorPostEvent(b *presets.Builder) {
	b.GetWebBuilder().RegisterEventFunc("selectFavorPost", func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var id uint
		if v := ctx.R.FormValue("id"); v != "" {
			iv, _ := strconv.Atoi(v)
			id = uint(iv)
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "favorPostSelector",
			Body: favorPostSelector(id),
		})
		web.AppendVarsScripts(&r, presets.CloseListingDialogVarScript)
		return
	})
}
