package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/qor5/x/v3/ui/vuetifyx"
	. "github.com/theplant/htmlgo"
	"github.com/tnclong/go-que/pg"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
)

type Builder struct {
	db                   *gorm.DB
	q                    Queue
	jpb                  *presets.Builder // for render job form
	pb                   *presets.Builder
	jbs                  []*JobBuilder
	mb                   *presets.ModelBuilder
	getCurrentUserIDFunc func(r *http.Request) string
	ab                   *activity.Builder
}

// Options contains configuration options for worker Builder.
type Options struct {
	DB          *gorm.DB
	AutoMigrate bool // Auto migrate worker tables
}

// New creates a new worker Builder with auto-migration enabled (default behavior).
func New(db *gorm.DB) *Builder {
	return newWithConfigs(db, NewGoQueQueue(db), true)
}

// NewWithQueue creates a new worker Builder with a custom queue and auto-migration enabled.
func NewWithQueue(db *gorm.DB, q Queue) *Builder {
	return newWithConfigs(db, q, true)
}

// NewWithOptions creates a new worker Builder with custom options.
// AutoMigrate defaults to false. Other settings can be configured via builder methods.
func NewWithOptions(opts *Options) *Builder {
	if opts == nil {
		panic("options cannot be nil")
	}
	if opts.DB == nil {
		panic("db cannot be nil")
	}

	q := newGoQueQueue(opts.DB)
	return newWithConfigs(opts.DB, q, opts.AutoMigrate)
}

// AutoMigrate creates or updates all worker-related tables:
// qor_jobs, qor_job_instances, qor_job_logs, go_que_errors, goque_jobs.
// This is automatically called by New() and NewWithQueue().
func AutoMigrate(db *gorm.DB) error {
	// Migrate worker tables
	if err := db.AutoMigrate(&QorJob{}, &QorJobInstance{}, &QorJobLog{}, &GoQueError{}); err != nil {
		return err
	}

	// Migrate goque_jobs table
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if _, err = pg.New(sqlDB); err != nil {
		return err
	}

	return nil
}

func newWithConfigs(db *gorm.DB, q Queue, autoMigrate bool) *Builder {
	if db == nil {
		panic("db can not be nil")
	}

	if autoMigrate {
		if err := AutoMigrate(db); err != nil {
			panic(err)
		}
	}

	r := &Builder{
		db:  db,
		q:   q,
		jpb: presets.New(),
	}

	return r
}

// default queue is go-que queue
func (b *Builder) Queue(q Queue) *Builder {
	b.q = q
	return b
}

func (b *Builder) GetCurrentUserIDFunc(f func(r *http.Request) string) *Builder {
	b.getCurrentUserIDFunc = f
	return b
}

// Activity sets Activity Builder to log activities
func (b *Builder) Activity(ab *activity.Builder) *Builder {
	b.ab = ab
	return b
}

func (b *Builder) NewJob(name string) *JobBuilder {
	for _, jb := range b.jbs {
		if jb.name == name {
			panic(fmt.Sprintf("worker %s already exists", name))
		}
	}

	j := newJob(b, name)
	b.jbs = append(b.jbs, j)

	return j
}

func (b *Builder) getJobBuilder(name string) *JobBuilder {
	for _, jb := range b.jbs {
		if jb.name == name {
			return jb
		}
	}

	return nil
}

func (b *Builder) mustGetJobBuilder(name string) *JobBuilder {
	jb := b.getJobBuilder(name)

	if jb == nil {
		panic(fmt.Sprintf("no job %s", name))
	}

	return jb
}

func (b *Builder) getJobBuilderByQorJobID(id uint) (*JobBuilder, error) {
	j := QorJob{}
	err := b.db.Where("id = ?", id).First(&j).Error
	if err != nil {
		return nil, err
	}

	return b.getJobBuilder(j.Job), nil
}

func (b *Builder) setStatus(id uint, status string) error {
	return b.db.Model(&QorJob{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
		}).
		Error
}

var permVerifier *perm.Verifier

