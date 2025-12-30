package pagebuilder

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

type pageBuilderModelKey struct{}

func (b *ModelBuilder) registerFuncs() {
	b.eventMiddleware = b.defaultWrapEvent
	b.preview = web.Page(b.previewContent)
}

func (b *ModelBuilder) registerCustomFuncs() {
	b.editor.RegisterEventFunc(ShowSortedContainerDrawerEvent, b.eventMiddleware(b.showSortedContainerDrawer))
	b.editor.RegisterEventFunc(AddContainerEvent, b.eventMiddleware(b.addContainer))
	b.editor.RegisterEventFunc(DeleteContainerConfirmationEvent, b.eventMiddleware(b.deleteContainerConfirmation))
	b.editor.RegisterEventFunc(DeleteContainerEvent, b.eventMiddleware(b.deleteContainer))
	b.editor.RegisterEventFunc(MoveContainerEvent, b.eventMiddleware(b.moveContainer))
	b.editor.RegisterEventFunc(MoveUpDownContainerEvent, b.eventMiddleware(b.moveUpDownContainer))
	b.editor.RegisterEventFunc(ToggleContainerVisibilityEvent, b.eventMiddleware(b.toggleContainerVisibility))
	b.editor.RegisterEventFunc(RenameContainerEvent, b.eventMiddleware(b.renameContainer))
	b.editor.RegisterEventFunc(ReloadRenderPageOrTemplateEvent, b.reloadRenderPageOrTemplate)
	b.editor.RegisterEventFunc(ReloadRenderPageOrTemplateBodyEvent, b.reloadRenderPageOrTemplateBody)
	b.editor.RegisterEventFunc(MarkAsSharedContainerEvent, b.eventMiddleware(b.markAsSharedContainer))
	b.editor.RegisterEventFunc(ContainerPreviewEvent, b.eventMiddleware(b.containerPreview))
	b.editor.RegisterEventFunc(ReplicateContainerEvent, b.eventMiddleware(b.replicateContainer))
	b.editor.RegisterEventFunc(EditContainerEvent, b.eventMiddleware(b.editContainer))
	b.editor.RegisterEventFunc(UpdateContainerEvent, b.eventMiddleware(b.updateContainer))
	b.editor.RegisterEventFunc(ReloadAddContainersListEvent, b.eventMiddleware(b.reloadAddContainersList))

	preview := web.Page(b.previewContent)
	preview.Wrap(func(in web.PageFunc) web.PageFunc {
		return func(ctx *web.EventContext) (r web.PageResponse, err error) {
			defer func() {
				if v := recover(); v != nil {
					if render, ok := v.(presets.PageRenderIface); ok {
						if rerr, ok := v.(error); ok {
							log.Printf("catch previewrender err: %+v", rerr)
						}
						r, err = render.Render(ctx)
						return
					}
					panic(v)
				}
			}()
			return in(ctx)
		}
	})
	b.preview = preview
}

func (b *ModelBuilder) setPageBuilderModel(obj interface{}, ctx *web.EventContext) {
	ctx.WithContextValue(pageBuilderModelKey{}, obj)
}

func (b *ModelBuilder) pageBuilderModel(ctx *web.EventContext) (obj interface{}, err error) {
	obj = ctx.ContextValue(pageBuilderModelKey{})
	if obj == nil {
		obj = b.mb.NewModel()
		pageID, pageVersion, locale := b.getPrimaryColumnValuesBySlug(ctx)
		if pageID == 0 {
			return
		}
		g := b.db.Where("id = ? ", pageID)
		if locale != "" {
			g.Where("locale_code = ?", locale)
		}
		if pageVersion != "" {
			g.Where("version = ?", pageVersion)
		}
		err = g.First(obj).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			panic(presets.ErrNotFound(err.Error()))
			return
		} else if err != nil {
			return
		}
		b.setPageBuilderModel(obj, ctx)
	}
	return
}

func (b *ModelBuilder) defaultWrapEvent(in web.EventFunc) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		var (
			obj  interface{}
			msgr = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		)
		if obj, err = b.pageBuilderModel(ctx); err != nil {
			return
		}
		if p, ok := obj.(publish.StatusInterface); ok {
			if p.EmbedStatus().Status == publish.StatusOnline || p.EmbedStatus().Status == publish.StatusOffline {
				presets.ShowMessage(&r, msgr.TheResourceCanNotBeModified, ColorError)
				web.AppendRunScripts(&r, web.Plaid().Reload().Go())
				return
			}
		}
		return in(ctx)
	}
}

func (b *ModelBuilder) showSortedContainerDrawer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var body h.HTMLComponent
	if body, err = b.renderContainersSortedList(ctx); err != nil {
		return
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: pageBuilderLayerContainerPortal, Body: body})
	return
}

