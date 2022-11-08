package presets

import (
	"fmt"
	"strings"

	"github.com/goplaid/web"
	"github.com/goplaid/x/i18n"
	"github.com/goplaid/x/perm"
	"github.com/qor/qor5/presets/actions"
	"github.com/goplaid/ui/stripeui"
	. "github.com/goplaid/ui/vuetify"
	"github.com/jinzhu/inflection"
	h "github.com/theplant/htmlgo"
)

type EditingBuilder struct {
	mb               *ModelBuilder
	Fetcher          FetchFunc
	Setter           SetterFunc
	Saver            SaveFunc
	Deleter          DeleteFunc
	Validator        ValidateFunc
	tabPanels        []ObjectComponentFunc
	hiddenFuncs      []ObjectComponentFunc
	sidePanel        ComponentFunc
	actionsFunc      ObjectComponentFunc
	editingTitleFunc EditingTitleComponentFunc
	FieldsBuilder
}

// string / []string / *FieldsSection
func (mb *ModelBuilder) Editing(vs ...interface{}) (r *EditingBuilder) {
	r = mb.editing
	if len(vs) == 0 {
		return
	}

	r.Only(vs...)
	return r
}

// string / []string / *FieldsSection
func (b *EditingBuilder) Only(vs ...interface{}) (r *EditingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Only(vs...)
	return
}

func (b *EditingBuilder) Creating(vs ...interface{}) (r *EditingBuilder) {
	if b.mb.creating == nil {
		b.mb.creating = &EditingBuilder{
			mb:        b.mb,
			Fetcher:   b.Fetcher,
			Setter:    b.Setter,
			Saver:     b.Saver,
			Deleter:   b.Deleter,
			Validator: b.Validator,
		}
	}
	r = b.mb.creating

	r.FieldsBuilder = *b.mb.writeFields.Only(vs...)

	return r
}

func (b *EditingBuilder) FetchFunc(v FetchFunc) (r *EditingBuilder) {
	b.Fetcher = v
	return b
}

func (b *EditingBuilder) SaveFunc(v SaveFunc) (r *EditingBuilder) {
	b.Saver = v
	return b
}

func (b *EditingBuilder) DeleteFunc(v DeleteFunc) (r *EditingBuilder) {
	b.Deleter = v
	return b
}

func (b *EditingBuilder) ValidateFunc(v ValidateFunc) (r *EditingBuilder) {
	b.Validator = v
	return b
}

func (b *EditingBuilder) SetterFunc(v SetterFunc) (r *EditingBuilder) {
	b.Setter = v
	return b
}

func (b *EditingBuilder) AppendTabsPanelFunc(v ObjectComponentFunc) (r *EditingBuilder) {
	b.tabPanels = append(b.tabPanels, v)
	return b
}

func (b *EditingBuilder) SidePanelFunc(v ComponentFunc) (r *EditingBuilder) {
	b.sidePanel = v
	return b
}

func (b *EditingBuilder) AppendHiddenFunc(v ObjectComponentFunc) (r *EditingBuilder) {
	b.hiddenFuncs = append(b.hiddenFuncs, v)
	return b
}

func (b *EditingBuilder) ActionsFunc(v ObjectComponentFunc) (r *EditingBuilder) {
	b.actionsFunc = v
	return b
}

func (b *EditingBuilder) EditingTitleFunc(v EditingTitleComponentFunc) (r *EditingBuilder) {
	b.editingTitleFunc = v
	return b
}

func (b *EditingBuilder) formNew(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermCreate).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	creatingB := b
	if b.mb.creating != nil {
		creatingB = b.mb.creating
	}

	b.mb.p.overlay(ctx.R.FormValue(ParamOverlay), &r, creatingB.editFormFor(nil, ctx), b.mb.rightDrawerWidth)
	return
}

