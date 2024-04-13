package views

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const WrapHandlerKey = "l10nWrapHandlerKey"
const MenuTopItemFunc = "l10nMenuTopItemFunc"

func Configure(b *presets.Builder, db *gorm.DB, lb *l10n.Builder, ab *activity.ActivityBuilder, models ...*presets.ModelBuilder) {
	for _, m := range models {
		obj := m.NewModel()
		_ = obj.(presets.SlugEncoder)
		_ = obj.(presets.SlugDecoder)
		_ = obj.(l10n.L10nInterface)
		if l10nONModel, exist := obj.(l10n.L10nONInterface); exist {
			l10nONModel.L10nON()
		}
		m.Listing().Field("Locale")
		m.Editing().Field("Locale")

		searcher := m.Listing().Searcher
		m.Listing().SearchFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error) {
			if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
				con := presets.SQLCondition{
					Query: "locale_code = ?",
					Args:  []interface{}{localeCode},
				}
				params.SQLConditions = append(params.SQLConditions, &con)
			}

			return searcher(model, params, ctx)
		})

		setter := m.Editing().Setter
		m.Editing().SetterFunc(func(obj interface{}, ctx *web.EventContext) {
			if ctx.R.FormValue(presets.ParamID) == "" {
				if localeCode := ctx.R.Context().Value(l10n.LocaleCode); localeCode != nil {
					if err := reflectutils.Set(obj, "LocaleCode", localeCode); err != nil {
						return
					}
				}
			}
			if setter != nil {
				setter(obj, ctx)
			}
		})

		deleter := m.Editing().Deleter
		m.Editing().DeleteFunc(func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			if err = deleter(obj, id, ctx); err != nil {
				return
			}
			locale := obj.(presets.SlugDecoder).PrimaryColumnValuesBySlug(id)["locale_code"]
			locale = fmt.Sprintf("%s(del:%d)", locale, time.Now().UnixMilli())

			withoutKeys := []string{}
			if ctx.R.URL.Query().Get("all_versions") == "true" {
				withoutKeys = append(withoutKeys, "version")
			}

			if err = utils.PrimarySluggerWhere(db.Unscoped(), obj, id, withoutKeys...).Update("locale_code", locale).Error; err != nil {
				return
			}
			return
		})

		rmb := m.Listing().RowMenu()
		rmb.RowMenuItem("Localize").ComponentFunc(localizeRowMenuItemFunc(m.Info(), "", url.Values{}))

		registerEventFuncs(db, m, lb, ab)
	}

	b.FieldDefaults(presets.LIST).
		FieldType(l10n.Locale{}).
		ComponentFunc(localeListFunc(db, lb))
	b.FieldDefaults(presets.WRITE).
		FieldType(l10n.Locale{}).
		ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var value string
			id, err := reflectutils.Get(obj, "ID")
			if err == nil && len(fmt.Sprint(id)) > 0 && fmt.Sprint(id) != "0" {
				value = field.Value(obj).(l10n.Locale).GetLocale()
			} else {
				value = lb.GetCorrectLocaleCode(ctx.R)
			}

			return h.Input("").Type("hidden").Attr(web.VField("LocaleCode", value)...)
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			value := field.Value(obj).(l10n.Locale).GetLocale()
			if !utils.Contains(lb.GetSupportLocaleCodesFromRequest(ctx.R), value) {
				return errors.New("Incorrect locale.")
			}

			return nil
		})

	b.AddWrapHandler(WrapHandlerKey, lb.EnsureLocale)
	b.AddMenuTopItemFunc(MenuTopItemFunc, runSwitchLocaleFunc(lb))
	b.I18n().
		RegisterForModule(language.English, I18nLocalizeKey, Messages_en_US).
		RegisterForModule(language.SimplifiedChinese, I18nLocalizeKey, Messages_zh_CN).
		RegisterForModule(language.Japanese, I18nLocalizeKey, Messages_ja_JP)
}

