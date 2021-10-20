package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/perm"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
	. "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type Builder struct {
	db         *gorm.DB
	q          Queue
	jpb        *presets.Builder
	jbs        []*JobBuilder
	configured bool
}

func New(db *gorm.DB) *Builder {
	if db == nil {
		panic("db can not be nil")
	}

	err := db.AutoMigrate(&QorJob{}, &QorJobInstance{})
	if err != nil {
		panic(err)
	}

	r := &Builder{
		db:  db,
		q:   NewGoQueQueue(db),
		jpb: presets.New(),
	}

	return r
}

// default queue is go-que queue
func (b *Builder) Queue(q Queue) *Builder {
	b.q = q
	return b
}

func (b *Builder) NewJob(name string) *JobBuilder {
	if b.configured {
		panic(fmt.Sprintf("Job should be registered before Worker configured into admin, but %v is registered after that", name))
	}

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

func (b *Builder) Configure(pb *presets.Builder) {
	b.configured = true
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

	permVerifier = perm.NewVerifier("workers", pb.GetPermission())
	pb.I18n().
		RegisterForModule(language.English, I18nWorkerKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nWorkerKey, Messages_zh_CN)

	mb := pb.Model(&QorJob{}).
		Label("Workers").
		URIName("workers").
		MenuIcon("smart_toy")

	lb := mb.Listing("ID", "Job", "Status", "CreatedAt")
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
	lb.Field("Status").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)
		qorJob := obj.(*QorJob)
		return Td(Text(getStatusTranslation(msgr, qorJob.Status)))
	})

	eb := mb.Editing("Job")
	eb.Field("Job").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		ctx.Hub.RegisterEventFunc("worker_renderJobEditingContent", b.eventRenderJobEditingContent)

		jobNames := make([]string, 0, len(b.jbs))
		for _, jb := range b.jbs {
			if editIsAllowed(ctx.R, jb.name) == nil {
				jobNames = append(jobNames, jb.name)
			}
		}
		return Div(
			VSelect().
				Items(jobNames).
				Attr(web.VFieldName("Job")...).
				On("input", web.Plaid().EventFunc("worker_renderJobEditingContent").Go()),
			web.Portal().Name("worker_jobEditingContent"),
		)
	})
	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		qorJob := obj.(*QorJob)
		if pErr := editIsAllowed(ctx.R, qorJob.Job); pErr != nil {
			return pErr
		}

		jb := b.mustGetJobBuilder(qorJob.Job)
		args, err := jb.unmarshalForm(ctx)
		if err != nil {
			return err
		}

		return b.db.Transaction(func(tx *gorm.DB) error {
			j := QorJob{
				Job:    qorJob.Job,
				Status: JobStatusNew,
			}
			err = b.db.Create(&j).Error
			if err != nil {
				return err
			}

			inst, err := jb.newJobInstance(j.ID, qorJob.Job, args)
			if err != nil {
				return err
			}

			return b.q.Add(inst)
		})
	})

	mb.Detailing("DetailingPage").Field("DetailingPage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

		ctx.Hub.RegisterEventFunc("worker_abortJob", b.eventAbortJob)
		ctx.Hub.RegisterEventFunc("worker_rerunJob", b.eventRerunJob)
		ctx.Hub.RegisterEventFunc("worker_updateJob", b.eventUpdateJob)
		ctx.Hub.RegisterEventFunc("worker_updateJobProgressing", b.eventUpdateJobProgressing)

		qorJob := obj.(*QorJob)
		inst, err := getModelQorJobInstance(b.db, qorJob.ID)
		if err != nil {
			return Text(err.Error())
		}

		var scheduledJobDetailing []HTMLComponent
		if inst.Status == JobStatusScheduled {
			jb := b.getJobBuilder(qorJob.Job)
			if jb != nil && jb.r != nil {
				args := jb.newResourceObject()
				err := json.Unmarshal([]byte(inst.Args), &args)
				if err != nil {
					return Text(err.Error())
				}
				body := jb.rmb.Editing().ToComponent(jb.rmb, args, nil, ctx)
				scheduledJobDetailing = []HTMLComponent{
					body,
					If(editIsAllowed(ctx.R, qorJob.Job) == nil,
						Div().Class("d-flex mt-3").Children(
							VSpacer(),
							VBtn(msgr.ActionCancelJob).Color("error").Class("mr-2").
								OnClick("worker_abortJob", fmt.Sprintf("%d", qorJob.ID), qorJob.Job),
							VBtn(msgr.ActionUpdateJob).Color("primary").
								OnClick("worker_updateJob", fmt.Sprintf("%d", qorJob.ID), qorJob.Job),
						),
					),
				}
			} else {
				scheduledJobDetailing = []HTMLComponent{
					VAlert().Dense(true).Type("warning").Children(
						Text(msgr.NoticeJobWontBeExecuted),
					),
					Div(Text("args: " + inst.Args)),
				}
			}
		}

		return Div(
			Div(Text(qorJob.Job)).Class("mb-3 text-h6 font-weight-regular"),
			If(inst.Status == JobStatusScheduled,
				scheduledJobDetailing...,
			).Else(
				Div(
					web.Portal().
						EventFunc("worker_updateJobProgressing", fmt.Sprintf("%d", qorJob.ID), qorJob.Job).
						AutoReloadInterval("vars.worker_updateJobProgressingInterval"),
				).Attr(web.InitContextVars, "{worker_updateJobProgressingInterval: 2000}"),
			),
			web.Portal().Name("worker_snackbar"),
		)
	})
}