func (b *Builder) Install(pb *presets.Builder) error {
	b.pb = pb
	permVerifier = perm.NewVerifier("workers", pb.GetPermission())
	pb.GetI18n().
		RegisterForModule(language.English, I18nWorkerKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nWorkerKey, Messages_zh_CN)

	mb := pb.Model(&QorJob{}).
		Label("Workers").
		URIName("workers").
		MenuIcon("mdi-briefcase")

	b.mb = mb
	mb.RegisterEventFunc("worker_selectJob", b.eventSelectJob)
	mb.RegisterEventFunc("worker_abortJob", b.eventAbortJob)
	mb.RegisterEventFunc("worker_rerunJob", b.eventRerunJob)
	mb.RegisterEventFunc("worker_updateJob", b.eventUpdateJob)
	mb.RegisterEventFunc("worker_updateJobProgressing", b.eventUpdateJobProgressing)
	mb.RegisterEventFunc("worker_loadHiddenLogs", b.eventLoadHiddenLogs)
	mb.RegisterEventFunc(ActionJobInputParams, b.eventActionJobInputParams)
	mb.RegisterEventFunc(ActionJobCreate, b.eventActionJobCreate)
	mb.RegisterEventFunc(ActionJobResponse, b.eventActionJobResponse)
	mb.RegisterEventFunc(ActionJobClose, b.eventActionJobClose)
	mb.RegisterEventFunc(ActionJobProgressing, b.eventActionJobProgressing)

	lb := mb.Listing("ID", "Job", "Status", "CreatedAt")
	lb.RowMenu().Empty()
	lb.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
		return []*vuetifyx.FilterItem{
			{
				Key:          "status",
				Label:        "Status",
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `status %s ?`,
				Options: []*vuetifyx.SelectItem{
					{Text: msgr.StatusNew, Value: JobStatusNew},
					{Text: msgr.StatusScheduled, Value: JobStatusScheduled},
					{Text: msgr.StatusRunning, Value: JobStatusRunning},
					{Text: msgr.StatusCancelled, Value: JobStatusCancelled},
					{Text: msgr.StatusDone, Value: JobStatusDone},
					{Text: msgr.StatusException, Value: JobStatusException},
					{Text: msgr.StatusKilled, Value: JobStatusKilled},
				},
			},
		}
	})
	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
		return []*presets.FilterTab{
			{
				Label: msgr.FilterTabAll,
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: msgr.FilterTabRunning,
				Query: url.Values{"status": []string{JobStatusRunning}},
			},
			{
				Label: msgr.FilterTabScheduled,
				Query: url.Values{"status": []string{JobStatusScheduled}},
			},
			{
				Label: msgr.FilterTabDone,
				Query: url.Values{"status": []string{JobStatusDone}},
			},
			{
				Label: msgr.FilterTabErrors,
				Query: url.Values{"status": []string{JobStatusException}},
			},
		}
	})
	lb.Field("Job").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		qorJob := obj.(*QorJob)
		return Td(Text(getTJob(ctx.R, qorJob.Job)))
	})
	lb.Field("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
		qorJob := obj.(*QorJob)
		return Td(Text(getTStatus(msgr, qorJob.Status)))
	})

	eb := mb.Editing("Job", "Args")

	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
		qorJob := obj.(*QorJob)
		if qorJob.Job == "" {
			err.FieldError("Job", msgr.PleaseSelectJob)
		}

		return err
	})

	type JobSelectItem struct {
		Label string
		Value string
	}
	eb.Field("Job").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		qorJob := obj.(*QorJob)
		return web.Portal(b.jobSelectList(ctx, qorJob.Job)).Name("worker_jobSelectList")
	})
	eb.Field("Args").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
			if fvErr := vErr.GetFieldErrors(field.Name); len(fvErr) > 0 {
				errM := make(map[string][]string)
				if err := json.Unmarshal([]byte(fvErr[0]), &errM); err == nil {
					for f, es := range errM {
						for _, e := range es {
							ve.FieldError(f, e)
						}
					}
				}
			}
		}

		qorJob := obj.(*QorJob)
		return web.Portal(b.jobEditingContent(ctx, qorJob.Job, qorJob.Args)).Name("worker_jobEditingContent")
	})

	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		qorJob := obj.(*QorJob)
		if qorJob.Job == "" {
			return errors.New("job is required")
		}
		j, err := b.createJob(ctx, qorJob)
		if err != nil {
			return err
		}
		if b.ab != nil {
			b.ab.OnCreate(ctx.R.Context(), j)
		}
		return
	})

	mb.Detailing("DetailingPage").Field("DetailingPage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

		qorJob := obj.(*QorJob)
		inst, err := getModelQorJobInstance(b.db, qorJob.ID)
		if err != nil {
			return Text(err.Error())
		}

		var scheduledJobDetailing []HTMLComponent
		eURL := path.Join(b.mb.Info().ListingHref(), fmt.Sprint(qorJob.ID))
		if inst.Status == JobStatusScheduled {
			jb := b.getJobBuilder(qorJob.Job)
			if jb != nil && jb.r != nil {
				args := jb.newResourceObject()
				err := json.Unmarshal([]byte(inst.Args), &args)
				if err != nil {
					return Text(err.Error())
				}
				body := jb.rmb.Editing().ToComponent(jb.rmb.Info(), args, ctx)
				scheduledJobDetailing = []HTMLComponent{
					body,
					If(editIsAllowed(ctx.R, qorJob.Job) == nil,
						Div().Class("d-flex mt-3").Children(
							VSpacer(),
							VBtn(msgr.ActionCancelJob).Color("error").Class("mr-2").
								Attr("@click", web.Plaid().
									URL(eURL).
									EventFunc("worker_abortJob").
									Query("jobID", fmt.Sprintf("%d", qorJob.ID)).
									Query("job", qorJob.Job).
									Go()),
							VBtn(msgr.ActionUpdateJob).Color("primary").
								Attr("@click", web.Plaid().
									URL(eURL).
									EventFunc("worker_updateJob").
									Query("jobID", fmt.Sprintf("%d", qorJob.ID)).
									Query("job", qorJob.Job).
									Go()),
						),
					),
				}
			} else {
				scheduledJobDetailing = []HTMLComponent{
					VAlert().Density(DensityCompact).Type("warning").Children(
						Text(msgr.NoticeJobWontBeExecuted),
					),
					Div(Text("args: " + inst.Args)),
				}
			}
		}

		// Set initial refresh interval based on job status
		initialInterval := 0
		if inst.Status == JobStatusNew || inst.Status == JobStatusRunning {
			initialInterval = 2000
		}

		return Div(
			Div(Text(getTJob(ctx.R, qorJob.Job))).Class("mb-3 text-h6 font-weight-regular"),
			web.Scope(
				If(inst.Status == JobStatusScheduled,
					scheduledJobDetailing...,
				).Else(
					web.Portal().
						Loader(web.Plaid().EventFunc("worker_updateJobProgressing").
							URL(eURL).
							Query("jobID", fmt.Sprintf("%d", qorJob.ID)).
							Query("job", qorJob.Job),
						).
						AutoReloadInterval("locals.worker_updateJobProgressingInterval"),
				),
			).VSlot("{ locals }").Init(fmt.Sprintf("{worker_updateJobProgressingInterval: %d}", initialInterval)),
			web.Portal().Name("worker_snackbar"),
		)
	})

	if b.ab != nil {
		b.ab.RegisterModel(mb).SkipCreate().SkipEdit().SkipDelete().
			AddTypeHanders(time.Time{}, func(old, now interface{}, prefixField string) []activity.Diff {
				fm := "2006-01-02 15:04:05"
				oldString := old.(time.Time).Format(fm)
				nowString := now.(time.Time).Format(fm)
				if oldString != nowString {
					return []activity.Diff{
						{Field: prefixField, Old: oldString, New: nowString},
					}
				}
				return []activity.Diff{}
			}).
			AddTypeHanders(Schedule{}, func(old, now interface{}, prefixField string) []activity.Diff {
				fm := "2006-01-02 15:04:05"
				oldString := old.(Schedule).ScheduleTime.Format(fm)
				nowString := now.(Schedule).ScheduleTime.Format(fm)
				if oldString != nowString {
					return []activity.Diff{
						{Field: prefixField, Old: oldString, New: nowString},
					}
				}
				return []activity.Diff{}
			})
	}

	return nil
}

