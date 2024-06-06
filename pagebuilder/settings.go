package pagebuilder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/x/v3/i18n"

	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
)

func overview(b *Builder, templateM *presets.ModelBuilder, pm *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			start, end, se string
			onlineHint     h.HTMLComponent
			isTemplate     bool
			ps             string
			id             uint
		)
		versionComponent := publish.DefaultVersionComponentFunc(pm)(obj, field, ctx)
		if templateM != nil {
			isTemplate = strings.Contains(ctx.R.RequestURI, "/"+templateM.Info().URIName()+"/")
		}
		if v, ok := obj.(PrimarySlugInterface); ok {
			ps = v.PrimarySlug()
		}
		if v, ok := obj.(interface {
			GetID() uint
		}); ok {
			id = v.GetID()
			ctx.R.Form.Set(paramPageID, strconv.Itoa(int(id)))
		}
		if v, ok := obj.(publish.VersionInterface); ok {
			ctx.R.Form.Set(paramPageVersion, v.EmbedVersion().Version)
		}
		if l, ok := obj.(l10n.LocaleInterface); ok {
			ctx.R.Form.Set(paramLocale, l.EmbedLocale().LocaleCode)
		}
		if isTemplate {
			ctx.R.Form.Set(paramsTpl, "1")
		}

		previewDevelopUrl := b.previewHref(ctx, pm, ps)

		if schedule, ok := obj.(publish.ScheduleInterface); ok {
			if em := schedule.EmbedSchedule().ScheduledStartAt; em != nil {
				start = em.Format("2006-01-02 15:04")
			}
			if em := schedule.EmbedSchedule().ScheduledEndAt; em != nil {
				end = em.Format("2006-01-02 15:04")
			}
			if start != "" || end != "" {
				se = "Scheduled at: " + start + " ~ " + end
			}
		}

		if p, ok := obj.(publish.StatusInterface); ok {
			if p.EmbedStatus().Status == publish.StatusOnline {
				onlineHint = VAlert(h.Text("The version cannot be edited directly after it is released. Please copy the version and edit it.")).Density(DensityCompact).Type(TypeInfo).Variant(VariantTonal).Closable(true).Class("mb-2")
			}
		}
		return h.Div(
			onlineHint,
			versionComponent,
			h.Div(
				h.Iframe().Src(previewDevelopUrl).Attr("scrolling", "no", "frameborder", "0").Style(`height:320px;width:100%;pointer-events: none;`),
				h.Div(
					h.Div(
						h.Text(se),
					).Class(fmt.Sprintf("bg-%s", ColorSecondaryLighten2)),
					VBtn("Page Builder").PrependIcon("mdi-pencil").Color(ColorSecondary).
						Class("rounded-sm").Height(40).Variant(VariantFlat),
				).Class("pa-6 w-100 d-flex justify-space-between align-center").Style(`position:absolute;top:0;left:0`),
			).Style(`position:relative`).Class("w-100").
				Attr("@click",
					web.Plaid().URL(fmt.Sprintf("%s/%s/editors/%v", b.prefix, pm.Info().URIName(), ps)).PushState(true).Go(),
				),
			h.Div(
				h.A(h.Text(previewDevelopUrl)).Href(previewDevelopUrl),
				VBtn("").Icon("mdi-file-document-multiple").Variant(VariantText).Size(SizeXSmall).Class("ml-1").
					Attr("@click", fmt.Sprintf(`$event.view.window.navigator.clipboard.writeText($event.view.window.location.origin+"%s");vars.presetsMessage = { show: true, message: "success", color: "%s"}`, previewDevelopUrl, ColorSuccess)),
			).Class("d-inline-flex align-center"),
		).Class("my-10")
	}
}

func templateSettings(_ *gorm.DB, pm *presets.ModelBuilder) presets.FieldComponentFunc {
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Template)

		overview := vx.DetailInfo(
			vx.DetailColumn(
				vx.DetailField(vx.OptionalText(p.Name)).Label("Title"),
				vx.DetailField(vx.OptionalText(p.Description)).Label("Description"),
			),
		)

		editBtn := VBtn("Edit").Variant(VariantFlat).
			Attr("@click", web.POST().
				EventFunc(actions.Edit).
				Query(presets.ParamOverlay, actions.Dialog).
				Query(presets.ParamID, p.PrimarySlug()).
				URL(pm.Info().ListingHref()).Go(),
			)

		return VContainer(
			VRow(
				VCol(
					vx.Card(overview).HeaderTitle("Overview").
						Actions(
							h.If(editBtn != nil, editBtn),
						).Class("mb-4 rounded-lg").Variant(VariantOutlined),
				).Cols(8),
			),
		)
	}
}

func detailingRow(label string, showComp h.HTMLComponent) (r *h.HTMLTagBuilder) {
	return h.Div(
		h.Div(h.Text(label)).Class("text-subtitle-2").Style("width:180px;height:20px"),
		h.Div(showComp).Class("text-body-1 ml-2 w-100"),
	).Class("d-flex align-center ma-2").Style("height:40px")
}

func detailPageEditor(dp *presets.DetailingBuilder, db *gorm.DB) {
	dp.Section("Page").
		Editing("Title", "Slug", "CategoryID").
		ViewComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			p := obj.(*Page)
			var (
				category Category
				err      error
			)
			if category, err = p.GetCategory(db); err != nil {
				panic(err)
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			return h.Div(
				h.Div(h.Text(msgr.PageOverView)).Class("text-h4"),
				detailingRow("Title", h.Text(p.Title)).Attr(web.VAssign("vars", fmt.Sprintf(`{pageTitle:"%s"}`, p.Title))...),
				detailingRow("Slug", h.Text(p.Slug)),
				detailingRow(msgr.Category, h.Text(category.Path)),
			)
		}).EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
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
		return h.Components(
			h.Div(h.Text(msgr.PageOverView)).Class("text-h4"),
			detailingRow("Title",
				VTextField().
					Variant(VariantOutlined).
					Density(DensityCompact).
					HideDetails(true).
					Attr(web.VField("Page.Title", p.Title)...),
			),
			detailingRow("Slug",
				VTextField().
					Variant(VariantOutlined).
					Density(DensityCompact).
					HideDetails(true).
					Attr(web.VField("Page.Slug", strings.TrimPrefix(p.Slug, "/"))...).
					Prefix("/").
					ErrorMessages(vErr.GetFieldErrors("Page.Category")...),
			),
			detailingRow(msgr.Category,
				VAutocomplete().
					Variant(VariantOutlined).
					Density(DensityCompact).
					HideDetails(true).
					Attr(web.VField("Page.CategoryID", p.CategoryID)...).
					Multiple(false).Chips(false).
					Items(categories).ItemTitle("Path").ItemValue("ID").
					ErrorMessages(vErr.GetFieldErrors("Page.CategoryID")...),
			),
		)
	})
	return
}
