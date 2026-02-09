package presets

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets/actions"
)

type EditingBuilder struct {
	mb                       *ModelBuilder
	Fetcher                  FetchFunc
	Setter                   SetterFunc
	Saver                    SaveFunc
	Deleter                  DeleteFunc
	Validator                ValidateFunc
	tabPanels                []TabComponentFunc
	hiddenFuncs              []ObjectComponentFunc
	sidePanel                ObjectComponentFunc
	actionsFunc              ObjectComponentFunc
	editingTitleFunc         EditingTitleComponentFunc
	idCurrentActiveProcessor IdCurrentActiveProcessor
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

func (b *EditingBuilder) Except(vs ...string) (r *EditingBuilder) {
	r = b
	r.FieldsBuilder = *r.FieldsBuilder.Except(vs...)
	return
}

func (b *EditingBuilder) Creating(vs ...interface{}) (r *EditingBuilder) {
	if b.mb.creating != nil && len(vs) == 0 {
		return b.mb.creating
	}

	if b.mb.creating == nil {
		b.mb.creating = &EditingBuilder{
			mb: b.mb,
			Fetcher: func(obj interface{}, id string, ctx *web.EventContext) (interface{}, error) {
				return b.Fetcher(obj, id, ctx)
			},
			Setter: func(obj interface{}, ctx *web.EventContext) {
				if b.Setter != nil {
					b.Setter(obj, ctx)
				}
			},
			Saver: func(obj interface{}, id string, ctx *web.EventContext) error {
				return b.Saver(obj, id, ctx)
			},
			Deleter: func(obj interface{}, id string, ctx *web.EventContext) error {
				return b.Deleter(obj, id, ctx)
			},
			Validator: func(obj interface{}, ctx *web.EventContext) (r web.ValidationErrors) {
				if b.Validator == nil {
					return r
				}
				return b.Validator(obj, ctx)
			},
		}
	}

	r = b.mb.creating
	if len(vs) == 0 {
		if len(b.fieldsLayout) == 0 {
			for _, f := range b.fields {
				vs = append(vs, f.name)
			}
		} else {
			vs = CloneFieldsLayout(b.fieldsLayout)
		}
	}
	if len(vs) == 0 {
		r.FieldsBuilder = *b.FieldsBuilder.Clone()
	} else {
		r.FieldsBuilder = *b.FieldsBuilder.Only(vs...)
	}
	return r
}

func (b *EditingBuilder) FetchFunc(v FetchFunc) (r *EditingBuilder) {
	b.Fetcher = v
	return b
}

func (b *EditingBuilder) WrapFetchFunc(w func(in FetchFunc) FetchFunc) (r *EditingBuilder) {
	b.Fetcher = w(b.Fetcher)
	return b
}

func (b *EditingBuilder) SaveFunc(v SaveFunc) (r *EditingBuilder) {
	b.Saver = v
	return b
}

func (b *EditingBuilder) WrapSaveFunc(w func(in SaveFunc) SaveFunc) (r *EditingBuilder) {
	b.Saver = w(b.Saver)
	return b
}

func (b *EditingBuilder) DeleteFunc(v DeleteFunc) (r *EditingBuilder) {
	b.Deleter = v
	return b
}

func (b *EditingBuilder) WrapDeleteFunc(w func(in DeleteFunc) DeleteFunc) (r *EditingBuilder) {
	b.Deleter = w(b.Deleter)
	return b
}

func (b *EditingBuilder) ValidateFunc(v ValidateFunc) (r *EditingBuilder) {
	b.Validator = v
	return b
}

func (b *EditingBuilder) WrapValidateFunc(w func(in ValidateFunc) ValidateFunc) (r *EditingBuilder) {
	b.Validator = w(b.Validator)
	return b
}

func (b *EditingBuilder) SetterFunc(v SetterFunc) (r *EditingBuilder) {
	b.Setter = v
	return b
}

func (b *EditingBuilder) WrapSetterFunc(w func(in SetterFunc) SetterFunc) (r *EditingBuilder) {
	b.Setter = w(b.Setter)
	return b
}

func (b *EditingBuilder) AppendTabsPanelFunc(v TabComponentFunc) (r *EditingBuilder) {
	b.tabPanels = append(b.tabPanels, v)
	return b
}

func (b *EditingBuilder) TabsPanels(vs ...TabComponentFunc) (r *EditingBuilder) {
	b.tabPanels = vs
	return b
}

func (b *EditingBuilder) SidePanelFunc(v ObjectComponentFunc) (r *EditingBuilder) {
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

func (b *EditingBuilder) WrapIdCurrentActive(w func(in IdCurrentActiveProcessor) IdCurrentActiveProcessor) (r *EditingBuilder) {
	if b.idCurrentActiveProcessor == nil {
		b.idCurrentActiveProcessor = w(func(_ *web.EventContext, current string) (string, error) {
			return current, nil
		})
	} else {
		b.idCurrentActiveProcessor = w(b.idCurrentActiveProcessor)
	}
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

	if b.idCurrentActiveProcessor != nil {
		ctx.WithContextValue(ctxKeyIdCurrentActiveProcessor{}, b.idCurrentActiveProcessor)
	}
	b.mb.p.overlay(ctx, &r, creatingB.editFormFor(nil, ctx), b.mb.rightDrawerWidth)
	return
}

func (b *EditingBuilder) formEdit(ctx *web.EventContext) (r web.EventResponse, err error) {
	if b.mb.Info().Verifier().Do(PermGet).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}
	if b.idCurrentActiveProcessor != nil {
		ctx.WithContextValue(ctxKeyIdCurrentActiveProcessor{}, b.idCurrentActiveProcessor)
	}
	b.mb.p.overlay(ctx, &r, b.editFormFor(nil, ctx), b.mb.rightDrawerWidth)
	return
}

func (b *EditingBuilder) singletonPageFunc(ctx *web.EventContext) (r web.PageResponse, err error) {
	if b.mb.Info().Verifier().Do(PermGet).WithReq(ctx.R).IsAllowed() != nil {
		err = perm.PermissionDenied
		return
	}

	msgr := b.mb.mustGetMessages(ctx.R)
	title := msgr.EditingObjectTitle(b.mb.Info().LabelName(ctx, true), "")
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

	if b.editingTitleFunc != nil {
		titleText := msgr.EditingObjectTitle(b.mb.Info().LabelName(ctx, true), getPageTitle(obj, ""))
		ctx.WithContextValue(CtxPageTitleComponent, b.editingTitleFunc(obj, titleText, ctx))
	}

	r.Body = web.Portal(b.editFormFor(obj, ctx)).Name(singletonEditingPortalName)
	return
}

func (b *EditingBuilder) editFormFor(obj interface{}, ctx *web.EventContext) h.HTMLComponent {
	var (
		msgr          = b.mb.mustGetMessages(ctx.R)
		id            = ctx.R.FormValue(ParamID)
		overlayType   = ctx.R.FormValue(ParamOverlay)
		onChangeEvent = fmt.Sprintf(`if (vars.%s) { vars.%s.editing=true };`, VarsPresetsDataChanged, VarsPresetsDataChanged)
		autosave      = overlayType == actions.Content
	)
	if b.mb.singleton {
		id = vx.ObjectID(obj)
	}

	buttonLabel := msgr.Create
	labelName := b.mb.Info().LabelName(ctx, true)
	var noPerm bool
	var title h.HTMLComponent
	title = h.Text(msgr.CreatingObjectTitle(
		labelName,
	))
	if id != "" {
		if obj == nil {
			var err error
			obj, err = b.Fetcher(b.mb.NewModel(), id, ctx)
			if err != nil {
				panic(err)
			}
		}
		noPerm = b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
		buttonLabel = msgr.Update
		editingTitleText := msgr.EditingObjectTitle(
			labelName,
			getPageTitle(obj, id))
		if b.editingTitleFunc != nil {
			title = b.editingTitleFunc(obj, editingTitleText, ctx)
		} else {
			title = h.Text(editingTitleText)
		}
	} else {
		noPerm = b.mb.Info().Verifier().Do(PermCreate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil
	}

	if obj == nil {
		obj = b.mb.NewModel()
	}

	if b.mb.singleton && id != "" {
		if ctx.R.Form == nil {
			_ = ctx.R.ParseForm()
		}
		ctx.R.Form.Set(ParamID, id)
	}

	var notice h.HTMLComponent
	{
		var text string
		var color string
		if msg, ok := ctx.Flash.(string); ok {
			if msg != "" {
				text = msg
				color = "success"
			}
		}
		vErr, ok := ctx.Flash.(*web.ValidationErrors)
		if ok {
			gErr := vErr.GetGlobalError()
			if gErr != "" {
				text = gErr
				color = "error"
			}
		}
		if text != "" {
			notice = web.Scope(
				VSnackbar(
					h.Div().Style("white-space: pre-wrap").Text(fmt.Sprintf(`{{ %q }}`, text)),
				).Location("top").
					Timeout(2000).
					Color(color).
					Attr("v-model", "locals.show"),
			).VSlot("{ locals }").Init(`{ show: true }`)
		}
	}

	queries := ctx.Queries()
	if b.mb.singleton {
		queries.Set(ParamID, id)
	}
	var actionButtons h.HTMLComponent = h.Components(
		VSpacer(),
		h.Iff(!noPerm, func() h.HTMLComponent {
			return web.Scope(
				VBtn(buttonLabel).
					Color("primary").
					Variant(VariantFlat).
					Attr(":disabled", "xLocals.isFetching").
					Attr(":loading", "xLocals.isFetching").
					Attr("@click", web.Plaid().
						BeforeScript("xLocals.isFetching=true").
						EventFunc(actions.Update).
						Queries(queries).
						ThenScript("xLocals.isFetching=false;").
						URL(b.mb.Info().ListingHref()).
						Go()),
			).VSlot("{locals:xLocals}").Init("{isFetching:false}")
		}),
	)

	if b.actionsFunc != nil {
		actionButtons = b.actionsFunc(obj, ctx)
	}

	var hiddenComps []h.HTMLComponent
	for _, hf := range b.hiddenFuncs {
		hiddenComps = append(hiddenComps, hf(obj, ctx))
	}

	if id == "" {
		ctx = ctx.WithContextValue(ctxKeyForceForCreating{}, true)
	}
	formContent := web.Scope(h.Components(
		VCardText(
			h.Components(hiddenComps...),
			web.Listen(b.mb.NotifModelsValidate(), setFieldErrorsScript),
			b.ToComponent(b.mb.Info(), obj, ctx),
		),
		h.If(!autosave, VCardActions(actionButtons)),
	))

	asideContent := defaultToPage(commonPageConfig{
		formContent: formContent,
		tabPanels:   b.tabPanels,
		sidePanel:   b.sidePanel,
	}, obj, ctx)

	closeBtnVarScript := CloseRightDrawerVarConfirmScript
	if overlayType == actions.Dialog {
		closeBtnVarScript = CloseDialogVarScript
	}
	scope := web.Scope(
		notice,
		VLayout(
			h.If(!b.mb.singleton,
				VAppBar(
					VToolbarTitle("").Class("pl-2").
						Children(title).ClassIf("pr-5", autosave),
					h.If(!autosave, VBtn("").Icon(true).Children(
						VIcon("mdi-close"),
					).Attr("@click.stop", closeBtnVarScript)),
				).Color("white").Elevation(0),
			),
			VMain(
				VSheet(
					VCard(asideContent).Variant(VariantFlat),
				).Class("pa-2 detailing-page-wrap"),
			),
		),
	).VSlot("{ form}")
	operateID := fmt.Sprint(time.Now().UnixNano())
	onChangeEvent += checkFormChangeScript
	if autosave {
		onChangeEvent += web.Plaid().URL(ctx.R.URL.Path).
			BeforeScript(fmt.Sprintf(`dash.__currentValidateKeys=null;dash.__ValidateOperateID=%q`, operateID)).
			EventFunc(actions.Validate).
			Query(ParamID, id).
			Query(ParamOperateID, operateID).
			Query(ParamOverlay, ctx.Param(ParamOverlay)).
			Go()
	} else {
		onChangeEvent += setValidateKeysScript +
			web.Plaid().URL(ctx.R.URL.Path).
				BeforeScript(fmt.Sprintf(`dash.__ValidateOperateID=%q;`, operateID)).
				EventFunc(actions.Validate).
				Query(ParamID, id).
				Query(ParamOperateID, operateID).
				Query(ParamOverlay, ctx.Param(ParamOverlay)).
				Go()
	}
	return web.Scope(scope.OnChange(onChangeEvent).UseDebounce(500)).VSlot("{dash}").DashInit("{errorMessages:{},disabled:{}}")
}

func (b *EditingBuilder) doValidate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		id        = ctx.Param(ParamID)
		operateID = ctx.Param(ParamOperateID)
		obj       = b.mb.NewModel()
		vErr      web.ValidationErrors
		usingB    = b
	)

	if b.mb.creating != nil && id == "" {
		usingB = b.mb.creating
	}
	defer func() {
		web.AppendRunScripts(&r,
			fmt.Sprintf(`if (dash.__ValidateOperateID==%q){%s}`, operateID,
				web.Emit(
					b.mb.NotifModelsValidate(),
					PayloadModelsSetter{
						FieldErrors: vErr.FieldErrors(),
						Id:          id,
						Passed:      !vErr.HaveErrors(),
					},
				)),
		)

		if vErr.HaveErrors() && vErr.HaveGlobalErrors() {
			web.AppendRunScripts(&r, ShowSnackbarScript(strings.Join(vErr.GetGlobalErrors(), ";"), ColorWarning))
		}
	}()
	obj, vErr = usingB.FetchAndUnmarshal(id, false, ctx)
	if vErr.HaveErrors() && vErr.HaveGlobalErrors() {
		return
	}
	vErrSetter := vErr
	if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
		vErr.GlobalError(perm.PermissionDenied.Error())
		return
	}
	if usingB.Validator != nil {
		vErr = usingB.Validator(obj, ctx)
		_ = vErrSetter.Merge(&vErr)
	}
	vErr = vErrSetter
	return
}

