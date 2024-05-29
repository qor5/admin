package publish

import (
	"cmp"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	v "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	PortalSchedulePublishDialog = "publish_PortalSchedulePublishDialog"
	PortalPublishCustomDialog   = "publish_PortalPublishCustomDialog"

	VarCurrentDisplayID = "vars.publish_VarCurrentDisplayID"
)

func duplicateVersionAction(db *gorm.DB, mb *presets.ModelBuilder, _ *Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		toObj := mb.NewModel()
		slugger := toObj.(presets.SlugDecoder)
		paramID := ctx.Param(presets.ParamID)
		currentVersionName := slugger.PrimaryColumnValuesBySlug(paramID)["version"]
		me := mb.Editing()

		fromObj := mb.NewModel()
		// TODO: use fetcher?
		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), paramID).First(fromObj).Error; err != nil {
			return
		}
		if err = utils.SetPrimaryKeys(fromObj, toObj, db, paramID); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		if vErr := me.SetObjectFields(fromObj, toObj, &presets.FieldContext{
			ModelInfo: mb.Info(),
		}, false, presets.ContextModifiedIndexesBuilder(ctx).FromHidden(ctx.R), ctx); vErr.HaveErrors() {
			presets.ShowMessage(&r, vErr.Error(), "error")
			return
		}

		if err = reflectutils.Set(toObj, "Version.ParentVersion", currentVersionName); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		if me.Validator != nil {
			if vErr := me.Validator(toObj, ctx); vErr.HaveErrors() {
				presets.ShowMessage(&r, vErr.Error(), "error")
				return
			}
		}

		if err = me.Saver(toObj, paramID, ctx); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		se := toObj.(presets.SlugEncoder)
		id := se.PrimarySlug()

		if !mb.HasDetailing() {
			// close dialog and open editing
			web.AppendRunScripts(&r,
				presets.CloseListingDialogVarScript,
				web.Plaid().EventFunc(actions.Edit).Query(presets.ParamID, id).Go(),
			)
			return
		}
		if !mb.Detailing().GetDrawer() {
			// open detailing without drawer
			// jump URL to support referer
			r.PushState = web.Location(nil).URL(mb.Info().DetailingHref(id))
			return
		}
		// close dialog and open detailingDrawer
		web.AppendRunScripts(&r,
			presets.CloseListingDialogVarScript,
			presets.CloseRightDrawerVarScript,
			web.Plaid().EventFunc(actions.DetailingDrawer).Query(presets.ParamID, id).Go(),
		)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")

		r.RunScript = web.Plaid().ThenScript(r.RunScript).Go()
		return
	}
}

func selectVersion(pm *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue("select_id")

		if !pm.HasDetailing() {
			// close dialog and open editing
			web.AppendRunScripts(&r,
				presets.CloseListingDialogVarScript,
				web.Plaid().EventFunc(actions.Edit).Query(presets.ParamID, id).Go(),
			)
			return
		}
		if !pm.Detailing().GetDrawer() {
			// open detailing without drawer
			// jump URL to support referer
			r.PushState = web.Location(nil).URL(pm.Info().DetailingHref(id))
			return
		}
		// close dialog and open detailingDrawer
		web.AppendRunScripts(&r,
			presets.CloseListingDialogVarScript,
			fmt.Sprintf("if (!!%s && %s != %q) { %s }", VarCurrentDisplayID, VarCurrentDisplayID, id, presets.CloseRightDrawerVarScript+";"+web.Plaid().EventFunc(actions.DetailingDrawer).Query(presets.ParamID, id).Go()),
		)
		return
	}
}

func renameVersionDialog(_ *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		versionName := ctx.R.FormValue("version_name")
		okAction := web.Plaid().
			URL(ctx.R.URL.Path).
			EventFunc(eventRenameVersion).
			Queries(ctx.Queries()).Go()

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
								On("click", "locals.renameVersionDialog = false"),

							v.VBtn("OK").
								Color("primary").
								Variant(v.VariantFlat).
								Theme(v.ThemeDark).
								Attr("@click", "locals.renameVersionDialog = false; "+okAction),
						),
					),
				).MaxWidth("420px").Attr("v-model", "locals.renameVersionDialog"),
			).Init("{renameVersionDialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func renameVersion(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		if mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
			presets.ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
			return
		}

		id := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}

		name := ctx.R.FormValue("VersionName")
		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, id, ctx); err != nil {
			return
		}

		listQueries := ctx.Queries().Get(presets.ParamListingQueries)
		r.RunScript = web.Plaid().URL(ctx.R.URL.Path).StringQuery(listQueries).EventFunc(actions.UpdateListingDialog).Go()
		return
	}
}

