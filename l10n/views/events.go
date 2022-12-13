package views

import (
	"reflect"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/utils"
	v "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	Localize   = "l10n_Localize"
	DoLocalize = "l10n_DoLocalize"
)

func registerEventFuncs(db *gorm.DB, mb *presets.ModelBuilder, lb *l10n.Builder, ab *activity.ActivityBuilder) {
	mb.RegisterEventFunc(Localize, localizeToConfirmation(db, lb, mb))
	mb.RegisterEventFunc(DoLocalize, doLocalizeTo(db, mb))
}

type SelectLocale struct {
	Label string
	Code  string
}

func localizeToConfirmation(db *gorm.DB, lb *l10n.Builder, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		presetsMsgr := presets.MustGetMessages(ctx.R)

		paramID := ctx.R.FormValue(presets.ParamID)
		cs := mb.NewModel().(presets.SlugDecoder).PrimaryColumnValuesBySlug(paramID)
		id := cs["id"]

		//todo get current locale
		fromLocale := lb.GetCorrectLocale(ctx.R)

		//todo search distinct locale_code except current locale
		var obj = mb.NewModelSlice()
		err = db.Distinct("locale_code").Where("id = ? AND locale_code <> ?", id, lb.GetLocaleCode(fromLocale)).Find(obj).Error
		if err != nil {
			return
		}
		vo := reflect.ValueOf(obj).Elem()
		var existLocales []string
		for i := 0; i < vo.Len(); i++ {
			existLocales = append(existLocales, vo.Index(i).Elem().FieldByName("LocaleCode").String())
		}
		toLocales := lb.GetSupportLocalesFromRequest(ctx.R)
		var selectLocales []SelectLocale
		for _, locale := range toLocales {
			if locale == fromLocale {
				continue
			}
			if !utils.Contains(existLocales, lb.GetLocaleCode(locale)) || vo.Len() == 0 {
				selectLocales = append(selectLocales, SelectLocale{Label: MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)), Code: lb.GetLocaleCode(locale)})
			}
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: presets.DialogPortalName,
			Body: v.VDialog(
				v.VCard(
					v.VCardTitle(h.Text(MustGetTranslation(ctx.R, "Localize"))),

					v.VCardText(
						h.Div(
							h.Div(
								h.Div(
									h.Label(MustGetTranslation(ctx.R, "LocalizeFrom")).Class("v-label v-label--active theme--light").Style("left: 0px; right: auto; position: absolute;"),
									//h.Br(),
									h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale))),
								).Class("v-text-field__slot"),
							).Class("v-input__slot"),
						).Class("v-input v-input--is-label-active v-input--is-dirty theme--light v-text-field v-text-field--is-booted"),
						v.VSelect().FieldName("localize_to").
							Label(MustGetTranslation(ctx.R, "LocalizeTo")).
							Multiple(true).Chips(true).
							Items(selectLocales).
							ItemText("Label").
							ItemValue("Code"),
					).Attr("style", "height: 200px;"),

					v.VCardActions(
						v.VSpacer(),
						v.VBtn(presetsMsgr.Cancel).
							Depressed(true).
							Class("ml-2").
							On("click", "vars.localizeConfirmation = false"),

						v.VBtn(presetsMsgr.OK).
							Color("primary").
							Depressed(true).
							Dark(true).
							Attr("@click", web.Plaid().
								EventFunc(DoLocalize).
								Query(presets.ParamID, paramID).
								Query("localize_from", lb.GetLocaleCode(fromLocale)).
								URL(ctx.R.URL.Path).
								Go()),
					),
				),
			).MaxWidth("600px").
				Attr("v-model", "vars.localizeConfirmation").
				Attr(web.InitContextVars, `{localizeConfirmation: false}`),
		})

		r.VarsScript = "setTimeout(function(){ vars.localizeConfirmation = true }, 100)"
		return
	}
}

func doLocalizeTo(db *gorm.DB, mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		fromParamID := ctx.R.FormValue(presets.ParamID)
		to, exist := ctx.R.Form["localize_to"]
		if !exist {
			return
		}

		var fromObj = mb.NewModel()

		if err = utils.PrimarySluggerWhere(db, mb.NewModel(), fromParamID).First(fromObj).Error; err != nil {
			return
		}

		me := mb.Editing()

		for _, toLocale := range to {
			var toObj = mb.NewModel()
			var fakeToObj = fromObj
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
					me.UpdateOverlayContent(ctx, &r, toObj, "", &vErr)
					return
				}
			}

			if err = db.Save(toObj).Error; err != nil {
				return
			}
		}

		presets.ShowMessage(&r, MustGetTranslation(ctx.R, "SuccessfullyLocalized"), "")

		// refresh current page
		r.Reload = true
		return
	}
}
