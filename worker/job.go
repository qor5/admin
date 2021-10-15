package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/goplaid/x/presets"
	"gorm.io/gorm"
)

type JobBuilder struct {
	b    *Builder
	name string
	r    interface{}
	rmb  *presets.ModelBuilder
	h    JobHandler
}

func newJob(b *Builder, name string) *JobBuilder {
	return &JobBuilder{
		b:    b,
		name: name,
	}
}

type JobHandler func(context.Context, HQorJob) error

func (jb *JobBuilder) Resource(r interface{}) *JobBuilder {
	jb.r = r
	jb.rmb = jb.b.jpb.Model(r)
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

func (jb *JobBuilder) newJobInstance(qorJobID uint, args interface{}) (*QorJobInstance, error) {
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
		Status:   jobStatusNew,
	}
	err := jb.b.db.Create(&inst).Error
	if err != nil {
		return nil, err
	}

	return jb.getJobInstance(qorJobID)
}

type QorJobInterface interface {
	HQorJob

	GetJobID() string
	GetStatus() string
	SetStatus(string) error

	StartReferesh()
	StopReferesh()

	GetHandler() JobHandler
}

// for job handler
type HQorJob interface {
	GetArgument() (interface{}, error)
	SetProgress(uint) error
	SetProgressText(string) error
	AddLog(string) error
}

var _ QorJobInterface = (*QorJobInstance)(nil)

func (job *QorJobInstance) GetJobID() string {
	return fmt.Sprint(job.QorJobID)
}

func (job *QorJobInstance) GetStatus() string {
	return job.Status
}

func (job *QorJobInstance) SetStatus(status string) error {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Status = status
	if status == jobStatusDone {
		job.Progress = 100
	}

	if job.shouldCallSave() {
		err := job.jb.b.setStatus(job.QorJobID, job.Status)
		if err != nil {
			return err
		}
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

	fmt.Println(log)
	job.Log += "\n" + log
	if job.shouldCallSave() {
		return job.callSave()
	}

	return nil
}

func (job *QorJobInstance) StartReferesh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	if !job.inReferesh {
		job.inReferesh = true
		job.stopReferesh = false

		go func() {
			job.referesh()
		}()
	}
}

func (job *QorJobInstance) StopReferesh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	err := job.callSave()
	if err != nil {
		log.Println(err)
	}

	job.stopReferesh = true
}

func (job *QorJobInstance) GetHandler() JobHandler {
	return job.jb.h
}

func (job *QorJobInstance) GetArgument() (interface{}, error) {
	return job.jb.parseArgs(job.Args)
}

func (job *QorJobInstance) shouldCallSave() bool {
	return !job.inReferesh || job.stopReferesh
}

func (job *QorJobInstance) callSave() error {
	return job.jb.b.db.Save(job).Error
}

func (job *QorJobInstance) referesh() {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	err := job.callSave()
	if err != nil {
		log.Println(err)
	}

	if job.stopReferesh {
		job.inReferesh = false
		job.stopReferesh = false
	} else {
		time.AfterFunc(5*time.Second, job.referesh)
	}
}
