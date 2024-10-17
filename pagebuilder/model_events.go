package pagebuilder

import (
	"encoding/json"
	"fmt"
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

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

func (b *ModelBuilder) registerFuncs() {
	b.editor.RegisterEventFunc(ShowSortedContainerDrawerEvent, b.showSortedContainerDrawer)
	b.editor.RegisterEventFunc(AddContainerEvent, b.addContainer)
	b.editor.RegisterEventFunc(DeleteContainerConfirmationEvent, b.deleteContainerConfirmation)
	b.editor.RegisterEventFunc(DeleteContainerEvent, b.deleteContainer)
	b.editor.RegisterEventFunc(MoveContainerEvent, b.moveContainer)
	b.editor.RegisterEventFunc(MoveUpDownContainerEvent, b.moveUpDownContainer)
	b.editor.RegisterEventFunc(ToggleContainerVisibilityEvent, b.toggleContainerVisibility)
	b.editor.RegisterEventFunc(RenameContainerDialogEvent, b.renameContainerDialog)
	b.editor.RegisterEventFunc(RenameContainerEvent, b.renameContainer)
	b.editor.RegisterEventFunc(ReloadRenderPageOrTemplateEvent, b.reloadRenderPageOrTemplate)
	b.editor.RegisterEventFunc(MarkAsSharedContainerEvent, b.markAsSharedContainer)
	b.editor.RegisterEventFunc(ContainerPreviewEvent, b.containerPreview)
	b.editor.RegisterEventFunc(ReplicateContainerEvent, b.replicateContainer)
	b.preview = web.Page(b.previewContent)
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
		status                      = ctx.R.FormValue(paramStatus)
		isReadonly                  = status != publish.StatusDraft && b.tb == nil
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
		wc["locale_code"] = locale
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
				ContainerDataID: fmt.Sprintf(`%s_%s`, inflection.Plural(strcase.ToKebab(c.ModelName)), strconv.Itoa(int(c.ModelID))),
			},
		)
	}
	pushState := web.Plaid().PushState(true).MergeQuery(true).
		Query(paramContainerDataID, web.Var(`element.label+"_"+element.model_id`))
	var clickColumnEvent string
	if !isReadonly {
		pushState.Query(paramContainerID, web.Var("element.param_id"))
		clickColumnEvent = fmt.Sprintf(`vars.%s=element.container_data_id;`, paramContainerDataID) +
			web.Plaid().
				URL(web.Var(fmt.Sprintf(`"%s/"+element.label`, b.builder.prefix))).
				EventFunc(actions.Edit).
				Query(presets.ParamOverlay, actions.Content).
				Query(presets.ParamPortalName, pageBuilderRightContentPortal).
				Query(presets.ParamID, web.Var("element.model_id")).
				Go() + ";" + pushState.RunPushState() +
			";" + scrollToContainer(fmt.Sprintf(`element.label+"_"+element.model_id`))
	}
	renameEvent := web.Plaid().
		URL(b.editorURL()).
		EventFunc(RenameContainerEvent).Query(paramStatus, status).Query(paramContainerID, web.Var("element.param_id")).Go()

	// container functions
	containerOperations :=
		h.Div(
			VBtn("").Variant(VariantText).Icon("mdi-content-copy").Size(SizeSmall).Attr("@click",
				web.Plaid().
					EventFunc(ReplicateContainerEvent).
					Query(paramContainerID, web.Var("element.param_id")).
					Go(),
			).Attr("v-show", "isHovering"),
			VMenu(
				web.Slot(
					VBtn("").Children(
						VIcon("mdi-dots-horizontal"),
					).Attr("v-bind", "props").Variant(VariantText).Size(SizeSmall),
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
					VListItem(h.Text(pMsgr.Delete)).PrependIcon("mdi-delete").Attr("@click",
						web.Plaid().
							URL(ctx.R.URL.Path).
							EventFunc(DeleteContainerConfirmationEvent).
							Query(paramContainerID, web.Var("element.param_id")).
							Query(paramContainerName, web.Var("element.display_name")).
							Go(),
					)),
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
														Attr("v-model", fmt.Sprintf("form.%s", paramsDisplayName)).
														Attr("v-if", "element.editShow").
														Attr("@blur", "element.editShow=false;"+renameEvent).
														Attr("@keyup.enter", renameEvent),
													VListItemTitle(h.Text("{{element.display_name}}")).Attr(":style", "[element.shared ? {'color':'green'}:{}]").Attr("v-if", "!element.editShow"),
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
						),
					).Attr("#item", " { element } "),
				),
			),
		).Class("px-4 overflow-y-auto").MaxHeight("86vh"),
		VBtn("").Children(
			web.Slot(
				VIcon("mdi-plus-circle-outline"),
			).Name(VSlotPrepend),
			h.Span(msgr.AddContainer).Class("ml-5"),
		).BaseColor(ColorPrimary).Variant(VariantText).Class(W100, "pl-14", "justify-start").
			Height(50).
			Attr("@click", appendVirtualElement()+web.Plaid().PushState(true).ClearMergeQuery([]string{paramContainerID}).RunPushState()+";vars.containerPreview=false;vars.overlay=true;vars.overlayEl.refs.overlay.showByElement($event)"),
	).Init(h.JSONString(sorterData)).VSlot("{ locals:sortLocals,form }")
	return
}

