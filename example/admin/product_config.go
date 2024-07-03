package admin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/qor5/admin/v3/media"
	"github.com/qor5/admin/v3/media/base"
	"github.com/qor5/admin/v3/publish"

	"github.com/qor5/admin/v3/example/models"
	"github.com/qor5/admin/v3/media/media_library"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"

	"github.com/qor5/admin/v3/worker"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configProduct(b *presets.Builder, _ *gorm.DB, wb *worker.Builder, publisher *publish.Builder) *presets.ModelBuilder {
	p := b.Model(&models.Product{}).Use(publisher)
	eb := p.Editing("StatusBar", "ScheduleBar", "Code", "Name", "Price", "Image")
	listing := p.Listing("Code", "Name", "Price", "Image").SearchColumns("Code", "Name").SelectableColumns(true)
	listing.ActionsAsMenu(true)

	noParametersJob := wb.ActionJob(
		"No parameters",
		p,
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show that an no parameter job can be executed")

	parametersBoxJob := wb.ActionJob(
		"Parameter input box",
		p,
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show that an input box when there are parameters").
		Params(&struct{ Name string }{})

	displayLogJob := wb.ActionJob(
		"Display log",
		p,
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.SetProgress(uint(i * 10))
					job.AddLog(fmt.Sprintf("%v", i))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show the log section of this job").
		Params(&struct{ Name string }{}).
		DisplayLog(true).
		ProgressingInterval(4000)

	getArgsJob := wb.ActionJob(
		"Get Args",
		p,
		func(ctx context.Context, job worker.QorJobInterface) error {
			jobInfo, err := job.GetJobInfo()
			if err != nil {
				return err
			}

			job.AddLog(fmt.Sprintf("Action Params Name is  %#+v", jobInfo.Argument.(*struct{ Name string }).Name))
			job.AddLog(fmt.Sprintf("Origina Context AuthInfo is  %#+v", jobInfo.Context["AuthInfo"]))
			job.AddLog(fmt.Sprintf("Origina Context URL is  %#+v", jobInfo.Context["URL"]))

			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					return nil
				default:
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show how to get the action's arguments and original page context").
		Params(&struct{ Name string }{}).
		DisplayLog(true).
		ContextHandler(func(ctx *web.EventContext) map[string]interface{} {
			auth, err := ctx.R.Cookie("auth")
			if err == nil {
				return map[string]interface{}{"AuthInfo": auth.Value}
			}
			return nil
		})

	listing.BulkAction("Action Job - No parameters").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - No parameters").Color("secondary").Variant(vuetify.VariantFlat).Class("ml-2").
					Attr("@click", noParametersJob.URL())
			})

	listing.BulkAction("Action Job - Parameter input box").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Parameter input box").Color("secondary").Variant(vuetify.VariantFlat).Class("ml-2").
					Attr("@click", parametersBoxJob.URL())
			})
	listing.BulkAction("Action Job - Display log").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Display log").Color("secondary").Variant(vuetify.VariantFlat).Class("ml-2").
					Attr("@click", displayLogJob.URL())
			})

	listing.BulkAction("Action Job - Get Args").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Get Args").Color("secondary").Variant(vuetify.VariantFlat).Class("ml-2").
					Attr("@click", getArgsJob.URL())
			})

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		u := obj.(*models.Product)
		if u.Code == "" {
			err.FieldError("Code", "Code is required")
		}
		if u.Name == "" {
			err.FieldError("Name", "Name is required")
		}
		return
	})

	eb.Field("Image").
		WithContextValue(
			media.MediaBoxConfig,
			&media_library.MediaBoxConfig{
				AllowType: "image",
				Sizes: map[string]*base.Size{
					"thumb": {
						Width:  100,
						Height: 100,
					},
				},
			})

	return p
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
