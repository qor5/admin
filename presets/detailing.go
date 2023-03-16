package presets

import (
	"net/url"

	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/presets/actions"
	. "github.com/qor5/ui/vuetify"
	"github.com/qor5/web"
	"github.com/qor5/x/perm"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
)

type DetailingBuilder struct {
	mb        *ModelBuilder
	actions   []*ActionBuilder
	pageFunc  web.PageFunc
	fetcher   FetchFunc
	tabPanels []ObjectComponentFunc
	drawer    bool
	FieldsBuilder
}

type pageTitle interface {
	PageTitle() string
}

// string / []string / *FieldsSection
func (mb *ModelBuilder) Detailing(vs ...interface{}) (r *DetailingBuilder) {
	r = mb.detailing
	mb.hasDetailing = true
	if len(vs) == 0 {
		return
	}

	r.Only(vs...)
	return r
}

// string / []string / *FieldsSection
func (b *DetailingBuilder) Only(vs ...interface{}) (r *DetailingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Only(vs...)
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

func (b *DetailingBuilder) Drawer(v bool) (r *DetailingBuilder) {
	b.drawer = v
	return b
}

func (b *DetailingBuilder) GetPageFunc() web.PageFunc {
	if b.pageFunc != nil {
		return b.pageFunc
	}
	return b.defaultPageFunc
}

func (b *DetailingBuilder) AppendTabsPanelFunc(v ObjectComponentFunc) (r *DetailingBuilder) {
	b.tabPanels = append(b.tabPanels, v)
	return b
}

func (b *DetailingBuilder) defaultPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {

	var id string
	if b.drawer {
		id = ctx.R.FormValue(ParamID)
	} else {
		id = pat.Param(ctx.R, "id")
	}
	r.Body = VContainer(h.Text(id))

	var obj = b.mb.NewModel()

	if id == "" {
		panic("not found")
	}

	obj, err = b.fetcher(obj, id, ctx)
	if err != nil {
		if err == ErrRecordNotFound {
			return b.mb.p.defaultNotFoundPageFunc(ctx)
		}
		return
	}

	if b.mb.Info().Verifier().Do(PermGet).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		r.Body = h.Div(h.Text(perm.PermissionDenied.Error()))
		return
	}

	msgr := MustGetMessages(ctx.R)
	r.PageTitle = msgr.DetailingObjectTitle(inflection.Singular(b.mb.label), getPageTitle(obj, id))

	var notice h.HTMLComponent
	if msg, ok := ctx.Flash.(string); ok {
		notice = VSnackbar(h.Text(msg)).Value(true).Top(true).Color("success").Value(true)
	}

	comp := b.ToComponent(b.mb.Info(), obj, ctx)

	var tabsContent h.HTMLComponent = comp

	if len(b.tabPanels) != 0 {
		var tabs []h.HTMLComponent
		for _, panelFunc := range b.tabPanels {
			value := panelFunc(obj, ctx)
			if value != nil {
				tabs = append(tabs, value)
			}
		}

		if len(tabs) != 0 {
			tabsContent = VTabs(
				VTab(h.Text(msgr.FormTitle)),
				VTabItem(web.Scope(comp).VSlot("{plaidForm}")),
				h.Components(tabs...),
			).Class("v-tabs--fixed-tabs")
		}
	}

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
	closeBtnVarScript := closeRightDrawerVarScript
	if overlayType == actions.Dialog {
		closeBtnVarScript = closeDialogVarScript
	}
	comp := web.Scope(
		VAppBar(
			VToolbarTitle("").Class("pl-2").
				Children(h.Text(pr.PageTitle)),
			VSpacer(),
			VBtn("").Icon(true).Children(
				VIcon("close"),
			).Attr("@click.stop", closeBtnVarScript),
		).Color("white").Elevation(0).Dense(true),

		VSheet(
			VCard(pr.Body).Flat(true).Class("pa-1"),
		).Class("pa-2"),
	).VSlot("{ plaidForm }")

	b.mb.p.overlay(overlayType, &r, comp, b.mb.rightDrawerWidth)
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
	if err := action.updateFunc(id, ctx); err != nil || ctx.Flash != nil {
		if ctx.Flash == nil {
			ctx.Flash = err
		}

		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: rightDrawerContentPortalName,
			Body: b.actionForm(action, ctx),
		})
		return r, nil
	}

	r.PushState = web.Location(url.Values{})
	r.VarsScript = closeRightDrawerVarScript

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
					Dark(true).
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
