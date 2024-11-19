package pagebuilder

import (
	"cmp"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"github.com/theplant/relay"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

const (
	ParamPage               = "page"
	ParamPerPage            = "per_page"
	ParamTemplateSelectedID = "select_id"
	ParamSearchKeyword      = "keyword"
	iframeCardHeight        = 180
	dialogIframeCardHeight  = 120
	cardContentHeight       = 88
)

type (
	TemplateBuilder struct {
		mb                 *presets.ModelBuilder
		tm                 *presets.ModelBuilder
		model              *ModelBuilder
		builder            *Builder
		useDefaultTemplate bool
	}
	TemplateInterface interface {
		GetName(ctx *web.EventContext) string
		GetDescription(ctx *web.EventContext) string
	}
)

func (b *Builder) template(mb *presets.ModelBuilder, tm *presets.ModelBuilder) {
	b.templates = append(b.templates, &TemplateBuilder{
		mb:      mb,
		tm:      tm,
		builder: b,
	})
}

func (b *Builder) RegisterModelBuilderTemplate(mb *presets.ModelBuilder, tm *presets.ModelBuilder) *Builder {
	if !b.templateEnabled {
		return b
	}
	b.template(mb, tm)
	return b
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
	wrapper := presets.WrapperFieldLabel(func(evCtx *web.EventContext, obj interface{}, field *presets.FieldContext) (name2label map[string]string, err error) {
		msgr := i18n.MustGetModuleMessages(evCtx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		return map[string]string{
			"Name":        msgr.Name,
			"Description": msgr.Description,
		}, nil
	})
	creating.Field("Name").LazyWrapComponentFunc(wrapper)
	creating.Field("Description").LazyWrapComponentFunc(wrapper)

	b.templateModel = template
	b.RegisterModelBuilderTemplate(pm, template)

	return
}

func (b *TemplateBuilder) configList() {
	listing := b.model.mb.Listing()
	oldPageFunc := listing.GetPageFunc()
	listing.PageFunc(func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var pr web.PageResponse
		if pr, err = oldPageFunc(ctx); err != nil {
			return
		}
		r.PageTitle = pr.PageTitle
		ctx.WithContextValue(presets.CtxPageTitleComponent, h.Div(
			VAppBarTitle(h.Text(b.model.mb.Info().LabelName(ctx, false))),
		).Class(W100, "d-flex align-center"))
		r.Body = web.Portal(
			b.templateContent(ctx),
		).Name(PageTemplatePortalName)
		return
	})
}

func (b *TemplateBuilder) configModelWithTemplate() {
	mb := b.mb
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
					Attr("@click", web.Plaid().URL(mb.Info().ListingHref()).EventFunc(OpenTemplateDialogEvent).Query(presets.ParamOverlay, actions.Dialog).Go()),
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
					tplID, _, _ = b.model.primaryColumnValuesBySlug(selectID)
					if b.builder.l10n == nil {
						localeCode = ""
					}
					if err = b.builder.GetModelBuilder(b.mb).copyContainersToAnotherPage(b.builder.db, tplID, "", localeCode, int(pageID.(uint)), version, localeCode, b.model.name, b.builder.GetModelBuilder(b.mb).name); err != nil {
						panic(err)
					}
				}
				return
			}
		})
		filed.ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return h.Components(
				web.Listen(NotifTemplateSelected(b.model.mb),
					web.Plaid().EventFunc(ReloadSelectedTemplateEvent).FieldValue(ParamTemplateSelectedID, web.Var("payload.slug")).Go(),
				),
				web.Portal(b.selectedTemplate(ctx)).Name(TemplateSelectedPortalName),
			)
		})
	}
}

