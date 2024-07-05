package presets

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
)

type DetailingBuilder struct {
	mb                 *ModelBuilder
	actions            []*ActionBuilder
	pageFunc           web.PageFunc
	fetcher            FetchFunc
	tabPanels          []TabComponentFunc
	sidePanel          ObjectComponentFunc
	afterTitleCompFunc ObjectComponentFunc
	drawer             bool
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
	r.PageTitle = msgr.DetailingObjectTitle(inflection.Singular(b.mb.label), getPageTitle(obj, id))
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

	r.Body = VContainer(
		notice,
	).AppendChildren(tabsContent).Fluid(true)

	return
}

func (b *DetailingBuilder) showInDrawer(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermGet).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	pr, err := b.GetPageFunc()(ctx)
	if err != nil {
		return
	}

	overlayType := ctx.R.FormValue(ParamOverlay)
	closeBtnVarScript := CloseRightDrawerVarScript
	if overlayType == actions.Dialog {
		closeBtnVarScript = closeDialogVarScript
	}

	title := h.Div(h.Text(pr.PageTitle)).Class("d-flex")
	if v, ok := GetComponentFromContext(ctx, ctxDetailingAfterTitleComponent); ok {
		title.AppendChildren(VSpacer(), v)
	}

	comp := web.Scope(
		VLayout(
			VAppBar(
				VAppBarTitle(title).Class("pl-2"),
				VBtn("").Icon("mdi-close").
					Attr("@click.stop", closeBtnVarScript),
			).Color("white").Elevation(0),

			VMain(
				VSheet(
					VCard(pr.Body).Flat(true).Class("pa-1"),
				).Class("pa-2"),
			),
		),
	).VSlot("{ form }")

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

	r.PushState = web.Location(url.Values{})
	r.RunScript = CloseRightDrawerVarScript

	return
}

func (b *DetailingBuilder) formDrawerAction(ctx *web.EventContext) (r web.EventResponse, err error) {
	action := getAction(b.actions, ctx.R.FormValue(ParamAction))
	if action == nil {
		panic("action required")
	}

	b.mb.p.rightDrawer(&r, b.actionForm(action, ctx), "")
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
					Theme("dark").
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
	obj, err = b.GetFetchFunc()(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
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
	key := ctx.Queries().Get(SectionFieldName)
	id := ctx.Queries().Get(ParamID)

	f := b.Section(key)

	obj := b.mb.NewModel()
	obj, err = b.GetFetchFunc()(obj, id, ctx)
	if err != nil {
		return
	}
	if f.setter != nil {
		f.setter(obj, ctx)
	}

	err = f.saver(obj, id, ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.viewComponent(obj, &FieldContext{
			ModelInfo: b.mb.modelInfo,
			FormKey:   f.name,
			Name:      f.name,
			Label:     f.label,
		}, ctx),
	})

	r.Emit(b.mb.NotifModelsUpdated(), PayloadModelsUpdated{
		Ids:    []string{id},
		Models: []any{obj},
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

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, nil, ctx, int(deleteIndex), int(index), -1),
	})

	return
}

// SaveDetailListField Event: click detail list field element Save button
func (b *DetailingBuilder) SaveDetailListField(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		fieldName string
		index     int64
	)

	fieldName = ctx.Queries().Get(SectionFieldName)
	f := b.Section(fieldName)

	index, err = strconv.ParseInt(ctx.Queries().Get(f.SaveBtnKey()), 10, 64)
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

	err = f.saver(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, nil, ctx, -1, -1, int(index)),
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

	// delete from slice
	var list any
	if list, err = reflectutils.Get(obj, f.name); err != nil {
		return
	}
	listValue := reflect.ValueOf(list)
	if listValue.Kind() != reflect.Slice {
		err = errors.New("field is not a slice")
		return
	}
	newList := reflect.MakeSlice(reflect.TypeOf(list), 0, 0)
	for i := 0; i < listValue.Len(); i++ {
		if i != int(index) {
			newList = reflect.Append(newList, listValue.Index(i))
		}
	}
	if err = reflectutils.Set(obj, f.name, newList.Interface()); err != nil {
		return
	}

	err = f.saver(obj, ctx.Queries().Get(ParamID), ctx)
	if err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, nil, ctx, int(index), -1, -1),
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

	var list any
	if list, err = reflectutils.Get(obj, f.name); err != nil {
		return
	}

	listLen := 0
	if list != nil {
		listValue := reflect.ValueOf(list)
		if listValue.Kind() != reflect.Slice {
			err = errors.New(fmt.Sprintf("the kind of list field is %s, not slice", listValue.Kind()))
			return
		}
		listLen = listValue.Len()
	}

	if err = reflectutils.Set(obj, f.name+"[]", f.editingFB.model); err != nil {
		return
	}

	if err = f.saver(obj, ctx.Queries().Get(ParamID), ctx); err != nil {
		ShowMessage(&r, err.Error(), "warning")
		return r, nil
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: f.FieldPortalName(),
		Body: f.listComponent(obj, nil, ctx, -1, listLen, -1),
	})

	return
}
