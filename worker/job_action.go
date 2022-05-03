package worker

import (
	"fmt"

	"github.com/goplaid/web"
	"github.com/goplaid/x/presets"
	"github.com/goplaid/x/presets/actions"
	"github.com/goplaid/x/vuetify"
	h "github.com/theplant/htmlgo"
)

const (
	JobActionCreate     = "worker_job_action_create"
	JobActionShowParams = "worker_job_action_show_params"
)

type JobActionConfig struct {
	Name   string
	Hander JobHandler //optional

	Params       interface{}           // optional
	ModelBuilder *presets.ModelBuilder // Params ModelBuilder optional
	HideLog      bool                  // optional
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
			eventName = JobActionShowParams
		}

		return vuetify.VBtn(cfg.Name).
			Color("primary").
			Depressed(true).
			Class("ml-2").
			Attr("@click", web.Plaid().
				URL(b.mb.Info().ListingHref()).
				EventFunc(eventName).
				Query("jobName", cfg.Name).
				Query("hideLog", cfg.HideLog).
				Query("hasParams", cfg.Params != nil).
				Go())
	}
}

func (b *Builder) eventJobActionCreate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName = ctx.R.FormValue("jobName")
		hideLog = ctx.R.FormValue("hideLog")
		qorJob  = &QorJob{Job: jobName}
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
		EventFunc(actions.Show).
		Query(presets.ParamOverlay, actions.Dialog).
		Query(presets.ParamID, fmt.Sprint(job.ID)).
		Query("hideLog", hideLog).
		Go()
	return
}

func (b *Builder) eventJobActionShowParams(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		jobName   = ctx.R.FormValue("jobName")
		hideLog   = ctx.R.FormValue("hideLog")
		hasParams = ctx.R.FormValue("hasParams")
		msgr      = presets.MustGetMessages(ctx.R)
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: "presets_DialogPortalName",
		Body: web.Scope(
			vuetify.VDialog(
				vuetify.VCard(
					vuetify.VCardTitle(h.Text(jobName)),
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
								Query("hideLog", hideLog).
								Query("hasParams", hasParams).
								Go()),
					),
				)).
				Attr("v-model", "vars.presetsDialog").
				Width("600"),
		).VSlot("{ plaidForm }"),
	})
	r.VarsScript = "setTimeout(function(){vars.presetsDialog = true; }, 100)"
	return
}
