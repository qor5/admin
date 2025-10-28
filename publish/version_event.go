package publish

import (
	"errors"
	"time"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"
)

const (
	PortalSchedulePublishDialog = "publish_PortalSchedulePublishDialog"
	PortalPublishCustomDialog   = "publish_PortalPublishCustomDialog"

	paramVersionName = "version_name"
)

func duplicateVersionAction(mb *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		defer func() {
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				err = nil
			}
		}()

		slug := ctx.Param(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return
		}

		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, presets.PermUpdate, PermDuplicate) {
			return r, perm.PermissionDenied
		}

		version := EmbedVersion(obj)
		if version == nil {
			err = errInvalidObject
			return
		}

		oldVersion := version.Version
		newVersion, err := version.CreateVersion(db, slug, mb.NewModel())
		if err != nil {
			return
		}
		*version = Version{newVersion, newVersion, oldVersion}

		status := EmbedStatus(obj)
		if status != nil {
			*status = Status{Status: StatusDraft}
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
			return
		}

		web.AppendRunScripts(&r, "locals.commonConfirmDialog = false")
		r.Emit(mb.NotifModelsCreated(), presets.PayloadModelsCreated{
			Models: []any{obj},
		})
		r.Emit(NotifVersionSelected(mb), PayloadVersionSelected{Slug: slug})

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)
		presets.ShowMessage(&r, msgr.SuccessfullyCreated, "")
		return
	}
}

func renameVersionDialog(_ *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		utilMsgr := i18n.MustGetModuleMessages(ctx.R, utils.I18nUtilsKey, Messages_en_US).(*utils.Messages)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPublishKey, Messages_en_US).(*Messages)

		versionName := ctx.R.FormValue(paramVersionName)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DialogPortalName,
			Body: web.Scope(
				vx.VXDialog(
					vx.VXField().Attr(web.VField("VersionName", versionName)...).HideDetails(true),
				).Title(msgr.RenameVersion).
					CancelText(utilMsgr.Cancel).
					OkText(utilMsgr.OK).
					Attr("@click:ok", web.Plaid().
						URL(ctx.R.URL.Path).
						EventFunc(eventRenameVersion).
						Queries(ctx.Queries()).Go()).
					Attr("v-model", "locals.renameVersionDialog"),

				// vx.VXDialog(
				// 	v.VCard(
				// 		v.VCardTitle(h.Text(msgr.RenameVersion)),
				// 		v.VCardText(
				// 			v.VTextField().Attr(web.VField("VersionName", versionName)...).Variant(v.FieldVariantUnderlined),
				// 		),
				// 		v.VCardActions(
				// 			v.VSpacer(),
				// 			v.VBtn(utilMsgr.Cancel).
				// 				Variant(v.VariantFlat).
				// 				Class("ml-2").
				// 				On("click", "locals.renameVersionDialog = false"),

				// 			v.VBtn(utilMsgr.OK).
				// 				Color("primary").
				// 				Variant(v.VariantFlat).
				// 				Theme(v.ThemeDark).
				// 				Attr("@click", web.Plaid().
				// 					URL(ctx.R.URL.Path).
				// 					EventFunc(eventRenameVersion).
				// 					Queries(ctx.Queries()).Go(),
				// 				),
				// 		),
				// 	),
				// ).MaxWidth("420px").Attr("v-model", "locals.renameVersionDialog"),
			).Init("{renameVersionDialog:true}").VSlot("{locals}"),
		})
		return
	}
}

func renameVersion(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		id := ctx.R.FormValue(presets.ParamID)
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, id, ctx)
		if err != nil {
			return
		}

		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, presets.PermUpdate) {
			return r, perm.PermissionDenied
		}

		name := ctx.R.FormValue("VersionName")
		if err = reflectutils.Set(obj, "Version.VersionName", name); err != nil {
			return
		}

		if err = mb.Editing().Saver(obj, id, ctx); err != nil {
			return
		}

		web.AppendRunScripts(&r, "locals.renameVersionDialog = false")
		r.Emit(mb.NotifModelsUpdated(), presets.PayloadModelsUpdated{
			Ids:    []string{id},
			Models: map[string]any{id: obj},
		})
		return
	}
}

func deleteVersionDialog(_ *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		versionName := ctx.R.FormValue(paramVersionName)

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

func deleteVersion(mb *presets.ModelBuilder, db *gorm.DB) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		defer func() {
			if err != nil {
				presets.ShowMessage(&r, err.Error(), "error")
				err = nil
			}
		}()

		slug := ctx.R.FormValue(presets.ParamID)
		if slug == "" {
			return r, errors.New("no delete_id")
		}
		obj := mb.NewModel()
		obj, err = mb.Editing().Fetcher(obj, slug, ctx)
		if err != nil {
			return
		}

		if DeniedDo(mb.Info().Verifier(), obj, ctx.R, presets.PermDelete) {
			return r, perm.PermissionDenied
		}

		if err := mb.Editing().Deleter(mb.NewModel(), slug, ctx); err != nil {
			return r, err
		}

		// find the older version first then find the max version
		deletedVersion := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(slug)["version"]
		nextVersion := mb.NewModel()
		db := utils.PrimarySluggerWhere(db, nextVersion, slug, "version").Order("version DESC").WithContext(ctx.R.Context())
		err = db.Where("version < ?", deletedVersion).First(nextVersion).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return r, err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := db.First(nextVersion).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return r, err
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				nextVersion = nil
			}
		}

		web.AppendRunScripts(&r, "locals.deleteConfirmation = false")

		addon := PayloadModelsDeletedAddon{}
		if nextVersion != nil {
			addon.NextVersionSlug = nextVersion.(presets.SlugEncoder).PrimarySlug()
		}
		r.Emit(
			mb.NotifModelsDeleted(),
			presets.PayloadModelsDeleted{Ids: []string{slug}},
			addon,
		)
		return r, nil
	}
}
