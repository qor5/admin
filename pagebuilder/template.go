package pagebuilder

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

const (
	ParamTemplateSelectedID = "select_id"
	iframeCardHeight        = 180
	dialogIframeCardHeight  = 120
	cardContentHeight       = 88
	cardDialogContentHeight = 34
)

type (
	TemplateBuilder struct {
		tm      *ModelBuilder
		builder *Builder
	}
)

func (b *Builder) template(tm *presets.ModelBuilder) *TemplateBuilder {
	r := &TemplateBuilder{
		tm:      b.Model(tm),
		builder: b,
	}
	r.tm.isTemplate = true
	return r
}

func (b *Builder) defaultTemplateInstall(pb *presets.Builder, pm *presets.ModelBuilder) (err error) {
	if !b.templateEnabled {
		return
	}
	template := pb.Model(&Template{}).URIName("page_templates").Label("Templates")
	defer func() {
		if b.ab != nil {
			template.Use(b.ab)
		}
	}()
	template.LabelName(func(evCtx *web.EventContext, singular bool) string {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if singular {
			return msgr.ModelLabelTemplate
		}
		return msgr.ModelLabelTemplates
	})
	creating := template.Editing("Name", "Description").ValidateFunc(func(obj interface{}, ctx *web.EventContext) (err web.ValidationErrors) {
		p := obj.(*Template)

		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

		if p.Name == "" {
			err.FieldError("Name", msgr.InvalidNameMsg)
			return
		}
		return
	})
	wrapper := presets.WrapperFieldLabel(func(evCtx *web.EventContext, _ interface{}, _ *presets.FieldContext) (name2label map[string]string, err error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"Name":        msgr.Name,
			"Description": msgr.Description,
		}, nil
	})
	creating.Field("Name").LazyWrapComponentFunc(wrapper)
	creating.Field("Description").LazyWrapComponentFunc(wrapper)

	b.templateBuilder = b.template(template)
	pm.Use(b.templateBuilder)
	return
}

func (b *TemplateBuilder) configModelWithTemplate(mb *presets.ModelBuilder) {
	creating := mb.Editing().Creating()
	filed := creating.GetField(PageTemplateSelectionFiled)
	if filed != nil && filed.GetCompFunc() == nil {
		mb.Listing().NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
			msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
			return h.Components(
				web.Portal().Name(TemplateSelectDialogPortalName),
				VBtn(msgr.New).
					Color(ColorPrimary).
					Variant(VariantElevated).
					Theme("light").Class("ml-2").
					Attr("@click", web.Plaid().URL(b.tm.mb.Info().ListingHref()).EventFunc(actions.OpenListingDialog).Query(presets.ParamOverlay, actions.Dialog).Go()),
			)
		})

		creating.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
			return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
				if err = in(obj, id, ctx); err != nil {
					return
				}
				var (
					selectID            = ctx.Param(ParamTemplateSelectedID)
					pageID              interface{}
					localeCode, version string
				)
				if p, ok := obj.(l10n.LocaleInterface); ok {
					localeCode = p.EmbedLocale().LocaleCode
				}
				if p, ok := obj.(publish.VersionInterface); ok {
					version = p.EmbedVersion().Version
					pageID = reflectutils.MustGet(obj, "ID")
				}
				if selectID != "" {
					var tplID int

					tplID, _, _ = b.tm.primaryColumnValuesBySlug(selectID)
					if b.builder.l10n == nil {
						localeCode = ""
					}
					if err = b.builder.GetModelBuilder(mb).copyContainersToAnotherPage(b.builder.db, tplID, "", localeCode, int(pageID.(uint)), version, localeCode, b.tm.name, b.builder.GetModelBuilder(mb).name); err != nil {
						panic(err)
					}
				}
				return
			}
		})
		filed.ComponentFunc(func(_ interface{}, _ *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Components(
				web.Listen(NotifTemplateSelected(b.tm.mb),
					web.Plaid().EventFunc(ReloadSelectedTemplateEvent).FieldValue(ParamTemplateSelectedID, web.Var("payload.slug")).Go(),
				),
				web.Portal(b.selectedTemplate(ctx)).Name(TemplateSelectedPortalName),
			)
		})
	}
}

