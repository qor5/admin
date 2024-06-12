package l10n

import (
	"context"
	"reflect"
	"slices"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
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
	mb.RegisterEventFunc(Localize, localizeToConfirmation(db, lb, mb))
	mb.RegisterEventFunc(DoLocalize, doLocalizeTo(db, mb, lb, ab))
}

type SelectLocale struct {
	Label string
	Code  string
}

func localizeToConfirmation(db *gorm.DB, lb *Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		presetsMsgr := presets.MustGetMessages(ctx.R)

		paramID := ctx.Param(presets.ParamID)
		cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(paramID)
		id := cs["id"]

		fromLocale := lb.GetCorrectLocaleCode(ctx.R)

		obj := mb.NewModelSlice()
		err = db.Distinct("locale_code").Where("id = ? AND locale_code <> ?", id, fromLocale).Find(obj).Error
		if err != nil {
			return
		}
		vo := reflect.ValueOf(obj).Elem()
		var existLocales []string
		for i := 0; i < vo.Len(); i++ {
			existLocales = append(existLocales, vo.Index(i).Elem().FieldByName("LocaleCode").String())
		}
		toLocales := lb.GetSupportLocaleCodesFromRequest(ctx.R)
		var selectLocales []SelectLocale
		for _, locale := range toLocales {
			if locale == fromLocale {
				continue
			}
			if !slices.Contains(existLocales, locale) || vo.Len() == 0 {
				selectLocales = append(selectLocales, SelectLocale{Label: MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)), Code: locale})
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DialogPortalName,
			Body: VDialog(
				VCard(
					VCardTitle(h.Text(MustGetTranslation(ctx.R, "Localize"))),

					VCardText(
						h.Div(
							h.Label(MustGetTranslation(ctx.R, "LocalizeFrom")).Class("v-label v-field-label v-field-label--floating"),
							// h.Br(),
							h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale))),
						).Class("v-field v-field--active v-field--appended v-field--dirty v-field--variant-underlined v-theme--light v-locale--is-ltr "),
						VSelect().
							Attr(web.VField("localize_to", nil)...).
							Variant(FieldVariantUnderlined).
							Label(MustGetTranslation(ctx.R, "LocalizeTo")).
							Multiple(true).Chips(true).
							Items(selectLocales).
							ItemTitle("Label").
							ItemValue("Code"),
					).Attr("style", "height: 200px;"),

					VCardActions(
						VSpacer(),
						VBtn(presetsMsgr.Cancel).
							Variant(VariantFlat).
							Class("ml-2").
							On("click", "vars.localizeConfirmation = false"),

						VBtn(presetsMsgr.OK).
							Color("primary").
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", web.Plaid().
								EventFunc(DoLocalize).
								Query(presets.ParamID, paramID).
								Query("localize_from", fromLocale).
								URL(ctx.R.URL.Path).
								Go()),
					),
				),
			).MaxWidth("600px").
				Attr("v-model", "vars.localizeConfirmation").
				Attr(web.VAssign("vars", `{localizeConfirmation: false}`)...),
		})

		r.RunScript = "setTimeout(function(){ vars.localizeConfirmation = true }, 100)"
		return
	}
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
			if _, ok := ab.GetModelBuilder(fromObj); !ok {
				return
			}
			if len(toObjs) > 0 {
				if err = ab.AddCustomizedRecord(LocalizeFrom, false, ctx.R.Context(), fromObj); err != nil {
					return
				}
				for _, toObj := range toObjs {
					if err = ab.AddCustomizedRecord(LocalizeTo, false, ctx.R.Context(), toObj); err != nil {
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

			if err = me.Saver(toObj, toParamID, ctx); err != nil {
				return
			}
			toObjs = append(toObjs, toObj)
		}

		presets.ShowMessage(&r, MustGetTranslation(ctx.R, "SuccessfullyLocalized"), "")

		// refresh current page
		r.Reload = true
		return
	}
}