func (b *ModelBuilder) renderContainersSortedList(ctx *web.EventContext) (r h.HTMLComponent, err error) {
	var (
		cons                        []*Container
		status                      = ctx.Param(paramStatus)
		isReadonly                  = status != publish.StatusDraft && !b.isTemplate
		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
		msgr                        = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		pMsgr                       = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
	)
	wc := map[string]interface{}{
		"page_model_name": b.name,
		"page_id":         pageID,
		"page_version":    pageVersion,
	}
	if locale != "" {
		wc[l10n.SlugLocaleCode] = locale
	}
	err = b.db.Order("display_order ASC").Where(wc).Find(&cons).Error
	if err != nil {
		return
	}

	var sorterData ContainerSorter
	sorterData.Items = []ContainerSorterItem{}

	for i, c := range cons {
		vicon := "mdi-eye"
		if c.Hidden {
			vicon = "mdi-eye-off"
		}
		sorterData.Items = append(sorterData.Items,
			ContainerSorterItem{
				Index:           i,
				Label:           inflection.Plural(strcase.ToKebab(c.ModelName)),
				ModelName:       c.ModelName,
				ModelID:         strconv.Itoa(int(c.ModelID)),
				DisplayName:     c.DisplayName,
				ContainerID:     strconv.Itoa(int(c.ID)),
				URL:             b.builder.ContainerByName(c.ModelName).mb.Info().ListingHref(),
				Shared:          c.Shared,
				VisibilityIcon:  vicon,
				ParamID:         c.PrimarySlug(),
				Locale:          locale,
				Hidden:          c.Hidden,
				ContainerDataID: fmt.Sprintf(`%s_%s_%s`, inflection.Plural(strcase.ToKebab(c.ModelName)), strconv.Itoa(int(c.ModelID)), c.PrimarySlug()),
			},
		)
	}
	pushState := web.Plaid().PushState(true).MergeQuery(true).
		Query(paramContainerDataID, web.Var(`element.container_data_id`))
	var clickColumnEvent string
	if !isReadonly {
		pushState.Query(paramContainerID, web.Var("element.param_id"))
		clickColumnEvent = fmt.Sprintf(`vars.%s=element.container_data_id;`, paramContainerDataID) +
			web.Plaid().
				EventFunc(EditContainerEvent).
				MergeQuery(true).
				Query(paramContainerDataID, web.Var("element.container_data_id")).
				Query(presets.ParamOverlay, actions.Content).
				Query(presets.ParamPortalName, pageBuilderRightContentPortal).
				Go() + ";" + pushState.RunPushState() +
			";" + scrollToContainer(`element.container_data_id`)
	}
	renameEvent := web.Plaid().
		EventFunc(RenameContainerEvent).Query(paramStatus, status).Query(paramContainerID, web.Var("element.param_id")).Go()

	// container functions
	containerOperations := h.Div(
		VChip().Text(msgr.Shared).Color(ColorPrimary).Size(SizeXSmall).Attr("v-if", "element.shared"),
		VMenu(
			web.Slot(
				VBtn("").Children(
					VIcon("mdi-dots-horizontal"),
				).Attr("v-bind", "props").Attr("@click", clickColumnEvent).Variant(VariantText).Size(SizeSmall),
			).Name("activator").Scope("{ props }"),
			VList(
				VListItem(h.Text(msgr.Rename)).PrependIcon("mdi-pencil").Attr("@click",
					"element.editShow=!element.editShow",
				),
				VListItem(h.Text(fmt.Sprintf("{{element.hidden?%q:%q}}", msgr.Show, msgr.Hide))).Attr(":prepend-icon", "element.visibility_icon").Attr("@click",
					web.Plaid().
						EventFunc(ToggleContainerVisibilityEvent).
						Query(paramContainerID, web.Var("element.param_id")).
						Query(paramStatus, status).
						Go(),
				),
				VListItem(h.Text(msgr.Copy)).PrependIcon("mdi-content-copy").Attr("@click",
					web.Plaid().
						EventFunc(ReplicateContainerEvent).
						Query(paramContainerID, web.Var("element.param_id")).
						Query(paramStatus, ctx.Param(paramStatus)).
						Go(),
				),
				VListItem(h.Text(pMsgr.Delete)).PrependIcon("mdi-delete").Attr("@click",
					web.Plaid().
						EventFunc(DeleteContainerConfirmationEvent).
						Query(paramContainerID, web.Var("element.param_id")).
						Query(paramContainerName, web.Var("element.display_name")).
						Query(paramStatus, status).
						Go(),
				),
				h.If(!b.builder.disabledShared,
					VListItem(h.Text(msgr.MarkAsShared)).PrependIcon("mdi-share").Attr("@click",
						web.Plaid().
							EventFunc(MarkAsSharedContainerEvent).
							Query(paramContainerID, web.Var("element.param_id")).
							Go(),
					).Attr("v-if", "!element.shared"),
				),
			),
		),
	).Attr("v-show", "!element.editShow")
	r = web.Scope(
		VSheet(
			VList(
				vx.VXDraggable().ItemKey("model_id").Handle(".handle").Attr("v-model", "sortLocals.items").Animation(300).
					Attr("@end", web.Plaid().
						URL(ctx.R.URL.Path).
						EventFunc(MoveContainerEvent).
						Queries(ctx.R.Form).
						FieldValue(paramMoveResult, web.Var("JSON.stringify(sortLocals.items)")).
						Go()).Children(
					h.Template(
						h.Div(
							VHover(
								web.Slot(
									VListItem(
										web.Slot(
											h.If(!isReadonly,
												VBtn("").Variant(VariantText).Icon("mdi-drag").Class("my-2 ml-1 mr-1").Attr(":class", `element.hidden?"":"handle"`),
											),
										).Name("prepend"),
										VListItemTitle(
											VListItem(
												web.Scope(
													vx.VXField().Autofocus(true).
														Attr(":hide-details", "true").
														Attr("v-model", fmt.Sprintf("form.%s", paramDisplayName)).
														Attr("v-if", "element.editShow").
														Attr("@blur", "element.editShow=false;"+renameEvent).
														Attr("@keyup.enter", renameEvent),
													VListItemTitle(h.Text("{{element.display_name}}")).Attr("v-if", "!element.editShow"),
												).VSlot("{form}").FormInit("{ DisplayName:element.display_name }"),
											),
										),
										web.Slot(
											h.If(!isReadonly,
												containerOperations,
											),
										).Name("append"),
									).Attr(":variant", fmt.Sprintf(` element.hidden &&!isHovering && !element.editShow?%q:%q`, VariantPlain, VariantText)).
										Attr(":class", fmt.Sprintf(`element.container_data_id==vars.%s && !element.hidden?"bg-%s":""`, paramContainerDataID, ColorPrimaryLighten2)).
										Attr("v-bind", "props", "@click", clickColumnEvent).
										Attr(web.VAssign("vars",
											fmt.Sprintf(`{%s:%q}`, paramContainerDataID, ctx.Param(paramContainerDataID)))...),
								).Name("default").Scope("{ isHovering, props }"),
							),
							VDivider(),
						).Attr(":data-container-id", "element.container_data_id"),
					).Attr("#item", " { element } "),
				),
			),
		).Class("px-4 overflow-y-auto").MaxHeight("86vh").Attr("v-on-mounted", `({ el, window }) => {
      locals.__pageBuilderLeftContentKeepScroll = (container_data_id) => {
			if (container_data_id){
				const container = el.querySelector("div[data-container-id='" + container_data_id + "']")
				if (container){
               		 el.scrollTop = container.offsetTop;
					return
				}
			}
            const scrollTop = locals.__pageBuilderLeftContentScrollTop;
            window.setTimeout(() => {
                el.scrollTop = scrollTop;
            }, 0)
        }
    el.__handleScroll = (event) => {
        locals.__pageBuilderLeftContentScrollTop = event.target.scrollTop;
    }
    el.addEventListener('scroll', el.__handleScroll)
}`).
			Attr("v-on-unmounted", `({el}) => {
				el.removeEventListener('scroll', el.__handleScroll);
						}`),
		VBtn("").Children(
			web.Slot(
				VIcon("mdi-plus-circle-outline"),
			).Name(VSlotPrepend),
			h.Span(msgr.AddContainer).Class("ml-5"),
		).BaseColor(ColorPrimary).Variant(VariantText).Class(W100, "pl-14", "justify-start").
			Height(50).Attr("v-on-mounted", fmt.Sprintf(`()=>{
			if (!!locals.__pageBuilderLeftContentKeepScroll &&locals.__pageBuilderLeftContentKeepScrollFlag) {
       		 	locals.__pageBuilderLeftContentKeepScroll(%q);
             }
			locals.__pageBuilderLeftContentKeepScrollFlag=true;
			}`, ctx.Param(paramContainerDataID))).
			Attr(":disabled", "vars.__pageBuilderAddContainerBtnDisabled").
			Attr("@click", appendVirtualElement()+web.Plaid().PushState(true).ClearMergeQuery([]string{paramContainerID}).RunPushState()+";vars.containerPreview=false;vars.overlay=true;vars.overlayEl.refs.overlay.showByElement($event)"),
	).Init(h.JSONString(sorterData)).VSlot("{ locals:sortLocals,form }")
	return
}