func (b *Builder) Listen() {
	var jds []*QorJobDefinition
	for _, jb := range b.jbs {
		jds = append(jds, &QorJobDefinition{
			Name:    jb.name,
			Handler: jb.h,
		})
	}
	err := b.q.Listen(jds, func(qorJobID uint) (QueJobInterface, error) {
		jb, err := b.getJobBuilderByQorJobID(qorJobID)
		if err != nil {
			return nil, err
		}
		if jb == nil {
			return nil, errors.New("failed to find job (job name modified?)")
		}

		return jb.getJobInstance(qorJobID)
	})
	if err != nil {
		panic(err)
	}
}

func (b *Builder) Shutdown(ctx context.Context) error {
	return b.q.Shutdown(ctx)
}

func (b *Builder) createJob(ctx *web.EventContext, qorJob *QorJob) (j *QorJob, err error) {
	if err = editIsAllowed(ctx.R, qorJob.Job); err != nil {
		return
	}

	jb := b.mustGetJobBuilder(qorJob.Job)

	// encode args
	args, vErr := jb.unmarshalForm(ctx)
	if vErr.HaveErrors() {
		errM := make(map[string][]string)
		argsT := reflect.TypeOf(jb.r).Elem()
		for i := 0; i < argsT.NumField(); i++ {
			fName := argsT.Field(i).Name
			errM[fName] = vErr.GetFieldErrors(fName)
		}
		bErrM, _ := json.Marshal(errM)
		err = errors.New(string(bErrM))
		return
	}

	// encode context
	context := make(map[string]interface{})
	for key, v := range DefaultOriginalPageContextHandler(ctx) {
		context[key] = v
	}

	if jb.contextHandler != nil {
		for key, v := range jb.contextHandler(ctx) {
			context[key] = v
		}
	}

	err = b.db.Transaction(func(tx *gorm.DB) error {
		j = &QorJob{
			Job:    qorJob.Job,
			Status: JobStatusNew,
		}
		err = tx.Create(j).Error
		if err != nil {
			return err
		}
		var inst *QorJobInstance
		inst, err = jb.newJobInstance(ctx.R, j.ID, qorJob.Job, args, context)
		if err != nil {
			return err
		}
		return b.q.Add(ctx.R.Context(), inst)
	})
	return
}

