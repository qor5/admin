package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/presets"
	vx "github.com/goplaid/x/vuetifyx"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

//go:generate moq -pkg mock -out mock/mock.go . QorJobInterface

type JobBuilder struct {
	b    *Builder
	name string
	r    interface{}
	rmb  *presets.ModelBuilder
	h    JobHandler
}

func newJob(b *Builder, name string) *JobBuilder {
	if b == nil {
		panic("builder is nil")
	}
	if strings.TrimSpace(name) == "" {
		panic("name is empty")
	}

	return &JobBuilder{
		b:    b,
		name: name,
	}
}

type JobHandler func(context.Context, QorJobInterface) error

// r should be ptr to struct
func (jb *JobBuilder) Resource(r interface{}) *JobBuilder {
	{
		v := reflect.TypeOf(r)
		if v.Kind() != reflect.Ptr {
			panic("resource is not ptr to struct")
		}
		if v.Elem().Kind() != reflect.Struct {
			panic("resource is not ptr to struct")
		}
	}

	jb.r = r
	jb.rmb = jb.b.jpb.Model(r)

	if _, ok := r.(Scheduler); ok {
		jb.rmb.Editing().Field("ScheduleTime").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
			t := obj.(Scheduler).GetScheduleTime()
			var v string
			if t != nil {
				v = t.Local().Format("2006-01-02 15:04")
			}
			return vx.VXDateTimePicker().FieldName(field.Name).Label(msgr.ScheduleTime).
				Value(v).
				TimePickerProps(vx.TimePickerProps{
					Format:     "24hr",
					Scrollable: true,
				}).
				ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText)
		}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			v := ctx.R.Form.Get(field.Name)
			if v == "" {
				return nil
			}
			t, err := time.ParseInLocation("2006-01-02 15:04", v, time.Local)
			if err != nil {
				return err
			}
			obj.(Scheduler).SetScheduleTime(&t)
			return nil
		})
	}
	return jb
}

func (jb *JobBuilder) GetResourceBuilder() *presets.ModelBuilder {
	return jb.rmb
}

func (jb *JobBuilder) Handler(h JobHandler) *JobBuilder {
	jb.h = h
	return jb
}

func (jb *JobBuilder) newResourceObject() interface{} {
	if jb.r == nil {
		return nil
	}
	return reflect.New(reflect.TypeOf(jb.r).Elem()).Interface()
}

func (jb *JobBuilder) unmarshalForm(ctx *web.EventContext) (args interface{}, vErr web.ValidationErrors) {
	args = jb.newResourceObject()
	if args != nil {
		vErr = jb.rmb.Editing().RunSetterFunc(ctx, false, args)
	}

	return args, vErr
}

func (jb *JobBuilder) parseArgs(in string) (args interface{}, err error) {
	if jb.r == nil {
		return nil, nil
	}

	args = jb.newResourceObject()
	err = json.Unmarshal([]byte(in), args)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func getModelQorJobInstance(db *gorm.DB, qorJobID uint) (*QorJobInstance, error) {
	var insts []*QorJobInstance
	err := db.Where("qor_job_id = ?", qorJobID).
		Order("created_at desc").
		Limit(1).
		Find(&insts).
		Error
	if err != nil {
		return nil, err
	}
	if len(insts) == 0 {
		return nil, errors.New("no qor job instance")
	}

	return insts[0], nil
}

func (jb *JobBuilder) getJobInstance(qorJobID uint) (*QorJobInstance, error) {
	inst, err := getModelQorJobInstance(jb.b.db, qorJobID)
	if err != nil {
		return nil, err
	}

	inst.jb = jb

	return inst, nil
}

func (jb *JobBuilder) newJobInstance(
	r *http.Request,
	qorJobID uint,
	qorJobName string,
	args interface{},
) (*QorJobInstance, error) {
	var mArgs string
	if v, ok := args.(string); ok {
		mArgs = v
	} else {
		bArgs, err := json.Marshal(args)
		if err != nil {
			return nil, err
		}
		mArgs = string(bArgs)
	}
	inst := QorJobInstance{
		QorJobID: qorJobID,
		Args:     mArgs,
		Job:      qorJobName,
		Status:   JobStatusNew,
	}
	if jb.b.operatorGetter != nil {
		inst.Operator = jb.b.operatorGetter(r)
	}
	err := jb.b.db.Create(&inst).Error
	if err != nil {
		return nil, err
	}

	return jb.getJobInstance(qorJobID)
}

type QueJobInterface interface {
	QorJobInterface

	GetStatus() string
	FetchAndSetStatus() (string, error)
	SetStatus(string) error

	StartRefresh()
	StopRefresh()

	GetHandler() JobHandler
}

// for job handler
type QorJobInterface interface {
	GetJobID() string
	GetJobName() string
	GetOperator() string
	GetArgument() (interface{}, error)
	SetProgress(uint) error
	SetProgressText(string) error
	AddLog(string) error
	AddLogf(format string, a ...interface{}) error
}

var _ QueJobInterface = (*QorJobInstance)(nil)

func (job *QorJobInstance) GetJobName() string {
	return job.Job
}

func (job *QorJobInstance) GetJobID() string {
	return fmt.Sprint(job.QorJobID)
}

func (job *QorJobInstance) GetStatus() string {
	return job.Status
}

func (job *QorJobInstance) FetchAndSetStatus() (string, error) {
	var status string
	{
		db, err := job.jb.b.db.DB()
		if err != nil {
			return job.Status, err
		}

		err = db.QueryRow("select status from qor_job_instances where id = $1", job.ID).Scan(&status)
		if err != nil {
			return job.Status, err
		}
		if status == "" {
			return job.Status, errors.New("failed to fetch qor_job_instance status")
		}
	}

	if job.Status != status {
		err := job.SetStatus(status)
		if err != nil {
			return job.Status, err
		}
	}

	return job.Status, nil
}

func (job *QorJobInstance) SetStatus(status string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Status = status
	if status == JobStatusDone {
		job.Progress = 100
	}

	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) SetProgress(progress uint) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	if progress > 100 {
		progress = 100
	}
	job.Progress = progress

	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) SetProgressText(s string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.ProgressText = s
	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) AddLog(log string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Log += "\n" + log
	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) AddLogf(format string, a ...interface{}) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Log += "\n" + fmt.Sprintf(format, a...)
	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) StartRefresh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	if !job.inRefresh {
		job.inRefresh = true
		job.stopRefresh = false

		go func() {
			job.refresh()
		}()
	}
}

func (job *QorJobInstance) StopRefresh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	err := job.callSave()
	if err != nil {
		log.Println(err)
	}

	job.stopRefresh = true
}

func (job *QorJobInstance) GetHandler() JobHandler {
	return job.jb.h
}

func (job *QorJobInstance) GetArgument() (interface{}, error) {
	return job.jb.parseArgs(job.Args)
}

func (job *QorJobInstance) GetOperator() string {
	return job.Operator
}

func (job *QorJobInstance) shouldCallSave() bool {
	return !job.inRefresh || job.stopRefresh
}

func (job *QorJobInstance) callSave() error {
	err := job.jb.b.setStatus(job.QorJobID, job.Status)
	if err != nil {
		return err
	}
	return job.jb.b.db.Save(job).Error
}

func (job *QorJobInstance) refresh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	err := job.callSave()
	if err != nil {
		log.Println(err)
	}

	if job.stopRefresh {
		job.inRefresh = false
		job.stopRefresh = false
	} else {
		time.AfterFunc(5*time.Second, job.refresh)
	}
}
