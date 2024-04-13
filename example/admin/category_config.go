package admin

import (
	"strconv"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/presets"
	v "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configCategory(b *presets.Builder, db *gorm.DB) *presets.ModelBuilder {
	p := b.Model(&models.Category{})

	eb := p.Editing("StatusBar", "ScheduleBar", "Name", "Products")
	p.Listing("Name")

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.Category)
		if u.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		return
	})

	p.RegisterEventFunc("products_selector", productsSelector(db))

	eb.Field("Products").
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var selectedItems = []productItem{}
			c, ok := obj.(*models.Category)
			if ok {
				var ps []models.Product
				db.Where("id in (?)", []string(c.Products)).Find(&ps)
				for _, k := range []string(c.Products) {
					for _, p := range ps {
						id := strconv.Itoa(int(p.ID))
						if k == id {
							selectedItems = append(selectedItems, productItem{
								ID:    id,
								Name:  p.Name,
								Image: p.Image.URL("thumb"),
							})
							break
						}
					}
				}
			}

			return v.VXSelectMany().Label(field.Label).AddItemLabel("add").
				ItemText("name").
				// TODO (fix it ) FieldName(field.Name).
				SelectedItems(selectedItems).
				SearchItemsFunc("products_selector")
		})

	return p
}