func (b *TemplateBuilder) templateContent(ctx *web.EventContext) h.HTMLComponent {
	var (
		err            error
		cardClickEvent string
		model          = b.model
		mb             = model.mb
		obj            = mb.NewModel()
		ml             = mb.Listing()
		pMsgr          = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
		msgr           = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		perPage        = cmp.Or(int64(ctx.ParamAsInt(ParamPerPage)), presets.PerPageDefault)
		page           = cmp.Or(int64(ctx.ParamAsInt(ParamPage)), 1)
		inDialog       = ctx.Param(presets.ParamOverlay) == actions.Dialog
		cols           = 3
		cardHeight     = iframeCardHeight
	)
	searchParams := &presets.SearchParams{
		Model:          mb.NewModel(),
		PageURL:        ctx.R.URL,
		Keyword:        ctx.Param(ParamSearchKeyword),
		KeywordColumns: []string{"Name"},
		OrderBys:       []relay.OrderBy{{Field: "CreatedAt", Desc: true}},
	}
	searchParams.PerPage = perPage
	searchParams.Page = page
	result, err := ml.Searcher(ctx, searchParams)
	if err != nil {
		panic(errors.Wrap(err, "searcher error"))
	}
	totalCount := int64(0)
	if result.TotalCount != nil {
		totalCount = int64(*result.TotalCount)
	}
	pagesCount := totalCount/perPage + 1
	if totalCount%(perPage) == 0 {
		pagesCount--
	}

	rows := VRow()
	if inDialog {
		cardHeight = dialogIframeCardHeight
		cardClickEvent, cols = b.getEventCols(inDialog, "")
		if page == 1 {
			rows.AppendChildren(VCol(
				VCard(
					VCardItem(
						VCard(
							VCardText(
								VIcon("mdi-plus").Class("mr-1"), h.Text(msgr.AddBlankPage),
							).Class("pa-0", H100, "text-"+ColorPrimary, "text-body-2", "d-flex", "justify-center", "align-center"),
						).Height(cardHeight).Elevation(0).Class("bg-"+ColorGreyLighten4),
					).Class("pa-0", W100),
					VCardText(h.Text(msgr.BlankPage)).Class("text-caption"),
				).Attr("@click", cardClickEvent).Elevation(0),
			).Cols(cols))
		}

	}
	reflectutils.ForEach(result.Nodes, func(obj interface{}) {
		var (
			name        string
			description string
			ps          string
			ojID        interface{}
			menus       []h.HTMLComponent
		)
		name, description = b.getTemplateNameDescription(obj, ctx)
		if p, ok := obj.(presets.SlugEncoder); ok {
			ps = p.PrimarySlug()
		}
		if ojID, err = reflectutils.Get(obj, "ID"); err != nil {
			panic(err)
		}
		cardClickEvent, cols = b.getEventCols(inDialog, ps)
		if mb.Info().Verifier().Do(presets.PermDelete).WithReq(ctx.R).IsAllowed() == nil {
			menus = append(menus,
				VListItem(h.Text(pMsgr.Delete)).Attr("@click", web.Plaid().
					URL(mb.Info().ListingHref()).
					EventFunc(actions.DeleteConfirmation).
					Query(presets.ParamID, ojID).
					Go(),
				))
		}
		if mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() == nil {
			menus = append(menus,
				VListItem(h.Text(pMsgr.Edit)).Attr("@click", web.Plaid().
					URL(mb.Info().ListingHref()).
					EventFunc(actions.Edit).
					Query(presets.ParamID, ojID).
					Go(),
				))
		}
		rows.AppendChildren(
			VCol(
				VCard(
					VCardItem(
						VCard(
							VCardText(
								h.Iframe().Src(model.PreviewHref(ctx, ps)).
									Attr("scrolling", "no", "frameborder", "0").
									Style(`pointer-events: none;transform-origin: 0 0; transform:scale(0.2);width:500%;height:500%`),
							).Class("pa-0", H100, "bg-"+ColorGreyLighten4),
						).Height(cardHeight).Elevation(0),
					).Class("pa-0", W100),
					h.If(!inDialog,
						VCardItem(
							VCard(
								VCardItem(
									h.Components(
										web.Slot(
											h.Text(name),
										).Name("title"),
										web.Slot(
											h.Text(description),
										).Name("subtitle"),
									),
									h.Div(
										h.Div(),
										h.If(!inDialog && len(menus) > 0,
											VMenu(
												web.Slot(
													VBtn("").Children(
														VIcon("mdi-dots-horizontal"),
													).Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
												).Name("activator").Scope("{ props }"),
												VList(
													menus...,
												),
											),
										),
									).Class(W100, "d-flex", "justify-space-between", "align-center"),
								).Class("pa-2"),
							).Color(ColorGreyLighten5).Height(cardContentHeight),
						).Class("pa-0"),
					),
					h.If(inDialog, VCardTitle(h.Text(name)).Class("text-caption")),
				).Attr("@click", cardClickEvent).Elevation(0),
			).Cols(cols),
		)
	})
	simpleReload := web.Plaid().
		URL(mb.Info().ListingHref()).
		EventFunc(ReloadTemplateContentEvent).
		MergeQuery(true).
		Queries(ctx.Queries()).
		Go()
	return h.Div(

		VContainer(
			h.If(!inDialog,
				VRow(
					VCol(
						h.Div(
							b.searchComponent(ctx),
							h.If(mb.Info().Verifier().Do(presets.PermCreate).WithReq(ctx.R).IsAllowed() == nil,
								VBtn(msgr.AddPageTemplate).
									Color(ColorPrimary).
									Variant(VariantElevated).
									Theme("light").
									Attr("@click", web.Plaid().URL(mb.Info().ListingHref()).EventFunc(actions.New).Go()),
							),
						).Class("d-flex justify-space-between align-center"),
					).Cols(12),
				).Class("position-sticky top-0", "bg-"+ColorBackground).Attr("style", "z-index:2"),
			),
			rows,
			h.If(totalCount > perPage,
				VRow(
					VCol(
						VPagination().
							Length(pagesCount).
							TotalVisible(5).
							NextIcon("mdi-page-last").
							PrevIcon("mdi-page-first").
							ModelValue(int(page)).
							Attr("@update:model-value", web.Plaid().
								URL(mb.Info().ListingHref()).
								EventFunc(ReloadTemplateContentEvent).
								Query(presets.ParamOverlay, ctx.Param(presets.ParamOverlay)).
								Query(ParamPage, web.Var("$event")).
								Go()).Class("float-right"),
					).Cols(12),
				).Class("position-sticky bottom-0", "bg-"+ColorBackground),
			)).Fluid(true), web.Listen(
			presets.NotifModelsCreated(obj), simpleReload,
			presets.NotifModelsUpdated(obj), simpleReload,
			presets.NotifModelsDeleted(obj), simpleReload,
		))
}

