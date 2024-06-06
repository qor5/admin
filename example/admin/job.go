package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/worker"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
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
			jobInfo, _ := job.GetJobInfo()
			job.AddLog(fmt.Sprintf("Argument %#+v", jobInfo.Argument))
			job.AddLog(fmt.Sprintf("Context %#+v", jobInfo.Context))
			return nil
		})
	ajb.GetResourceBuilder().Editing().Field("F1").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}
		return VTextField().Attr(web.VField(field.Name, field.Value(obj))...).Label(field.Label).ErrorMessages(vErr.GetFieldErrors(field.Name)...)
	}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
		v := ctx.R.FormValue("F1")
		obj.(*ArgJobResource).F1 = v

		if v == "aaa" {
			return errors.New("cannot be aaa")
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