func localeListFunc(db *gorm.DB, lb *l10n.Builder) func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		id, err := reflectutils.Get(obj, "ID")
		if err != nil {
			return nil
		}
		fromLocale := lb.GetCorrectLocaleCode(ctx.R)

		objs := reflect.New(reflect.SliceOf(reflect.TypeOf(obj).Elem())).Interface()
		err = db.Distinct("locale_code").Where("id = ? AND locale_code <> ?", id, fromLocale).Find(objs).Error
		if err != nil {
			return nil
		}
		vo := reflect.ValueOf(objs).Elem()
		var existLocales []string
		for i := 0; i < vo.Len(); i++ {
			existLocales = append(existLocales, vo.Index(i).FieldByName("LocaleCode").String())
		}

		allLocales := lb.GetSupportLocaleCodesFromRequest(ctx.R)
		var otherLocales []string
		for _, locale := range allLocales {
			if utils.Contains(existLocales, locale) {
				otherLocales = append(otherLocales, locale)
			}
		}

		var chips []h.HTMLComponent
		chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale)))).Color("green").Variant(VariantFlat).Label(true).Size(SizeSmall))

		for _, locale := range otherLocales {
			chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)))).Label(true).Size(SizeSmall))
		}
		return h.Td(
			chips...,
		)
	}
}

func runSwitchLocaleFunc(lb *l10n.Builder) func(ctx *web.EventContext) (r h.HTMLComponent) {
	return func(ctx *web.EventContext) (r h.HTMLComponent) {
		var supportLocales = lb.GetSupportLocaleCodesFromRequest(ctx.R)

		if len(lb.GetSupportLocaleCodes()) <= 1 || len(supportLocales) == 0 {
			return nil
		}

		localeQueryName := lb.GetQueryName()

		if len(supportLocales) == 1 {
			return h.Template().Children(
				h.Div(
					VList(
						VListItem(
							web.Slot(
								// icon was language
								VIcon("mdi-translate").Size(SizeSmall).Class("ml-1").Attr("style", "margin-right: 16px"),
							).Name("prepend"),
							VListItemTitle(
								h.Div(h.Text(fmt.Sprintf("%s%s %s", MustGetTranslation(ctx.R, "Location"), MustGetTranslation(ctx.R, "Colon"), MustGetTranslation(ctx.R, lb.GetLocaleLabel(supportLocales[0]))))).Role("button"),
							),
						).Class("pa-0").Density(DensityCompact),
					).Class("pa-0 ma-n4 mt-n6"),
				).Attr("@click", web.Plaid().Query(localeQueryName, supportLocales[0]).Go()),
			)
		}

		locale := lb.GetCorrectLocaleCode(ctx.R)

		var locales []h.HTMLComponent
		for _, contry := range supportLocales {
			locales = append(locales,
				h.Div(
					VListItem(
						VListItemTitle(
							h.Div(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(contry)))),
						),
					).Attr("@click", web.Plaid().Query(localeQueryName, contry).Go()),
				),
			)
		}

		return VMenu().Children(
			h.Template().Attr("v-slot:activator", "{props}").Children(
				h.Div(
					VList(
						VListItem(
							web.Slot(
								// icon was language
								VIcon("mdi-translate").Size(SizeSmall).Class("ml-1"),
							).Name("prepend"),
							VListItemTitle(
								h.Text(fmt.Sprintf("%s%s %s", MustGetTranslation(ctx.R, "Location"), MustGetTranslation(ctx.R, "Colon"), MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)))),
							),
							web.Slot(
								// icon was arrow_drop_down
								VIcon("mdi-arrow-down").Class("mr-1"),
							).Name("append"),
						).Class("pa-0").Density(DensityCompact),
					).Class("pa-0 ma-n4 mt-n6"),
				).Attr("v-bind", "props"),
			),

			VList(
				locales...,
			).Density(DensityCompact),
		)
	}
}

func localizeRowMenuItemFunc(mi *presets.ModelInfo, url string, editExtraParams url.Values) vx.RowMenuItemFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		if mi.Verifier().Do(presets.PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		return VListItem(
			web.Slot(
				// icon was language
				VIcon("mdi-translate"),
			).Name("prepend"),
			VListItemTitle(h.Text(MustGetTranslation(ctx.R, "Localize"))),
		).Attr("@click", web.Plaid().
			EventFunc(Localize).
			Queries(editExtraParams).
			Query(presets.ParamID, id).
			URL(url).
			Go())
	}
}
