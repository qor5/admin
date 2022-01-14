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
	h "github.com/theplant/htmlgo"
)

func addJobs(w *worker.Builder) {
	w.NewJob("noArgJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			job.AddLog("hoho1")
			job.AddLog("hoho2")
			job.AddLog("hoho3")
			return nil
		})
	w.NewJob("progressTextJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
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
	ajb := w.NewJob("argJob").
		Resource(&ArgJobResource{}).
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			args, _ := job.GetArgument()
			job.AddLog(fmt.Sprintf("%#+v", args))
			return nil
		})
	ajb.GetResourceBuilder().Editing().Field("F1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}
		return VTextField().FieldName(field.Name).Label(field.Label).Value(field.Value(obj)).ErrorMessages(vErr.GetFieldErrors(field.Name)...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.FormValue("F1")
		obj.(*ArgJobResource).F1 = v

		if v == "aaa" {
			return errors.New("cannot be aaa")
		}
		return nil
	})
	w.NewJob("longRunningJob").
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
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
		})
	type ScheduleJobResource struct {
		F1 string
		worker.Schedule
	}
	w.NewJob("scheduleJob").
		Resource(&ScheduleJobResource{}).
		Handler(func(ctx context.Context, job worker.QorJobInterface) error {
			args, _ := job.GetArgument()
			job.AddLog(fmt.Sprintf("%#+v", args))
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
