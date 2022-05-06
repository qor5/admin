package worker

import (
	"fmt"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
)

const (
	JobActionCreate           = "worker_job_action_create"
	JobActionCreateWithParams = "worker_job_action_create_with_params"
	JobActionResponse         = "worker_job_action_response"
	JobActionClose            = "worker_job_action_close"
	JobActionProgressing      = "worker_job_action_progressing"
)

var (
	DefaultOriginalPageContextHandles = map[string]func(*web.EventContext) interface{}{
		"URL": func(ctx *web.EventContext) interface{} {
			return ctx.R.Header.Get("Referer")
		},
	}

	actionConfigs = map[string]*actionConfig{}
)

type JobActionArgs struct {
	OriginalPageContext map[string]interface{}
	ActionParams        interface{}
}

type JobActionConfig struct {
	Name   string
	Hander JobHandler //optional

	OriginalPageContextHandles map[string]func(*web.EventContext) interface{} //optional
	Params                     interface{}                                    // optional
	ParamsModelBuilder         *presets.ModelBuilder                          // Params ModelBuilder optional

	DisplayLog bool // optional
}

type actionConfig struct {
	hasParams      bool
	displayLog     bool
	contextHandles map[string]func(*web.EventContext) interface{}
}

func (b *Builder) JobAction(cfg *JobActionConfig) presets.ComponentFunc {
	if cfg.Name == "" {
		panic("job name is required")
	}

	actionConfigs[cfg.Name] = &actionConfig{
		hasParams:      cfg.Params != nil,
		displayLog:     cfg.DisplayLog,
		contextHandles: cfg.OriginalPageContextHandles,
	}

	if job := b.getJobBuilder(cfg.Name); job != nil {
		cfg.Params = job.r
		return b.actionComponentFunc(cfg)
	}

	if cfg.Hander == nil {
		panic("job handler is required")
	}

	jb := b.NewJob(cfg.Name).Handler(cfg.Hander)

	if cfg.Params != nil {
		jb.Resource(cfg.Params)
	}

	if cfg.ParamsModelBuilder != nil {
		if cfg.Params == nil {
			cfg.Params = cfg.ParamsModelBuilder.NewModel()
		}
		jb.b.mb = cfg.ParamsModelBuilder
	}
	return b.actionComponentFunc(cfg)
}

func (b *Builder) actionComponentFunc(cfg *JobActionConfig) presets.ComponentFunc {
	return func(ctx *web.EventContext) h.HTMLComponent {
		var eventName = JobActionCreate
		if cfg.Params != nil {
			eventName = JobActionCreateWithParams
		}

		return vuetify.VBtn(cfg.Name).
			Color("primary").
			Depressed(true).
			Class("ml-2").
			Attr("@click", web.Plaid().
				URL(b.mb.Info().ListingHref()).
				EventFunc(eventName).
				Query("jobName", cfg.Name).
				Go())
	}
}

func (b *Builder) eventJobActionCreate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		config  = actionConfigs[jobName]
		qorJob  = &QorJob{Job: jobName}
	)

	if config == nil {
		return r, fmt.Errorf("job %s not found", jobName)
	}

	args := JobActionArgs{
		OriginalPageContext: make(map[string]interface{}),
	}
	for key, f := range DefaultOriginalPageContextHandles {
		args.OriginalPageContext[key] = f(ctx)
	}
	for key, f := range config.contextHandles {
		args.OriginalPageContext[key] = f(ctx)
	}
	if config.hasParams {
		jb := b.mustGetJobBuilder(jobName)
		if actionParams, err := jb.unmarshalForm(ctx); !err.HaveErrors() {
			args.ActionParams = actionParams
		}
	}
	qorJob.args = args

	job, err := b.createJob(ctx, qorJob)
	if err != nil {
		return
	}

	r.VarsScript = web.Plaid().
		URL(b.mb.Info().ListingHref()).
		EventFunc(JobActionResponse).
		Query(presets.ParamID, fmt.Sprint(job.ID)).
		Query("jobID", fmt.Sprintf("%d", job.ID)).
		Query("jobName", job.Job).
		Go()
	return
}

func (b *Builder) eventJobActionCreateWithParams(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		msgr    = presets.MustGetMessages(ctx.R)
		config  = actionConfigs[jobName]
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
						h.Text(jobName),
						vuetify.VSpacer(),
						vuetify.VBtn("").Icon(true).Children(
							vuetify.VIcon("close"),
						).Attr("@click.stop", "vars.presetsDialog=false"),
					),
					vuetify.VCardText(
						b.jobEditingContent(ctx, jobName, nil),
					),
					vuetify.VCardActions(
						vuetify.VSpacer(),
						vuetify.VBtn(msgr.Cancel).Elevation(0).Attr("@click", "vars.presetsDialog=false"),
						vuetify.VBtn(msgr.OK).Color("primary").Large(true).
							Attr("@click", web.Plaid().
								URL(b.mb.Info().ListingHref()).
								EventFunc(JobActionCreate).
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

func (b *Builder) eventJobActionResponse(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		jobID   = ctx.R.FormValue("jobID")
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "presets_DialogPortalName",
		Body: web.Scope(
			vuetify.VDialog(
				vuetify.VAppBar(
					vuetify.VToolbarTitle(jobName).Class("pl-2"),
					vuetify.VSpacer(),
					vuetify.VBtn("").Icon(true).Children(
						vuetify.VIcon("close"),
					).Attr("@click.stop", web.Plaid().
						URL(b.mb.Info().ListingHref()).
						EventFunc(JobActionClose).
						Query("jobID", jobID).
						Query("jobName", jobName).
						Go()),
				).Color("white").Elevation(0).Dense(true),

				vuetify.VCard(
					vuetify.VCardText(
						h.Div(
							web.Portal().Loader(
								web.Plaid().EventFunc(JobActionProgressing).
									URL(b.mb.Info().ListingHref()).
									Query("jobID", jobID).
									Query("jobName", jobName),
							).AutoReloadInterval("vars.jobActionProgressingInterval"),
						).Attr(web.InitContextVars, "{jobActionProgressingInterval: 2000, jobActionFinshed: false}"),
					),
				).Tile(true).Attr("style", "box-shadow: none;")).
				Attr("v-model", "vars.presetsDialog").
				Width("600").Persistent(true),
		).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){vars.presetsDialog = true; }, 100)"
	return
}

func (b *Builder) eventJobActionClose(ctx *web.EventContext) (er web.EventResponse, err error) {
	var (
		qorJobID   = uint(ctx.QueryAsInt("jobID"))
		qorJobName = ctx.R.FormValue("jobName")
	)

	er.VarsScript = "vars.presetsDialog = false;"
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

func (b *Builder) eventJobActionProgressing(ctx *web.EventContext) (er web.EventResponse, err error) {
	var (
		qorJobID   = uint(ctx.QueryAsInt("jobID"))
		qorJobName = ctx.R.FormValue("jobName")
		config     = actionConfigs[qorJobName]
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
		h.If(config.displayLog, jobActionLog(inst.Log)),
		h.If(inst.ProgressText != "",
			h.Div().Class("mb-3").Children(
				h.RawHTML(inst.ProgressText),
			),
		),
	)

	if inst.Status == JobStatusDone || inst.Status == JobStatusException {
		er.VarsScript = "vars.jobActionProgressingInterval = 0; vars.jobActionFinshed=true;"
	} else {
		er.VarsScript = "vars.jobActionProgressingInterval = 2000;"
	}
	return er, nil
}

func jobActionLog(log string) h.HTMLComponent {
	var logLines []h.HTMLComponent
	logs := strings.Split(log, "\n")
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
