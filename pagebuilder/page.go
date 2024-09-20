package pagebuilder

import (
	"fmt"
	"path"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/seo"
)

func (b *Builder) defaultPageInstall(pb *presets.Builder, pm *presets.ModelBuilder) (err error) {
	db := b.db

	listingFields := []string{"ID", "Title", publish.ListingFieldLive, "Path"}
	if b.ab != nil {
		listingFields = append(listingFields, activity.ListFieldNotes)
	}
	lb := pm.Listing(listingFields...)
	lb.FilterDataFunc(func(ctx *web.EventContext) vx.FilterData {
		liveFilterItem, err := publish.NewLiveFilterItem(ctx.R.Context(), "")
		if err != nil {
			panic(liveFilterItem)
		}
		return []*vx.FilterItem{liveFilterItem}
	})
	pm.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ModelLabelPage
		}
		return msgr.ModelLabelPages
	})
	lb.Field("Path").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		page := obj.(*Page)
		category, err := page.GetCategory(db)
		if err != nil {
			panic(err)
		}
		return h.Td(h.Text(page.getAccessUrl(page.getPublishUrl(b.l10n.GetLocalePath(page.LocaleCode), category.Path))))
	})

	detailList := []interface{}{"Title", PageBuilderPreviewCard, "Page"}
	if b.seoBuilder != nil {
		detailList = append(detailList, seo.SeoDetailFieldName)
	}

	dp := pm.Detailing(detailList...)
	dp.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		var versionBadge *VChipBuilder
		if v, ok := obj.(PrimarySlugInterface); ok {
			ps := v.PrimaryColumnValuesBySlug(v.PrimarySlug())

			versionBadge = VChip(h.Text(fmt.Sprintf("%d %s", versionCount(b.db, pm.NewModel(), ps["id"], ps[l10n.SlugLocaleCode]), msgr.Versions))).
				Color(ColorPrimary).Size(SizeSmall).Class("px-1 mx-1").Attr("style", "height:20px")
		}

		// listingHref := pm.Info().ListingHref()
		return h.Div(
			// VBtn("").Size(SizeXSmall).Icon("mdi-arrow-left").Tile(true).Variant(VariantOutlined).Attr("@click",
			// 	fmt.Sprintf(`
			// 		const last = vars.__history.last();
			// 		if (last && last.url && last.url.startsWith(%q)) {
			// 			$event.view.window.history.back();
			// 			return;
			// 		}
			// 		%s`, listingHref, web.GET().URL(listingHref).PushState(true).Go(),
			// 	),
			// ),
			h.H1("{{vars.pageTitle}}").Class("page-main-title"),
			versionBadge.Class("mt-2 ml-2"),
		).Class("d-inline-flex align-center")
	})
	// register modelBuilder
	names := b.filterFields([]interface{}{"Title", "CategoryID", "Slug"})
	if b.templateEnabled {
		names = append([]interface{}{PageTemplateSelectionFiled}, names...)
	}
	lb.WrapColumns(presets.CustomizeColumnLabel(func(evCtx *web.EventContext) (map[string]string, error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"ID":    msgr.ListHeaderID,
			"Title": msgr.ListHeaderTitle,
			"Path":  msgr.ListHeaderPath,
		}, nil
	}))
	eb := pm.Editing().Creating(names...)
	eb.ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		c := obj.(*Page)
		err = pageValidator(ctx, c, db, b.l10n)
		return
	})
	titleFiled := eb.GetField("Title")
	if titleFiled != nil {
		titleFiled.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

			return VTextField().
				Label(msgr.ListHeaderTitle).
				Variant(FieldVariantUnderlined).
				Attr(web.VField(field.Name, field.Value(obj))...).
				ErrorMessages(vErr.GetFieldErrors("Page.Title")...)
		})
	}
	slugFiled := eb.GetField("Slug")
	if slugFiled != nil {
		slugFiled.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

			return VTextField().
				Variant(FieldVariantUnderlined).
				Label(msgr.Slug).
				Attr(web.VField(field.Name, strings.TrimPrefix(field.Value(obj).(string), "/"))...).
				Prefix("/").
				ErrorMessages(vErr.GetFieldErrors("Page.Slug")...)
		}).SetterFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
			m := obj.(*Page)
			m.Slug = path.Join("/", m.Slug)
			return nil
		})
	}
	categoryIDFiled := eb.GetField("CategoryID")
	if categoryIDFiled != nil {
		categoryIDFiled.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			p := obj.(*Page)
			categories := []*Category{}
			locale, _ := l10n.IsLocalizableFromContext(ctx.R.Context())
			if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
				panic(err)
			}

			var vErr web.ValidationErrors
			if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
				vErr = *ve
			}

			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

			complete := VAutocomplete().
				Label(msgr.Category).
				Variant(FieldVariantUnderlined).
				Multiple(false).Chips(false).
				Items(categories).ItemTitle("Path").ItemValue("ID").
				ErrorMessages(vErr.GetFieldErrors("Page.Category")...)
			if p.CategoryID > 0 {
				complete.Attr(web.VField(field.Name, p.CategoryID)...)
			} else {
				complete.Attr(web.VField(field.Name, "")...)
			}
			return complete
		})
	}

	detailPageEditor(dp, b)

	b.configDetailLayoutFunc(pb, pm, db)

	return
}
