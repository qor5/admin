package publish

import (
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	"gorm.io/gorm"
)

func publishAction(_ *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.Param(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		reqCtx := publisher.WithContextValues(ctx.R.Context())
		err = publisher.Publish(obj, reqCtx)
		if err != nil {
			return
		}
		if ab != nil {
			if _, exist := ab.GetModelBuilder(obj); exist {
				ab.AddCustomizedRecord(actionName, false, ctx.R.Context(), obj)
			}
		}

		if script := ctx.R.FormValue(ParamScriptAfterPublish); script != "" {
			web.AppendRunScripts(&r, script)
		} else {
			presets.ShowMessage(&r, "success", "")
			r.Reload = true
		}
		return
	}
}

func unpublishAction(_ *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.Param(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		reqCtx := publisher.WithContextValues(ctx.R.Context())
		err = publisher.UnPublish(obj, reqCtx)
		if err != nil {
			return
		}
		if ab != nil {
			if _, exist := ab.GetModelBuilder(obj); exist {
				ab.AddCustomizedRecord(actionName, false, ctx.R.Context(), obj)
			}
		}

		presets.ShowMessage(&r, "success", "")
		r.Reload = true
		return
	}
}
