package presets

import (
	"cmp"
	"errors"
	"fmt"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	v "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets/actions"
)

type (
	DetailingStyle          string
	DetailingLayout         string
	DetailingBreadcrumbFunc func(ctx *web.EventContext, obj any, id string) (BreadcrumbItemsFunc, error)
)

const (
	DetailingStylePage   DetailingStyle = "Page"
	DetailingStyleDrawer DetailingStyle = "Drawer"
	DetailingStyleDialog DetailingStyle = "Dialog"
)

const (
	LayoutCenter DetailingLayout = "layout-center"
)

type DetailingBuilder struct {
	mb                       *ModelBuilder
	actions                  []*ActionBuilder
	pageFunc                 web.PageFunc
	fetcher                  FetchFunc
	tabPanels                []TabComponentFunc
	sidePanel                ObjectComponentFunc
	titleFunc                func(evCtx *web.EventContext, obj any, style DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error)
	afterTitleCompFunc       ObjectComponentFunc
	drawer                   bool
	layouts                  []DetailingLayout
	idCurrentActiveProcessor IdCurrentActiveProcessor
	FieldsBuilder
	breadcrumbFunc DetailingBreadcrumbFunc
}

type pageTitle interface {
	PageTitle() string
}

// Detailing configures the detailing builder with the given fields.
// Accepts string / []string / *FieldsSection
func (mb *ModelBuilder) Detailing(vs ...interface{}) (r *DetailingBuilder) {
	r = mb.detailing
	if !mb.hasDetailing && len(vs) == 0 {
		vs = mb.editing.FieldNames()
	}

	mb.hasDetailing = true

	if len(vs) == 0 {
		return
	}

	r.Only(vs...)
	return r
}

func (b *DetailingBuilder) GetDrawer() bool {
	return b.drawer
}

// ContainerClass let u easier to adjust the detailing page by each project
func (b *DetailingBuilder) ContainerClass(layoutVal DetailingLayout) (r *DetailingBuilder) {
	b.layouts = append(b.layouts, layoutVal)
	return b
}

// Only specifies which fields to include in the detailing builder.
// Accepts string / []string / *FieldsSection
func (b *DetailingBuilder) Only(vs ...interface{}) (r *DetailingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Only(vs...)
	return
}

func (b *DetailingBuilder) Prepend(vs ...interface{}) (r *DetailingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Prepend(vs...)
	return
}

func (b *DetailingBuilder) Except(vs ...string) (r *DetailingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Except(vs...)
	return
}

func (b *DetailingBuilder) PageFunc(pf web.PageFunc) (r *DetailingBuilder) {
	b.pageFunc = pf
	return b
}

func (b *DetailingBuilder) WrapPageFunc(w func(in web.PageFunc) web.PageFunc) (r *DetailingBuilder) {
	b.pageFunc = w(b.pageFunc)
	return b
}

func (b *DetailingBuilder) FetchFunc(v FetchFunc) (r *DetailingBuilder) {
	b.fetcher = v
	return b
}

func (b *DetailingBuilder) WrapFetchFunc(w func(in FetchFunc) FetchFunc) (r *DetailingBuilder) {
	b.fetcher = w(b.fetcher)
	return b
}

func (b *DetailingBuilder) GetFetchFunc() FetchFunc {
	return b.fetcher
}

func (b *DetailingBuilder) Drawer(v bool) (r *DetailingBuilder) {
	b.drawer = v
	return b
}

// Title must not return empty, and titleCompo can return nil
func (b *DetailingBuilder) Title(f func(evCtx *web.EventContext, obj any, style DetailingStyle, defaultTitle string) (title string, titleCompo h.HTMLComponent, err error)) (r *DetailingBuilder) {
	b.titleFunc = f
	return b
}

func (b *DetailingBuilder) AfterTitleCompFunc(v ObjectComponentFunc) (r *DetailingBuilder) {
	if v == nil {
		panic("value required")
	}
	b.afterTitleCompFunc = v
	return b
}

func (b *DetailingBuilder) GetPageFunc() web.PageFunc {
	return b.pageFunc
}

func (b *DetailingBuilder) AppendTabsPanelFunc(v TabComponentFunc) (r *DetailingBuilder) {
	b.tabPanels = append(b.tabPanels, v)
	return b
}

func (b *DetailingBuilder) TabsPanelFunc() (r []TabComponentFunc) {
	return b.tabPanels
}

