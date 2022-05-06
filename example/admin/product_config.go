package admin

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/example/models"
	"github.com/qor/qor5/media"
	"github.com/qor/qor5/media/media_library"
	media_view "github.com/qor/qor5/media/views"

	"github.com/qor/qor5/worker"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

func configProduct(b *presets.Builder, db *gorm.DB, wb *worker.Builder) {
	p := b.Model(&models.Product{})
	eb := p.Editing("Code", "Name", "Price", "Image")
	listing := p.Listing("Code", "Name", "Price", "Image").SearchColumns("Code", "Name").SelectableColumns(true)
	listing.ActionsAsMenu(true)

	noParametersJob := wb.JobAction(
		"No parameters",
		p.Info().URIName(),
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.AddLog(fmt.Sprintf("%v", i))
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show that an no parameter job can be executed")

	parametersBoxJob := wb.JobAction(
		"Parameter input box",
		p.Info().URIName(),
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.AddLog(fmt.Sprintf("%v", i))
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show that an input box when there are parameters").Params(&struct{ Name string }{})

	displayLogJob := wb.JobAction(
		"Display log",
		p.Info().URIName(),
		func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 10; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.AddLog(fmt.Sprintf("%v", i))
					job.SetProgress(uint(i * 10))
					time.Sleep(time.Second)
				}
			}
			job.SetProgressText(`<a href="https://qor5-test.s3.ap-northeast-1.amazonaws.com/system/media_libraries/37/file.@qor_preview.png">Please download this file</a>`)
			return nil
		},
	).Description("This test demo is used to show the log section of this job").Params(&struct{ Name string }{}).DisplayLog(true)

	getArgsJob := wb.JobAction(
		"Get Args",
		p.Info().URIName(),
		func(ctx context.Context, job worker.QorJobInterface) error {
			args, err := job.GetArgument()
			if err != nil {
				return err
			}

			jobActionArgs := args.(*worker.JobActionArgs)
			actionParams := jobActionArgs.ActionParams.(*struct{ Name string })
			job.AddLog(fmt.Sprintf("Action Params Name is  %#+v", actionParams.Name))

			job.AddLog(fmt.Sprintf("Origina Context AuthInfo is  %#+v", jobActionArgs.OriginalPageContext["AuthInfo"]))
			job.AddLog(fmt.Sprintf("Origina Context URL is  %#+v", jobActionArgs.OriginalPageContext["URL"]))

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
	).Description("This test demo is used to show how to get the action's arguments and original page context").Params(&struct{ Name string }{}).DisplayLog(true).
		ContextHandler(func(ctx *web.EventContext) map[string]interface{} {
			auth, err := ctx.R.Cookie("auth")
			if err == nil {
				return map[string]interface{}{"AuthInfo": auth.Value}
			}
			return nil
		})

	listing.BulkAction("Job Action - No parameters").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Job Action - No parameters").Color("primary").Depressed(true).Class("ml-2").Attr("@click", noParametersJob.URL())
			})

	listing.BulkAction("Job Action - Parameter input box").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Job Action - Parameter input box").Color("primary").Depressed(true).Class("ml-2").
					Attr("@click", parametersBoxJob.URL())
			})
	listing.BulkAction("Job Action - Display log").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Job Action - Display log").Color("primary").Depressed(true).Class("ml-2").
					Attr("@click", displayLogJob.URL())
			})

	listing.BulkAction("Job Action - Get Args").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Job Action - Get Args").Color("primary").Depressed(true).Class("ml-2").
					Attr("@click", getArgsJob.URL())
			})

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
