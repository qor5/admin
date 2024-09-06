package publish

import (
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	"gorm.io/gorm"
)

func publishAction(_ *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.Param(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, PermPublish) {
			return r, perm.PermissionDenied
		}

		reqCtx := publisher.WithContextValues(ctx.R.Context())
		err = publisher.Publish(reqCtx, obj)
		if err != nil {
			return
		}
		if publisher.ab != nil {
			if amb, exist := publisher.ab.GetModelBuilder(mb); exist {
				amb.Log(ctx.R.Context(), actionName, obj, nil)
			}
		}

		if script := ctx.R.FormValue(ParamScriptAfterPublish); script != "" {
			web.AppendRunScripts(&r, script)
		} else {
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
			web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).
				ThenScript(presets.ShowSnackbarScript(msgr.SuccessfullyPublish, v.ColorSuccess)).
				Go(),
			)
		}
		return
	}
}

func unpublishAction(_ *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.Param(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, PermUnpublish) {
			return r, perm.PermissionDenied
		}

		reqCtx := publisher.WithContextValues(ctx.R.Context())
		err = publisher.UnPublish(reqCtx, obj)
		if err != nil {
			return
		}
		if publisher.ab != nil {
			if amb, exist := publisher.ab.GetModelBuilder(mb); exist {
				amb.Log(ctx.R.Context(), actionName, obj, nil)
			}
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).
			ThenScript(presets.ShowSnackbarScript(msgr.SuccessfullyUnPublish, v.ColorSuccess)).
			Go(),
		)
		return
	}
}