func (b *ModelBuilder) addContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		modelName       = ctx.Param(paramModelName)
		sharedContainer = ctx.Param(paramSharedContainer)
		modelID         = ctx.ParamAsInt(paramModelID)
		containerID     = ctx.Param(paramContainerID)
		newContainerID  string

		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
	)

	if sharedContainer == "true" {
		newContainerID, err = b.addSharedContainerToPage(pageID, containerID, pageVersion, locale, modelName, uint(modelID))
	} else {
		var newModelId uint
		newModelId, newContainerID, err = b.addContainerToPage(ctx, pageID, containerID, pageVersion, locale, modelName)
		modelID = int(newModelId)
	}
	cb := b.builder.ContainerByName(modelName)
	containerDataId := cb.getContainerDataID(modelID)
	r.RunScript = web.Plaid().PushState(true).MergeQuery(true).
		Query(paramContainerDataID, containerDataId).
		Query(paramContainerID, newContainerID).
		Form(map[string]string{paramContainerNew: "1"}).
		Go()
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
			EventFunc(ReloadRenderPageOrTemplateEvent).
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
	)
	paramID := ctx.R.FormValue(paramContainerID)
	direction := ctx.R.FormValue(paramMoveDirection)

	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]

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
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).MergeQuery(true).Go(),
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).MergeQuery(true).Query(paramStatus, ctx.Param(paramStatus)).Go(),
	)
	return
}

func (b *ModelBuilder) toggleContainerVisibility(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	var (
		paramID     = ctx.R.FormValue(paramContainerID)
		cs          = container.PrimaryColumnValuesBySlug(paramID)
		containerID = cs["id"]
		locale      = cs["locale_code"]
	)

	err = b.db.Exec("UPDATE page_builder_containers SET hidden = NOT(coalesce(hidden,FALSE)) WHERE id = ? AND locale_code = ?", containerID, locale).Error

	web.AppendRunScripts(&r,
		web.Plaid().
			EventFunc(ReloadRenderPageOrTemplateEvent).
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
					Go()).
				Attr("v-model", "locals.deleteConfirmation"),
		).VSlot(`{ locals  }`).Init(`{deleteConfirmation: true}`),
	})

	return
}

func (b *ModelBuilder) deleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	cs := container.PrimaryColumnValuesBySlug(ctx.Param(paramContainerID))
	containerID := cs["id"]
	locale := cs["locale_code"]

	err = b.db.Delete(&Container{}, "id = ? AND locale_code = ?", containerID, locale).Error
	if err != nil {
		return
	}
	r.RunScript = web.Plaid().PushState(true).ClearMergeQuery([]string{paramContainerID, paramContainerDataID}).Go()
	return
}