func deleteVersionDialog(_ *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		versionName := ctx.R.FormValue("version_name")

		utilMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)
		// msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DeleteConfirmPortalName,
			Body: utils.DeleteDialog(
				// TODO i18
				fmt.Sprintf("Are you sure you want to delete %s?", versionName),
				"locals.deleteConfirmation = false;"+web.Plaid().
					URL(ctx.R.URL.Path).
					EventFunc(eventDeleteVersion).
					Queries(ctx.Queries()).Go(),
				utilMsgr),
		})
		return
	}
}

func findVersionItems(db *gorm.DB, mb *presets.ModelBuilder, paramId string) (list interface{}, err error) {
	list = mb.NewModelSlice()
	primaryKeys, err := utils.GetPrimaryKeys(mb.NewModel(), db)
	if err != nil {
		return
	}
	err = utils.PrimarySluggerWhere(db.Session(&gorm.Session{NewDB: true}).Select(strings.Join(primaryKeys, ",")), mb.NewModel(), paramId, "version").
		Order("version DESC").
		Find(list).
		Error
	return list, err
}

func deleteVersion(mb *presets.ModelBuilder, pm *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		deleteID := ctx.R.FormValue(presets.ParamID)
		if len(deleteID) <= 0 {
			presets.ShowMessage(&r, "no delete_id", "warning")
			return
		}

		if mb.Info().Verifier().Do(presets.PermDelete).WithReq(ctx.R).IsAllowed() != nil {
			presets.ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
			return
		}

		if err = mb.Editing().Deleter(mb.NewModel(), deleteID, ctx); err != nil {
			presets.ShowMessage(&r, err.Error(), "warning")
			return
		}

		currentDisplayID := ctx.R.FormValue("current_display_id")
		if deleteID == currentDisplayID {
			items, _ := findVersionItems(db, mb, deleteID)
			rv := reflect.ValueOf(items).Elem()
			if rv.Len() == 0 {
				r.PushState = web.Location(nil).URL(pm.Info().ListingHref())
				return
			}

			deletedVersion := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(deleteID)["version"]

			version := rv.Index(0).Interface()
			if rv.Len() > 1 {
				hasOlderVersion := false
				for i := 0; i < rv.Len(); i++ {
					v := rv.Index(i).Interface()
					if v.(VersionInterface).GetVersion() < deletedVersion {
						hasOlderVersion = true
						version = v
						break
					}
				}
				if !hasOlderVersion {
					version = rv.Index(rv.Len() - 1)
				}
			}

			currentDisplayID = version.(presets.SlugEncoder).PrimarySlug()
			web.AppendRunScripts(&r, fmt.Sprintf("%s = %q", VarCurrentDisplayID, currentDisplayID))

			if !pm.HasDetailing() {
				web.AppendRunScripts(&r,
					web.Plaid().EventFunc(actions.Edit).Query(presets.ParamID, currentDisplayID).Go(),
				)
			} else {
				if !pm.Detailing().GetDrawer() {
					r.PushState = web.Location(nil).URL(pm.Info().DetailingHref(currentDisplayID))
					return
				}
				web.AppendRunScripts(&r,
					presets.CloseRightDrawerVarScript,
					web.Plaid().EventFunc(actions.DetailingDrawer).Query(presets.ParamID, currentDisplayID).Go(),
				)
			}
		}

		listQuery, err := url.ParseQuery(ctx.Queries().Get(presets.ParamListingQueries))
		if err != nil {
			return
		}
		if deleteID == cmp.Or(listQuery.Get("select_id"), listQuery.Get("f_select_id")) {
			listQuery.Set("select_id", currentDisplayID)
		}

		web.AppendRunScripts(&r,
			web.Plaid().URL(ctx.R.URL.Path).Queries(listQuery).EventFunc(actions.UpdateListingDialog).Go(),
			// web.Plaid().URL(pm.Info().ListingHref()).EventFunc(actions.ReloadList).Go(), // TODO: dont know how to reload res list now
		)
		return
	}
}
