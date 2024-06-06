package examples_admin

// @snippet_begin(ActionWorkerExample)
import (
	"context"
	"fmt"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/worker"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type ExampleResource struct {
	gorm.Model
	Name string
}

func MountActionWorker(b *presets.Builder) {
	DB := ExampleDB()
	mb := b.Model(&ExampleResource{})
	mb.Listing().ActionsAsMenu(true)

	wb := worker.New(DB)
	wb.Install(b)
	defer wb.Listen()

	addActionJobs(mb, wb)
}

func addActionJobs(mb *presets.ModelBuilder, wb *worker.Builder) {
	lb := mb.Listing()

	noParametersJob := wb.ActionJob(
		"No parameters",
		mb,
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
		mb,
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
		mb,
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
		mb,
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

	lb.Action("Action Job - No parameters").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - No parameters").Color("secondary").Class("ml-2").
					Attr("@click", noParametersJob.URL())
			})

	lb.Action("Action Job - Parameter input box").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Parameter input box").Color("secondary").Class("ml-2").
					Attr("@click", parametersBoxJob.URL())
			})
	lb.Action("Action Job - Display log").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Display log").Color("secondary").Class("ml-2").
					Attr("@click", displayLogJob.URL())
			})

	lb.Action("Action Job - Get Args").
		ButtonCompFunc(
			func(ctx *web.EventContext) h.HTMLComponent {
				return vuetify.VBtn("Action Job - Get Args").Color("secondary").Class("ml-2").
					Attr("@click", getArgsJob.URL())
			})
}

// @snippet_end