func (b *ModelBuilder) addContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		modelName                   = ctx.Param(paramModelName)
		sharedContainer             = ctx.Param(paramSharedContainer)
		modelID                     = ctx.ParamAsInt(paramModelID)
		containerID                 = ctx.Param(paramContainerID)
		newContainerID              string
		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
		obj                         interface{}
	)
	if obj, err = b.getObjFromSlug(ctx); err != nil {
		return
	}

	if sharedContainer == "true" {
		newContainerID, err = b.addSharedContainerToPage(ctx, obj, pageID, containerID, pageVersion, locale, modelName, uint(modelID))
	} else {
		var newModelId uint
		newModelId, newContainerID, err = b.addContainerToPage(ctx, obj, pageID, containerID, pageVersion, locale, modelName)
		modelID = int(newModelId)
	}
	cb := b.builder.ContainerByName(modelName)
	containerDataId := cb.getContainerDataID(modelID, newContainerID)
	web.AppendRunScripts(&r,
		web.Plaid().PushState(true).MergeQuery(true).
			Query(paramContainerDataID, containerDataId).
			Query(paramContainerID, newContainerID).RunPushState(),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).Query(paramContainerDataID, containerDataId).
			Query(paramStatus, ctx.Param(paramStatus)).Go(),
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).Query(paramContainerDataID, containerDataId).Go(),
		web.Plaid().EventFunc(EditContainerEvent).MergeQuery(true).Query(paramContainerDataID, containerDataId).Go(),
		"vars.emptyIframe=false;",
	)
	return
}

