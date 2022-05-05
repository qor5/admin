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

type JobActionConfig struct {
	Name   string
	Hander JobHandler //optional

	Params       interface{}           // optional
	ModelBuilder *presets.ModelBuilder // Params ModelBuilder optional
	DisplayLog   bool                  // optional
}

func (b *Builder) JobAction(cfg *JobActionConfig) presets.ComponentFunc {
	if cfg.Name == "" {
		panic("job name is required")
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

	if cfg.ModelBuilder != nil {
		if cfg.Params == nil {
			cfg.Params = cfg.ModelBuilder.NewModel()
		}
		jb.b.mb = cfg.ModelBuilder
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
				Query("displayLog", cfg.DisplayLog).
				Query("hasParams", cfg.Params != nil).
				Go())
	}
}

func (b *Builder) eventJobActionCreate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName    = ctx.R.FormValue("jobName")
		displayLog = ctx.R.FormValue("displayLog")
		qorJob     = &QorJob{Job: jobName}
	)

	if ctx.R.FormValue("hasParams") == "true" {
		jb := b.mustGetJobBuilder(jobName)
		if args, err := jb.unmarshalForm(ctx); !err.HaveErrors() {
			qorJob.args = args
		}
	}

	job, err := b.createJob(ctx, qorJob)
	if err != nil {
		return
	}

	r.VarsScript = web.Plaid().
		URL(b.mb.Info().ListingHref()).
		EventFunc(JobActionResponse).
		Query(presets.ParamID, fmt.Sprint(job.ID)).
		Query("displayLog", displayLog).
		Query("jobID", fmt.Sprintf("%d", job.ID)).
		Query("jobName", job.Job).
		Go()
	return
}

func (b *Builder) eventJobActionCreateWithParams(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName    = ctx.R.FormValue("jobName")
		displayLog = ctx.R.FormValue("displayLog")
		hasParams  = ctx.R.FormValue("hasParams")
		msgr       = presets.MustGetMessages(ctx.R)
	)

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
								Query("displayLog", displayLog).
								Query("hasParams", hasParams).
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
		jobName    = ctx.R.FormValue("jobName")
		jobID      = ctx.R.FormValue("jobID")
		displayLog = ctx.R.FormValue("displayLog")
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
									Query("jobName", jobName).
									Query("displayLog", displayLog),
							).AutoReloadInterval("vars.jobActionProgressingInterval"),
						).Attr(web.InitContextVars, "{jobActionProgressingInterval: 2000, jobActionFinshed: false}"),
					),
				)).
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
		displayLog = ctx.R.FormValue("displayLog")
	)

	inst, err := getModelQorJobInstance(b.db, qorJobID)
	if err != nil {
		return er, err
	}

	er.Body = h.Div(
		h.Div(vuetify.VProgressLinear(
			h.Strong(fmt.Sprintf("%d%%", inst.Progress)),
		).Value(int(inst.Progress)).Height(20)).Class("mb-5"),
		h.If(displayLog == `true`, jobActionLog(inst.Log)),
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
