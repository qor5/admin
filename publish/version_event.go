package publish

import (
	"cmp"
	"errors"
	"net/url"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	PortalSchedulePublishDialog = "publish_PortalSchedulePublishDialog"
	PortalPublishCustomDialog   = "publish_PortalPublishCustomDialog"

	VarCurrentDisplaySlug = "vars.publish_VarCurrentDisplaySlug"
)

func duplicateVersionAction(mb *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		if mb.Info().Verifier().Do(presets.PermCreate).WithReq(ctx.R).IsAllowed() != nil {
			return r, perm.PermissionDenied
		}

		slug := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), slug).First(obj).Error; err != nil {
			return
		}

		ver := EmbedVersion(obj)
		if ver == nil {
			err = errInvalidObject
			return
		}

		oldVersion := ver.Version
		newVersion, err := ver.CreateVersion(db, slug, mb.NewModel())
		if err != nil {
			return
		}
		*ver = Version{newVersion, newVersion, oldVersion}

		st := EmbedStatus(obj)
		if st != nil {
			*st = Status{Status: StatusDraft}
		}

		sched := EmbedSchedule(obj)
		if sched != nil {
			*sched = Schedule{}
		}

		_, err = reflectutils.Get(obj, "CreatedAt")
		if err == nil {
			if err = reflectutils.Set(obj, "CreatedAt", time.Time{}); err != nil {
				return
			}
		}
		_, err = reflectutils.Get(obj, "UpdatedAt")
		if err == nil {
			if err = reflectutils.Set(obj, "UpdatedAt", time.Time{}); err != nil {
				return
			}
		}

		slug = obj.(presets.SlugEncoder).PrimarySlug()
		if err = mb.Editing().Creating().Saver(obj, slug, ctx); err != nil {
			presets.ShowMessage(&r, err.Error(), "error")
			return
		}

		web.AppendRunScripts(&r,
			"locals.commonConfirmDialog = false",
			Notify(PayloadVersionSelected{
				ToPayloadItem(obj, mb.Info().Label()),
			}),
		)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")
		return
	}
}

func selectVersion(mb *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		slug := cmp.Or(ctx.R.FormValue("select_id"), ctx.R.FormValue("f_select_id"))

		obj := mb.NewModel()
		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), slug).First(obj).Error; err != nil {
			return
		}

		web.AppendRunScripts(&r,
			presets.CloseListingDialogVarScript,
			Notify(PayloadVersionSelected{
				ToPayloadItem(obj, mb.Info().Label()),
			}),
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
								Attr("@click", okAction),
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

		web.AppendRunScripts(&r,
			"locals.renameVersionDialog = false",
			Notify(PayloadItemUpdated{
				ToPayloadItem(obj, mb.Info().Label()),
			}),
		)

		listQueries := ctx.Queries().Get(presets.ParamListingQueries)
		r.RunScript = web.Plaid().URL(ctx.R.URL.Path).StringQuery(listQueries).EventFunc(actions.UpdateListingDialog).Go()
		return
	}
}

func deleteVersionDialog(_ *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		versionName := ctx.R.FormValue("version_name")

		utilMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DeleteConfirmPortalName,
			Body: utils.DeleteDialog(
				msgr.DeleteVersionConfirmationText(versionName),
				web.Plaid().
					URL(ctx.R.URL.Path).
					EventFunc(eventDeleteVersion).
					Queries(ctx.Queries()).Go(),
				utilMsgr),
		})
		return
	}
}

const paramCurrentDisplaySlug = "current_display_id"

func deleteVersion(mb *presets.ModelBuilder, pm *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return wrapEventFuncWithShowError(func(ctx *web.EventContext) (web.EventResponse, error) {
		var r web.EventResponse

		if mb.Info().Verifier().Do(presets.PermDelete).WithReq(ctx.R).IsAllowed() != nil {
			return r, perm.PermissionDenied
		}

		slug := ctx.R.FormValue(presets.ParamID)
		if len(slug) <= 0 {
			return r, errors.New("no delete_id")
		}

		if err := mb.Editing().Deleter(mb.NewModel(), slug, ctx); err != nil {
			return r, err
		}

		web.AppendRunScripts(&r,
			"locals.deleteConfirmation = false",
			Notify(PayloadItemDeleted{mb.Info().Label(), slug}),
		)

		currentDisplaySlug := ctx.R.FormValue(paramCurrentDisplaySlug)
		if slug == currentDisplaySlug {
			deletedVersion := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(slug)["version"]

			// find the older version first then find the max version
			version := mb.NewModel()
			db := utils.PrimarySluggerWhere(db, version, slug, "version").Order("version DESC").WithContext(ctx.R.Context())
			err := db.Where("version < ?", deletedVersion).First(version).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return r, err
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := db.First(version).Error
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return r, err
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					r.PushState = web.Location(nil).URL(pm.Info().ListingHref())
					return r, nil
				}
			}

			currentDisplaySlug = version.(presets.SlugEncoder).PrimarySlug()
			web.AppendRunScripts(&r,
				Notify(PayloadVersionSelected{
					ToPayloadItem(version, mb.Info().Label()),
				}),
			)
		}

		listQuery, err := url.ParseQuery(ctx.Queries().Get(presets.ParamListingQueries))
		if err != nil {
			return r, err
		}
		if slug == cmp.Or(listQuery.Get("select_id"), listQuery.Get("f_select_id")) {
			listQuery.Set("select_id", currentDisplaySlug)
		}

		web.AppendRunScripts(&r,
			web.Plaid().URL(ctx.R.URL.Path).Queries(listQuery).EventFunc(actions.UpdateListingDialog).Go(),
			// web.Plaid().EventFunc(actions.ReloadList).Go(), // TODO: This will reload the dialog list, I don't know how to reload the main list yet.
		)
		return r, nil
	})
}
