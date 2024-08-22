package presets

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
)

type (
	DetailingStyle  string
	DetailingLayout string
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
	SectionsBuilder
}

type pageTitle interface {
	PageTitle() string
}

// string / []string / *FieldsSection
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

// let u easier to adjust the detailing page by each project
func (b *DetailingBuilder) ContainerClass(layoutVal DetailingLayout) (r *DetailingBuilder) {
	b.layouts = append(b.layouts, layoutVal)
	return b
}

// string / []string / *FieldsSection
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

// The title must not return empty, and titleCompo can return nil
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
	if b.pageFunc != nil {
		return b.pageFunc
	}
	return b.defaultPageFunc
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

type ctxKeyDetailingStyle struct{}

func (b *DetailingBuilder) defaultPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	id := ctx.Param(ParamID)
	r.Body = VContainer(h.Text(id))

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

	msgr := MustGetMessages(ctx.R)
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
		notice = VSnackbar(h.Text(msg)).ModelValue(true).Location("top").Color("success")
	}

	comp := web.Scope(b.ToComponent(b.mb.Info(), obj, ctx)).VSlot("{form}")
	var tabsContent h.HTMLComponent = defaultToPage(commonPageConfig{
		formContent: comp,
		tabPanels:   b.tabPanels,
		sidePanel:   b.sidePanel,
	}, obj, ctx)

	actionButtons := []h.HTMLComponent{}
	for _, ba := range b.actions {
		if b.mb.Info().Verifier().SnakeDo(permActions, ba.name).WithReq(ctx.R).IsAllowed() != nil {
			continue
		}

		if ba.buttonCompFunc != nil {
			actionButtons = append(actionButtons, ba.buttonCompFunc(ctx))
			continue
		}

		actionButtons = append(actionButtons, VBtn(b.mb.getLabel(ba.NameLabel)).
			Color(cmp.Or(ba.buttonColor, ColorPrimary)).Variant(VariantFlat).
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
		actionButtonsCompo = h.Div(VSpacer()).Class("d-flex flex-row ga-2").AppendChildren(actionButtons...)
	}

	layoutClass := make([]string, len(b.layouts))
	for i, layout := range b.layouts {
		layoutClass[i] = string(layout)
	}

	r.Body = VContainer().Children(
		notice,
		h.Div().Class("d-flex flex-column ga-10", strings.Join(layoutClass, ", ")).Children(
			actionButtonsCompo,
			tabsContent,
		),
	).Fluid(true).Class("px-0 pt-0")

	return
}

func (b *DetailingBuilder) WrapIdCurrentActive(w func(IdCurrentActiveProcessor) IdCurrentActiveProcessor) (r *DetailingBuilder) {
	if b.idCurrentActiveProcessor == nil {
		b.idCurrentActiveProcessor = w(func(ctx *web.EventContext, current string) (string, error) {
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
	if v, ok := GetComponentFromContext(ctx, ctxDetailingAfterTitleComponent); ok {
		header.AppendChildren(VSpacer(), v)
	}

	comp := web.Scope(
		VLayout(
			VAppBar(
				VAppBarTitle(header).Class("pl-2"),
				VBtn("").Icon("mdi-close").
					Attr("@click.stop", closeBtnVarScript),
			).Color("white").Elevation(0),

			VMain(
				VSheet(
					VCard(pr.Body).Flat(true).Class("pa-1"),
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

func (b *DetailingBuilder) doAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamAction))
	if action == nil {
		panic("action required")
	}
	id := ctx.R.FormValue(ParamID)
	if err := action.updateFunc(id, ctx, &r); err != nil || ctx.Flash != nil {
		if ctx.Flash == nil {
			ctx.Flash = err
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: RightDrawerContentPortalName,
			Body: b.actionForm(action, ctx),
		})
		return r, nil
	}
	web.AppendRunScripts(&r, CloseDialogVarScript)
	return
}

func (b *DetailingBuilder) openActionDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamAction))
	if action == nil {
		panic("action required")
	}

	b.mb.p.dialog(ctx, &r, b.actionForm(action, ctx), "")
	return
}

func (b *DetailingBuilder) actionForm(action *ActionBuilder, ctx *web.EventContext) h.HTMLComponent {
	msgr := MustGetMessages(ctx.R)

	id := ctx.R.FormValue(ParamID)
	if id == "" {
		panic("id required")
	}

	return VContainer(
		VCard(
			VCardText(
				action.compFunc(id, ctx),
			),
			VCardActions(
				VSpacer(),
				VBtn(msgr.Update).
					Theme("light").
					Color(ColorPrimary).
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

// EditDetailField EventFunc: click detail field component edit button
func (b *DetailingBuilder) EditDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	key := ctx.Queries().Get(SectionFieldName)

	f := b.Section(key)

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, ctx.Param(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.editComponent(obj, &FieldContext{
			ModelInfo: b.mb.modelInfo,
			FormKey:   f.name,
			Name:      f.name,
			Label:     f.label,
		}, ctx),
	})
	return r, nil
}

// SaveDetailField EventFunc: click save button
func (b *DetailingBuilder) SaveDetailField(ctx *web.EventContext) (r web.EventResponse, err error) {
	key := ctx.Param(SectionFieldName)
	id := ctx.Param(ParamID)
	isCancel := ctx.ParamAsBool(SectionIsCancel)

	f := b.Section(key)

	field := &FieldContext{
		ModelInfo: b.mb.modelInfo,
		FormKey:   f.name,
		Name:      f.name,
		Label:     f.label,
	}

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, id, ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if isCancel {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: f.FieldPortalName(),
			Body: f.viewComponent(obj, field, ctx),
		})
		return
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	err = f.saver(obj, id, ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	if _, ok := ctx.Flash.(*web.ValidationErrors); ok {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: f.FieldPortalName(),
			Body: f.editComponent(obj, field, ctx),
		})
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.viewComponent(obj, field, ctx),
	})

	r.Emit(b.mb.NotifModelsUpdated(), PayloadModelsUpdated{
		Ids:    []string{id},
		Models: map[string]any{id: obj},
	})
	return r, nil
}