func (b *ModelBuilder) moveContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	moveResult := ctx.R.FormValue(paramMoveResult)

	var result []ContainerSorterItem
	err = json.Unmarshal([]byte(moveResult), &result)
	if err != nil {
		return
	}
	err = b.db.Transaction(func(tx *gorm.DB) (inerr error) {
		for i, r := range result {
			if inerr = tx.Model(&Container{}).Where("id = ? AND locale_code = ?", r.ContainerID, r.Locale).Update("display_order", i+1).Error; inerr != nil {
				return
			}
		}
		return
	})
	web.AppendRunScripts(&r,
		web.Plaid().PushState(true).
			EventFunc(ReloadRenderPageOrTemplateBodyEvent).
			AfterScript(
				web.Plaid().
					Form(nil).
					EventFunc(ShowSortedContainerDrawerEvent).
					MergeQuery(true).
					Query(paramStatus,
						ctx.Param(paramStatus)).
					Go()).
			MergeQuery(true).Go(),
	)
	return
}

func (b *ModelBuilder) moveUpDownContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		container    Container
		preContainer Container
		paramID      = ctx.R.FormValue(paramContainerID)
		direction    = ctx.R.FormValue(paramMoveDirection)
		cs           = container.PrimaryColumnValuesBySlug(paramID)
		containerID  = cs[presets.ParamID]
		locale       = cs[l10n.SlugLocaleCode]
	)

	err = b.db.Transaction(func(tx *gorm.DB) (inerr error) {
		if inerr = tx.Where("id = ? AND locale_code = ?", containerID, locale).First(&container).Error; inerr != nil {
			return
		}
		g := tx.Model(&Container{}).Where("page_id = ? AND page_version = ? AND locale_code = ? ", container.PageID, container.PageVersion, container.LocaleCode)
		if direction == EventUp {
			g = g.Where("display_order < ? ", container.DisplayOrder).Order(" display_order desc ")
		} else {
			g = g.Where("display_order > ? ", container.DisplayOrder).Order(" display_order asc ")
		}
		g.First(&preContainer)
		if preContainer.ID <= 0 {
			return
		}
		if inerr = tx.Model(&Container{}).Where("id = ? AND locale_code = ?", containerID, locale).Update("display_order", preContainer.DisplayOrder).Error; inerr != nil {
			return
		}
		if inerr = tx.Model(&Container{}).Where("id = ? AND locale_code = ?", preContainer.ID, preContainer.LocaleCode).Update("display_order", container.DisplayOrder).Error; inerr != nil {
			return
		}
		return
	})
	web.AppendRunScripts(&r,
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).MergeQuery(true).Go(),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).MergeQuery(true).Query(paramStatus, ctx.Param(paramStatus)).Go(),
	)
	return
}

func (b *ModelBuilder) toggleContainerVisibility(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		container   Container
		paramID     = ctx.R.FormValue(paramContainerID)
		cs          = container.PrimaryColumnValuesBySlug(paramID)
		containerID = cs[presets.ParamID]
		locale      = cs[l10n.SlugLocaleCode]
		obj         interface{}
	)
	if obj, err = b.getObjFromSlug(ctx); err != nil {
		return
	}

	if err = b.db.Where("id = ? AND locale_code = ?", containerID, locale).First(&container).Error; err != nil {
		return
	}
	diffs := []activity.Diff{
		{Field: fmt.Sprintf("[%s %v].Hidden", container.DisplayName, container.ModelID), Old: fmt.Sprint(container.Hidden), New: fmt.Sprint(!container.Hidden)},
	}
	container.Hidden = !container.Hidden
	if err = b.db.Model(&Container{}).Where("id = ? AND locale_code = ?", containerID, locale).Updates(map[string]interface{}{"hidden": container.Hidden}).Error; err != nil {
		return
	}
	defer func() {
		if b.builder.ab != nil && b.builder.editorActivityProcessor != nil {
			detail := &EditorLogInput{
				Action:     activity.ActionEdit,
				PageObject: obj,
				Container:  container,
				Detail:     diffs,
			}
			if b.builder.editorActivityProcessor != nil {
				detail = b.builder.editorActivityProcessor(ctx, detail)
			}
			if detail == nil {
				return
			}
			mb, ok := b.builder.ab.GetModelBuilder(b.mb)
			if !ok {
				return
			}
			mb.Log(ctx.R.Context(), detail.Action, detail.PageObject, detail.Detail)
		}
	}()

	web.AppendRunScripts(&r,
		web.Plaid().
			EventFunc(ReloadRenderPageOrTemplateBodyEvent).
			MergeQuery(true).
			Go(),
		web.Plaid().
			EventFunc(ShowSortedContainerDrawerEvent).
			MergeQuery(true).
			Query(paramStatus, ctx.Param(paramStatus)).
			Go(),
	)
	return
}