func (b *TemplateBuilder) selectedTemplate(ctx *web.EventContext) h.HTMLComponent {
	var (
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		template    = b.tm.mb.NewModel()
		selectID    = ctx.Param(ParamTemplateSelectedID)
		err         error
		name        string
		previewHref string
	)
	if selectID != "" {
		if err = utils.PrimarySluggerWhere(b.builder.db, template, selectID).First(template).Error; err != nil {
			panic(err)
		}
		p := template.(*Template)
		name = p.Name
		previewHref = b.tm.PreviewHref(ctx, selectID)

	}
	return h.Div(
		h.Div(
			h.Span(msgr.ModelLabelTemplate).Class("text-body-1"),
		),
		VBtn(msgr.ChangeTemplate).Color(ColorPrimary).
			Variant(VariantTonal).
			PrependIcon("mdi-cached").
			Attr("@click",
				web.Plaid().
					URL(b.tm.mb.Info().ListingHref()).
					Query(templateSelectedID, selectID).
					Query(presets.ParamOverlay, actions.Dialog).
					EventFunc(actions.OpenListingDialog).Go()),
		VCard(
			VCardText(
				h.Iframe().Src(previewHref).
					Attr("scrolling", "no", "frameborder", "0").
					Style(`pointer-events: none;transform-origin: 0 0; transform:scale(0.2);width:500%;height:500%`),
			).Class("pa-0", H100, "border-xl"),
		).Height(106).Width(224).Elevation(0).Class("mt-2"),
		h.Div(
			h.Span(name).Class("text-caption"),
		).Class("mt-2"),
	).Class("mb-6").Attr(web.VAssign("form", fmt.Sprintf(`{%s:%q}`, ParamTemplateSelectedID, selectID))...)
}

type TemplateSelected struct {
	Slug string `json:"slug"`
}

func NotifTemplateSelected(mb *presets.ModelBuilder) string {
	return fmt.Sprintf("pagebuilder_NotifTemplateSelected_%T", mb.NewModel())
}

func (b *TemplateBuilder) ModelInstall(_ *presets.Builder, mb *presets.ModelBuilder) error {
	b.configModelWithTemplate(mb)
	b.registerFunctions(mb)
	return nil
}

func (b *TemplateBuilder) Install(_ *presets.Builder) error {
	builder := b.builder
	tm := b.tm
	builder.useAllPlugin(tm.mb, tm.name)
	builder.configEditor(tm)
	b.configList()
	return nil
}

