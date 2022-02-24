package admin

import (
	"fmt"
	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"
	"gorm.io/gorm"
	"strconv"
)

func configProduct(b *presets.Builder, db *gorm.DB) {
	p := b.Model(&models.Product{})

	eb := p.Editing("Code", "Name", "Image")
	p.Listing("Code", "Name", "Image")

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.Product)
		if u.Code == "" {
			err.FieldError("Name", "Code is required")
		}
		if u.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		return
	})

	eb.Field("Image").
		WithContextValue(
			media_view.MediaBoxConfig,
			&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*media.Size{
					"thumb": {
						Width:  100,
						Height: 100,
					},
				},
			})

}

type productItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

func productsSelector(db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var ps []models.Product
		var items []productItem
		searchKey := ctx.R.FormValue("keyword")
		sql := db.Order("created_at desc").Limit(10)
		if searchKey != "" {
			key := fmt.Sprintf("%%%s%%", searchKey)
			sql = sql.Where("name ILIKE ? or code ILIKE ?", key, key)
		}
		sql.Find(&ps)
		for _, p := range ps {
			items = append(items, productItem{
				ID:    strconv.Itoa(int(p.ID)),
				Name:  p.Name,
				Image: p.Image.URL("thumb"),
			})
		}
		r.Data = items
		return
	}
}