func (b *ModelBuilder) deleteContainerConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		containerID   = ctx.R.FormValue(paramContainerID)
		containerName = ctx.R.FormValue(paramContainerName)
		msgr          = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		pMsgr         = presets.MustGetMessages(ctx.R)
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: presets.DeleteConfirmPortalName,
		Body: web.Scope(
			vx.VXDialog(h.Text(msgr.AreWantDeleteContainer(containerName))).
				Title(msgr.ModalTitleConfirm).
				CancelText(pMsgr.Cancel).
				OkText(pMsgr.Delete).
				Attr("@click:ok", web.Plaid().
					EventFunc(DeleteContainerEvent).
					Query(paramContainerID, containerID).
					Query(paramStatus, ctx.Param(paramStatus)).
					ThenScript("locals.deleteConfirmation=false").
					Go()).
				Attr("v-model", "locals.deleteConfirmation"),
		).VSlot(`{ locals  }`).Init(`{deleteConfirmation: true}`),
	})

	return
}

func (b *ModelBuilder) deleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		container              Container
		pageID, pageVersion, _ = b.getPrimaryColumnValuesBySlug(ctx)
		cs                     = container.PrimaryColumnValuesBySlug(ctx.Param(paramContainerID))
		containerID            = cs[presets.ParamID]
		locale                 = cs[l10n.SlugLocaleCode]
		count                  int64
		obj                    interface{}
	)
	if obj, err = b.getObjFromSlug(ctx); err != nil {
		return
	}
	if err = b.db.Transaction(func(tx *gorm.DB) (dbErr error) {
		if dbErr = tx.Where("id = ? AND locale_code = ?", containerID, locale).First(&container).Error; dbErr != nil {
			return
		}
		if dbErr = tx.Delete(&Container{}, "id = ? AND locale_code = ?", containerID, locale).Error; err != nil {
			return
		}
		return tx.Model(&Container{}).Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ?", pageID, pageVersion, locale, b.name).Count(&count).Error
	}); err != nil {
		return
	}
	if b.builder.ab != nil && b.builder.editorActivityProcessor != nil {
		mb, ok := b.builder.ab.GetModelBuilder(b.mb)
		if !ok {
			return
		}
		detail := &EditorLogInput{
			Action:     ActionDeleteContainer,
			PageObject: obj,
			Container:  container,
			Detail:     fmt.Sprintf("%s %s", container.DisplayName, container.PrimarySlug()),
		}
		if b.builder.editorActivityProcessor != nil {
			detail = b.builder.editorActivityProcessor(ctx, detail)
		}
		if detail == nil {
			return
		}
		mb.Log(ctx.R.Context(), detail.Action, detail.PageObject, detail.Detail)
	}
	web.AppendRunScripts(&r,
		web.Plaid().PushState(true).ClearMergeQuery([]string{paramContainerID, paramContainerDataID}).RunPushState(),
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).Go(),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).Query(paramStatus, ctx.Param(paramStatus)).Go(),
		web.Plaid().EventFunc(EditContainerEvent).Go(),
		fmt.Sprintf("vars.emptyIframe=%v", count == 0),
	)
	return
}

func (b *Builder) renameContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		paramID  = ctx.R.FormValue(paramContainerID)
		name     = ctx.R.FormValue(paramContainerName)
		msgr     = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		pMsgr    = presets.MustGetMessages(ctx.R)
		okAction = web.Plaid().
				ThenScript("locals.renameDialog=false").
				EventFunc(RenameContainerFromDialogEvent).Query(paramContainerID, paramID).Go()
		portalName = dialogPortalName
	)

	if ctx.R.FormValue("portal") == "presets" {
		portalName = presets.DialogPortalName
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: portalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(h.Text(msgr.Rename)),
					VCardText(
						VTextField().Attr(web.VField("DisplayName", name)...).Variant(FieldVariantUnderlined),
					),
					VCardActions(
						VSpacer(),
						VBtn(pMsgr.Cancel).
							Variant(VariantFlat).
							Class("ml-2").
							On("click", "locals.renameDialog = false"),

						VBtn(pMsgr.OK).
							Color(ColorPrimary).
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", okAction),
					),
				),
			).MaxWidth("400px").
				Attr("v-model", "locals.renameDialog"),
		).Init("{renameDialog:true}").VSlot("{locals}"),
	})
	return
}

func (b *ModelBuilder) renderContainerHover(cb *ContainerBuilder, ctx *web.EventContext, msgr *Messages) h.HTMLComponent {
	containerName := cb.name
	if b.builder.pb.GetI18n() != nil {
		containerName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, cb.name)
	}
	addContainerEvent := web.Plaid().EventFunc(AddContainerEvent).
		MergeQuery(true).
		BeforeScript("pLocals.creating=true").
		Query(paramModelName, cb.name).
		Query(paramStatus, ctx.Param(paramStatus)).
		ThenScript("vars.overlay=false;xLocals.add=true").
		Go()
	return VHover(
		web.Slot(
			web.Scope(
				VListItem(
					VListItemTitle(h.Text(containerName)),
					web.Slot(VBtn(msgr.Add).Color(ColorPrimary).Size(SizeSmall).Attr("v-if", "isHovering")).Name(VSlotAppend),
				).Attr("v-bind", "props", ":active", "isHovering").
					Class("cursor-pointer").
					Attr("@click", fmt.Sprintf(`if (isHovering &&!pLocals.creating){%s}`, addContainerEvent)).
					Color(ColorPrimary),
			).VSlot("{ locals: pLocals }").Init(`{ creating : false }`),
		).Name("default").Scope(`{isHovering, props }`),
	).Attr("@update:model-value", fmt.Sprintf(`(val)=>{if (val){%s} }`,
		web.Plaid().EventFunc(ContainerPreviewEvent).
			Query(paramModelName, cb.name).
			Go(),
	))
}

