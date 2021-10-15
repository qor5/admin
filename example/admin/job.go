package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/qor/qor5/worker"
	. "github.com/theplant/htmlgo"
)

func addJobs(w *worker.Builder) {
	w.NewJob("noArgJob").
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			job.AddLog("hoho1")
			job.AddLog("hoho2")
			job.AddLog("hoho3")
			return nil
		})
	w.NewJob("progressTextJob").
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			job.AddLog("hoho1")
			job.AddLog("hoho2")
			job.AddLog("hoho3")
			job.SetProgressText(`<a href="https://www.google.com">Download users</a>`)
			return nil
		})
	type ArgJobResource struct {
		F1 string
		F2 int
		F3 bool
	}
	w.NewJob("argJob").
		Resource(&ArgJobResource{}).
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			args, _ := job.GetArgument()
			job.AddLog(fmt.Sprintf("%#+v", args))
			return nil
		})
	w.NewJob("longRunningJob").
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			for i := 1; i <= 20; i++ {
				job.AddLog(fmt.Sprintf("%v", i))
				job.SetProgress(uint(i * 5))
				time.Sleep(time.Second)
			}
			return nil
		})
	type ScheduleJobResource struct {
		F1 string
		worker.Schedule
	}
	sjb := w.NewJob("scheduleJob").
		Resource(&ScheduleJobResource{}).
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			args, _ := job.GetArgument()
			job.AddLog(fmt.Sprintf("%#+v", args))
			return nil
		})
	sjbeb := sjb.GetResourceBuilder().Editing()
	sjbeb.Field("ScheduleTime").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		args := obj.(*ScheduleJobResource)
		return VTextField().Label("ScheduleTime").Attr(web.VFieldName("ScheduleTime")...).Value(args.ScheduleTime)
	})

	w.NewJob("errorJob").
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			job.AddLog("=====perform error job")
			return errors.New("imError")
		})

	w.NewJob("panicJob").
		Handler(func(ctx context.Context, job worker.HQorJob) error {
			job.AddLog("=====perform panic job")
			panic("letsPanic")
		})
}