func (b *ModelBuilder) renameContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		paramID  = ctx.R.FormValue(paramContainerID)
		name     = ctx.R.FormValue(paramContainerName)
		msgr     = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		pMsgr    = presets.MustGetMessages(ctx.R)
		okAction = web.Plaid().
			URL(b.editorURL()).
			EventFunc(RenameContainerEvent).Query(paramContainerID, paramID).Go()
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
			containerName := cb.name
			if b.builder.ps.GetI18n() != nil {
				containerName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, cb.name)
			}
			addContainerEvent := web.Plaid().EventFunc(AddContainerEvent).
				MergeQuery(true).
				Query(paramModelName, cb.name).
				Go()
			containers = append(containers,
				VHover(
					web.Slot(
						VListItem(
							VListItemTitle(h.Text(containerName)),
							web.Slot(VBtn(msgr.Add).Color(ColorPrimary).Size(SizeSmall).Attr("v-if", "isHovering")).Name(VSlotAppend),
						).Attr("v-bind", "props", ":active", "isHovering").
							Class("cursor-pointer").
							Attr("@click", fmt.Sprintf(`isHovering?%s:null`, addContainerEvent)).
							ActiveColor(ColorPrimary),
					).Name("default").Scope(`{isHovering, props }`),
				).Attr("@update:model-value", fmt.Sprintf(`(val)=>{if (val){%s} }`,
					web.Plaid().EventFunc(ContainerPreviewEvent).
						Query(paramModelName, cb.name).
						Go(),
				),
				),
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

			if b.builder.ps.GetI18n() != nil && groupName != "" {
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
				containerName := builder.name
				if b.builder.ps.GetI18n() != nil {
					containerName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, builder.name)
				}
				addContainerEvent := web.Plaid().EventFunc(AddContainerEvent).
					MergeQuery(true).
					Query(paramModelName, builder.name).
					Go()
				listItems = append(listItems,
					VHover(
						web.Slot(
							VListItem(
								VListItemTitle(h.Text(containerName)),
								web.Slot(VBtn(msgr.Add).Color(ColorPrimary).Size(SizeSmall).Attr("v-if", "isHovering")).Name(VSlotAppend),
							).Attr("v-bind", "props", ":active", "isHovering").
								Class("cursor-pointer").
								Attr("@click", fmt.Sprintf(`isHovering?%s:null`, addContainerEvent)).
								ActiveColor(ColorPrimary),
						).Name("default").Scope(`{isHovering, props }`),
					).Attr("@update:model-value", fmt.Sprintf(`(val)=>{if (val){%s} }`,
						web.Plaid().EventFunc(ContainerPreviewEvent).
							Query(paramModelName, builder.name).
							Go(),
					),
					),
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

	var cons []*Container

	b.db.Select("display_name,model_name,model_id").Where("shared = true AND locale_code = ? and page_model_name = ? ", locale, b.name).Group("display_name,model_name,model_id").Find(&cons)
	sort.Slice(cons, func(i, j int) bool {
		return b.builder.ContainerByName(cons[i].ModelName).group != "" && b.builder.ContainerByName(cons[j].ModelName).group == ""
	})
	for _, group := range utils.GroupBySlice[*Container, string](cons, func(builder *Container) string {
		return b.builder.ContainerByName(builder.ModelName).group
	}) {
		if len(group) == 0 {
			break
		}
		groupName := msgr.Shared
		if b.builder.expendContainers {
			groupsNames = append(groupsNames, groupName)
		}
		var listItems []h.HTMLComponent
		for _, builder := range group {
			c := b.builder.ContainerByName(builder.ModelName)
			containerName := c.name
			if b.builder.ps.GetI18n() != nil {
				containerName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, c.name)
			}
			listItems = append(listItems,
				VListItem(
					VListItemTitle(h.Text(containerName)).Attr("@click", web.Plaid().
						EventFunc(AddContainerEvent).
						MergeQuery(true).
						Query(paramContainerName, builder.ModelName).
						Query(paramModelName, builder.ModelName).
						Query(paramModelID, builder.ModelID).
						Query(paramSharedContainer, "true").
						Go()),
				).Value(containerName))
		}

		containers = append(containers, VListGroup(
			web.Slot(
				VListItem(
					VListItemTitle(h.Text(groupName)),
				).Attr("v-bind", "props"),
			).Name("activator").Scope(" {  props }"),
			h.Components(listItems...),
		).Value(groupName))
	}
	component = VList(containers...).Opened(groupsNames)
	return
}

func (b *ModelBuilder) renameContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	var (
		paramID     = ctx.R.FormValue(paramContainerID)
		cs          = container.PrimaryColumnValuesBySlug(paramID)
		containerID = cs["id"]
		locale      = cs["locale_code"]
		name        = ctx.R.FormValue(paramsDisplayName)
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
	web.AppendRunScripts(&r,
		web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).MergeQuery(true).Query(paramStatus, ctx.Param(paramStatus)).Go(),
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).MergeQuery(true).Go(),
	)
	return
}

func (b *ModelBuilder) reloadRenderPageOrTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var body h.HTMLComponent
	obj := b.mb.NewModel()

	if body, err = b.renderPageOrTemplate(ctx, obj, true, true); err != nil {
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
		iframe := b.rendering(h.Components(previewContainer), ctx, obj, locale, false, true, true)
		body = h.Div(iframe).
			Style("pointer-events: none;transform-origin: 0 0; transform:scale(0.25);width:400%")

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
		var (
			model = containerMb.NewModel()
		)
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
	r.RunScript = web.Plaid().Query(paramContainerDataID, containerMb.getContainerDataID(modelID)).Query(paramContainerID, newContainerID).PushState(true).Go()
	return
}