func (b *Builder) eventSelectJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	job := ctx.R.FormValue("jobName")
	er.UpdatePortals = append(er.UpdatePortals,
		&web.PortalUpdate{
			Name: "worker_jobEditingContent",
			Body: b.jobEditingContent(ctx, job, nil),
		},
		&web.PortalUpdate{
			Name: "worker_jobSelectList",
			Body: b.jobSelectList(ctx, job),
		},
	)

	return
}

func (b *Builder) eventAbortJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.ParamAsInt("jobID"))
	qorJobName := ctx.R.FormValue("job")

	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	inst, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}
	isScheduled := inst.Status == JobStatusScheduled

	err = b.doAbortJob(ctx.R.Context(), inst)
	if err != nil {
		_, ok := err.(*cannotAbortError)
		if !ok {
			return er, err
		}
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "worker_snackbar",
			Body: VSnackbar().ModelValue(true).Timeout(3000).Color("warning").Children(
				Text(msgr.NoticeJobCannotBeAborted),
			),
		})
	}

	er.Reload = true

	if b.ab != nil {
		action := "Abort"
		if isScheduled {
			action = "Cancel"
		}
		b.ab.Log(ctx.R.Context(), action, &QorJob{
			Model: gorm.Model{
				ID: inst.QorJobID,
			},
		}, nil)
	}

	return er, nil
}

type cannotAbortError struct {
	err error
}

func (e *cannotAbortError) Error() string {
	return e.err.Error()
}

func (b *Builder) doAbortJob(ctx context.Context, inst *QorJobInstance) (err error) {
	switch inst.Status {
	case JobStatusRunning:
		return b.q.Kill(ctx, inst)
	case JobStatusNew, JobStatusScheduled:
		return b.q.Remove(ctx, inst)
	default:
		return &cannotAbortError{
			err: fmt.Errorf("job status is %s, cannot be aborted/canceled", inst.Status),
		}
	}
}

func (b *Builder) eventRerunJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.ParamAsInt("jobID"))
	qorJobName := ctx.R.FormValue("job")

	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	old, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}
	if old.Status != JobStatusDone {
		return er, errors.New("job is not done")
	}

	inst, err := jb.newJobInstance(ctx.R, qorJobID, qorJobName, old.Args, old.Context)
	if err != nil {
		return er, err
	}
	err = b.setStatus(qorJobID, JobStatusNew)
	if err != nil {
		return er, err
	}
	err = b.q.Add(ctx.R.Context(), inst)
	if err != nil {
		return er, err
	}

	er.Reload = true

	if b.ab != nil {
		b.ab.Log(ctx.R.Context(), "Rerun", &QorJob{
			Model: gorm.Model{
				ID: inst.QorJobID,
			},
		}, nil)
	}
	return
}

