package examples_admin

// @snippet_begin(WorkerExample)
import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/worker"
)

func MountWorker(b *presets.Builder) {
	DB := ExampleDB()

	wb := worker.New(DB)
	wb.Install(b)
	defer wb.Listen()

	addJobs(wb)
}

func addJobs(w *worker.Builder) {
	w.NewJob("noArgJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			job.AddLog("hoho1")
			job.AddLog("hoho2")
			job.AddLog("hoho3")
			return nil
		})

	type ArgJobResource struct {
		F1 string
		F2 int
		F3 bool
	}
	argJb := w.NewJob("argJob").
		Resource(&ArgJobResource{}).
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			jobInfo, _ := job.GetJobInfo()
			job.AddLog(fmt.Sprintf("Argument %#+v", jobInfo.Argument))
			return nil
		})
	// you can to customize the resource Editing via GetResourceBuilder()
	argJb.GetResourceBuilder().Editing()

	w.NewJob("progressTextJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			job.AddLog("hoho1")
			job.AddLog("hoho2")
			job.AddLog("hoho3")
			job.SetProgressText(`<a href="https://www.google.com">Download users</a>`)
			return nil
		})

	// check ctx.Done() to stop the handler
	w.NewJob("longRunningJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			for i := 1; i <= 5; i++ {
				select {
				case <-ctx.Done():
					job.AddLog("job aborted")
					return nil
				default:
					job.AddLog(fmt.Sprintf("%v", i))
					job.SetProgress(uint(i * 20))
					time.Sleep(time.Second)
				}
			}
			return nil
		})

	// insert worker.Schedule to resource to make a job schedulable
	type ScheduleJobResource struct {
		F1 string
		worker.Schedule
	}
	w.NewJob("scheduleJob").
		Resource(&ScheduleJobResource{}).
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			jobInfo, _ := job.GetJobInfo()
			job.AddLog(fmt.Sprintf("%#+v", jobInfo.Argument))
			return nil
		})

	w.NewJob("errorJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			job.AddLog("=====perform error job")
			return errors.New("imError")
		})

	w.NewJob("panicJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			job.AddLog("=====perform panic job")
			panic("letsPanic")
		})
}

// @snippet_end