func (b *ModelBuilder) renderSharedContainerHover(displayName, modelName string, modelID uint, ctx *web.EventContext, msgr *Messages) h.HTMLComponent {
	addContainerEvent := web.Plaid().EventFunc(AddContainerEvent).
		MergeQuery(true).
		BeforeScript("pLocals.creating=true").
		Query(paramModelName, modelName).
		Query(paramModelID, modelID).
		Query(paramSharedContainer, "true").
		Query(paramStatus, ctx.Param(paramStatus)).
		ThenScript("vars.overlay=false;xLocals.add=true").
		Go()
	return VHover(
		web.Slot(
			web.Scope(
				VListItem(
					VListItemTitle(h.Text(displayName)),
					web.Slot(VBtn(msgr.Add).Color(ColorPrimary).Size(SizeSmall).Attr("v-if", "isHovering")).Name(VSlotAppend),
				).Attr("v-bind", "props", ":active", "isHovering").
					Class("cursor-pointer").
					Attr("@click", fmt.Sprintf(`if (isHovering &&!pLocals.creating){%s}`, addContainerEvent)).
					Color(ColorPrimary),
			).VSlot("{ locals: pLocals }").Init(`{ creating : false }`),
		).Name("default").Scope(`{isHovering, props }`),
	).Attr("@update:model-value", fmt.Sprintf(`(val)=>{if (val){%s} }`,
		web.Plaid().EventFunc(ContainerPreviewEvent).
			Query(paramModelName, modelName).
			Query(paramModelID, modelID).
			Query(paramSharedContainer, "true").
			Go(),
	))
}

func (b *ModelBuilder) renderContainersList(ctx *web.EventContext) (component h.HTMLComponent) {
	var (
		msgr         = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		_, _, locale = b.getPrimaryColumnValuesBySlug(ctx)
	)
	var (
		containers        []h.HTMLComponent
		groupsNames       []string
		containerBuilders []*ContainerBuilder
	)
	containerBuilders = b.getContainerBuilders()
	if b.builder.disabledNormalContainersGroup {
		for _, cb := range containerBuilders {
			containers = append(containers,
				b.renderContainerHover(cb, ctx, msgr),
			)
		}
	} else {
		sort.Slice(containerBuilders, func(i, j int) bool {
			return containerBuilders[i].group != "" && containerBuilders[j].group == ""
		})
		groupContainers := utils.GroupBySlice[*ContainerBuilder, string](containerBuilders, func(builder *ContainerBuilder) string {
			return builder.group
		})
		for _, group := range groupContainers {
			if len(group) == 0 {
				break
			}
			groupName := group[0].group

			if b.builder.pb.GetI18n() != nil && groupName != "" {
				groupName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, groupName)
			}
			if groupName == "" {
				groupName = msgr.Others
			}
			if b.builder.expendContainers {
				groupsNames = append(groupsNames, groupName)
			}
			var listItems []h.HTMLComponent
			for _, builder := range group {
				listItems = append(listItems,
					b.renderContainerHover(builder, ctx, msgr),
				)
			}
			containers = append(containers, VListGroup(
				web.Slot(
					VListItem(
						VListItemTitle(
							h.Text(groupName),
						),
					).Attr("v-bind", "props"),
				).Name("activator").Scope(" { props}"),
				h.Components(listItems...),
			).Value(groupName))
		}
	}

	var (
		cons      []*Container
		listItems []h.HTMLComponent
	)

	b.db.Select("display_name,model_name,model_id").Where("shared = true AND locale_code = ?  ", locale).Group("display_name,model_name,model_id").Find(&cons)

	for _, con := range cons {
		listItems = append(listItems, b.renderSharedContainerHover(con.DisplayName, con.ModelName, con.ModelID, ctx, msgr))
	}
	containers = append(containers, VListGroup(
		web.Slot(
			VListItem(
				VListItemTitle(h.Text(msgr.Shared)),
			).Attr("v-bind", "props"),
		).Name("activator").Scope(" {  props }"),
		h.Components(listItems...),
	).Value(msgr.Shared))

	component = VList(containers...).Opened(groupsNames)
	return
}