func (b *EditingBuilder) formEdit(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermGet).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}
	b.mb.p.overlay(ctx.R.FormValue(ParamOverlay), &r, b.editFormFor(nil, ctx), b.mb.rightDrawerWidth)
	return
}

func (b *EditingBuilder) singletonPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	if b.mb.Info().Verifier().Do(PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	msgr := MustGetMessages(ctx.R)
	title := msgr.EditingObjectTitle(i18n.T(ctx.R, ModelsI18nModuleKey, inflection.Singular(b.mb.label)), "")
	r.PageTitle = title
	obj, err := b.Fetcher(b.mb.NewModel(), "", ctx)
	if err == ErrRecordNotFound {
		if err = b.Saver(b.mb.NewModel(), "", ctx); err != nil {
			return
		}
		obj, err = b.Fetcher(b.mb.NewModel(), "", ctx)
	}
	if err != nil {
		return
	}
	r.Body = b.editFormFor(obj, ctx)
	return
}

func (b *EditingBuilder) editFormFor(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
	msgr := MustGetMessages(ctx.R)

	id := ctx.R.FormValue(ParamID)
	if b.mb.singleton {
		id = stripeui.ObjectID(obj)
	}

	var buttonLabel = msgr.Create
	var disableUpdateBtn bool
	var title h.HTMLComponent
	title = h.Text(msgr.CreatingObjectTitle(
		i18n.T(ctx.R, ModelsI18nModuleKey, inflection.Singular(b.mb.label)),
	))
	if len(id) > 0 {
		if obj == nil {
			var err error
			obj, err = b.Fetcher(b.mb.NewModel(), id, ctx)
			if err != nil {
				panic(err)
			}
		}
		disableUpdateBtn = b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
		buttonLabel = msgr.Update
		editingTitleText := msgr.EditingObjectTitle(
			i18n.T(ctx.R, ModelsI18nModuleKey, inflection.Singular(b.mb.label)),
			getPageTitle(obj, id))
		if b.editingTitleFunc != nil {
			title = b.editingTitleFunc(obj, editingTitleText, ctx)
		} else {
			title = h.Text(editingTitleText)
		}
	}

	if obj == nil {
		obj = b.mb.NewModel()
	}

	var notice h.HTMLComponent
	if msg, ok := ctx.Flash.(string); ok {
		if len(msg) > 0 {
			notice = VAlert(h.Text(msg)).
				Border("left").
				Type("success").
				Elevation(2).
				ColoredBorder(true)
		}
	}

	vErr, ok := ctx.Flash.(*web.ValidationErrors)
	if ok {
		gErr := vErr.GetGlobalError()
		if len(gErr) > 0 {
			notice = VAlert(h.Text(gErr)).
				Border("left").
				Type("error").
				Elevation(2).
				ColoredBorder(true)
		}
	}

	queries := ctx.Queries()
	if b.mb.singleton {
		queries.Add(ParamID, id)
	}
	updateBtn := VBtn(buttonLabel).
		Color("primary").
		Attr("@click", web.Plaid().
			EventFunc(actions.Update).
			Queries(queries).
			URL(b.mb.Info().ListingHref()).
			Go())
	if disableUpdateBtn {
		updateBtn = updateBtn.Disabled(disableUpdateBtn)
	} else {
		updateBtn = updateBtn.Attr(":disabled", "isFetching").
			Attr(":loading", "isFetching")
	}
	var actionButtons h.HTMLComponent = h.Components(
		VSpacer(),
		updateBtn,
	)

	if b.actionsFunc != nil {
		actionButtons = b.actionsFunc(obj, ctx)
	}

	var hiddenComps []h.HTMLComponent
	for _, hf := range b.hiddenFuncs {
		hiddenComps = append(hiddenComps, hf(obj, ctx))
	}

	formContent := h.Components(
		VCardText(
			notice,
			h.Components(hiddenComps...),
			b.ToComponent(b.mb.Info(), obj, ctx),
		),
		VCardActions(actionButtons),
	)

	var asideContent h.HTMLComponent = formContent

	if len(b.tabPanels) != 0 {
		var tabs []h.HTMLComponent

		for _, panelFunc := range b.tabPanels {
			value := panelFunc(obj, ctx)
			if value != nil {
				tabs = append(tabs, value)
			}
		}

		if len(tabs) != 0 {
			asideContent = VTabs(
				VTab(h.Text(msgr.FormTitle)),
				VTabItem(web.Scope(formContent).VSlot("{plaidForm}")),
				h.Components(tabs...),
			).Class("v-tabs--fixed-tabs")
		}
	}

	if b.sidePanel != nil {
		sidePanel := b.sidePanel(ctx)
		if sidePanel != nil {
			asideContent = VContainer(
				VRow(
					VCol(asideContent).Cols(8),
					VCol(sidePanel).Cols(4),
				),
			)
		}
	}

	overlayType := ctx.R.FormValue(ParamOverlay)
	closeBtnVarScript := closeRightDrawerVarScript
	if overlayType == actions.Dialog {
		closeBtnVarScript = closeDialogVarScript
	}

	return web.Scope(
		h.If(!b.mb.singleton,
			VAppBar(
				VToolbarTitle("").Class("pl-2").
					Children(title),
				VSpacer(),
				VBtn("").Icon(true).Children(
					VIcon("close"),
				).Attr("@click.stop", closeBtnVarScript),
			).Color("white").Elevation(0).Dense(true),
		),

		VSheet(
			VCard(asideContent).Flat(true),
		).Class("pa-2"),
	).VSlot("{ plaidForm }")
}

