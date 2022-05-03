package admin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"
	"github.com/qor/qor5/worker"
	"gorm.io/gorm"
)

func configProduct(b *presets.Builder, db *gorm.DB, wb *worker.Builder) {
	p := b.Model(&models.Product{})

	eb := p.Editing("Code", "Name", "Price", "Image")
	listing := p.Listing("Code", "Name", "Price", "Image").SearchColumns("Code", "Name").SelectableColumns(true)
	listing.ActionsAsMenu(true)
	listing.BulkAction("Job Action - Without Params").
		ButtonCompFunc(wb.JobAction(&worker.JobActionConfig{
			Name: "Job Action Without Params",
			Hander: func(ctx context.Context, job worker.QorJobInterface) error {
				for i := 1; i <= 20; i++ {
					select {
					case <-ctx.Done():
						job.AddLog("job aborted")
						return nil
					default:
						job.AddLog(fmt.Sprintf("%v", i))
						job.SetProgress(uint(i * 5))
						time.Sleep(time.Second)
					}
				}
				return nil
			},
		}))

	listing.BulkAction("Job Action - Having Params").
		ButtonCompFunc(wb.JobAction(&worker.JobActionConfig{
			Name:   "Job Action Having Params",
			Params: &struct{ Name string }{},
			Hander: func(ctx context.Context, job worker.QorJobInterface) error {
				params, _ := job.GetArgument()
				job.AddLog(fmt.Sprintf("Params is  %#+v", params))

				for i := 1; i <= 20; i++ {
					select {
					case <-ctx.Done():
						job.AddLog("job aborted")
						return nil
					default:
						job.AddLog(fmt.Sprintf("%v", i))
						job.SetProgress(uint(i * 5))
						time.Sleep(time.Second)
					}
				}
				return nil
			},
		}))

	listing.BulkAction("Job Action - Hide Log").
		ButtonCompFunc(wb.JobAction(&worker.JobActionConfig{
			Name:    "Job Action Hide Log",
			Params:  &struct{ Name string }{},
			HideLog: true,
			Hander: func(ctx context.Context, job worker.QorJobInterface) error {
				params, _ := job.GetArgument()
				job.AddLog(fmt.Sprintf("Params is  %#+v", params))

				for i := 1; i <= 20; i++ {
					select {
					case <-ctx.Done():
						job.AddLog("job aborted")
						return nil
					default:
						job.AddLog(fmt.Sprintf("%v", i))
						job.SetProgress(uint(i * 5))
						time.Sleep(time.Second)
					}
				}
				return nil
			},
		}))

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
