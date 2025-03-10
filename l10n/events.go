package l10n

import (
	"context"
	"reflect"

	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
)

const (
	Localize   = "l10n_LocalizeEvent"
	DoLocalize = "l10n_DoLocalizeEvent"

	FromID      = "l10n_DoLocalize_FromID"
	FromVersion = "l10n_DoLocalize_FromVersion"
	FromLocale  = "l10n_DoLocalize_FromLocale"

	LocalizeFrom = "Localize From"
	LocalizeTo   = "Localize To"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, lb *Builder, ab *activity.Builder) {
	mb.RegisterEventFunc(DoLocalize, doLocalizeTo(db, mb, lb, ab)) // Execute localization
}

type SelectLocale struct {
	// label is a descriptive name used to display the name of a language or region to the user.
	// It is usually a user-friendly string, such as "English", "French", "Chinese", etc. In the user interface,
	// it is used to display selectable localization options.
	Label string

	// code is an identifier that represents a language or locale.
	// It is usually a string that represents a specific language or region code, such as "en" (English), "fr" (French), "zh" (Chinese), etc. In code,
	// it is used to distinguish between different localized versions.
	Code string
}

func doLocalizeTo(db *gorm.DB, mb *presets.ModelBuilder, lb *Builder, ab *activity.Builder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		fromParamID := ctx.Param(presets.ParamID)
		cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(fromParamID)
		fromID := cs["id"]
		fromVersion := cs["version"]
		fromLocale := cs["locale_code"]
		to := make(map[string]struct{})
		for _, v := range ctx.R.Form["localize_to"] {
			for _, lc := range lb.GetSupportLocaleCodes() {
				if v == lc {
					to[v] = struct{}{}
					break
				}
			}
		}
		if len(to) == 0 {
			web.AppendRunScripts(&r, "vars.localizeConfirmation = false")
			return
		}

		fromObj := mb.NewModel()

		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), fromParamID).First(fromObj).Error; err != nil {
			return
		}

		var toObjs []interface{}
		defer func(fromObj interface{}) {
			if ab == nil {
				return
			}
			amb, ok := ab.GetModelBuilder(mb)
			if !ok {
				return
			}
			if len(toObjs) > 0 {
				_, err = amb.Log(ctx.R.Context(), LocalizeFrom, fromObj, nil)
				if err != nil {
					return
				}
				for _, toObj := range toObjs {
					_, err = amb.Log(ctx.R.Context(), LocalizeTo, toObj, nil)
					if err != nil {
						return
					}
				}
			}
		}(reflect.Indirect(reflect.ValueOf(fromObj)).Interface())
		me := mb.Editing()

		for toLocale := range to {
			toObj := mb.NewModel()
			fakeToObj := fromObj
			if err = reflectutils.Set(fakeToObj, "LocaleCode", toLocale); err != nil {
				return
			}

			toParamID := fakeToObj.(presets.SlugEncoder).PrimarySlug()
			if err = utils.SetPrimaryKeys(fromObj, toObj, db, toParamID); err != nil {
				return
			}

			me.SetObjectFields(fromObj, toObj, &presets.FieldContext{
				ModelInfo: mb.Info(),
			}, false, presets.ContextModifiedIndexesBuilder(ctx).FromHidden(ctx.R), ctx)

			if me.Validator != nil {
				if vErr := me.Validator(toObj, ctx); vErr.HaveErrors() {
					presets.ShowMessage(&r, vErr.Error(), "error")
					return
				}
			}
			newContext := context.WithValue(ctx.R.Context(), FromID, fromID)
			newContext = context.WithValue(newContext, FromVersion, fromVersion)
			newContext = context.WithValue(newContext, FromLocale, fromLocale)
			ctx.R = ctx.R.WithContext(newContext)

			if err = me.Saver(toObj, "", ctx); err != nil {
				return
			}
			toObjs = append(toObjs, toObj)
		}

		presets.ShowMessage(&r, MustGetTranslation(ctx.R, "SuccessfullyLocalized"), "")

		// refresh current page
		web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).Go())
		return
	}
}
