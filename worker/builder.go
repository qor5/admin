package worker

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	. "github.com/goplaid/x/vuetify"
	"github.com/goplaid/x/vuetifyx"
	. "github.com/theplant/htmlgo"
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
		q:   NewCronQueue(),
		jpb: presets.New(),
	}

	return r
}

// default queue is cron queue
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
			panic(fmt.Sprintf("worker %s already exists\n", name))
		}
	}

	j := newJob(b, name)
	b.jbs = append(b.jbs, j)

	return j
}

func (b *Builder) mustGetJobBuilder(name string) *JobBuilder {
	for _, jb := range b.jbs {
		if jb.name == name {
			return jb
		}
	}

	panic(fmt.Sprintf("no job %s", name))
}

func (b *Builder) getJobBuilderByQorJobID(id uint) (*JobBuilder, error) {
	j := QorJob{}
	err := b.db.Where("id = ?", id).First(&j).Error
	if err != nil {
		return nil, err
	}

	return b.mustGetJobBuilder(j.Job), nil
}

func (b *Builder) runJob(qorJobID uint) error {
	jb, err := b.getJobBuilderByQorJobID(qorJobID)
	if err != nil {
		return err
	}

	inst, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			inst.AddLog(string(debug.Stack()))
			inst.SetProgressText(fmt.Sprint(r))
			inst.SetStatus(jobStatusException)
		}
	}()

	if inst.GetStatus() != jobStatusNew && inst.GetStatus() != jobStatusScheduled {
		return errors.New("invalid job status, current status: " + inst.GetStatus())
	}

	if err = inst.SetStatus(jobStatusRunning); err == nil {
		if err = b.q.Run(inst); err == nil {
			return inst.SetStatus(jobStatusDone)
		}

		inst.SetProgressText(err.Error())
		inst.SetStatus(jobStatusException)
	}

	return nil
}

func (b *Builder) setStatus(id uint, status string) error {
	return b.db.Model(&QorJob{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
		}).
		Error
}

func (b *Builder) Configure(pb *presets.Builder) {
	{
		// Parse job
		cmdLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		qorJobID := cmdLine.String("qor-job", "", "Qor Job ID")
		cmdLine.Parse(os.Args[1:])
		b.configured = true

		if *qorJobID != "" {
			id, err := strconv.ParseUint(*qorJobID, 10, 64)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if err := b.runJob(uint(id)); err == nil {
				os.Exit(0)
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	mb := pb.Model(&QorJob{}).
		Label("Workers").
		URIName("workers").
		MenuIcon("smart_toy")

	lb := mb.Listing("ID", "Job", "Status", "CreatedAt")
	lb.FilterDataFunc(func(ctx *web.EventContext) vuetifyx.FilterData {
		return []*vuetifyx.FilterItem{
			{
				Key:          "status",
				Label:        "Status",
				ItemType:     vuetifyx.ItemTypeSelect,
				SQLCondition: `status %s ?`,
				Options: []*vuetifyx.SelectItem{
					{Text: "New", Value: jobStatusNew},
					{Text: "Scheduled", Value: jobStatusScheduled},
					{Text: "Running", Value: jobStatusRunning},
					{Text: "Cancelled", Value: jobStatusCancelled},
					{Text: "Done", Value: jobStatusDone},
					{Text: "Exception", Value: jobStatusException},
					{Text: "Killed", Value: jobStatusKilled},
				},
			},
		}
	})
	lb.FilterTabsFunc(func(ctx *web.EventContext) []*presets.FilterTab {
		return []*presets.FilterTab{
			{
				Label: "All Jobs",
				Query: url.Values{"all": []string{"1"}},
			},
			{
				Label: "Running",
				Query: url.Values{"status": []string{jobStatusRunning}},
			},
			{
				Label: "Scheduled",
				Query: url.Values{"status": []string{jobStatusScheduled}},
			},
			{
				Label: "Done",
				Query: url.Values{"status": []string{jobStatusDone}},
			},
			{
				Label: "Errors",
				Query: url.Values{"status": []string{jobStatusException}},
			},
		}
	})

	eb := mb.Editing("Job")
	eb.Field("Job").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		ctx.Hub.RegisterEventFunc("worker_renderJobEditingContent", b.eventRenderJobEditingContent)

		jobNames := make([]string, 0, len(b.jbs))
		for _, jb := range b.jbs {
			jobNames = append(jobNames, jb.name)
		}
		return Div(
			VSelect().
				Items(jobNames).
				Attr(web.VFieldName("Job")...).
				On("input", web.Plaid().EventFunc("worker_renderJobEditingContent").Go()),
			web.Portal().Name("jobEditingContent"),
		)
	})
	eb.SaveFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
		qorJob := obj.(*QorJob)
		jb := b.mustGetJobBuilder(qorJob.Job)
		args := jb.newResourceObject()
		if args != nil {
			ctx.MustUnmarshalForm(args)
		}

		j := QorJob{
			Job:    qorJob.Job,
			Status: jobStatusNew,
		}
		err = b.db.Create(&j).Error
		if err != nil {
			return err
		}

		inst, err := jb.newJobInstance(j.ID, args)
		if err != nil {
			return err
		}
		err = b.q.Add(inst)
		if err != nil {
			return err
		}

		return nil
	})

	dtb := mb.Detailing("DetailingPage")
	dtb.Field("DetailingPage").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) HTMLComponent {
		ctx.Hub.RegisterEventFunc("worker_abortJob", b.eventAbortJob)
		ctx.Hub.RegisterEventFunc("worker_rerunJob", b.eventRerunJob)
		ctx.Hub.RegisterEventFunc("worker_updateJobProgressing", b.eventUpdateJobProgressing)

		qorJob := obj.(*QorJob)
		inst, err := getModelQorJobInstance(b.db, qorJob.ID)
		if err != nil {
			return Text(err.Error())
		}

		return Div(
			Div(Text(qorJob.Job)).Class("mb-2 text-h6 font-weight-regular"),
			If(inst.Status == jobStatusScheduled,
				Div(Text(inst.Args)),
				VBtn("cancel scheduled job").OnClick("worker_abortJob", fmt.Sprintf("%d", qorJob.ID), qorJob.Job),
			).Else(
				Div(
					web.Portal().
						EventFunc("worker_updateJobProgressing", fmt.Sprintf("%d", qorJob.ID), qorJob.Job).
						AutoReloadInterval("vars.worker_updateJobProgressingInterval"),
				).Attr(web.InitContextVars, "{worker_updateJobProgressingInterval: 2000}"),
			),
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
		Name: "jobEditingContent",
		Body: body,
	})

	return
}

