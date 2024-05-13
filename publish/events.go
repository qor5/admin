package publish

import (
	"fmt"
	"reflect"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	v "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
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

	schedulePublishDialogEventV2 = "publish_SchedulePublishDialogEvent_v2"
	schedulePublishEventV2       = "publish_SchedulePublishEvent_v2"
	selectVersionEventV2         = "publish_selectVersionEvent_v2"
	renameVersionDialogEventV2   = "publish_renameVersionDialogEvent_v2"
	renameVersionEventV2         = "publish_renameVersionEvent_v2"
	deleteVersionDialogEventV2   = "publish_deleteVersionDialogEvent_v2"

	ActivityPublish   = "Publish"
	ActivityRepublish = "Republish"
	ActivityUnPublish = "UnPublish"

	ParamScriptAfterPublish = "publish_param_script_after_publish"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder) {
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

func publishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)

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

func unpublishAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)

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

func renameVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
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

func selectVersionsAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder, ab *activity.Builder, actionName string) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

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

func afterDeleteVersionAction(db *gorm.DB, mb *presets.ModelBuilder, publisher *Builder) web.EventFunc {
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
					if v.(VersionInterface).GetVersion() < deletedVersion {
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

func schedulePublishDialogV2(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		s, ok := obj.(ScheduleInterface)
		if !ok {
			return
		}

		var start, end string
		if s.GetScheduledStartAt() != nil {
			start = s.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if s.GetScheduledEndAt() != nil {
			end = s.GetScheduledEndAt().Format("2006-01-02 15:04")
		}

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		cmsgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		updateBtn := v.VBtn(cmsgr.Update).
			Color("primary").
			Attr(":disabled", "isFetching").
			Attr(":loading", "isFetching").
			Attr("@click", web.Plaid().
				EventFunc(schedulePublishEventV2).
				// Queries(queries).
				Query(presets.ParamID, paramID).
				Query(presets.ParamOverlay, actions.Dialog).
				URL(mb.Info().ListingHref()).
				Go())

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: SchedulePublishDialogPortalName,
			Body: web.Scope(
				v.VDialog(
					v.VCard(
						v.VCardTitle(h.Text("Schedule Publish Time")),
						v.VCardText(
							v.VRow(
								v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledStartAt", start)...).Label(msgr.ScheduledStartAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledStartAt" value="%s" v-field-name='"ScheduledStartAt"'> </vx-datetimepicker>`, start)),
								).Cols(6),
								v.VCol(
									vx.VXDateTimePicker().Attr(web.VField("ScheduledEndAt", end)...).Label(msgr.ScheduledEndAt).
										TimePickerProps(vx.TimePickerProps{Format: "24hr", Scrollable: true}).
										ClearText(msgr.DateTimePickerClearText).OkText(msgr.DateTimePickerOkText),
									// h.RawHTML(fmt.Sprintf(`<vx-datetimepicker label="ScheduledEndAt" value="%s" v-field-name='"ScheduledEndAt"'> </vx-datetimepicker>`, end)),
								).Cols(6),
							),
						),
						v.VCardActions(
							v.VSpacer(),
							updateBtn,
						),
					),
				).MaxWidth("480px").
					Attr("v-model", "locals.schedulePublishDialogV2"),
			).Init("{schedulePublishDialogV2:true}").VSlot("{locals}"),
		})
		return
	}
}

func schedulePublishV2(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}
		err = ScheduleEditSetterFunc(obj, nil, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		err = mb.Editing().Saver(obj, paramID, ctx)
		if err != nil {
			mb.Editing().UpdateOverlayContent(ctx, &r, obj, "", err)
			return
		}
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: SchedulePublishDialogPortalName,
			Body: nil,
		})
		return
	}
}

func renameVersionDialogV2(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("rename_id")
		versionName := ctx.R.FormValue("version_name")
		okAction := web.Plaid().
			URL(mb.Info().ListingHref()).
			EventFunc(renameVersionEventV2).
			Queries(ctx.Queries()).
			Query("rename_id", id).Go()

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DialogPortalName,
			Body: web.Scope(
				v.VDialog(
					v.VCard(
						v.VCardTitle(h.Text("Version")),
						v.VCardText(
							v.VTextField().Attr(web.VField("VersionName", versionName)...).Variant(v.FieldVariantUnderlined),
						),
						v.VCardActions(
							v.VSpacer(),
							v.VBtn("Cancel").
								Variant(v.VariantFlat).
								Class("ml-2").
								On("click", "locals.renameVersionDialogV2 = false"),

							v.VBtn("OK").
								Color("primary").
								Variant(v.VariantFlat).
								Theme(v.ThemeDark).
								Attr("@click", "locals.renameVersionDialogV2 = false; "+okAction),
						),
					),
				).MaxWidth("420px").Attr("v-model", "locals.renameVersionDialogV2"),
			).Init("{renameVersionDialogV2:true}").VSlot("{locals}"),
		})
		return
	}
}

func renameVersionV2(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		paramID := ctx.R.FormValue("rename_id")
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, paramID, ctx)
		if err != nil {
			return
		}

		name := ctx.R.FormValue("VersionName")
		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, paramID, ctx); err != nil {
			return
		}
		qs := ctx.Queries()
		delete(qs, "version_name")
		delete(qs, "rename_id")

		r.RunScript = web.Plaid().URL(ctx.R.RequestURI).Queries(qs).EventFunc(actions.UpdateListingDialog).Go()
		return
	}
}

func deleteVersionDialogV2(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("delete_id")
		versionName := ctx.R.FormValue("version_name")

		utilMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)
		// msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DeleteConfirmPortalName,
			Body: utils.DeleteDialog(
				// TODO i18
				fmt.Sprintf("Are you sure you want to delete %s?", versionName),
				web.Plaid().
					URL(mb.Info().ListingHref()).
					EventFunc(actions.DoDelete).
					Queries(ctx.Queries()).
					Query(presets.ParamInDialog, "true").
					Query(presets.ParamID, id).Go(),
				utilMsgr),
		})
		return
	}
}
