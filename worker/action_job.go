package worker

import (
	"fmt"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/presets"
	"github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	h "github.com/theplant/htmlgo"
)

const (
	ActionJobInputParams = "worker_action_job_input_params"
	ActionJobCreate      = "worker_action_job_create"
	ActionJobResponse    = "worker_action_job_response"
	ActionJobClose       = "worker_action_job_close"
	ActionJobProgressing = "worker_action_job_progressing"
)

var (
	DefaultOriginalPageContextHandler = func(ctx *web.EventContext) map[string]interface{} {
		return map[string]interface{}{
			"URL": ctx.R.Header.Get("Referer"),
		}
	}

	actionJobs = map[string]*ActionJobBuilder{}
)

type ActionJobBuilder struct {
	fullname            string
	shortname           string
	description         string //optional
	hasParams           bool
	displayLog          bool //optional
	progressingInterval int

	b  *Builder    // worker builder
	jb *JobBuilder // job builder
}

func (b *Builder) ActionJob(jobName string, model *presets.ModelBuilder, hander JobHandler) *ActionJobBuilder {
	if jobName == "" {
		panic("job name is required")
	}

	if hander == nil {
		panic("job handler is required")
	}

	fullname := fmt.Sprintf("Action Job - %s - %s", model.Info().Label(), jobName)

	if actionJobs[fullname] != nil {
		return actionJobs[fullname]
	}

	action := &ActionJobBuilder{
		fullname:            fullname,
		shortname:           jobName,
		progressingInterval: 2000,
		jb:                  b.NewJob(fullname).Handler(hander),
		b:                   b,
	}
	actionJobs[fullname] = action
	action.jb.global = false
	return action
}

func (action *ActionJobBuilder) Params(params interface{}) *ActionJobBuilder {
	action.hasParams = true
	action.jb.Resource(params)
	return action
}

func (action *ActionJobBuilder) Global(b bool) *ActionJobBuilder {
	action.jb.global = b
	return action
}

func (action *ActionJobBuilder) ProgressingInterval(interval int) *ActionJobBuilder {
	action.progressingInterval = interval
	return action
}

func (action *ActionJobBuilder) ContextHandler(handler func(*web.EventContext) map[string]interface{}) *ActionJobBuilder {
	action.jb.contextHandler = handler
	return action
}

func (action *ActionJobBuilder) DisplayLog(b bool) *ActionJobBuilder {
	action.displayLog = b
	return action
}

func (action *ActionJobBuilder) Description(description string) *ActionJobBuilder {
	action.description = description
	return action
}

func (action ActionJobBuilder) GetParamsModelBuilder() *presets.ModelBuilder {
	return action.jb.rmb
}

func (action ActionJobBuilder) URL() string {
	return web.Plaid().URL(action.b.mb.Info().ListingHref()).EventFunc(ActionJobInputParams).Query("jobName", action.fullname).Go()
}

func (b *Builder) eventActionJobCreate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		config  = actionJobs[jobName]
		qorJob  = &QorJob{Job: jobName}
	)

	if config == nil {
		return r, fmt.Errorf("job %s not found", jobName)
	}

	job, err := b.createJob(ctx, qorJob)
	if err != nil {
		return
	}
	if b.ab != nil {
		b.ab.AddRecords(activity.ActivityCreate, ctx.R.Context(), job)
	}

	r.VarsScript = web.Plaid().
		URL(b.mb.Info().ListingHref()).
		EventFunc(ActionJobResponse).
		Query(presets.ParamID, fmt.Sprint(job.ID)).
		Query("jobID", fmt.Sprintf("%d", job.ID)).
		Query("jobName", job.Job).
		Go()
	return
}

func (b *Builder) eventActionJobInputParams(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		msgr    = presets.MustGetMessages(ctx.R)
		config  = actionJobs[jobName]
	)

	if config == nil {
		return r, fmt.Errorf("job %s not found", jobName)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "presets_DialogPortalName",
		Body: web.Scope(
			vuetify.VDialog(
				vuetify.VCard(
					vuetify.VCardTitle(
						h.Text(config.shortname),
						vuetify.VSpacer(),
						vuetify.VBtn("").Icon(true).Children(
							vuetify.VIcon("close"),
						).Attr("@click.stop", "vars.presetsDialog=false"),
					),

					h.If(config.description != "", vuetify.VCardSubtitle(
						h.Text(config.description),
					)),

					h.If(config.hasParams, vuetify.VCardText(
						b.jobEditingContent(ctx, jobName, nil),
					)),

					vuetify.VCardActions(
						vuetify.VSpacer(),
						vuetify.VBtn(msgr.Cancel).Elevation(0).Attr("@click", "vars.presetsDialog=false"),
						vuetify.VBtn(msgr.OK).Color("primary").Large(true).
							Attr("@click", web.Plaid().
								URL(b.mb.Info().ListingHref()).
								EventFunc(ActionJobCreate).
								Query("jobName", jobName).
								Go()),
					),
				)).
				Attr("v-model", "vars.presetsDialog").
				Width("600").Persistent(true),
		).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){vars.presetsDialog = true; }, 100)"
	return
}