func (b *DetailingBuilder) TabsPanels(vs ...TabComponentFunc) (r *DetailingBuilder) {
	b.tabPanels = vs
	return b
}

func (b *DetailingBuilder) SidePanelFunc(v ObjectComponentFunc) (r *DetailingBuilder) {
	b.sidePanel = v
	return b
}

func (b *DetailingBuilder) WrapSidePanel(w func(in ObjectComponentFunc) ObjectComponentFunc) (r *DetailingBuilder) {
	if b.sidePanel == nil {
		b.sidePanel = func(_ interface{}, _ *web.EventContext) h.HTMLComponent {
			return nil
		}
	}
	b.sidePanel = w(b.sidePanel)
	return b
}

type ctxKeyDetailingStyle struct{}

func (b *DetailingBuilder) defaultPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	id := ctx.Param(ParamID)
	r.Body = v.VContainer(h.Text(id))

	obj := b.mb.NewModel()

	if id == "" {
		panic("not found")
	}

	obj, err = b.GetFetchFunc()(obj, id, ctx)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return b.mb.p.DefaultNotFoundPageFunc(ctx)
		}
		return
	}

	if b.mb.Info().Verifier().Do(PermGet).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		r.Body = h.Div(h.Text(perm.PermissionDenied.Error()))
		return
	}

	msgr := b.mb.mustGetMessages(ctx.R)
	title := msgr.DetailingObjectTitle(b.mb.Info().LabelName(ctx, true), getPageTitle(obj, id))
	if b.titleFunc != nil {
		style, ok := ctx.ContextValue(ctxKeyDetailingStyle{}).(DetailingStyle)
		if !ok {
			style = DetailingStylePage
		}

		title, titleCompo, err := b.titleFunc(ctx, obj, style, title)
		if err != nil {
			return r, err
		}
		if titleCompo != nil {
			ctx.WithContextValue(CtxPageTitleComponent, titleCompo)
		}
		r.PageTitle = title
	} else {
		r.PageTitle = title
	}
	if b.afterTitleCompFunc != nil {
		ctx.WithContextValue(ctxDetailingAfterTitleComponent, b.afterTitleCompFunc(obj, ctx))
	}

	var notice h.HTMLComponent
	if msg, ok := ctx.Flash.(string); ok {
		notice = v.VSnackbar(
			h.Div().Style("white-space: pre-wrap").Text(fmt.Sprintf(`{{%q}}`, msg)),
		).ModelValue(true).Location("top").Color("success")
	}

	comp := web.Scope(
		b.ToComponent(b.mb.Info(), obj, ctx),
	).VSlot("{form}")
	tabsContent := defaultToPage(commonPageConfig{
		formContent: comp,
		tabPanels:   b.tabPanels,
		sidePanel:   b.sidePanel,
	}, obj, ctx)

	var actionButtons []h.HTMLComponent
	for _, ba := range b.actions {
		if b.mb.Info().Verifier().SnakeDo(permActions, ba.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		if ba.buttonCompFunc != nil {
			actionButtons = append(actionButtons, ba.buttonCompFunc(ctx))
			continue
		}

		actionButtons = append(actionButtons, v.VBtn(b.mb.getLabel(ba.NameLabel)).
			Color(cmp.Or(ba.buttonColor, v.ColorPrimary)).Variant(v.VariantFlat).
			Attr("@click", web.Plaid().
				EventFunc(actions.Action).
				Query(ParamID, id).
				Query(ParamAction, ba.name).
				URL(b.mb.Info().DetailingHref(id)).
				Go(),
			),
		)
	}
	var actionButtonsCompo h.HTMLComponent
	if len(actionButtons) > 0 {
		actionButtonsCompo = h.Div(v.VSpacer()).Class("d-flex flex-row ga-2").AppendChildren(actionButtons...)
	}

	layoutClass := make([]string, len(b.layouts))
	for i, layout := range b.layouts {
		layoutClass[i] = string(layout)
	}
	if b.breadcrumbFunc != nil {
		itemFunc, err := b.breadcrumbFunc(ctx, obj, id)
		if err != nil {
			return r, err
		}
		ctx.WithContextValue(BreadcrumbItemsFuncKey{}, itemFunc)
	}
	r.Body = v.VContainer().Children(
		notice,
		h.Div().Class("d-flex flex-column", strings.Join(layoutClass, ", ")).Children(
			actionButtonsCompo,
			tabsContent,
		),
	).Fluid(true).Class("px-0 pt-0 detailing-page-wrap")

	return
}