func (b *EditingBuilder) doDelete(ctx *web.EventContext) (r web.EventResponse, err1 error) {
	if b.mb.Info().Verifier().Do(PermDelete).WithReq(ctx.R).IsAllowed() != nil {
		ShowMessage(&r, perm.PermissionDenied.Error(), "warning")
		return
	}

	ids := ctx.R.FormValue(ParamID)
	if ids != "" {
		deletedIds := strings.Split(ids, ",")
		for _, id := range deletedIds {
			obj := b.mb.NewModel()
			err := b.Deleter(obj, id, ctx)
			if err != nil {
				ShowMessage(&r, err.Error(), "warning")
				return
			}
		}

		r.Emit(
			b.mb.NotifModelsDeleted(),
			PayloadModelsDeleted{Ids: deletedIds},
		)
	}

	web.AppendRunScripts(&r, "locals.deleteConfirmation = false")

	if event := ctx.Queries().Get(ParamAfterDeleteEvent); event != "" {
		web.AppendRunScripts(&r,
			web.Plaid().
				EventFunc(event).
				Queries(ctx.Queries()).
				Go(),
		)
	}
	return
}

func (b *EditingBuilder) FetchAndUnmarshal(id string, removeDeletedAndSort bool, ctx *web.EventContext) (obj interface{}, vErr web.ValidationErrors) {
	obj = b.mb.NewModel()
	if id != "" {
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
) (created bool, err error) {
	id := ctx.R.FormValue(ParamID)
	created = id == ""

	usingB := b
	if b.mb.creating != nil && id == "" {
		usingB = b.mb.creating
	}

	obj, vErr := usingB.FetchAndUnmarshal(id, true, ctx)

	modifiedIndexes := ContextModifiedIndexesBuilder(ctx).FromHidden(ctx.R)
	modifiedIndexes.deletedValues = make(map[string][]string)
	modifiedIndexes.sortedValues = make(map[string][]string)
	if vErr.HaveErrors() && vErr.HaveGlobalErrors() {
		usingB.UpdateOverlayContent(ctx, r, obj, "", &vErr)
		return created, &vErr
	}
	vErrSetter := vErr
	if id != "" {
		if b.mb.Info().Verifier().Do(PermUpdate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			b.UpdateOverlayContent(ctx, r, obj, "", perm.PermissionDenied)
			return created, perm.PermissionDenied
		}
	} else {
		if b.mb.Info().Verifier().Do(PermCreate).ObjectOn(obj).WithReq(ctx.R).IsAllowed() != nil {
			b.UpdateOverlayContent(ctx, r, obj, "", perm.PermissionDenied)
			return created, perm.PermissionDenied
		}
	}

	if usingB.Validator != nil {
		vErr = usingB.Validator(obj, ctx)
		_ = vErrSetter.Merge(&vErr)
		vErr = vErrSetter

	}
	if vErr.HaveErrors() {
		usingB.UpdateOverlayContent(ctx, r, obj, "", &vErr)
		return created, &vErr
	}

	err1 := usingB.Saver(obj, id, ctx)
	if err1 != nil {
		var ve *web.ValidationErrors
		if errors.As(err1, &ve) {
			usingB.UpdateOverlayContent(ctx, r, obj, "", ve)
		} else {
			usingB.UpdateOverlayContent(ctx, r, obj, "", err1)
		}
		return created, err1
	}

	if id == "" {
		r.Emit(
			b.mb.NotifModelsCreated(),
			PayloadModelsCreated{
				Models: []any{obj},
			},
		)
	} else {
		r.Emit(
			b.mb.NotifModelsUpdated(),
			PayloadModelsUpdated{Ids: []string{id}, Models: map[string]any{id: obj}},
		)
	}

	overlayType := ctx.R.FormValue(ParamOverlay)
	script := CloseRightDrawerVarScript
	if overlayType == actions.Dialog {
		script = CloseDialogVarScript
	}
	if silent {
		script = ""
	}

	afterUpdateScript := ctx.R.FormValue(ParamOverlayAfterUpdateScript)
	if afterUpdateScript != "" {
		web.AppendRunScripts(r, script, strings.NewReplacer(".go()",
			fmt.Sprintf(".query(%s, %s).go()",
				h.JSONString(ParamOverlayUpdateID),
				h.JSONString(vx.ObjectID(obj)),
			)).Replace(afterUpdateScript),
		)

		return
	}
	web.AppendRunScripts(r, script)
	return
}

func (b *EditingBuilder) defaultUpdate(ctx *web.EventContext) (r web.EventResponse, err error) {
	created, uErr := b.doUpdate(ctx, &r, false)
	if uErr == nil {
		msgr := b.mb.mustGetMessages(ctx.R)
		if created {
			ShowMessage(&r, msgr.SuccessfullyCreated, "")
		} else {
			ShowMessage(&r, msgr.SuccessfullyUpdated, "")
		}
	}
	return r, nil
}

func (b *EditingBuilder) SaveOverlayContent(
	ctx *web.EventContext,
	r *web.EventResponse,
) (err error) {
	_, err = b.doUpdate(ctx, r, true)
	return err
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
	portalName := ctx.R.FormValue(ParamPortalName)
	p := RightDrawerContentPortalName
	if overlayType == actions.Dialog {
		p = dialogContentPortalName
	}
	if b.mb.singleton {
		p = singletonEditingPortalName
	}
	if portalName != "" {
		p = portalName
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: p,
		Body: b.editFormFor(obj, ctx),
	})
}

func (b *EditingBuilder) Section(sections ...*SectionBuilder) *EditingBuilder {
	for _, sb := range sections {
		if sb.isList {
			panic("list section can not in edit")
		}
		if sb.isUsed {
			panic("section is used")
		}
		sb.isUsed = true
		sb.registerEvent()
		sb.isEdit = true

		sb.WrapSaveBtnFunc(func(_ ObjectBoolFunc) ObjectBoolFunc {
			return func(_ interface{}, _ *web.EventContext) bool {
				return false
			}
		})

		sb.ComponentFunc(func(obj interface{}, field *FieldContext, ctx *web.EventContext) h.HTMLComponent {
			return web.Portal(
				sb.editComponent(obj, field, ctx),
			).Name(sb.FieldPortalName())
		})

		b.Field(sb.name).Component(sb).
			SetterFunc(func(obj interface{}, _ *FieldContext, ctx *web.EventContext) (err error) {
				err = sb.defaultUnmarshalFunc(obj, ctx)
				if err != nil {
					return err
				}
				if sb.setter != nil {
					err = sb.setter(obj, ctx)
				}
				return err
			})
	}

	return b
}
