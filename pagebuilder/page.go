package pagebuilder

import (
	"fmt"
	"path"
	"strings"

	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"

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
	dp.WrapPageFunc(func(in web.PageFunc) web.PageFunc {
		return func(ctx *web.EventContext) (r web.PageResponse, err error) {
			r, err = in(ctx)
			if err != nil {
				return
			}
			r.Body = h.Div(r.Body).Class("px-6")
			return
		}
	})
	dp.Field("Title").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		var (
			versionBadge *VChipBuilder
			ps           string
		)
		if v, ok := obj.(presets.SlugEncoder); ok {
			ps = v.PrimarySlug()
		}
		if v, ok := obj.(presets.SlugDecoder); ok && ps != "" {
			cs := v.PrimaryColumnValuesBySlug(ps)
			versionBadge = VChip(h.Text(fmt.Sprintf("%d %s", versionCount(b.db, pm.NewModel(), cs[presets.ParamID], cs[l10n.SlugLocaleCode]), msgr.Versions))).
				Color(ColorPrimary).Size(SizeSmall).Class("px-1 mx-1").Attr("style", "height:20px")
		}

		return h.Div(
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

	eb := pm.Editing().
		WrapValidateFunc(func(in presets.ValidateFunc) presets.ValidateFunc {
			return func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
				var (
					p    = obj.(*Page)
					msgr = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
				)

				if p.Status.Status == publish.StatusOnline || p.Status.Status == publish.StatusOffline {
					err.GlobalError(msgr.TheResourceCanNotBeModified)
					return
				}
				if err = pageValidator(ctx, p, db, b.l10n); err.HaveErrors() {
					return
				}
				return
			}
		}).Creating(names...)

	titleFiled := eb.GetField("Title")
	if titleFiled != nil {
		titleFiled.LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
				comp := in(obj, field, ctx)
				return comp.(*vx.VXFieldBuilder).Label(msgr.ListHeaderTitle)
			}
		})
	}
	slugFiled := eb.GetField("Slug")
	if slugFiled != nil {
		slugFiled.LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				comp := in(obj, field, ctx)
				p := obj.(*Page)
				msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
				return comp.(*vx.VXFieldBuilder).
					Label(msgr.Slug).
					Attr(presets.VFieldError(field.FormKey, strings.TrimPrefix(p.Slug, "/"), field.Errors)...).
					Disabled(field.Disabled).Attr("prefix", "/")
			}
		}).LazyWrapSetterFunc(func(in presets.FieldSetterFunc) presets.FieldSetterFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
				p := obj.(*Page)
				p.Slug = path.Join("/", p.Slug)
				return in(obj, field, ctx)
			}
		})
	}
	categoryIDFiled := eb.GetField("CategoryID")
	if categoryIDFiled != nil {
		categoryIDFiled.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var (
				p          = obj.(*Page)
				categories []*Category
				locale, _  = l10n.IsLocalizableFromContext(ctx.R.Context())
			)
			if innerErr := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; innerErr != nil {
				panic(innerErr)
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			complete := presets.SelectField(obj, field, ctx).
				Multiple(false).Chips(false).
				Label(msgr.Category).
				Clearable(true).
				Items(categories).ItemTitle("Path").ItemValue("ID")
			if p.CategoryID > 0 {
				complete.Attr(presets.VFieldError(field.FormKey, p.CategoryID, field.Errors)...)
			} else {
				complete.Attr(presets.VFieldError(field.FormKey, "", field.Errors)...)
			}
			return complete
		})
	}

	detailPageEditor(dp, pm, b)

	b.configDetailLayoutFunc(pb, pm, db)

	return
}
