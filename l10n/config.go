package l10n

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/utils"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	WrapHandlerKey  = "l10nWrapHandlerKey"
	MenuTopItemFunc = "l10nMenuTopItemFunc"
)

func localeListFunc(db *gorm.DB, lb *Builder) func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
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
		chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale)))).Color("success").Variant(VariantFlat).Label(true).Size(SizeSmall))

		for _, locale := range otherLocales {
			chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale)))).Label(true).Size(SizeSmall))
		}
		return h.Td(
			chips...,
		)
	}
}

func runSwitchLocaleFunc(lb *Builder) func(ctx *web.EventContext) (r h.HTMLComponent) {
	return func(ctx *web.EventContext) (r h.HTMLComponent) {
		supportLocales := lb.GetSupportLocaleCodesFromRequest(ctx.R)

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