func (b *DetailingBuilder) WrapIdCurrentActive(w func(IdCurrentActiveProcessor) IdCurrentActiveProcessor) (r *DetailingBuilder) {
	if b.idCurrentActiveProcessor == nil {
		b.idCurrentActiveProcessor = w(func(_ *web.EventContext, current string) (string, error) {
			return current, nil
		})
	} else {
		b.idCurrentActiveProcessor = w(b.idCurrentActiveProcessor)
	}
	return b
}

func (b *DetailingBuilder) showInDrawer(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermGet).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}
	onChangeEvent := fmt.Sprintf("if (vars.%s) { vars.%s.detailing=true };", VarsPresetsDataChanged, VarsPresetsDataChanged)

	overlayType := ctx.R.FormValue(ParamOverlay)
	closeBtnVarScript := CloseRightDrawerVarConfirmScript
	style := DetailingStyleDrawer
	if overlayType == actions.Dialog {
		closeBtnVarScript = CloseDialogVarScript
		style = DetailingStyleDialog
	}
	ctx.WithContextValue(ctxKeyDetailingStyle{}, style)
	pr, err := b.GetPageFunc()(ctx)
	if err != nil {
		return
	}
	titleCompo, ok := ctx.ContextValue(CtxPageTitleComponent).(h.HTMLComponent)
	if !ok {
		titleCompo = h.Text(pr.PageTitle)
	} else {
		ctx.WithContextValue(CtxPageTitleComponent, nil)
	}
	header := h.Div(titleCompo).Class("d-flex")
	if val, ok := GetComponentFromContext(ctx, ctxDetailingAfterTitleComponent); ok {
		header.AppendChildren(v.VSpacer(), val)
	}

	comp := web.Scope(
		v.VLayout(
			v.VAppBar(
				v.VAppBarTitle(header).Class("pl-2 drawer-title"),
				v.VBtn("").Icon("mdi-close").
					Attr("@click.stop", closeBtnVarScript),
			).Color("white").Elevation(0),

			v.VMain(
				v.VSheet(
					v.VCard(pr.Body).Flat(true).Class("pa-1"),
				).Class("pa-2"),
			),
		),
	).VSlot("{ form }").OnChange(onChangeEvent).UseDebounce(150)

	if b.idCurrentActiveProcessor != nil {
		ctx.WithContextValue(ctxKeyIdCurrentActiveProcessor{}, b.idCurrentActiveProcessor)
	}
	b.mb.p.overlay(ctx, &r, comp, b.mb.rightDrawerWidth)
	return
}

func getPageTitle(obj interface{}, id string) string {
	title := id
	if pt, ok := obj.(pageTitle); ok {
		title = pt.PageTitle()
	}
	return title
}

func (b *DetailingBuilder) fetchAction(ctx *web.EventContext, name string) (*ActionBuilder, error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamAction))
	if action == nil {
		return nil, errors.New("cannot find requested action")
	}

	if action.updateFunc == nil {
		return nil, errors.New("action.updateFunc not set")
	}

	if action.compFunc == nil {
		return nil, errors.New("action.compFunc not set")
	}

	err := b.mb.Info().Verifier().SnakeDo(permActions, name).WithReq(ctx.R).IsAllowed()
	if err != nil {
		return nil, err
	}

	return action, nil
}

func (b *DetailingBuilder) doAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	action, err := b.fetchAction(ctx, ctx.R.FormValue(ParamAction))
	if err != nil {
		ShowMessage(&r, err.Error(), v.ColorError)
		return r, nil
	}

	id := ctx.R.FormValue(ParamID)
	if err := action.updateFunc(id, ctx, &r); err != nil || ctx.Flash != nil {
		if ctx.Flash == nil {
			ctx.Flash = err
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: dialogContentPortalName,
			Body: b.actionForm(action, ctx),
		})
		return r, nil
	}
	web.AppendRunScripts(&r, CloseDialogVarScript)
	return
}

func (b *DetailingBuilder) openActionDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	action, err := b.fetchAction(ctx, ctx.R.FormValue(ParamAction))
	if err != nil {
		ShowMessage(&r, err.Error(), v.ColorError)
		return r, nil
	}

	b.mb.p.dialog(ctx, &r, b.actionForm(action, ctx), "")
	return
}

