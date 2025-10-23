package l10n

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/qor5/web/v3"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
)

const (
	WrapHandlerKey  = "l10nWrapHandlerKey"
	MenuTopItemFunc = "l10nMenuTopItemFunc"
	SlugLocaleCode  = "locale_code"
)

type SwitchLocaleKey struct{}

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
			chips = append(chips, VChip(h.Text(MustGetTranslation(ctx.R, lb.GetLocaleLabel(fromLocale)))).Color("success").Variant(VariantFlat).Label(true).Size(SizeXSmall))
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
					Size(SizeXSmall),
			)
		}
		//menu := lb.localizeMenu(obj, chips, field, ctx, slices.DeleteFunc(allLocales, func(s string) bool {
		//	return s == fromLocale
		//}), existLocales)
		return h.Td(
			h.Div(chips...).Class("d-flex ga-2"),
		)
	}
}

func (b *Builder) localizeMenu(obj interface{}, chips h.HTMLComponents, field *presets.FieldContext, ctx *web.EventContext, allLocales, existLocales []string) h.HTMLComponent {
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
			web.Slot(VIcon("mdi-check").Attr("v-show", fmt.Sprintf(`menuLocals.locales.includes(%q)`, locale)).Size(SizeSmall)).Name(VSlotAppend),
		).Disabled(disable).Attr("@click",
			fmt.Sprintf(`menuLocals.locales.includes(%q)?menuLocals.locales.splice(menuLocals.locales.indexOf(%q), 1):menuLocals.locales.push(%q);`, locale, locale, locale)),
		)
	}
	return web.Scope(
		VMenu(
			web.Slot(
				h.Div(
					chips,
					VIcon("mdi-menu-down"),
				).Class("d-flex ga-2").AttrIf("v-bind", "props", !allSelected),
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
		Init(`{locales:[]}`)
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

func (b *Builder) runSwitchLocaleFunc(ctx *web.EventContext) (r h.HTMLComponent) {
	var (
		btn             h.HTMLComponent
		localsListItems []h.HTMLComponent

		allLocales = b.GetSupportLocaleCodesFromRequest(ctx.R)
		fromLocale = b.GetCorrectLocaleCode(ctx.R)
		fromImg    = b.GetLocaleImg(fromLocale)
		id         = ctx.Param(presets.ParamID)
	)
	if fromImg != "" {
		btn = VBtn("").Class("pa-0").Children(
			h.Div(
				h.RawHTML(fromImg),
				VIcon("mdi-menu-down"),
			).Class("d-flex ga-2"),
		).Attr("v-bind", "plaid().vue.mergeProps(menu,tooltip)").Size(SizeSmall).Variant(VariantText)
	} else {
		btn = VChip(
			h.Span(MustGetTranslation(ctx.R, b.GetLocaleLabel(fromLocale))).Attr("style", "max-width: 80px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;"),
		).Color(ColorSuccess).Variant(VariantFlat).Label(true).Size(SizeXSmall).Attr("v-bind", "plaid().vue.mergeProps(menu,tooltip)")
	}
	for _, locale := range allLocales {
		if locale == fromLocale {
			continue
		}
		clickEvent := web.Plaid().FieldValue(b.queryName, locale).Go()
		if id != "" {
			originPath := ctx.R.URL.Path
			val := ctx.ContextValue(SwitchLocaleKey{})
			if val != nil {
				clickEvent = web.Plaid().URL(val).PushState(true).FieldValue(b.queryName, locale).Go()
			} else {
				uri := strings.TrimSuffix(originPath, "/"+id)
				if uri != originPath {
					clickEvent = web.Plaid().URL(uri).PushState(true).FieldValue(b.queryName, locale).Go()
				}
			}
		}
		img := b.GetLocaleImg(locale)
		localsListItems = append(localsListItems, VListItem(
			VListItemTitle(
				h.If(img != "", h.RawHTML(img)),
				h.Text(MustGetTranslation(ctx.R, b.GetLocaleLabel(locale))),
			).Class("d-flex align-center ga-2"),
		).Attr("@click", clickEvent))
	}
	return web.Scope(
		VMenu(
			web.Slot(
				VTooltip(
					web.Slot(
						btn,
					).Name("activator").Scope(`{props:tooltip}`),
				).Location(LocationBottom).Text(MustGetTranslation(ctx.R, b.GetLocaleLabel(fromLocale))),
			).Name("activator").Scope(`{props:menu}`),
			VList(
				localsListItems...,
			),
		)).VSlot("{form}")
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
