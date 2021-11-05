package admin

import (
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/lib/pq"
	"github.com/qor/qor5/example/models"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configRole(b *presets.Builder, db *gorm.DB) {
	role := b.Model(&models.Role{})

	eb := role.Editing(
		"Name",
		"Permissions",
	)

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.Role)
		if u.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		return
	})

	eb.Field("Permissions").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return VCombobox().Label(field.Label).Value(field.Value(obj)).
				Attr(web.VFieldName(field.Name)...).
				Multiple(true).Chips(true).
				Items([]string{"Read", "Create", "Update", "Delete"})
		})

	lb := role.Listing("Name", "Permissions")
	lb.Field("Permissions").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			v := field.Value(obj)
			if v == nil {
				return nil
			}
			arr := v.(pq.StringArray)
			return h.Td(h.Text(strings.Join(arr, ",")))
		})
}