func (b *Builder) eventActionJobResponse(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		jobID   = ctx.R.FormValue("jobID")
		config  = actionJobs[jobName]
	)

	if config == nil {
		return r, fmt.Errorf("job %s not found", jobName)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "presets_DialogPortalName",
		Body: web.Scope(
			vuetify.VDialog(
				vuetify.VAppBar(
					vuetify.VToolbarTitle(config.shortname).Class("pl-2"),
					vuetify.VSpacer(),
					vuetify.VBtn("").Icon(true).Children(
						vuetify.VIcon("close"),
					).Attr("@click.stop", web.Plaid().
						URL(b.mb.Info().ListingHref()).
						EventFunc(ActionJobClose).
						Query("jobID", jobID).
						Query("jobName", jobName).
						Go()),
				).Color("white").Elevation(0).Dense(true),

				vuetify.VCard(
					vuetify.VCardText(
						h.Div(
							web.Portal().Loader(
								web.Plaid().EventFunc(ActionJobProgressing).
									URL(b.mb.Info().ListingHref()).
									Query("jobID", jobID).
									Query("jobName", jobName),
							).AutoReloadInterval("vars.actionJobProgressingInterval"),
						).Attr(web.InitContextVars, fmt.Sprintf("{actionJobProgressingInterval: %d}", config.progressingInterval)),
					),
				).Tile(true).Attr("style", "box-shadow: none;")).
				Attr("v-model", "vars.presetsDialog").
				Width("600").Persistent(true),
		).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){vars.presetsDialog = true; }, 100)"
	return
}

func (b *Builder) eventActionJobClose(ctx *web.EventContext) (er web.EventResponse, err error) {
	var (
		qorJobID   = uint(ctx.QueryAsInt("jobID"))
		qorJobName = ctx.R.FormValue("jobName")
	)

	er.VarsScript = "vars.presetsDialog = false;vars.actionJobProgressingInterval = 0;"
	if pErr := editIsAllowed(ctx.R, qorJobName); pErr != nil {
		return er, pErr
	}

	jb := b.mustGetJobBuilder(qorJobName)
	inst, err := jb.getJobInstance(qorJobID)
	if err != nil {
		return er, err
	}

	switch inst.Status {
	case JobStatusRunning:
		err = b.q.Kill(inst)
	case JobStatusNew, JobStatusScheduled:
		err = b.q.Remove(inst)
	}

	return er, err
}

func (b *Builder) eventActionJobProgressing(ctx *web.EventContext) (er web.EventResponse, err error) {
	var (
		qorJobID   = uint(ctx.QueryAsInt("jobID"))
		qorJobName = ctx.R.FormValue("jobName")
		config     = actionJobs[qorJobName]
	)

	if config == nil {
		return er, fmt.Errorf("job %s not found", qorJobName)
	}

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	er.Body = h.Div(
		h.Div(vuetify.VProgressLinear(
			h.Strong(fmt.Sprintf("%d%%", inst.Progress)),
		).Value(int(inst.Progress)).Height(20)).Class("mb-5"),
		h.If(config.displayLog, actionJobLog(*config.b, inst)),
		h.If(inst.ProgressText != "",
			h.Div().Class("mb-3").Children(
				h.RawHTML(inst.ProgressText),
			),
		),
	)

	if inst.Status == JobStatusDone || inst.Status == JobStatusException {
		er.VarsScript = "vars.actionJobProgressingInterval = 0;"
	} else {
		er.VarsScript = fmt.Sprintf("vars.actionJobProgressingInterval = %d;", config.progressingInterval)
	}
	return er, nil
}

func actionJobLog(b Builder, inst *QorJobInstance) h.HTMLComponent {
	var logLines []h.HTMLComponent
	logs := make([]string, 0, 100)

	var mLogs []*QorJobLog
	b.db.Where("qor_job_instance_id = ?", inst.ID).
		Order("created_at desc").
		Limit(100).
		Find(&mLogs)

	for i := len(mLogs) - 1; i >= 0; i-- {
		logs = append(logs, mLogs[i].Log)
	}

	var reverseStyle string
	if len(logs) > 18 {
		reverseStyle = "display: flex;flex-direction: column-reverse;"
		for i := len(logs) - 1; i >= 0; i-- {
			logLines = append(logLines, h.P().Style(`margin: 0;margin-bottom: 4px;`).Children(h.Text(logs[i])))
		}
	} else {
		for _, l := range logs {
			logLines = append(logLines, h.P().Style(`margin: 0;margin-bottom: 4px;`).Children(h.Text(l)))
		}
	}
	return h.Div().Class("mb-3").Style(fmt.Sprintf(`
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
	`, reverseStyle)).Children(logLines...)
}