func (b *Builder) eventRenderJobEditingContent(ctx *web.EventContext) (er web.EventResponse, err error) {
	jb := b.mustGetJobBuilder(ctx.Event.Value)
	var body HTMLComponent
	if jb.rmb != nil {
		body = jb.rmb.Editing().ToComponent(jb.rmb, jb.r, nil, ctx)
	}
	er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
		Name: "worker_jobEditingContent",
		Body: body,
	})

	return
}

func (b *Builder) eventAbortJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	inst, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}

	err = b.doAbortJob(inst)
	if err != nil {
		_, ok := err.(*cannotAbortError)
		if !ok {
			return er, err
		}
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "worker_snackbar",
			Body: VSnackbar().Value(true).Timeout(3000).Color("warning").Children(
				Text(msgr.NoticeJobCannotBeAborted),
			),
		})
	}

	er.Reload = true
	er.VarsScript = "vars.worker_updateJobProgressingInterval = 2000"
	return er, nil
}

type cannotAbortError struct {
	err error
}

func (e *cannotAbortError) Error() string {
	return e.err.Error()
}

func (b *Builder) doAbortJob(inst *QorJobInstance) (err error) {
	switch inst.Status {
	case JobStatusRunning:
		return b.q.Kill(inst)
	case JobStatusNew, JobStatusScheduled:
		return b.q.Remove(inst)
	default:
		return &cannotAbortError{
			err: fmt.Errorf("job status is %s, cannot be aborted/canceled", inst.Status),
		}
	}
}

func (b *Builder) eventRerunJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

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

	inst, err := jb.newJobInstance(qorJobID, qorJobName, old.Args)
	if err != nil {
		return er, err
	}
	err = b.q.Add(inst)
	if err != nil {
		return er, err
	}

	er.Reload = true
	er.VarsScript = "vars.worker_updateJobProgressingInterval = 2000"
	return
}

func (b *Builder) eventUpdateJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	newArgs, err := jb.unmarshalForm(ctx)
	if err != nil {
		return er, err
	}

	old, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}
	err = b.doAbortJob(old)
	if err != nil {
		_, ok := err.(*cannotAbortError)
		if !ok {
			return er, err
		}
		er.UpdatePortals = append(er.UpdatePortals, &web.PortalUpdate{
			Name: "worker_snackbar",
			Body: VSnackbar().Value(true).Timeout(3000).Color("warning").Children(
				Text(msgr.NoticeJobCannotBeAborted),
			),
		})
		er.Reload = true
		return er, nil
	}

	newInst, err := jb.newJobInstance(qorJobID, qorJobName, newArgs)
	if err != nil {
		return er, err
	}
	err = b.q.Add(newInst)
	if err != nil {
		return er, err
	}

	er.Reload = true
	er.VarsScript = "vars.worker_updateJobProgressingInterval = 2000"
	return er, nil
}

func (b *Builder) eventUpdateJobProgressing(ctx *web.EventContext) (er web.EventResponse, err error) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nWorkerKey, Messages_en_US).(*Messages)

	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	canEdit := editIsAllowed(ctx.R, qorJobName) == nil
	er.Body = jobProgressing(canEdit, msgr, qorJobID, qorJobName, inst.Status, inst.Progress, inst.Log, inst.ProgressText)
	if inst.Status != JobStatusNew && inst.Status != JobStatusRunning {
		er.VarsScript = "vars.worker_updateJobProgressingInterval = 0"
	} else {
		er.VarsScript = "vars.worker_updateJobProgressingInterval = 2000"
	}
	return er, nil
}

func jobProgressing(
	canEdit bool,
	msgr *Messages,
	id uint,
	job string,
	status string,
	progress uint,
	log string,
	progressText string,
) HTMLComponent {
	// https://stackoverflow.com/a/44051405/10150757
	var logLines []HTMLComponent
	logs := strings.Split(log, "\n")
	var reverseStyle string
	if len(logs) > 18 {
		reverseStyle = "display: flex;flex-direction: column-reverse;"
		for i := len(logs) - 1; i >= 0; i-- {
			logLines = append(logLines, P().Style(`
    margin: 0;
    margin-bottom: 4px;`).Children(Text(logs[i])))
		}
	} else {
		for _, l := range logs {
			logLines = append(logLines, P().Style(`
    margin: 0;
    margin-bottom: 4px;`).Children(Text(l)))
		}
	}
	inRefresh := status == JobStatusNew || status == JobStatusRunning
	return Div(
		Div(Text(msgr.DetailTitleStatus)).Class("text-caption"),
		Div().Class("d-flex align-center mb-5").Children(
			Div().Style("width: 120px").Children(
				Text(fmt.Sprintf("%s (%d%%)", getStatusTranslation(msgr, status), progress)),
			),
			VProgressLinear().Value(int(progress)),
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
						OnClick("worker_abortJob", fmt.Sprintf("%d", id), job),
				),
				If(status == JobStatusDone,
					VBtn(msgr.ActionRerunJob).Color("primary").
						OnClick("worker_rerunJob", fmt.Sprintf("%d", id), job),
				),
			),
		),
	)
}

func getStatusTranslation(msgr *Messages, status string) string {
	switch status {
	case JobStatusNew:
		return msgr.StatusNew
	case JobStatusScheduled:
		return msgr.StatusScheduled
	case JobStatusRunning:
		return msgr.StatusRunning
	case JobStatusCancelled:
		return msgr.StatusCancelled
	case JobStatusDone:
		return msgr.StatusDone
	case JobStatusException:
		return msgr.StatusException
	case JobStatusKilled:
		return msgr.StatusKilled
	}
	return status
}
