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

func overview(b *Builder, templateM *presets.ModelBuilder) presets.FieldComponentFunc {
	// TODO: refactor versionDialog to use publish/views
	// publish.ConfigureVersionListDialog(db, b.ps, pm)
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		// TODO: init default VersionComponent

		pm := b.mb

		var (
			start, end, se string
			categories     []*Category
			onlineHint     h.HTMLComponent
		)
		var (
			versionComponent = publish.DefaultVersionComponentFunc(pm)(obj, field, ctx)
			mi               = field.ModelInfo
			p                = obj.(*Page)
			c                = &Category{}
		)
		isTemplate := strings.Contains(ctx.R.RequestURI, "/"+templateM.Info().URIName()+"/")
		if v, ok := obj.(interface {
			GetID() uint
		}); ok {
			ctx.R.Form.Set(paramPageID, strconv.Itoa(int(v.GetID())))
		}
		if v, ok := obj.(publish.VersionInterface); ok {
			ctx.R.Form.Set(paramPageVersion, v.GetVersion())
		}
		if l, ok := obj.(l10n.L10nInterface); ok {
			ctx.R.Form.Set(paramLocale, l.GetLocale())
		}
		if isTemplate {
			ctx.R.Form.Set(paramsTpl, "1")
		}

		previewDevelopUrl := b.previewHref(ctx)

		b.db.First(c, "id = ? AND locale_code = ?", p.CategoryID, p.LocaleCode)
		if p.GetScheduledStartAt() != nil {
			start = p.GetScheduledStartAt().Format("2006-01-02 15:04")
		}
		if p.GetScheduledEndAt() != nil {
			end = p.GetScheduledEndAt().Format("2006-01-02 15:04")
		}
		if start != "" || end != "" {
			se = "Scheduled at: " + start + " ~ " + end
		}

		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		if err := b.db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
			panic(err)
		}

		versionBadge := VChip(h.Text(fmt.Sprintf("%d versions", versionCount(b.db, p)))).Color(ColorPrimary).Size(SizeSmall).Class("px-1 mx-1").Attr("style", "height:20px")
		if p.GetStatus() == publish.StatusOnline {
			onlineHint = VAlert(h.Text("The version cannot be edited directly after it is released. Please copy the version and edit it.")).Density(DensityCompact).Type(TypeInfo).Variant(VariantTonal).Closable(true).Class("mb-2")
		}
		return web.Scope(
			VLayout(
				VAppBar(
					web.Slot(
						VBtn("").Size(SizeXSmall).Icon("mdi-arrow-left").Tile(true).Variant(VariantOutlined).Attr("@click",
							web.GET().URL(mi.PresetsPrefix()+"/pages").PushState(true).Go(),
						),
					).Name(VSlotPrepend),
					web.Slot(
						h.Div(
							h.H1("{{vars.pageTitle}}"),
							versionBadge.Class("mt-2 ml-2"),
						).Class("d-inline-flex align-center"),
					).Name(VSlotTitle),
				).Color(ColorSurface).Elevation(0),
				VMain(
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
					).Style(`position:relative`).Class("w-100 mt-4").
						Attr("@click",
							web.Plaid().URL(fmt.Sprintf("%s/editors/%v", b.prefix, p.PrimarySlug())).PushState(true).Go(),
						),
					h.Div(
						h.A(h.Text(previewDevelopUrl)).Href(previewDevelopUrl),
						VBtn("").Icon("mdi-file-document-multiple").Variant(VariantText).Size(SizeXSmall).Class("ml-1").
							Attr("@click", fmt.Sprintf(`$event.view.window.navigator.clipboard.writeText($event.view.window.location.origin+"%s");vars.presetsMessage = { show: true, message: "success", color: "%s"}`, previewDevelopUrl, ColorSuccess)),
					).Class("d-inline-flex align-center"),
				).Class("my-10"),
			),
		).VSlot(" { locals  }").Init(`{ tab:"Page"} `)
	}
}

func templateSettings(db *gorm.DB, pm *presets.ModelBuilder) presets.FieldComponentFunc {
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

func detailPageEditor(dp *presets.DetailingBuilder, db *gorm.DB) {
	dp.Field("Page").SetSwitchable(true).Editing("Title", "Slug", "CategoryID").
		// TODO adjust layout
		ShowComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			p := obj.(*Page)
			var (
				category Category
				err      error
			)
			if category, err = p.GetCategory(db); err != nil {
				panic(err)
			}
			return h.Div(
				h.Div(h.Text("title:"+p.Title)).Attr(web.VAssign("vars", fmt.Sprintf(`{pageTitle:"%s"}`, p.Title))...),
				h.Div(h.Text("slug:"+p.Slug)),
				h.Div(h.Text("category:"+category.Path)),
			)
		}).EditComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		p := obj.(*Page)
		categories := []*Category{}
		locale, _ := l10n.IsLocalizableFromCtx(ctx.R.Context())
		if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
			panic(err)
		}

		var vErr web.ValidationErrors
		if ve, ok := ctx.Flash.(*web.ValidationErrors); ok {
			vErr = *ve
		}
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return h.Components(
			VTextField().
				Variant(FieldVariantUnderlined).Density(DensityCompact).
				Attr(web.VField("Page.Title", p.Title)...),

			VTextField().
				Variant(FieldVariantUnderlined).
				Attr(web.VField("Page.Slug", strings.TrimPrefix(p.Slug, "/"))...).
				Prefix("/").
				ErrorMessages(vErr.GetFieldErrors("Page.Category")...),

			vx.VXAutocomplete().Label(msgr.Category).
				Attr(web.VField("Page.CategoryID", p.CategoryID)...).
				Multiple(false).Chips(false).
				Items(categories).ItemText("Path").ItemValue("ID").
				ErrorMessages(vErr.GetFieldErrors("Page.CategoryID")...),
		)
	})
	return
}