func (b *Builder) eventAbortJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	jb := b.mustGetJobBuilder(qorJobName)
	inst, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}

	switch inst.Status {
	case jobStatusRunning:
		err = b.q.Kill(inst)
		if err != nil {
			return er, err
		}
		err = inst.SetStatus(jobStatusKilled)
		if err != nil {
			return er, err
		}
	case jobStatusNew, jobStatusScheduled:
		err = inst.SetStatus(jobStatusKilled)
		if err != nil {
			return er, err
		}
		err = b.q.Remove(inst)
		if err != nil {
			return er, err
		}
	default:
		return er, fmt.Errorf("job status is %s, cannot be aborted", inst.Status)
	}

	er.Reload = true
	return
}

func (b *Builder) eventRerunJob(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	jb := b.mustGetJobBuilder(qorJobName)
	old, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}
	if old.Status != jobStatusDone {
		return er, errors.New("job is not done")
	}

	inst, err := jb.newJobInstance(qorJobID, old.Args)
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

func (b *Builder) eventUpdateJobProgressing(ctx *web.EventContext) (er web.EventResponse, err error) {
	qorJobID := uint(ctx.Event.ParamAsInt(0))
	qorJobName := ctx.Event.Params[1]

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	er.Body = jobProgressing(qorJobID, qorJobName, inst.Status, inst.Progress, inst.Log, inst.ProgressText)
	if inst.Status != jobStatusNew && inst.Status != jobStatusRunning {
		er.VarsScript = "vars.worker_updateJobProgressingInterval = 0"
	}
	return er, nil
}

func jobProgressing(
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
	inRefresh := status == jobStatusNew || status == jobStatusRunning
	return Div(
		Div(Text("Status")).Class("text-caption"),
		Div().Class("d-flex align-center mb-3").Children(
			Div().Style("width: 120px").Children(
				Text(fmt.Sprintf("%s (%d%%)", status, progress)),
			),
			VProgressLinear().Value(int(progress)),
		),
		Div(Text("Job Log")).Class("text-caption"),
		Div().Class("mb-2").Style(fmt.Sprintf(`
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
			Div().Class("mb-2").Children(
				RawHTML(progressText),
			),
		),

		If(inRefresh,
			VBtn("abort job").OnClick("worker_abortJob", fmt.Sprintf("%d", id), job),
		),
		If(status == jobStatusDone,
			VBtn("rerun job").OnClick("worker_rerunJob", fmt.Sprintf("%d", id), job),
		),
	)
}
