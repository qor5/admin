package views

import (
	"reflect"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

const (
	PublishEvent            = "publish_PublishEvent"
	RepublishEvent          = "publish_RepublishEvent"
	UnpublishEvent          = "publish_UnpublishEvent"
	switchVersionEvent      = "publish_SwitchVersionEvent"
	SaveNewVersionEvent     = "publish_SaveNewVersionEvent"
	DuplicateVersionEvent   = "publish_DuplicateVersionEvent"
	renameVersionEvent      = "publish_RenameVersionEvent"
	selectVersionsEvent     = "publish_SelectVersionsEvent"
	afterDeleteVersionEvent = "publish_AfterDeleteVersionEvent"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"

	ParamScriptAfterPublish = "publish_param_script_after_publish"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder) {
	mb.RegisterEventFunc(PublishEvent, publishAction(db, mb, publisher, ab, ActivityPublish))
	mb.RegisterEventFunc(RepublishEvent, publishAction(db, mb, publisher, ab, ActivityRepublish))
	mb.RegisterEventFunc(UnpublishEvent, unpublishAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(switchVersionEvent, switchVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(SaveNewVersionEvent, saveNewVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(DuplicateVersionEvent, duplicateVersionAction(db, mb, publisher))
	mb.RegisterEventFunc(renameVersionEvent, renameVersionAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(selectVersionsEvent, selectVersionsAction(db, mb, publisher, ab, ActivityUnPublish))
	mb.RegisterEventFunc(afterDeleteVersionEvent, afterDeleteVersionAction(db, mb, publisher))

}

func publishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		publisher.WithEventContext(ctx)
		err = publisher.Publish(obj)
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

func unpublishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *publish.Builder, ab *activity.ActivityBuilder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
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
		paramID := ctx.R.FormValue(presets.ParamID)

		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		name := ctx.R.FormValue("name")

		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, paramID, ctx); err != nil {
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
		cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(ctx.R.FormValue("id"))
		deletedVersion := cs["version"]
		deletedID := qs.Get("id")
		currentSelectedID := qs.Get("current_selected_id")
		// switching version is one of the following in order that exists:
		// 1. current selected version
		// 2. next(older) version
		// 3. prev(newer) version
		switchingVersion := currentSelectedID
		if deletedID == currentSelectedID {
			versions, _ := findVersionItems(db, mb, ctx, deletedID)
			vO := reflect.ValueOf(versions).Elem()
			if vO.Len() == 0 {
				r.Reload = true
				return
			}

			version := vO.Index(0).Interface()
			if vO.Len() > 1 {
				hasOlderVersion := false
				for i := 0; i < vO.Len(); i++ {
					v := vO.Index(i).Interface()
					if v.(publish.VersionInterface).GetVersion() < deletedVersion {
						hasOlderVersion = true
						version = v
						break
					}
				}
				if !hasOlderVersion {
					version = vO.Index(vO.Len() - 1)
				}
			}

			switchingVersion = version.(presets.SlugEncoder).PrimarySlug()
		}

		web.AppendRunScripts(&r,
			web.Plaid().
				EventFunc(switchVersionEvent).
				Query(presets.ParamID, switchingVersion).
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