func (b *Builder) renameContainerFromDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	var (
		paramID     = ctx.R.FormValue(paramContainerID)
		cs          = container.PrimaryColumnValuesBySlug(paramID)
		containerID = cs[presets.ParamID]
		locale      = cs[l10n.SlugLocaleCode]
		name        = ctx.R.FormValue(paramDisplayName)
		pMsgr       = i18n.MustGetModuleMessages(ctx.R, presets.CoreI18nModuleKey, Messages_en_US).(*presets.Messages)
	)
	err = b.db.First(&container, "id = ? AND locale_code = ?  ", containerID, locale).Error
	if err != nil {
		return
	}
	if container.Shared {
		err = b.db.Model(&Container{}).Where("model_name = ? AND model_id = ? AND locale_code = ?", container.ModelName, container.ModelID, locale).Update("display_name", name).Error
		if err != nil {
			return
		}
	} else {
		err = b.db.Model(&Container{}).Where("id = ? AND locale_code = ?", containerID, locale).Update("display_name", name).Error
		if err != nil {
			return
		}

	}
	web.AppendRunScripts(&r, web.Plaid().MergeQuery(true).Go(), fmt.Sprintf(` setTimeout(function(){ %s }, 200)`,
		presets.ShowSnackbarScript(pMsgr.SuccessfullyUpdated, ColorSuccess)))
	return
}

func (b *ModelBuilder) renameContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	var (
		paramID     = ctx.R.FormValue(paramContainerID)
		cs          = container.PrimaryColumnValuesBySlug(paramID)
		containerID = cs[presets.ParamID]
		locale      = cs[l10n.SlugLocaleCode]
		name        = ctx.R.FormValue(paramDisplayName)
		obj         interface{}
	)
	if obj, err = b.getObjFromSlug(ctx); err != nil {
		return
	}
	err = b.db.First(&container, "id = ? AND locale_code = ?  ", containerID, locale).Error
	if err != nil {
		return
	}
	diffs := []activity.Diff{
		{Field: fmt.Sprintf("[%s %v].DisplayName", container.DisplayName, container.ModelID), Old: container.DisplayName, New: name},
	}
	err = b.db.Model(&Container{}).Where("model_name = ? AND model_id = ? AND locale_code = ?", container.ModelName, container.ModelID, locale).Update("display_name", name).Error
	if err != nil {
		return
	}
	defer func() {
		if container.DisplayName != name && b.builder.ab != nil && b.builder.editorActivityProcessor != nil {
			detail := &EditorLogInput{
				Action:          activity.ActionEdit,
				PageObject:      obj,
				Container:       container,
				ContainerObject: nil,
				Detail:          diffs,
			}
			if b.builder.editorActivityProcessor != nil {
				detail = b.builder.editorActivityProcessor(ctx, detail)
			}
			if detail == nil {
				return
			}
			mb, ok := b.builder.ab.GetModelBuilder(b.mb)
			if !ok {
				return
			}
			mb.Log(ctx.R.Context(), detail.Action, detail.PageObject, detail.Detail)
		}
	}()
	web.AppendRunScripts(&r,
		fmt.Sprintf("vars.__pageBuilderRightContentTitle=%q", name),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).MergeQuery(true).Query(paramStatus, ctx.Param(paramStatus)).Go(),
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).MergeQuery(true).Go(),
	)
	return
}

func (b *ModelBuilder) reloadRenderPageOrTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var body h.HTMLComponent

	if body, err = b.renderPageOrTemplate(ctx, true, true, false); err != nil {
		return
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: editorPreviewContentPortal, Body: body})
	return
}

func (b *ModelBuilder) containerPreview(ctx *web.EventContext) (r web.EventResponse, err error) {
	var previewContainer h.HTMLComponent
	var (
		ID, _, locale = b.getPrimaryColumnValuesBySlug(ctx)
		obj           = b.mb.NewModel()
	)
	if err = b.db.First(&obj, ID).Error; err != nil {
		return
	}
	var body h.HTMLComponent
	if !b.builder.previewContainer {
		containerBuilder := b.builder.ContainerByName(ctx.Param(paramModelName))
		cover := containerBuilder.cover
		if cover == "" {
			cover = path.Join(b.builder.prefix, b.builder.imagesPrefix, strings.ReplaceAll(containerBuilder.name, " ", "")+".svg")
		}
		body = VImg().Src(cover).Width("100%").Height(200)
	} else {
		previewContainer, err = b.renderPreviewContainer(ctx, obj, locale, false, true)
		if err != nil {
			return
		}
		iframe := b.renderScrollIframe(h.Components(previewContainer), ctx, obj, locale, false, true, false)
		iframeBody := h.MustString(iframe, ctx.R.Context())
		body = h.Div(
			h.Iframe().Attr(":srcdoc", h.JSONString(iframeBody)).
				Attr("@load", `const iframe= $event.target;iframe.style.height=iframe.contentWindow.document.documentElement.scrollHeight+"px"`).
				Attr("frameborder", "0").Style("width:100%"),
		).
			Style("pointer-events: none;transform-origin: 0 0; transform:scale(0.25);width:400%;height:400%")

	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: addContainerDialogContentPortal,
		Body: VCard(body).MaxHeight(200).Elevation(0).Flat(true).Tile(true).Color(ColorGreyLighten3),
	})
	r.RunScript = "vars.containerPreview = true"
	return
}