func (b *Builder) eventUpdateJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.ParamAsInt("jobID"))
	qorJobName := ctx.R.FormValue("job")

	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	newArgs, argsVErr := jb.unmarshalForm(ctx)
	if argsVErr.HaveErrors() {
		return er, errors.New("invalid arguments")
	}

	contexts := make(map[string]interface{})
	for key, v := range DefaultOriginalPageContextHandler(ctx) {
		contexts[key] = v
	}
	if jb.contextHandler != nil {
		for key, v := range jb.contextHandler(ctx) {
			contexts[key] = v
		}
	}

	old, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}
	oldArgs, _ := jb.parseArgs(old.Args)
	err = b.doAbortJob(ctx.R.Context(), old)
	if err != nil {
		_, ok := err.(*cannotAbortError)
		if !ok {
			return er, err
		}
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "worker_snackbar",
			Body: VSnackbar().ModelValue(true).Timeout(3000).Color("warning").Children(
				Text(msgr.NoticeJobCannotBeAborted),
			),
		})
		er.Reload = true
		return er, nil
	}

	newInst, err := jb.newJobInstance(ctx.R, qorJobID, qorJobName, newArgs, contexts)
	if err != nil {
		return er, err
	}
	err = b.q.Add(ctx.R.Context(), newInst)
	if err != nil {
		return er, err
	}

	er.Reload = true
	if b.ab != nil {
		b.ab.OnEdit(
			ctx.R.Context(),
			&QorJob{
				Model: gorm.Model{
					ID: newInst.QorJobID,
				},
				Args: oldArgs,
			},
			&QorJob{
				Model: gorm.Model{
					ID: newInst.QorJobID,
				},
				Args: newArgs,
			},
		)
	}
	return er, nil
}

func (b *Builder) eventUpdateJobProgressing(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.ParamAsInt("jobID"))
	qorJobName := ctx.R.FormValue("job")

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	canEdit := editIsAllowed(ctx.R, qorJobName) == nil
	logs := make([]string, 0, 100)
	hasMoreLogs := false
	{
		var count int64
		err = b.db.Model(&QorJobLog{}).
			Where("qor_job_instance_id = ?", inst.ID).
			Count(&count).
			Error
		if err != nil {
			return er, err
		}
		if count > 100 {
			hasMoreLogs = true
		}
		if count > 0 {
			var mLogs []*QorJobLog
			err = b.db.Where("qor_job_instance_id = ?", inst.ID).
				Order("created_at desc").
				Limit(100).
				Find(&mLogs).
				Error
			if err != nil {
				return er, err
			}
			for i := len(mLogs) - 1; i >= 0; i-- {
				logs = append(logs, mLogs[i].Log)
			}
		}
	}
	er.Body = b.jobProgressing(canEdit, msgr, qorJobID, qorJobName, inst.Status, inst.Progress, logs, hasMoreLogs, inst.ProgressText)
	return er, nil
}

func (b *Builder) eventLoadHiddenLogs(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.ParamAsInt("jobID"))
	currentCount := ctx.ParamAsInt("currentCount")

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	var logs []*QorJobLog
	err = b.db.Where("qor_job_instance_id = ?", inst.ID).
		Order("created_at desc").
		Offset(currentCount).
		Find(&logs).
		Error
	if err != nil {
		return er, err
	}
	logLines := make([]HTMLComponent, 0, len(logs))
	for i := len(logs) - 1; i >= 0; i-- {
		logLines = append(logLines, P().Style(`
    margin: 0;
    margin-bottom: 4px;`).Children(Text(logs[i].Log)))
	}
	er.UpdatePortals = append(er.UpdatePortals,
		&web.PortalUpdate{
			Name: "worker_hiddenLogs",
			Body: Div(logLines...),
		},
	)
	return er, nil
}

