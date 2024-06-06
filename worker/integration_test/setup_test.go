package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/worker"
	integration "github.com/qor5/admin/v3/worker/integration_test"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/testenv"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	pb *presets.Builder
)

func TestMain(m *testing.M) {
	env, err := testenv.New().DBEnable(true).SetUp()
	if err != nil {
		panic(err)
	}
	defer env.TearDown()
	db = env.DB

	pb = presets.New().
		DataOperator(gorm2op.DataOperator(db))

	wb := worker.NewWithQueue(db, integration.Que)
	pb.Use(wb)
	addJobs(wb)
	wb.Listen()

	m.Run()
}

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

func cleanData() {
	err := db.Exec(`
delete from qor_jobs;
delete from qor_job_instances;
delete from qor_job_logs;
    `).Error
	if err != nil {
		panic(err)
	}
}

func mustParseEventResponse(b []byte) web.EventResponse {
	r := web.EventResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		panic(err)
	}
	return r
}

func mustCreateJob(form map[string]string) {
	rBody := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(rBody)
	for k, v := range form {
		mw.WriteField(k, v)
	}
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/workers?__execute_event__=presets_Update", rBody)
	r.Header.Add("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", mw.Boundary()))
	w := httptest.NewRecorder()
	pb.ServeHTTP(w, r)
	body := w.Body.String()
	if !strings.Contains(body, "success") {
		panic("create job failed")
	}
}

func mustGetFirstJob() *worker.QorJob {
	r := &worker.QorJob{}
	if err := db.First(r).Error; err != nil {
		panic(err)
	}
	return r
}