func (b *EditingBuilder) doDelete(ctx *web.EventContext) (r web.EventResponse, err1 error) {
	if b.mb.Info().Verifier().Do(PermDelete).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	id := ctx.R.FormValue(ParamID)
	var obj = b.mb.NewModel()
	if len(id) > 0 {
		err := b.Deleter(obj, id, ctx)
		if err != nil {
			ShowMessage(&r, err.Error(), "warning")
			return
		}
	}

	removeSelectQuery := web.Var(fmt.Sprintf(`{value: %s, add: false, remove: true}`, h.JSONString(id)))
	if isInDialogFromQuery(ctx) {
		u := fmt.Sprintf("%s?%s", b.mb.Info().ListingHref(), ctx.Queries().Get(ParamListingQueries))
		web.AppendVarsScripts(&r,
			"vars.deleteConfirmation = false",
			web.Plaid().
				URL(u).
				EventFunc(actions.UpdateListingDialog).
				MergeQuery(true).
				Query(ParamSelectedIds, removeSelectQuery).
				Go(),
		)
	} else {
		// refresh current page

		// TODO: response location does not support `valueOp`
		// r.PushState = web.Location(nil).
		// 	MergeQuery(true).
		//  Query(ParamSelectedIds, removeSelectQuery)
		web.AppendVarsScripts(&r,
			web.Plaid().
				PushState(true).
				MergeQuery(true).
				Query(ParamSelectedIds, removeSelectQuery).
				Go(),
		)
	}
	return
}

func (b *EditingBuilder) FetchAndUnmarshal(id string, removeDeletedAndSort bool, ctx *web.EventContext) (obj interface{}, vErr web.ValidationErrors) {
	obj = b.mb.NewModel()
	if len(id) > 0 {
		var err1 error
		obj, err1 = b.Fetcher(obj, id, ctx)
		if err1 != nil {
			vErr.GlobalError(err1.Error())
			// b.UpdateOverlayContent(ctx, &r, obj, "", err1)
			return
		}
	}

	vErr = b.RunSetterFunc(ctx, removeDeletedAndSort, obj)
	return
}