func (b *TemplateBuilder) searchComponent(ctx *web.EventContext) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
	clickEvent := web.Plaid().PushState(true).MergeQuery(true).Query(ParamSearchKeyword, web.Var("vars.searchMsg")).RunPushState() + ";" + web.Plaid().
		EventFunc(ReloadTemplateContentEvent).
		Query(presets.ParamOverlay, ctx.Param(presets.ParamOverlay)).
		Query(ParamSearchKeyword, web.Var("vars.searchMsg")).Go()

	return vx.VXField().
		Placeholder(msgr.Search).
		HideDetails(true).
		Attr(":clearable", "true").
		Attr("v-model", "vars.searchMsg").
		Attr(web.VAssign("vars", fmt.Sprintf(`{searchMsg:%q}`, ctx.Param(ParamSearchKeyword)))...).
		Attr("@click:clear", `vars.searchMsg="";`+clickEvent).
		Attr("@keyup.enter", clickEvent).
		Children(
			web.Slot(VIcon("mdi-magnify").Attr("@click", clickEvent)).Name("append-inner"),
		).Width(320)
}

func (b *TemplateBuilder) selectedTemplate(ctx *web.EventContext) h.HTMLComponent {
	var (
		msgr     = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		template = b.model.mb.NewModel()
		selectID = ctx.Param(ParamTemplateSelectedID)
		err      error
		name     string
	)
	if selectID != "" {
		if err = utils.PrimarySluggerWhere(b.builder.db, template, selectID).First(template).Error; err != nil {
			panic(err)
		}
		name, _ = b.getTemplateNameDescription(template, ctx)
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
					Query(templateSelectedID, selectID).
					Query(presets.ParamOverlay, actions.Dialog).
					EventFunc(OpenTemplateDialogEvent).Go()),
		VCard(
			VCardText(
				h.Iframe().Src(b.model.PreviewHref(ctx, selectID)).
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

func (b *TemplateBuilder) getEventCols(inDialog bool, ps string) (cardClickEvent string, cols int) {
	cols = 3
	if inDialog {
		cols = 4
		emit := web.Emit(NotifTemplateSelected(b.model.mb), TemplateSelected{Slug: ps})
		cardClickEvent = fmt.Sprintf("%s;vars.pageBuilderSelectTemplateDialog=false;if(!vars.presetsDialog){%s};",
			emit,
			web.Plaid().URL(b.mb.Info().ListingHref()).EventFunc(actions.New).FieldValue(ParamTemplateSelectedID, ps).Query(presets.ParamOverlay, actions.Dialog).Go())
	} else {
		cardClickEvent = web.Plaid().URL(b.model.editorURLWithSlug(ps)).PushState(true).Go()
	}
	return
}

func (b *TemplateBuilder) getTemplateNameDescription(obj interface{}, ctx *web.EventContext) (name, description string) {
	if p, ok := obj.(TemplateInterface); ok {
		name = p.GetName(ctx)
		description = p.GetDescription(ctx)
		return
	}
	if v, err := reflectutils.Get(obj, "Name"); err == nil {
		name = v.(string)
	}
	if v, err := reflectutils.Get(obj, "Description"); err == nil {
		description = v.(string)
	}
	return
}

func (b *TemplateBuilder) Install() {
	builder := b.builder
	tm := b.tm
	if tm == nil {
		tm = builder.templateModel
	}
	defer builder.useAllPlugin(tm)
	model := builder.GetModelBuilder(tm)
	if model == nil {
		model = builder.Model(tm)
		if _, ok := tm.NewModel().(publish.VersionInterface); ok {
			panic("error template model")
		}
		builder.configEditor(model)
	}
	b.model = model
	model.tb = b
	b.configModelWithTemplate()
	b.configList()
	b.registerFunctions()
}
