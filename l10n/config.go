package l10n

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	WrapHandlerKey  = "l10nWrapHandlerKey"
	MenuTopItemFunc = "l10nMenuTopItemFunc"
	SlugLocaleCode  = "locale_code"
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
		existLocales := make([]string, 0)
		for i := 0; i < vo.Len(); i++ {
			existLocales = append(existLocales, vo.Index(i).FieldByName("LocaleCode").String())
		}
		allLocales := lb.GetSupportLocaleCodesFromRequest(ctx.R)
		var otherLocales []string
		for _, locale := range allLocales {
			if slices.Contains(existLocales, locale) {
				otherLocales = append(otherLocales, locale)
			}
		}
		var chips []h.HTMLComponent
		fromImg := lb.GetLocaleImg(fromLocale)
		if fromImg != "" {
			chips = append(chips, h.RawHTML(fromImg))
		} else {
			chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale)))).Color("success").Variant(VariantFlat).Label(true).Size(SizeSmall))
		}

		for _, locale := range otherLocales {
			img := lb.GetLocaleImg(locale)
			if img != "" {
				chips = append(chips, h.RawHTML(img))
				continue
			}
			chips = append(chips,
				VChip(
					h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(locale))),
				).
					Label(true).
					Size(SizeSmall),
			)
		}
		menu := lb.localizeMenu(obj, field, ctx, slices.DeleteFunc(allLocales, func(s string) bool {
			return s == fromLocale
		}), existLocales)
		if menu != nil {
			chips = append(chips, menu)
		}
		return h.Td(
			h.Div(
				chips...,
			).Class("d-flex ga-2"),
		)
	}
}

func (b *Builder) localizeMenu(obj interface{}, field *presets.FieldContext, ctx *web.EventContext, allLocales, existLocales []string) h.HTMLComponent {
	if field.ModelInfo.Verifier().Do(presets.PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		return nil
	}
	p, ok := obj.(presets.SlugEncoder)
	if !ok {
		return nil
	}
	var localsListItems []h.HTMLComponent
	allSelected := true

	for _, locale := range allLocales {
		disable := false
		if slices.Contains(existLocales, locale) {
			disable = true
		} else {
			allSelected = false
		}
		img := b.GetLocaleImg(locale)
		localsListItems = append(localsListItems, VListItem(
			VListItemTitle(
				h.If(img != "", h.RawHTML(img)),
				h.Text(MustGetTranslation(ctx.R, b.GetLocaleLabel(locale))),
			).Class("d-flex align-center ga-2"),
			web.Slot(VIcon("mdi-check").Attr("v-show", fmt.Sprintf(`menuLocals.locales.includes("%s")`, locale)).Size(SizeSmall)).Name(VSlotAppend),
		).Disabled(disable).Attr("@click",
			fmt.Sprintf(`menuLocals.locales.includes("%s")?menuLocals.locales.splice(menuLocals.locales.indexOf("%s"), 1):menuLocals.locales.push("%s");`, locale, locale, locale)),
		)
	}
	return web.Scope(
		VMenu(
			web.Slot(
				VIcon("mdi-menu-down").Attr("v-bind", "props").Disabled(allSelected),
			).Name("activator").Scope(`{props}`),
			VList(
				localsListItems...,
			),
		).CloseOnContentClick(false).Attr("@update:model-value",
			fmt.Sprintf(`$event?null:%s`,
				web.Plaid().
					URL(ctx.R.URL.Path).
					EventFunc(DoLocalize).
					Query(presets.ParamID, p.PrimarySlug()).
					Query("localize_from", b.GetCorrectLocaleCode(ctx.R)).
					FieldValue("localize_to", web.Var("menuLocals.locales")).
					Go(),
			)),
	).VSlot(`{locals:menuLocals}`).
		Init(fmt.Sprintf(`{locales:%v}`, h.JSONString(existLocales)))
}

func runSwitchLocaleFunc(lb *Builder) func(ctx *web.EventContext) (r h.HTMLComponent) {
	return func(ctx *web.EventContext) (r h.HTMLComponent) {
		supportLocales := lb.GetSupportLocaleCodesFromRequest(ctx.R)

		if len(lb.GetSupportLocaleCodes()) <= 1 || len(supportLocales) == 0 {
			return nil
		}

		localeQueryName := lb.queryName

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