// EditDetailListField Event: click detail list field element edit button
func (b *DetailingBuilder) EditDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		fieldName          string
		index, deleteIndex int64
	)

	fieldName = ctx.Queries().Get(SectionFieldName)
	f := b.Section(fieldName)

	unsaved := ctx.ParamAsBool(f.elementUnsavedKey())
	index, err = strconv.ParseInt(ctx.Queries().Get(f.EditBtnKey()), 10, 64)
	if err != nil {
		return
	}
	deleteIndex = -1
	if ctx.Queries().Get(f.DeleteBtnKey()) != "" {
		deleteIndex, err = strconv.ParseInt(ctx.Queries().Get(f.EditBtnKey()), 10, 64)
		if err != nil {
			return
		}
	}

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if unsaved {
		if _, err := f.appendElement(obj); err != nil {
			panic(err)
		}
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, ctx, int(deleteIndex), int(index), -1, unsaved),
	})

	return
}

// SaveDetailListField Event: click detail list field element Save button
func (b *DetailingBuilder) SaveDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		fieldName string
		index     int
		isCancel  bool
	)

	isCancel = ctx.ParamAsBool(SectionIsCancel)
	fieldName = ctx.Param(SectionFieldName)

	f := b.Section(fieldName)

	unsaved := ctx.ParamAsBool(f.elementUnsavedKey())
	index = ctx.ParamAsInt(f.SaveBtnKey())

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if isCancel {
		var listLen int
		if unsaved {
			if listLen, err = f.appendElement(obj); err != nil {
				panic(err)
			}
		}
		if listLen-1 == index {
			if err = f.removeElement(obj, index); err != nil {
				panic(err)
			}
			index = -1
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: f.FieldPortalName(),
			Body: f.listComponent(obj, ctx, -1, -1, index, unsaved),
		})
		return
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	err = f.saver(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	if ctx.ParamAsBool(f.elementUnsavedKey()) {
		if _, err := f.appendElement(obj); err != nil {
			panic(err)
		}
	}

	if _, ok := ctx.Flash.(*web.ValidationErrors); ok {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: f.FieldPortalName(),
			Body: f.listComponent(obj, ctx, -1, index, -1, unsaved),
		})
		return
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, ctx, -1, -1, index, unsaved),
	})

	return
}

// DeleteDetailListField Event: click detail list field element Delete button
func (b *DetailingBuilder) DeleteDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		fieldName string
		index     int64
	)

	fieldName = ctx.Queries().Get(SectionFieldName)
	f := b.Section(fieldName)

	unsaved := ctx.ParamAsBool(f.elementUnsavedKey())
	index, err = strconv.ParseInt(ctx.Queries().Get(f.DeleteBtnKey()), 10, 64)
	if err != nil {
		return
	}

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	err = f.removeElement(obj, int(index))
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	err = f.saver(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	if unsaved {
		if _, err := f.appendElement(obj); err != nil {
			panic(err)
		}
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, ctx, int(index), -1, -1, unsaved),
	})

	return
}

// CreateDetailListField Event: click detail list field element Add row button
func (b *DetailingBuilder) CreateDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	fieldName := ctx.Queries().Get(SectionFieldName)
	f := b.Section(fieldName)

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	var listLen int
	if ctx.ParamAsBool(f.elementUnsavedKey()) {
		if listLen, err = f.appendElement(obj); err != nil {
			panic(err)
		}
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, ctx, -1, listLen-1, -1, true),
	})

	return
}

const fieldRefreshOnUpdate = "__RefreshOnUpdate__"

func (b *DetailingBuilder) EnableRefreshOnUpdate() *DetailingBuilder {
	b.Field(fieldRefreshOnUpdate).ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
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