func (b *TemplateBuilder) configList() {
	var (
		listing = b.tm.mb.Listing().SearchColumns("Name")
		config  = &presets.CardDataTableConfig{}
	)
	defer listing.DataTableFunc(presets.CardDataTableFunc(listing, config))
	listing.NewButtonFunc(func(ctx *web.EventContext) h.HTMLComponent {
		var (
			lc   = presets.ListingCompoFromContext(ctx.R.Context())
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		)
		if lc.Popup {
			return nil
		}
		return VBtn(msgr.AddPageTemplate).
			Color(ColorPrimary).
			Variant(VariantElevated).
			Theme("light").Class("ml-2").
			Attr("@click", web.Plaid().EventFunc(actions.New).Go())
	})
	listing.DialogWidth(templateDialogWidth).Title(func(ctx *web.EventContext, _ presets.ListingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		if ctx.Param(web.EventFuncIDName) == actions.OpenListingDialog {
			return msgr.CreateFromTemplate, nil, nil
		}
		return defaultTitle, nil, nil
	})
	rowMenu := listing.RowMenu()
	rowMenu.RowMenuItem("Edit").ComponentFunc(func(_ interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
		var (
			pMsgr = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
			mb    = b.tm.mb
		)
		if mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
			return nil
		}

		return VListItem(VListItemTitle(h.Text(pMsgr.Edit))).PrependIcon("mdi-pencil").Attr("@click", web.Plaid().
			URL(mb.Info().ListingHref()).
			EventFunc(actions.Edit).
			Query(presets.ParamID, id).
			Go(),
		)
	})
	rowMenu.RowMenuItem("Localize").ComponentFunc(func(_ interface{}, _ string, _ *web.EventContext) h.HTMLComponent {
		return nil
	})
	config.CardTitle = func(ctx *web.EventContext, obj interface{}) (h.HTMLComponent, int) {
		var (
			cardHeight = iframeCardHeight
			lc         = presets.ListingCompoFromContext(ctx.R.Context())
		)
		if lc.Popup {
			cardHeight = dialogIframeCardHeight
		}
		return h.Iframe().Src(b.tm.PreviewHref(ctx, presets.ObjectID(obj))).
			Attr("scrolling", "no", "frameborder", "0").
			Style(`pointer-events: none;transform-origin: 0 0; transform:scale(0.2);width:500%;height:500%`), cardHeight
	}
	config.CardContent = func(ctx *web.EventContext, obj interface{}) (content h.HTMLComponent, height int) {
		var (
			lc = presets.ListingCompoFromContext(ctx.R.Context())
			p  = obj.(*Template)
		)
		if lc.Popup {
			return VCardTitle(h.Text(p.Name)).Class("text-caption"), cardDialogContentHeight
		} else {
			return h.Components(
				VCardTitle(h.Text(p.Name)),
				VCardSubtitle(h.Text(p.Description)),
			), cardContentHeight
		}
	}
	config.Cols = func(ctx *web.EventContext) int {
		lc := presets.ListingCompoFromContext(ctx.R.Context())
		if lc.Popup {
			return 4
		}
		return 3
	}
	config.WrapMultipleSelectedActions = func(_ *web.EventContext, _ h.HTMLComponents) h.HTMLComponents {
		return nil
	}
	config.ClickCardEvent = func(ctx *web.EventContext, obj interface{}) string {
		var (
			lc = presets.ListingCompoFromContext(ctx.R.Context())
			ps = presets.ObjectID(obj)
		)
		if lc.Popup {
			emit := web.Emit(NotifTemplateSelected(b.tm.mb), TemplateSelected{Slug: ps})
			return fmt.Sprintf("%s;%s;if(!vars.presetsDialog){%s};",
				emit,
				presets.CloseListingDialogVarScript,
				web.Plaid().EventFunc(actions.New).FieldValue(ParamTemplateSelectedID, ps).Query(presets.ParamOverlay, actions.Dialog).Go())
		} else {
			return web.Plaid().URL(b.tm.editorURLWithSlug(ps)).PushState(true).Go()
		}
	}
	config.WrapRows = func(ctx *web.EventContext, searchParams *presets.SearchParams, _ *presets.SearchResult, rows *VRowBuilder) *VRowBuilder {
		var (
			lc   = presets.ListingCompoFromContext(ctx.R.Context())
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		)
		if lc.Popup {
			emit := web.Emit(NotifTemplateSelected(b.tm.mb), TemplateSelected{Slug: ""})
			cardClickEvent := fmt.Sprintf("%s;%s;if(!vars.presetsDialog){%s};",
				emit,
				presets.CloseListingDialogVarScript,
				web.Plaid().EventFunc(actions.New).FieldValue(ParamTemplateSelectedID, "").Query(presets.ParamOverlay, actions.Dialog).Go())
			if searchParams.Page == 1 {
				rows.PrependChildren(
					VCol(
						VCard(
							VCardItem(
								VCard(
									VCardText(
										VIcon("mdi-plus").Class("mr-1"), h.Text(msgr.AddBlankPage),
									).Class("pa-0", H100, "text-"+ColorPrimary, "text-body-2", "d-flex", "justify-center", "align-center"),
								).Height(dialogIframeCardHeight).Elevation(0).Class("bg-"+ColorGreyLighten4),
							).Class("pa-0", W100),
							VCardItem(
								VCard(
									VCardItem(
										h.Div(
											h.Div(VCardTitle(h.Text(msgr.BlankPage)).Class("text-caption")),
										).Class(W100, "d-flex", "justify-space-between", "align-center"),
									).Class("pa-2"),
								).Color(ColorGreyLighten5).Height(cardDialogContentHeight),
							).Class("pa-0"),
						).Attr("@click", cardClickEvent).Elevation(0),
					).Cols(4))
			}
		}
		return rows
	}
	config.WrapRooters = func(ctx *web.EventContext, footers h.HTMLComponents) h.HTMLComponents {
		lc := presets.ListingCompoFromContext(ctx.R.Context())
		if lc.Popup {
			return nil
		}
		return footers
	}
	config.RemainingHeight = func(ctx *web.EventContext) string {
		lc := presets.ListingCompoFromContext(ctx.R.Context())
		if lc.Popup {
			return "400px"
		}
		return "180px"
	}
}
