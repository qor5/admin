package views

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/qor5/admin/activity"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/utils"
	v "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

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
			return nil
		}).
		SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			return nil
		})

	b.AddWrapHandler(lb.EnsureLocale)
	b.AddMenuTopItemFunc(runSwitchLocaleFunc(lb))
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
		chips = append(chips, v.VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale)))).Color("green").TextColor("white").Label(true).Small(true))

		for _, locale := range otherLocales {
			chips = append(chips, v.VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)))).Label(true).Small(true))
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
					v.VList(
						v.VListItem(
							v.VListItemIcon(
								v.VIcon("language").Small(true).Class("ml-1"),
							).Attr("style", "margin-right: 16px"),
							v.VListItemContent(
								v.VListItemTitle(
									h.Div(h.Text(fmt.Sprintf("%s%s %s", MustGetTranslation(ctx.R, "Location"), MustGetTranslation(ctx.R, "Colon"), MustGetTranslation(ctx.R, lb.GetLocaleLabel(supportLocales[0]))))).Role("button"),
								),
							),
						).Class("pa-0").Dense(true),
					).Class("pa-0 ma-n4 mt-n6"),
				).Attr("@click", web.Plaid().Query(localeQueryName, supportLocales[0]).Go()),
			)
		}

		locale := lb.GetCorrectLocaleCode(ctx.R)

		var locales []h.HTMLComponent
		for _, contry := range supportLocales {
			locales = append(locales,
				h.Div(
					v.VListItem(
						v.VListItemContent(
							v.VListItemTitle(
								h.Div(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(contry)))),
							),
						),
					).Attr("@click", web.Plaid().Query(localeQueryName, contry).Go()),
				),
			)
		}

		return v.VMenu().OffsetY(true).Children(
			h.Template().Attr("v-slot:activator", "{on, attrs}").Children(
				h.Div(
					v.VList(
						v.VListItem(
							v.VListItemIcon(
								v.VIcon("language").Small(true).Class("ml-1"),
							).Attr("style", "margin-right: 16px"),
							v.VListItemContent(
								v.VListItemTitle(
									h.Text(fmt.Sprintf("%s%s %s", MustGetTranslation(ctx.R, "Location"), MustGetTranslation(ctx.R, "Colon"), MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)))),
								),
							),
							v.VListItemIcon(
								v.VIcon("arrow_drop_down").Small(false).Class("mr-1"),
							),
						).Class("pa-0").Dense(true),
					).Class("pa-0 ma-n4 mt-n6"),
				).Attr("v-bind", "attrs").Attr("v-on", "on"),
			),

			v.VList(
				locales...,
			).Dense(true),
		)
	}
}

func localizeRowMenuItemFunc(mi *presets.ModelInfo, url string, editExtraParams url.Values) vx.RowMenuItemFunc {
	return func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		if mi.Verifier().Do(presets.PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		return v.VListItem(
			v.VListItemIcon(v.VIcon("language")),
			v.VListItemTitle(h.Text(MustGetTranslation(ctx.R, "Localize"))),
		).Attr("@click", web.Plaid().
			EventFunc(Localize).
			Queries(editExtraParams).
			Query(presets.ParamID, id).
			URL(url).
			Go())
	}
}