func (b *EditingBuilder) doUpdate(
	ctx *web.EventContext,
	r *web.EventResponse,
	// will not close drawer/dialog
	silent bool,
) (err error) {
	id := ctx.R.FormValue(ParamID)
	usingB := b
	if b.mb.creating != nil && id == "" {
		usingB = b.mb.creating
	}

	obj, vErr := usingB.FetchAndUnmarshal(id, true, ctx)
	if vErr.HaveErrors() {
		usingB.UpdateOverlayContent(ctx, r, obj, "", &vErr)
		return &vErr
	}

	if len(id) > 0 {
		if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			b.UpdateOverlayContent(ctx, r, obj, "", perm.PermissionDenied)
			return perm.PermissionDenied
		}
	} else {
		if b.mb.Info().Verifier().Do(PermCreate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			b.UpdateOverlayContent(ctx, r, obj, "", perm.PermissionDenied)
			return perm.PermissionDenied
		}
	}

	if usingB.Validator != nil {
		if vErr = usingB.Validator(obj, ctx); vErr.HaveErrors() {
			usingB.UpdateOverlayContent(ctx, r, obj, "", &vErr)
			return &vErr
		}
	}

	err1 := usingB.Saver(obj, id, ctx)
	if err1 != nil {
		usingB.UpdateOverlayContent(ctx, r, obj, "", err1)
		return err1
	}

	overlayType := ctx.R.FormValue(ParamOverlay)
	script := closeRightDrawerVarScript
	if overlayType == actions.Dialog {
		script = closeDialogVarScript
	}
	if silent {
		script = ""
	}

	afterUpdateScript := ctx.R.FormValue(ParamOverlayAfterUpdateScript)
	if afterUpdateScript != "" {
		web.AppendVarsScripts(r, script, strings.NewReplacer(".go()",
			fmt.Sprintf(".query(%s, %s).go()",
				h.JSONString(ParamOverlayUpdateID),
				h.JSONString(stripeui.ObjectID(obj)),
			)).Replace(afterUpdateScript),
		)

		return
	}

	if isInDialogFromQuery(ctx) {
		web.AppendVarsScripts(r,
			web.Plaid().
				URL(ctx.R.RequestURI).
				EventFunc(actions.UpdateListingDialog).
				StringQuery(ctx.R.URL.Query().Get(ParamListingQueries)).
				Go(),
		)
	} else {
		r.PushState = web.Location(nil)
	}
	web.AppendVarsScripts(r, script)
	return
}

func (b *EditingBuilder) defaultUpdate(ctx *web.EventContext) (r web.EventResponse, err error) {
	uErr := b.doUpdate(ctx, &r, false)
	if uErr == nil {
		msgr := MustGetMessages(ctx.R)
		ShowMessage(&r, msgr.SuccessfullyUpdated, "")
	}
	return r, nil
}

func (b *EditingBuilder) SaveOverlayContent(
	ctx *web.EventContext,
	r *web.EventResponse,
) (err error) {
	return b.doUpdate(ctx, r, true)
}

func (b *EditingBuilder) RunSetterFunc(ctx *web.EventContext, removeDeletedAndSort bool, toObj interface{}) (vErr web.ValidationErrors) {
	if b.Setter != nil {
		b.Setter(toObj, ctx)
	}

	vErr = b.Unmarshal(toObj, b.mb.Info(), removeDeletedAndSort, ctx)

	return
}

func (b *EditingBuilder) UpdateOverlayContent(
	ctx *web.EventContext,
	r *web.EventResponse,
	obj interface{},
	successMessage string,
	err error,
) {
	ctx.Flash = err

	if err != nil {
		if _, ok := err.(*web.ValidationErrors); !ok {
			vErr := &web.ValidationErrors{}
			vErr.GlobalError(err.Error())
			ctx.Flash = vErr
		}
	}

	if ctx.Flash == nil {
		ctx.Flash = successMessage
	}

	overlayType := ctx.R.FormValue(ParamOverlay)
	p := rightDrawerContentPortalName

	if overlayType == actions.Dialog {
		p = dialogContentPortalName
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: p,
		Body: b.editFormFor(obj, ctx),
	})

}