func (b *Builder) jobProgressing(
	canEdit bool,
	msgr *Messages,
	id uint,
	job string,
	status string,
	progress uint,
	logs []string,
	hasMoreLogs bool,
	progressText string,
) HTMLComponent {
	logLines := make([]HTMLComponent, 0, len(logs)+1)
	if hasMoreLogs {
		logLines = append(logLines, web.Portal(
			VBtn("Load hidden logs").Attr("@click", web.Plaid().EventFunc("worker_loadHiddenLogs").
				Query("jobID", id).
				Query("currentCount", len(logs)).Go()).
				Size(SizeSmall).
				Variant(VariantFlat).
				Class("mb-3"),
		).Name("worker_hiddenLogs"))
	}
	for _, l := range logs {
		logLines = append(logLines, P().Style(`
    margin: 0;
    margin-bottom: 4px;`).Children(Text(l)))
	}
	// https://stackoverflow.com/a/44051405/10150757
	var reverseStyle string
	if len(logs) > 18 {
		reverseStyle = "display: flex;flex-direction: column-reverse;"
		for i, j := 0, len(logLines)-1; i < j; i, j = i+1, j-1 {
			logLines[i], logLines[j] = logLines[j], logLines[i]
		}
	}
	inRefresh := status == JobStatusNew || status == JobStatusRunning
	eURL := path.Join(b.mb.Info().ListingHref(), fmt.Sprint(id))

	// Set refresh interval based on job status
	interval := 0
	if inRefresh {
		interval = 2000
	}

	return Div(
		// Portal passes parent Scope's locals to its body
		// Use v-on-mounted to set interval when Portal body renders
		Div().Style("display:none").Attr("v-on-mounted", fmt.Sprintf("() => { locals.worker_updateJobProgressingInterval = %d }", interval)),

		Div(Text(msgr.DetailTitleStatus)).Class("text-caption"),
		Div().Class("d-flex align-center mb-5").Children(
			Div().Style("width: 120px").Children(
				Text(fmt.Sprintf("%s (%d%%)", getTStatus(msgr, status), progress)),
			),
			VProgressLinear().ModelValue(int(progress)),
		),

		Div(Text(msgr.DetailTitleLog)).Class("text-caption"),
		Div().Class("mb-3").Style(fmt.Sprintf(`
		background-color: #222;
		color: #fff;
		font-family: menlo,Roboto,Helvetica,Arial,sans-serif;
		height: 300px;
		padding: 8px;
		overflow: auto;
		box-sizing: border-box;
		font-size: 12px;
		line-height: 1;
		%s
		`, reverseStyle)).Children(
			logLines...,
		),

		If(progressText != "",
			Div().Class("mb-3").Children(
				RawHTML(progressText),
			),
		),

		If(canEdit,
			Div().Class("d-flex mt-3").Children(
				VSpacer(),
				If(inRefresh,
					VBtn(msgr.ActionAbortJob).Color("error").
						Attr("@click", web.Plaid().
							URL(eURL).
							EventFunc("worker_abortJob").
							Query("jobID", fmt.Sprintf("%d", id)).
							Query("job", job).
							Go()),
				),
				If(status == JobStatusDone,
					VBtn(msgr.ActionRerunJob).Color("primary").
						Attr("@click", web.Plaid().
							URL(eURL).
							EventFunc("worker_rerunJob").
							Query("jobID", fmt.Sprintf("%d", id)).
							Query("job", job).
							Go()),
				),
			),
		),
	)
}

func (b *Builder) jobSelectList(
	ctx *web.EventContext,
	job string,
) HTMLComponent {
	var vErr web.ValidationErrors
	if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
		vErr = *ve
	}
	var alert HTMLComponent
	if v := vErr.GetFieldErrors("Job"); len(v) > 0 {
		alert = VAlert(Text(strings.Join(v, ","))).Type("error")
	}
	items := make([]HTMLComponent, 0, len(b.jbs))
	for _, jb := range b.jbs {
		if !jb.global {
			continue
		}
		label := getTJob(ctx.R, jb.name)
		if editIsAllowed(ctx.R, jb.name) == nil {
			items = append(items,
				VListItem(
					VListItemTitle(
						A(Text(label)).Attr("@click",
							web.Plaid().EventFunc("worker_selectJob").
								Query("jobName", jb.name).
								Go(),
						),
					)),
			)
		}
	}

	return Div(
		Input("").Type("hidden").Attr(web.VField("Job", job)...),
		If(job == "",
			alert,
			VList(items...).Nav(true).Density(DensityCompact),
		).Else(
			Div(
				VIcon("arrow_back").Attr("@click",
					web.Plaid().EventFunc("worker_selectJob").
						Query("jobName", "").
						Go(),
				),
			).Class("mb-3"),
			Div(Text(getTJob(ctx.R, job))).Class("mb-3 text-h6").Style("font-weight: inherit"),
		),
	)
}

func (b *Builder) jobEditingContent(
	ctx *web.EventContext,
	job string,
	args interface{},
) HTMLComponent {
	if job == "" {
		return Template()
	}

	jb := b.mustGetJobBuilder(job)
	var argsObj interface{}
	if args != nil {
		argsObj = args
	} else {
		argsObj = jb.r
	}

	if jb.rmb == nil {
		return Template()
	}
	return jb.rmb.Editing().ToComponent(jb.rmb.Info(), argsObj, ctx)
}
