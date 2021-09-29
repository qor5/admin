package admin

import (
	"net/url"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	v "github.com/goplaid/x/vuetifyx"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configUser(b *presets.Builder, db *gorm.DB) {
	user := b.Model(&models.User{})

	ed := user.Editing(
		"Name",
		"Company",
		"Email",
		"Permission",
		"Status",
	)

	ed.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.User)
		if u.Email == "" {
			err.FieldError("Email", "Email is required")
		}
		return
	})

	ed.Field("Status").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VSelect().FieldName(field.Name).
				Label(field.Label).Value(field.Value(obj)).
				Items([]string{"active", "inactive"})
		})

	cl := user.Listing("Name", "Email", "Status")

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