func (b *DetailingBuilder) actionForm(action *ActionBuilder, ctx *web.EventContext) h.HTMLComponent {
	msgr := b.mb.mustGetMessages(ctx.R)

	id := ctx.R.FormValue(ParamID)
	if id == "" {
		panic("id required")
	}

	return v.VContainer(
		v.VCard(
			v.VCardText(
				action.compFunc(id, ctx),
			),
			v.VCardActions(
				v.VSpacer(),
				v.VBtn(msgr.Update).
					Theme("light").
					Color(v.ColorPrimary).
					Attr("@click", web.Plaid().
						EventFunc(actions.DoAction).
						Query(ParamID, id).
						Query(ParamAction, ctx.R.FormValue(ParamAction)).
						URL(b.mb.Info().DetailingHref(id)).
						Go()),
			),
		).Flat(true),
	).Fluid(true)
}

const fieldRefreshOnUpdate = "__RefreshOnUpdate__"

func (b *DetailingBuilder) EnableRefreshOnUpdate() *DetailingBuilder {
	b.Field(fieldRefreshOnUpdate).ComponentFunc(func(obj interface{}, _ *FieldContext, ctx *web.EventContext) h.HTMLComponent {
		slug := obj.(SlugEncoder).PrimarySlug()

		qs := ctx.R.URL.Query()
		eventFuncID := qs.Get(web.EventFuncIDName)
		delete(qs, web.EventFuncIDName)

		refresh := web.Plaid().URL(ctx.R.URL.Path).Queries(qs)
		if eventFuncID != "" {
			refresh.EventFunc(eventFuncID)
		}
		return web.Listen(b.mb.NotifModelsUpdated(), fmt.Sprintf(`payload.ids.includes(%q) && %s`, slug, refresh.Go()))
	})
	return b
}

func (b *DetailingBuilder) Section(sections ...*SectionBuilder) *DetailingBuilder {
	for _, sb := range sections {
		if sb.isUsed {
			panic("section is used")
		}
		sb.isUsed = true
		sb.registerEvent()
		sb.isEdit = false

		b.Field(sb.name).Component(sb)
	}
	return b
}

func (b *DetailingBuilder) defaultBreadcrumbFunc(ctx *web.EventContext, obj any, id string) (BreadcrumbItemsFunc, error) {
	var (
		msgr      = b.mb.mustGetMessages(ctx.R)
		titleComp h.HTMLComponent
		title     = msgr.DetailingObjectTitle(b.mb.Info().LabelName(ctx, true), getPageTitle(obj, id))
	)
	if b.titleFunc != nil {
		style, ok := ctx.ContextValue(ctxKeyDetailingStyle{}).(DetailingStyle)
		if !ok {
			style = DetailingStylePage
		}
		xtitle, xtitleComp, err := b.titleFunc(ctx, obj, style, title)
		if err != nil {
			return nil, err
		}
		if xtitleComp != nil {
			titleComp = xtitleComp
		}
		if xtitle != "" {
			title = xtitle
		}
	}
	if titleComp == nil {
		titleComp = h.Text(title)
	}
	return func(ctx *web.EventContext, disableLast bool) (r []h.HTMLComponent) {
		listingHref := b.mb.Info().ListingHref()
		r = []h.HTMLComponent{
			v.VBreadcrumbsItem(h.Text(b.mb.Info().LabelName(ctx, false))).
				Href(listingHref),
		}
		if b.mb.hasDetailing && !b.drawer {
			detailingHref := b.mb.Info().DetailingHref(ctx.Param(ParamID))
			r = append(r, v.VBreadcrumbsItem(titleComp).
				Href(detailingHref).
				Disabled(disableLast))
		}
		return r
	}, nil
}

func (b *DetailingBuilder) Breadcrumb(f DetailingBreadcrumbFunc) *DetailingBuilder {
	b.breadcrumbFunc = f
	return b
}

func (b *DetailingBuilder) GetBreadcrumb() DetailingBreadcrumbFunc {
	return b.breadcrumbFunc
}

func (b *DetailingBuilder) WrapBreadcrumb(w func(DetailingBreadcrumbFunc) DetailingBreadcrumbFunc) *DetailingBuilder {
	b.breadcrumbFunc = w(b.breadcrumbFunc)
	return b
}