func (b *ModelBuilder) replicateContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		container      Container
		cs             = container.PrimaryColumnValuesBySlug(ctx.Param(paramContainerID))
		containerID    = cs[presets.ParamID]
		locale         = cs[l10n.SlugLocaleCode]
		containerMb    *ContainerBuilder
		modelID        int
		newContainerID string
	)
	if err = b.db.Transaction(func(tx *gorm.DB) (dbErr error) {
		if dbErr = tx.Where("id = ? AND locale_code = ?", containerID, locale).First(&container).Error; dbErr != nil {
			return
		}
		containerMb = b.builder.ContainerByName(container.ModelName)
		model := containerMb.NewModel()
		if container.Shared {
			container.Shared = false
			// presets.ShowMessage(&r, "", ColorWarning)
		}
		container.ID = 0
		if dbErr = tx.First(&model, container.ModelID).Error; dbErr != nil {
			return
		}
		if dbErr = reflectutils.Set(model, "ID", uint(0)); dbErr != nil {
			return
		}
		ctx.WithContextValue(gorm2op.CtxKeyDB{}, tx)
		defer ctx.WithContextValue(gorm2op.CtxKeyDB{}, nil)
		if dbErr = containerMb.Editing().Creating().Saver(model, "", ctx); dbErr != nil {
			return
		}
		if dbErr = withLocale(
			b.builder,
			tx.Model(&Container{}).
				Where("page_id = ? and page_version = ? and page_model_name = ? and display_order > ? ", container.PageID, container.PageVersion, container.PageModelName, container.DisplayOrder),
			locale,
		).
			UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; dbErr != nil {
			return
		}
		container.DisplayOrder += 1
		container.ModelID = reflectutils.MustGet(model, "ID").(uint)
		modelID = int(container.ModelID)
		container.Hidden = false
		if dbErr = tx.Save(&container).Error; dbErr != nil {
			return
		}
		newContainerID = container.PrimarySlug()
		return
	}); err != nil {
		return
	}
	cb := b.builder.ContainerByName(container.ModelName)
	containerDataId := cb.getContainerDataID(modelID, newContainerID)
	web.AppendRunScripts(&r,
		web.Plaid().PushState(true).MergeQuery(true).
			Query(paramContainerDataID, containerDataId).
			Query(paramContainerID, newContainerID).RunPushState(),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).Query(paramContainerDataID, containerDataId).
			Query(paramStatus, ctx.Param(paramStatus)).Go(),
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).Query(paramContainerDataID, containerDataId).Go(),
		web.Plaid().EventFunc(EditContainerEvent).MergeQuery(true).Query(containerDataId, containerDataId).Go(),
	)
	return
}

func (b *ModelBuilder) editContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	data := strings.Split(ctx.Param(paramContainerDataID), "_")
	if len(data) <= 2 {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: pageBuilderRightContentPortal,
			Body: b.builder.emptyEdit(ctx),
		})
		return
	}
	r.RunScript = web.Plaid().
		URL("/"+strings.TrimLeft(path.Join(b.builder.pb.GetURIPrefix(), data[0]), "/")).
		EventFunc(actions.Edit).
		Query(presets.ParamID, data[1]).
		Query(presets.ParamPortalName, pageBuilderRightContentPortal).
		Query(presets.ParamOverlay, actions.Content).
		Query(paramDevice, cmp.Or(ctx.Param(paramDevice), b.builder.defaultDevice)).
		Go()
	return
}

func (b *ModelBuilder) updateContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		containerUri = ctx.Param(paramContainerUri)
		containerID  = ctx.Param(paramContainerID)
	)
	r.RunScript = web.Plaid().URL(containerUri).
		EventFunc(actions.Update).
		Query(presets.ParamID, containerID).
		Query(presets.ParamPortalName, pageBuilderRightContentPortal).
		ThenScript(
			web.Plaid().EventFunc(ReloadRenderPageOrTemplateBodyEvent).
				Query(paramStatus, ctx.Param(paramStatus)).MergeQuery(true).
				Query(paramIsUpdate, true).Go()).
		Query(presets.ParamOverlay, actions.Content).
		Go()
	return
}

func (b *ModelBuilder) reloadRenderPageOrTemplateBody(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		data []byte
		body h.HTMLComponent
	)
	iframeEventName := ctx.Param(paramIframeEventName)
	if iframeEventName == "" {
		iframeEventName = updateBodyEventName
	}
	if body, err = b.renderPageOrTemplate(ctx, true, true, true); err != nil {
		return
	}
	if data, err = body.MarshalHTML(ctx.R.Context()); err != nil {
		return
	}
	web.AppendRunScripts(&r,
		web.Emit(b.notifIframeBodyUpdated(),
			notifIframeBodyUpdatedPayload{
				Body:            string(data),
				ContainerDataID: ctx.Param(paramContainerDataID),
				IsUpdate:        ctx.Param(paramIsUpdate) == "true",
				EventName:       iframeEventName,
			},
		),
	)
	return
}

func (b *ModelBuilder) reloadAddContainersList(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: pageBuilderAddContainersPortal,
		Body: b.renderContainersList(ctx),
	})
	return
}

func (b *ModelBuilder) notifIframeBodyUpdated() string {
	return fmt.Sprintf("pageBuilder_notifIframeBodyUpdated_%v", b.name)
}

type notifIframeBodyUpdatedPayload struct {
	Body            string `json:"body"`
	ContainerDataID string `json:"containerDataID"`
	IsUpdate        bool   `json:"isUpdate"`
	EventName       string `json:"eventName"`
}
