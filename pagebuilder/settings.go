package pagebuilder

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/publish"
)

func overview(m *ModelBuilder) presets.FieldComponentFunc {
	pm := m.mb
	b := m.builder
	return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
		var (
			start, end, se     string
			onlineHint         h.HTMLComponent
			ps                 string
			version            string
			id                 uint
			containerCount     int64
			updatedSharedCount int64
		)
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		versionComponent := publish.DefaultVersionComponentFunc(pm)(obj, field, ctx)
		if v, ok := obj.(presets.SlugEncoder); ok {
			ps = v.PrimarySlug()
		}

		id = reflectutils.MustGet(obj, "ID").(uint)
		UpdatedAt := reflectutils.MustGet(obj, "UpdatedAt").(time.Time)
		ctx.R.Form.Set(paramPageID, strconv.Itoa(int(id)))

		if v, ok := obj.(publish.VersionInterface); ok {
			version = v.EmbedVersion().Version
			ctx.R.Form.Set(paramPageVersion, version)
		}
		if l, ok := obj.(l10n.LocaleInterface); ok {
			ctx.R.Form.Set(paramLocale, l.EmbedLocale().LocaleCode)
		}

		previewDevelopUrl := m.PreviewHref(ctx, ps)

		if schedule, ok := obj.(publish.ScheduleInterface); ok {
			if em := schedule.EmbedSchedule().ScheduledStartAt; em != nil {
				start = em.Format("2006-01-02 15:04")
			}
			if em := schedule.EmbedSchedule().ScheduledEndAt; em != nil {
				end = em.Format("2006-01-02 15:04")
			}
			if start != "" || end != "" {
				se = msgr.ScheduledAt + ": " + start + " ~ " + end
			}
		}
		b.db.Model(&Container{}).
			Where("page_id = ? AND page_version = ? and page_model_name = ?", id, version, m.name).
			Count(&containerCount)
		b.db.Model(&Container{}).
			Where("page_id = ? AND page_version = ? and page_model_name = ? and shared=true and updated_at > ?", id, version, m.name, UpdatedAt).
			Count(&updatedSharedCount)
		var copyURL string
		coverBtn := VBtn(msgr.EditPage).AppendIcon("mdi-pencil")
		if p, ok := obj.(publish.StatusInterface); ok {
			copyURL = fmt.Sprintf(`$event.view.window.location.origin+%q`, previewDevelopUrl)
			status := p.EmbedStatus().Status
			if status != publish.StatusDraft {
				coverBtn = VBtn(msgr.ViewPage).AppendIcon("mdi-eye")
			}
			if status == publish.StatusOnline {
				onlineHint = h.Div(
					h.If(updatedSharedCount > 0, VAlert(h.Text(msgr.SharedContainerHasBeenUpdated)).
						Density(DensityCompact).Type(TypeInfo).Variant(VariantTonal).Closable(true).Class("my-2"),
					),
					VAlert(
						h.Text(msgr.OnlineHit)).
						Density(DensityCompact).Type(TypeInfo).Variant(VariantTonal).Closable(true).Class("my-2"),
				)
				var err error
				previewDevelopUrl, err = b.publisher.FullUrl(ctx.R.Context(), p.EmbedStatus().OnlineUrl)
				if err != nil {
					panic(err)
				}
				copyURL = fmt.Sprintf(`%q`, previewDevelopUrl)
			}
		}
		// Add pauseVideo=1 parameter to previewDevelopUrl for iframe
		iframeUrlWithVideoPause := previewDevelopUrl
		if strings.Contains(iframeUrlWithVideoPause, "?") {
			iframeUrlWithVideoPause += "&pauseVideo=1"
		} else {
			iframeUrlWithVideoPause += "?pauseVideo=1"
		}

		previewComp := h.A(h.Text(previewDevelopUrl)).Href(previewDevelopUrl)
		if m.builder.previewOpenNewTab {
			previewComp.Target("_blank")
		}
		return h.Div(
			onlineHint,

			versionComponent,
			web.Listen(m.mb.NotifModelsUpdated(),
				web.Plaid().URL(m.mb.Info().DetailingHref(ps)).Go()),
			h.Div(
				h.Div(
					h.If(containerCount == 0,
						h.Div(
							VCard(
								VCardTitle(h.RawHTML(previewIframeEmptySvg)).Class("d-flex justify-center"),
								VCardSubtitle(h.Text(msgr.NoContentHit)).
									Class("d-flex justify-center"),
							).Flat(true).Class("bg-"+ColorGreyLighten4),
						).Class("d-flex align-center justify-center", H100, "bg-"+ColorGreyLighten4),
					),
					h.If(containerCount > 0,
						h.Iframe().Src(iframeUrlWithVideoPause).
							Attr("scrolling", "no", "frameborder", "0").
							Style(`pointer-events: none; 
 -webkit-mask-image: radial-gradient(circle, black 80px, transparent);
  mask-image: radial-gradient(circle, black 80px, transparent);
transform-origin: 0 0; transform:scale(0.5);width:200%;height:200%`),
					),
				).Class(W100, H100, "overflow-hidden"),
				h.Div(
					h.Div(
						h.Text(se),
					).Class(fmt.Sprintf("bg-%s", ColorGreyLighten3)),
					coverBtn.Color(ColorBlack).
						Class("rounded").Height(36).Variant(VariantElevated),
				).Class("pa-6 w-100 d-flex justify-space-between align-center").Style(`position:absolute;bottom:0;left:0`),
			).Style(`position:relative;height:320px;width:100%`).Class("border-thin rounded-lg").
				Attr("@click",
					web.Plaid().URL(m.editorURLWithSlug(ps)).PushState(true).Go(),
				),
			h.Div(
				previewComp,
				VBtn("").Icon("mdi-content-copy").Color(ColorSecondary).Width(20).Height(20).Variant(VariantText).Size(SizeXSmall).Class("ml-1 fix-btn-icon").
					Attr("@click", fmt.Sprintf(`$event.view.window.navigator.clipboard.writeText(%s);vars.presetsMessage = { show: true, message: "success", color: %q}`, copyURL, ColorSuccess)),
			).Class("d-inline-flex align-center py-4"),
		).Class("my-10")
	}
}

