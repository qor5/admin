package views

import (
	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/publish"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

const (
	publishEvent            = "publish_PublishEvent"
	republishEvent          = "publish_republishEvent"
	unpublishEvent          = "publish_UnpublishEvent"
	switchVersionEvent      = "publish_SwitchVersionEvent"
	SaveNewVersionEvent     = "publish_SaveNewVersionEvent"
	saveNameVersionEvent    = "publish_SaveNameVersionEvent"
	renameVersionEvent      = "publish_RenameVersionEvent"
	selectVersionsEvent     = "publish_SelectVersionsEvent"
	afterDeleteVersionEvent = "publish_AfterDeleteVersionEvent"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder) {
	mb.RegisterEventFunc(publishEvent, publishAction(db, mb, publisher, ab, ActivityPublish))
	mb.RegisterEventFunc(republishEvent, publishAction(db, mb, publisher, ab, ActivityRepublish))
	mb.RegisterEventFunc(unpublishEvent, unpublishAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(switchVersionEvent, switchVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(SaveNewVersionEvent, saveNewVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(renameVersionEvent, renameVersionAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(selectVersionsEvent, selectVersionsAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(afterDeleteVersionEvent, afterDeleteVersionAction(db, mb, publisher))

}

func publishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}
		err = publisher.Publish(obj)
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

func unpublishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}

		err = publisher.UnPublish(obj)
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

func renameVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("id")

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}

		name := ctx.R.FormValue("name")

		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, ctx.R.FormValue("id"), ctx); err != nil {
			return
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyRename, "")
		return
	}
}

func selectVersionsAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		)

		table, _, err := versionListTable(db, mb, msgr, ctx)
		if err != nil {
			return
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: "versions-list",
			Body: table,
		})
		return
	}
}

func afterDeleteVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		qs := ctx.Queries()
		currentSelectedID := qs.Get("current_selected_id")

		web.AppendVarsScripts(&r,
			web.Plaid().
				EventFunc(switchVersionEvent).
				Query("id", currentSelectedID).
				Query("selected", qs.Get("selected")).
				Query("page", qs.Get("page")).
				Query("no_msg", true).
				Go(),
			web.Plaid().
				EventFunc(actions.ReloadList).
				Go(),
		)

		return
	}
}
