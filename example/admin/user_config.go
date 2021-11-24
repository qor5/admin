package admin

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/qor/qor5/note"
	"github.com/sunfmin/reflectutils"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	v "github.com/goplaid/x/vuetifyx"

	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type Item struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	disabled bool   `json:"disabled"`
	divider  bool   `json:"divider"`
	header   string `json:"header"`
}

func configUser(b *presets.Builder, db *gorm.DB) {
	user := b.Model(&models.User{})
	note.Configure(db, b, user)

	ed := user.Editing(
		"Name",
		"Company",
		"Email",
		"Permission",
		"Roles",
		"Status",
	)

	ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.User)
		if u.Email == "" {
			err.FieldError("Email", "Email is required")
		}
		return
	})
	user.RegisterEventFunc("roles_selector", rolesSelector(db))

	ed.Field("Roles").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var selectedItems = []Item{}
			var values = []string{}
			u, ok := obj.(*models.User)
			if ok {
				var roles []models.Role
				db.Model(u).Association("Roles").Find(&roles)
				for _, r := range roles {
					values = append(values, fmt.Sprint(r.ID))
					selectedItems = append(selectedItems, Item{
						Text:  r.Name,
						Value: fmt.Sprint(r.ID),
					})
				}
			}

			return v.VXAutocomplete().Label(field.Label).
				// ItemText("text").ItemValue("value").
				FieldName(field.Name).
				Multiple(true).Chips(true).Clearable(true).DeletableChips(true).
				Value(values).
				SelectedItems(selectedItems).
				// Items(items).
				CacheItems(true).
				ItemsEventFunc("roles_selector")
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			u, ok := obj.(*models.User)
			if !ok {
				return
			}
			rids := ctx.R.Form[field.Name]
			var roles []models.Role
			for _, id := range rids {
				uid, err1 := strconv.Atoi(id)
				if err1 != nil {
					continue
				}
				roles = append(roles, models.Role{
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

	cl := user.Listing("ID", "Name", "Email", "Status", "Notes").PerPage(10)

	cl.FilterDataFunc(func(ctx *web.EventContext) v.FilterData {
		return []*v.FilterItem{
			{
				Key:          "created",
				Label:        "Create Time",
				ItemType:     v.ItemTypeDate,
				SQLCondition: `cast(strftime('%%s', created_at) as INTEGER) %s ?`,
			},
			{
				Key:          "name",
				Label:        "Name",
				ItemType:     v.ItemTypeString,
				SQLCondition: `name %s ?`,
			},
			{
				Key:          "status",
				Label:        "Status",
				ItemType:     v.ItemTypeSelect,
				SQLCondition: `status %s ?`,
				Options: []*v.SelectItem{
					&v.SelectItem{Text: "Active", Value: "active"},
					&v.SelectItem{Text: "Inactive", Value: "inactive"},
				},
			},
		}
	})

	cl.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: "Felix",
				Query: url.Values{"name.ilike": []string{"felix"}},
			},
			{
				Label: "Active",
				Query: url.Values{"status": []string{"active"}},
			},
			{
				Label: "All",
				Query: url.Values{"all": []string{"1"}},
			},
		}
	})
}

func rolesSelector(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var roles []models.Role
		var items []Item
		searchKey := ctx.Event.Value
		sql := db.Order("name").Limit(3)
		if searchKey != "" {
			sql = sql.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", searchKey))
		}
		sql.Find(&roles)
		for _, r := range roles {
			items = append(items, Item{
				Text:  r.Name,
				Value: fmt.Sprint(r.ID),
			})
		}
		r.Data = items
		return
	}
}