func detailPageEditor(dp *presets.DetailingBuilder, mb *presets.ModelBuilder, b *Builder) {
	db := b.db
	fields := b.filterFields([]interface{}{"Title", "CategoryID", "Slug"})
	section := presets.NewSectionBuilder(mb, "Page").
		Editing(fields...).WrapValidator(func(in presets.ValidateFunc) presets.ValidateFunc {
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
	})
	if b.expectField("Title") {
		section.ViewingField("Title").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				comp := in(obj, field, ctx)
				p := obj.(*Page)
				return h.Div(comp).Attr(web.VAssign("vars", fmt.Sprintf(`{pageTitle:%q}`, p.Title))...)
			}
		})
	}
	if b.expectField("Slug") {
		section.EditingField("Slug").LazyWrapComponentFunc(func(in presets.FieldComponentFunc) presets.FieldComponentFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
				msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
				comp := in(obj, field, ctx)
				p := obj.(*Page)
				return comp.(*vx.VXFieldBuilder).Label(msgr.Slug).
					Attr(presets.VFieldError(field.FormKey, strings.TrimPrefix(p.Slug, "/"), field.Errors)...).
					Attr("prefix", "/")
			}
		}).LazyWrapSetterFunc(func(in presets.FieldSetterFunc) presets.FieldSetterFunc {
			return func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) (err error) {
				p := obj.(*Page)
				p.Slug = path.Join("/", p.Slug)
				if err = in(obj, field, ctx); err != nil {
					return
				}
				return
			}
		})
	}
	if b.expectField("CategoryID") {
		section.ViewingField("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			p := obj.(*Page)
			var (
				category Category
				err      error
			)
			if category, err = p.GetCategory(db); err != nil {
				panic(err)
			}
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

			return presets.ReadonlyText(obj, field, ctx).
				Label(msgr.Category).
				Value(category.Path)
		})
		section.EditingField("CategoryID").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			var (
				p          = obj.(*Page)
				categories []*Category
				locale, _  = l10n.IsLocalizableFromContext(ctx.R.Context())
			)
			if err := db.Model(&Category{}).Where("locale_code = ?", locale).Find(&categories).Error; err != nil {
				panic(err)
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
	dp.Section(section)
}
