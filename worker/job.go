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

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	. "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

//go:generate moq -pkg mock -out mock/qor_job.go . QorJobInterface

type JobBuilder struct {
	b              *Builder
	name           string
	r              interface{}
	rmb            *presets.ModelBuilder
	h              JobHandler
	contextHandler func(*web.EventContext) map[string]interface{} // optional
	global         bool
}

func newJob(b *Builder, name string) *JobBuilder {
	if b == nil {
		panic("builder is nil")
	}
	if strings.TrimSpace(name) == "" {
		panic("name is empty")
	}

	return &JobBuilder{
		b:      b,
		name:   name,
		global: true,
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
			return vx.VXDatepicker().Attr(web.VField(field.Name, v)...).Label(msgr.ScheduleTime).Type("datetimepicker").
				Format("YYYY-MM-DD HH:mm")
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

func (jb *JobBuilder) ContextHandler(handler func(*web.EventContext) map[string]interface{}) *JobBuilder {
	jb.contextHandler = handler
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
	context interface{},
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

	var ctx string
	if v, ok := context.(string); ok {
		ctx = v
	} else {
		bArgs, err := json.Marshal(context)
		if err != nil {
			return nil, err
		}
		ctx = string(bArgs)
	}

	inst := QorJobInstance{
		QorJobID: qorJobID,
		Args:     mArgs,
		Context:  ctx,
		Job:      qorJobName,
		Status:   JobStatusNew,
	}
	if jb.b.getCurrentUserIDFunc != nil {
		inst.Operator = jb.b.getCurrentUserIDFunc(r)
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

type JobInfo struct {
	JobID    string
	JobName  string
	Operator string
	Argument interface{}
	Context  map[string]interface{}
}

// for job handler
type QorJobInterface interface {
	GetJobInfo() (*JobInfo, error)
	SetProgress(uint) error
	SetProgressText(string) error
	AddLog(string) error
	AddLogf(format string, a ...interface{}) error
}

var _ QueJobInterface = (*QorJobInstance)(nil)

func (job *QorJobInstance) GetJobInfo() (ji *JobInfo, err error) {
	arg, err := job.getArgument()
	if err != nil {
		return
	}

	context, err := job.getContext()
	if err != nil {
		return
	}

	return &JobInfo{
		JobID:    fmt.Sprint(job.QorJobID),
		JobName:  job.Job,
		Operator: job.Operator,
		Argument: arg,
		Context:  context,
	}, nil
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
	return job.jb.b.db.Create(&QorJobLog{
		QorJobInstanceID: job.ID,
		Log:              log,
	}).Error
}

func (job *QorJobInstance) AddLogf(format string, a ...interface{}) error {
	return job.AddLog(fmt.Sprintf(format, a...))
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

func (job *QorJobInstance) getArgument() (interface{}, error) {
	return job.jb.parseArgs(job.Args)
}

func (job *QorJobInstance) getContext() (map[string]interface{}, error) {
	context := make(map[string]interface{})
	err := json.Unmarshal([]byte(job.Context), &context)
	return context, err
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
