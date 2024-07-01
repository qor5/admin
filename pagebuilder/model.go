package pagebuilder

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/sunfmin/reflectutils"

	"github.com/qor5/admin/v3/utils"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

type (
	ModelBuilder struct {
		name    string
		mb      *presets.ModelBuilder
		editor  *presets.ModelBuilder
		db      *gorm.DB
		builder *Builder
		preview http.Handler
	}
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
	b.editor.RegisterEventFunc(NewContainerDialogEvent, b.newContainerDialog)
	b.editor.RegisterEventFunc(ContainerPreviewEvent, b.containerPreview)
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
		isReadonly                  = status != publish.StatusDraft
		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
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
		displayName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, c.DisplayName)

		sorterData.Items = append(sorterData.Items,
			ContainerSorterItem{
				Index:           i,
				Label:           inflection.Plural(strcase.ToKebab(c.ModelName)),
				ModelName:       c.ModelName,
				ModelID:         strconv.Itoa(int(c.ModelID)),
				DisplayName:     displayName,
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
	r = web.Scope(
		VSheet(
			VList(
				h.Tag("vx-draggable").
					Attr("item-key", "model_id").
					Attr("v-model", "sortLocals.items", "handle", ".handle", "animation", "300").
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
													VTextField().HideDetails(true).Density(DensityCompact).Color(ColorPrimary).Autofocus(true).Variant(FieldVariantOutlined).
														Attr("v-model", fmt.Sprintf("form.%s", paramsDisplayName)).
														Attr("v-if", "element.editShow").
														Attr("@blur", "element.editShow=false").
														Attr("@keyup.enter", web.Plaid().
															URL(fmt.Sprintf("%s/%s-editors", b.builder.prefix, b.name)).
															EventFunc(RenameContainerEvent).Query(paramStatus, status).Query(paramContainerID, web.Var("element.param_id")).Go()),
													VListItemTitle(h.Text("{{element.display_name}}")).Attr(":style", "[element.shared ? {'color':'green'}:{}]").Attr("v-if", "!element.editShow"),
												).VSlot("{form}").FormInit("{ DisplayName:element.display_name }"),
											),
										),
										web.Slot(
											h.If(!isReadonly,
												h.Div(
													VBtn("").Variant(VariantText).Icon("mdi-pencil").Attr("@click",
														"element.editShow=!element.editShow",
													).Attr("v-show", "!element.editShow && !element.hidden && isHovering"),
													VBtn("").Variant(VariantText).Attr(":icon", "element.visibility_icon").Size(SizeSmall).Attr("@click",
														web.Plaid().
															EventFunc(ToggleContainerVisibilityEvent).
															Query(paramContainerID, web.Var("element.param_id")).
															Query(paramStatus, status).
															Go(),
													).Attr("v-show", "!element.editShow && (element.hidden || isHovering)"),
													VBtn("").Variant(VariantText).Icon("mdi-delete").Attr("@click",
														web.Plaid().
															URL(ctx.R.URL.Path).
															EventFunc(DeleteContainerConfirmationEvent).
															Query(paramContainerID, web.Var("element.param_id")).
															Query(paramContainerName, web.Var("element.display_name")).
															Go(),
													).Attr("v-show", "!element.editShow && !element.hidden && isHovering"),
												),
											),
										).Name("append"),
									).Attr(":variant", fmt.Sprintf(` element.hidden &&!isHovering && !element.editShow?"%s":"%s"`, VariantPlain, VariantText)).
										Attr(":class", fmt.Sprintf(`element.container_data_id==vars.%s && !element.hidden?"bg-%s":""`, paramContainerDataID, ColorPrimaryLighten2)).
										Attr("v-bind", "props", "@click", clickColumnEvent).
										Attr(web.VAssign("vars",
											fmt.Sprintf(`{%s:"%s"}`, paramContainerDataID, ctx.Param(paramContainerDataID)))...),
								).Name("default").Scope("{ isHovering, props }"),
							),
							VDivider(),
						),
					).Attr("#item", " { element } "),
				),
				VListItem(
					web.Slot(
						VIcon("mdi-plus-circle-outline"),
					).Name(VSlotPrepend),
					h.Span("New Element"),
				).BaseColor(ColorPrimary).Class(W100, "ml-4").
					Attr("@click", web.Plaid().EventFunc(NewContainerDialogEvent).Go()),
			),
		).Class("pa-4 pt-2"),
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
		newModelId, newContainerID, err = b.addContainerToPage(pageID, containerID, pageVersion, locale, modelName)
		modelID = int(newModelId)
	}
	cb := b.builder.ContainerByName(modelName)

	r.RunScript = web.Plaid().PushState(true).MergeQuery(true).
		Query(paramContainerDataID, cb.getContainerDataID(modelID)).
		Query(paramContainerID, newContainerID).Go()
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
	ctx.R.Form.Del(paramMoveResult)
	r.RunScript = web.Plaid().URL(ctx.R.URL.Path).EventFunc(ReloadRenderPageOrTemplateEvent).Form(nil).Queries(ctx.R.Form).Go()
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
	r.RunScript = web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).MergeQuery(true).Go()
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
	r.RunScript = web.Plaid().
		EventFunc(ReloadRenderPageOrTemplateEvent).
		MergeQuery(true).
		Go() +
		";" +
		web.Plaid().
			EventFunc(ShowSortedContainerDrawerEvent).
			MergeQuery(true).
			Query(paramStatus, ctx.Param(paramStatus)).
			Go()
	return
}

func (b *ModelBuilder) deleteContainerConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		containerID   = ctx.R.FormValue(paramContainerID)
		containerName = ctx.R.FormValue(paramContainerName)
	)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: presets.DeleteConfirmPortalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(h.Text(fmt.Sprintf("Are you sure you want to delete %s?", containerName))),
					VCardActions(
						VSpacer(),
						VBtn("Cancel").
							Variant(VariantFlat).
							Class("ml-2").
							Attr("@click", "locals.deleteConfirmation = false"),

						VBtn("Delete").
							Color("primary").
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", web.Plaid().
								EventFunc(DeleteContainerEvent).
								Query(paramContainerID, containerID).
								Go()),
					),
				),
			).MaxWidth("600px").
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
	paramID := ctx.R.FormValue(paramContainerID)
	name := ctx.R.FormValue(paramContainerName)
	okAction := web.Plaid().
		URL(fmt.Sprintf("%s/%s-editors", b.builder.prefix, b.name)).
		EventFunc(RenameContainerEvent).Query(paramContainerID, paramID).Go()
	portalName := dialogPortalName
	if ctx.R.FormValue("portal") == "presets" {
		portalName = presets.DialogPortalName
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: portalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(h.Text("Rename")),
					VCardText(
						VTextField().Attr(web.VField("DisplayName", name)...).Variant(FieldVariantUnderlined),
					),
					VCardActions(
						VSpacer(),
						VBtn("Cancel").
							Variant(VariantFlat).
							Class("ml-2").
							On("click", "locals.renameDialog = false"),

						VBtn("OK").
							Color("primary").
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
		if groupName == "" {
			groupName = "Others"
		}
		if b.builder.expendContainers {
			groupsNames = append(groupsNames, groupName)
		}
		var listItems []h.HTMLComponent
		for _, builder := range group {
			containerName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, builder.name)
			listItems = append(listItems,
				VListItem(
					VListItemTitle(h.Text(containerName)),
				).Attr("@click",
					web.Plaid().EventFunc(ContainerPreviewEvent).
						Query(paramModelName, builder.name).
						Go(),
				))
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
			containerName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, c.name)
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

	r.RunScript = web.Plaid().EventFunc(ShowSortedContainerDrawerEvent).Query(paramStatus, ctx.Param(paramStatus)).Go() + ";" +
		web.Plaid().EventFunc(ReloadRenderPageOrTemplateEvent).MergeQuery(true).Go()
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

func (b *ModelBuilder) getContainerBuilders() (cons []*ContainerBuilder) {
	for _, builder := range b.builder.containerBuilders {
		if builder.modelBuilder == nil || b.mb == builder.modelBuilder {
			cons = append(cons, builder)
		}
	}
	return
}

func (b *ModelBuilder) setName() {
	b.name = utils.GetObjectName(b.mb.NewModel())
}

func (b *ModelBuilder) addSharedContainerToPage(pageID int, containerID, pageVersion, locale, modelName string, modelID uint) (newContainerID string, err error) {
	var c Container
	err = b.db.First(&c, "model_name = ? AND model_id = ? AND shared = true and page_model_name = ? ", modelName, modelID, b.name).Error
	if err != nil {
		return
	}
	var (
		maxOrder     sql.NullFloat64
		displayOrder float64
	)
	err = b.db.Model(&Container{}).Select("MAX(display_order)").
		Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ?", pageID, pageVersion, locale, b.name).Scan(&maxOrder).Error
	if err != nil {
		return
	}
	err = b.db.Model(&Container{}).Select("MAX(display_order)").
		Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ? ", pageID, pageVersion, locale, b.name).Scan(&maxOrder).Error
	if err != nil {
		return
	}
	if containerID != "" {
		var lastContainer Container
		cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
		if dbErr := b.db.Where("id = ? AND locale_code = ? and page_model_name = ? ", cs["id"], locale, b.name).First(&lastContainer).Error; dbErr == nil {
			displayOrder = lastContainer.DisplayOrder
			if err = b.db.Model(&Container{}).Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ? and display_order > ? ", pageID, pageVersion, locale, b.name, displayOrder).
				UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; err != nil {
				return
			}
		}

	} else {
		displayOrder = maxOrder.Float64
	}
	container := Container{
		PageID:        uint(pageID),
		PageVersion:   pageVersion,
		ModelName:     modelName,
		PageModelName: b.name,
		DisplayName:   c.DisplayName,
		ModelID:       modelID,
		Shared:        true,
		DisplayOrder:  displayOrder + 1,
		Locale: l10n.Locale{
			LocaleCode: locale,
		},
	}
	err = b.db.Create(&container).Error
	if err != nil {
		return
	}
	containerID = container.PrimarySlug()

	return
}

func withLocale(builder *Builder, wh *gorm.DB, locale string) *gorm.DB {
	if builder.l10n == nil {
		return wh
	}
	return wh.Where("locale_code = ?", locale)
}

func (b *ModelBuilder) addContainerToPage(pageID int, containerID, pageVersion, locale, modelName string) (modelID uint, newContainerID string, err error) {
	model := b.builder.ContainerByName(modelName).NewModel()
	var dc DemoContainer
	b.db.Where("model_name = ? AND locale_code = ?", modelName, locale).First(&dc)
	if dc.ID != 0 && dc.ModelID != 0 {
		b.db.Where("id = ?", dc.ModelID).First(model)
		reflectutils.Set(model, "ID", uint(0))
	}
	err = b.db.Create(model).Error
	if err != nil {
		return
	}

	var (
		maxOrder     sql.NullFloat64
		displayOrder float64
	)
	wh := b.db.Model(&Container{}).Select("MAX(display_order)").
		Where("page_id = ? and page_version = ? and page_model_name = ? ", pageID, pageVersion, b.name)

	err = withLocale(b.builder, wh, locale).Scan(&maxOrder).Error
	if err != nil {
		return
	}
	if containerID != "" {
		var lastContainer Container
		cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
		if dbErr := b.db.Where("id = ? AND locale_code = ? and page_model_name = ?", cs["id"], locale, b.name).First(&lastContainer).Error; dbErr == nil {
			displayOrder = lastContainer.DisplayOrder
			if err = withLocale(
				b.builder,
				b.db.Model(&Container{}).
					Where("page_id = ? and page_version = ? and page_model_name = ? and display_order > ? ", pageID, pageVersion, b.name, displayOrder),
				locale,
			).
				UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; err != nil {
				return
			}
		}

	} else {
		displayOrder = maxOrder.Float64
	}
	modelID = reflectutils.MustGet(model, "ID").(uint)
	container := Container{
		PageID:        uint(pageID),
		PageVersion:   pageVersion,
		ModelName:     modelName,
		PageModelName: b.name,
		DisplayName:   modelName,
		ModelID:       modelID,
		DisplayOrder:  displayOrder + 1,
		Locale: l10n.Locale{
			LocaleCode: locale,
		},
	}
	err = b.db.Create(&container).Error
	newContainerID = container.PrimarySlug()
	return
}

func (b *ModelBuilder) pageContent(ctx *web.EventContext, obj interface{}) (r web.PageResponse, err error) {
	var body h.HTMLComponent

	if body, err = b.renderPageOrTemplate(ctx, obj, true, true); err != nil {
		return
	}
	r.Body = web.Portal(
		h.Div(body).Attr(web.VAssign("vars", "{el:$}")...),
	).Name(editorPreviewContentPortal)
	return
}

func (b *ModelBuilder) getPrimaryColumnValuesBySlug(ctx *web.EventContext) (pageID int, pageVersion string, locale string) {
	var (
		ps map[string]string

		obj = b.mb.NewModel()
	)
	if p, ok := obj.(PrimarySlugInterface); ok {
		ps = p.PrimaryColumnValuesBySlug(ctx.Param(presets.ParamID))
	}
	pageVersion = ps[publish.SlugVersion]
	locale = ps[l10n.SlugLocaleCode]
	pageIDi, _ := strconv.ParseInt(ps["id"], 10, 64)
	pageID = int(pageIDi)
	return
}

func (b *ModelBuilder) renderPageOrTemplate(ctx *web.EventContext, obj interface{}, isEditor, isIframe bool) (r h.HTMLComponent, err error) {
	var (
		isTpl                       = ctx.R.FormValue(paramsTpl) != ""
		status                      = publish.StatusDraft
		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
	)
	if isTpl {
		tpl := &Template{}
		if err = b.db.First(tpl, "id = ? and locale_code = ?", pageID, pageVersion).Error; err != nil {
			return
		}
		p := tpl.Page()
		pageVersion = p.Version.Version
	} else {
		g := b.db.Where("id = ? and version = ?", pageID, pageVersion)
		if locale != "" {
			g.Where("locale_code = ?", locale)
		}
		if err = g.First(obj).Error; err != nil {
			return
		}
	}
	if p, ok := obj.(l10n.LocaleInterface); ok {
		locale = p.EmbedLocale().LocaleCode
	}
	var isReadonly bool
	if p, ok := obj.(publish.StatusInterface); ok {
		status = p.EmbedStatus().Status
	}
	if status != publish.StatusDraft && isEditor {
		isReadonly = true
	}
	var comps []h.HTMLComponent
	comps, err = b.renderContainers(ctx, pageID, pageVersion, locale, isEditor, isReadonly)
	if err != nil {
		return
	}
	r = b.rendering(comps, ctx, obj, locale, isEditor, isIframe, isReadonly)
	return
}

func (b *ModelBuilder) rendering(comps []h.HTMLComponent, ctx *web.EventContext, obj interface{}, locale string, isEditor, isIframe, isReadonly bool) (r h.HTMLComponent) {
	r = h.Components(comps...)
	if b.builder.pageLayoutFunc != nil {
		var seoTags h.HTMLComponent
		if b.builder.seoBuilder != nil {
			seoTags = b.builder.seoBuilder.Render(obj, ctx.R)
		}
		input := &PageLayoutInput{
			LocaleCode: locale,
			IsEditor:   isEditor,
			IsPreview:  !isEditor,
			SeoTags:    seoTags,
		}

		if isEditor {
			input.EditorCss = append(input.EditorCss, h.RawHTML(`<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">`))
			input.EditorCss = append(input.EditorCss, h.Style(`
			.wrapper-shadow{
			  position: relative;
			  width: 100%;	
			}
			.inner-shadow {
			  position: absolute;
			  width: 100%;
			  height: 100%;
			  opacity: 0;
			  top: 0;
			  left: 0;
			  box-shadow: 3px 3px 0 0px #3E63DD inset, -3px 3px 0 0px #3E63DD inset;
			}
			
			
			.editor-add {
			  width: 100%;
			  position: absolute;
			  z-index: 9998;
			  opacity: 0;
			  text-align: center;
			}
			
			.editor-add div {
			  width: 100%;
			  background-color: #3E63DD;
			  height: 3px;
			}
			
			.editor-add button {
			  width: 32px;
			  cursor: pointer;
              height: 32px;	
			  color: #FFFFFF;
			  background-color: #3E63DD;
			  pointer-event: none;
              position: absolute;
              bottom: -14px;
              padding: 4px 0 4px 0;
			}
			.wrapper-shadow:hover {
			  cursor: pointer;
			}
			
			.wrapper-shadow:hover .editor-add {
			  opacity: 1;
			}
			
			.wrapper-shadow:hover .editor-add div {
			  height: 6px;
			}
			
			.editor-bar {
			  position: absolute;
			  z-index: 9999;
			  width: 32%;
			  height: 32px;	
			  opacity: 0;
              display: flex;
			  align-items: center;	
			  background-color: #3E63DD;
			  justify-content: space-between;
              pointer-events: none;
              padding : 0 8px;

			}
   			.editor-bar-buttons{
              height: 24px;
			}
			.editor-bar button {
			  color: #FFFFFF;
			  cursor: pointer;
			  background-color: #3E63DD; 
              height: 24px;	
			}
			
			.editor-bar .title {
			  color: #FFFFFF;
		      width: 30%;
			  overflow: hidden;	
			  font-size: 12px;
			  font-style: normal;
			  font-weight: 400;
			  line-height: 16px; 
              text-overflow: ellipsis;
			  letter-spacing: 0.04px;	
			}
			
			.highlight .editor-bar {
			  opacity: 1;
              pointer-events: auto;
			}
			
			.highlight .editor-add {
			  opacity: 1;
			}
			
			.highlight .inner-shadow {
			  opacity: 1;
			}
`))
		}
		if f := ctx.R.Context().Value(ContainerToPageLayoutKey); f != nil {
			pl, ok := f.(*PageLayoutInput)
			if ok {
				input.FreeStyleCss = append(input.FreeStyleCss, pl.FreeStyleCss...)
				input.FreeStyleTopJs = append(input.FreeStyleTopJs, pl.FreeStyleTopJs...)
				input.FreeStyleBottomJs = append(input.FreeStyleBottomJs, pl.FreeStyleBottomJs...)
				input.Hreflang = pl.Hreflang
			}
		}

		if isIframe {
			iframeHeightCookie, _ := ctx.R.Cookie(iframeHeightName)
			iframeValue := "1000px"
			_ = iframeValue
			if iframeHeightCookie != nil {
				iframeValue = iframeHeightCookie.Value
			}
			// use newCtx to avoid inserting page head to head outside of iframe
			newCtx := &web.EventContext{
				R:        ctx.R,
				Injector: &web.PageInjector{},
			}
			r = b.builder.pageLayoutFunc(h.Components(comps...), input, newCtx)
			newCtx.Injector.HeadHTMLComponent("style", b.builder.pageStyle, true)
			r = h.HTMLComponents{
				h.RawHTML("<!DOCTYPE html>\n"),
				h.Tag("html").Children(
					h.Head(
						newCtx.Injector.GetHeadHTMLComponent(),
					),
					h.Body(
						h.Div(
							r,
						).Id("app").Attr("v-cloak", true),
						newCtx.Injector.GetTailHTMLComponent(),
					).Class("front"),
				).Attr(newCtx.Injector.HTMLLangAttrs()...),
			}
			_, width := b.builder.getDevice(ctx)
			scrollIframe := h.Tag("vx-scroll-iframe").Attr(
				":srcdoc", h.JSONString(h.MustString(r, ctx.R.Context())),
				"iframe-height", iframeValue,
				"iframe-height-name", iframeHeightName,
				"width", width,
				":container-data-id", fmt.Sprintf(`vars.containerTab=="%s"?"%s":""`, EditorTabAdd, ctx.Param(paramContainerDataID)),
				"ref", "scrollIframe").
				Attr(web.VAssign("vars", `{el:$}`)...)
			r = scrollIframe
			if !isReadonly && len(comps) == 0 {
				r = h.Components(
					scrollIframe.Attr("v-show", fmt.Sprintf(`vars.containerTab=="%s"`, EditorTabAdd)),
					h.Div(
						h.RawHTML(defaultContainerEmptyIcon),
						h.Div(h.Text("Please add your elements first")),
					).Style("display:flex;justify-content:center;align-items:center;flex-direction:column;height:80vh").
						Attr("v-show", fmt.Sprintf(`vars.containerTab!="%s"`, EditorTabAdd)),
				)
				return
			}

		} else {
			r = b.builder.pageLayoutFunc(h.Components(comps...), input, ctx)
			ctx.Injector.HeadHTMLComponent("style", b.builder.pageStyle, true)
		}
	}
	return
}

func (b *ModelBuilder) renderContainers(ctx *web.EventContext, pageID int, pageVersion string, locale string, isEditor bool, isReadonly bool) (r []h.HTMLComponent, err error) {
	var cons []*Container
	err = withLocale(
		b.builder,
		b.db.
			Order("display_order ASC").
			Where("page_id = ? AND page_version = ? and page_model_name = ? ", pageID, pageVersion, b.name),
		locale,
	).
		Find(&cons).Error
	if err != nil {
		return
	}
	device, _ := b.builder.getDevice(ctx)
	cbs := b.builder.getContainerBuilders(cons)
	for i, ec := range cbs {
		if ec.container.Hidden {
			continue
		}
		obj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(obj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}
		displayName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, ec.container.DisplayName)
		input := RenderInput{
			IsEditor:    isEditor,
			IsReadonly:  isReadonly,
			Device:      device,
			ContainerId: ec.container.PrimarySlug(),
			DisplayName: displayName,
		}
		pure := ec.builder.renderFunc(obj, &input, ctx)

		r = append(r, b.builder.containerWrapper(pure.(*h.HTMLTagBuilder), ctx, isEditor, isReadonly, i == 0, i == len(cbs)-1,
			ec.builder.getContainerDataID(int(ec.container.ModelID)), ec.container.ModelName, &input))
	}

	return
}

func (b *ModelBuilder) renderPreviewContainer(ctx *web.EventContext, locale string, isEditor, IsReadonly bool) (r h.HTMLComponent, err error) {
	var (
		modelName       = ctx.Param(paramModelName)
		sharedContainer = ctx.Param(paramSharedContainer)
		modelID         = ctx.ParamAsInt(paramModelID)
	)
	if sharedContainer != "true" || modelID == 0 {
		var con *DemoContainer
		err = withLocale(
			b.builder,
			b.db.
				Where("model_name = ?", modelName),
			locale,
		).
			First(&con).Error
		if err != nil {
			return
		}
		modelID = int(con.ModelID)
	}

	containerBuilder := b.builder.ContainerByName(modelName)
	device, _ := b.builder.getDevice(ctx)

	displayName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, modelName)
	input := RenderInput{
		IsEditor:    isEditor,
		IsReadonly:  IsReadonly,
		Device:      device,
		ContainerId: "",
		DisplayName: displayName,
	}
	obj := containerBuilder.NewModel()
	err = b.db.FirstOrCreate(obj, "id = ?", modelID).Error
	if err != nil {
		return
	}
	pure := containerBuilder.renderFunc(obj, &input, ctx)
	r = b.builder.containerWrapper(pure.(*h.HTMLTagBuilder), ctx, isEditor, IsReadonly, false, false,
		containerBuilder.getContainerDataID(modelID), modelName, &input)
	return
}

func (b *ModelBuilder) previewContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	obj := b.mb.NewModel()
	r.Body, err = b.renderPageOrTemplate(ctx, obj, false, false)
	if err != nil {
		return
	}
	if p, ok := obj.(PageTitleInterface); ok {
		r.PageTitle = p.GetTitle()
	}
	return
}

func (b *ModelBuilder) markAsSharedContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	paramID := ctx.R.FormValue(paramContainerID)
	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]

	err = b.db.Model(&Container{}).Where("id = ? AND locale_code = ?", containerID, locale).Update("shared", true).Error
	if err != nil {
		return
	}
	r.PushState = web.Location(url.Values{})
	return
}

func (b *ModelBuilder) copyContainersToNewPageVersion(db *gorm.DB, pageID int, locale, oldPageVersion, newPageVersion string) (err error) {
	return b.copyContainersToAnotherPage(db, pageID, oldPageVersion, locale, pageID, newPageVersion, locale)
}

func (b *ModelBuilder) copyContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion, locale string, toPageID int, toPageVersion, toPageLocale string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ? and page_model_name =? ", pageID, pageVersion, locale, b.name).Error
	if err != nil {
		return
	}

	for _, c := range cons {
		newModelID := c.ModelID
		if !c.Shared {
			model := b.builder.ContainerByName(c.ModelName).NewModel()
			if err = db.First(model, "id = ?", c.ModelID).Error; err != nil {
				return
			}
			if err = reflectutils.Set(model, "ID", uint(0)); err != nil {
				return
			}
			if err = db.Create(model).Error; err != nil {
				return
			}
			newModelID = reflectutils.MustGet(model, "ID").(uint)
		}

		if err = db.Create(&Container{
			PageID:        uint(toPageID),
			PageVersion:   toPageVersion,
			PageModelName: b.name,
			ModelName:     c.ModelName,
			DisplayName:   c.DisplayName,
			ModelID:       newModelID,
			DisplayOrder:  c.DisplayOrder,
			Shared:        c.Shared,
			Locale: l10n.Locale{
				LocaleCode: toPageLocale,
			},
		}).Error; err != nil {
			return
		}
	}
	return
}

func (b *ModelBuilder) localizeContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion, locale string, toPageID int, toPageVersion, toPageLocale string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").
		Where("page_id = ? AND page_version = ? AND locale_code = ? and page_model_name = ? ", pageID, pageVersion, locale, b.name).
		Find(&cons).Error
	if err != nil {
		return
	}

	for _, c := range cons {
		newModelID := c.ModelID
		newDisplayName := c.DisplayName
		if !c.Shared {
			model := b.builder.ContainerByName(c.ModelName).NewModel()
			if err = db.First(model, "id = ?", c.ModelID).Error; err != nil {
				return
			}
			if err = reflectutils.Set(model, "ID", uint(0)); err != nil {
				return
			}
			if err = db.Create(model).Error; err != nil {
				return
			}
			newModelID = reflectutils.MustGet(model, "ID").(uint)
		} else {
			var count int64
			var sharedCon Container
			if err = db.Where("model_name = ? AND localize_from_model_id = ? AND locale_code = ? AND shared = ? and page_model_name = ? ",
				c.ModelName, c.ModelID, toPageLocale, true, b.name).
				First(&sharedCon).Count(&count).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return
			}

			if count == 0 {
				model := b.builder.ContainerByName(c.ModelName).NewModel()
				if err = db.First(model, "id = ?", c.ModelID).Error; err != nil {
					return
				}
				if err = reflectutils.Set(model, "ID", uint(0)); err != nil {
					return
				}
				if err = db.Create(model).Error; err != nil {
					return
				}
				newModelID = reflectutils.MustGet(model, "ID").(uint)
			} else {
				newModelID = sharedCon.ModelID
				newDisplayName = sharedCon.DisplayName
			}
		}

		var newCon Container
		err = db.Order("display_order ASC").Find(&newCon, "id = ? AND locale_code = ?", c.ID, toPageLocale).Error
		if err != nil {
			return
		}

		newCon.ID = c.ID
		newCon.PageID = uint(toPageID)
		newCon.PageVersion = toPageVersion
		newCon.ModelName = c.ModelName
		newCon.DisplayName = newDisplayName
		newCon.ModelID = newModelID
		newCon.DisplayOrder = c.DisplayOrder
		newCon.Shared = c.Shared
		newCon.LocaleCode = toPageLocale
		newCon.LocalizeFromModelID = c.ModelID
		newCon.PageModelName = b.name

		if err = db.Save(&newCon).Error; err != nil {
			return
		}
	}
	return
}

func (b *ModelBuilder) configDuplicate(mb *presets.ModelBuilder) {
	eb := mb.Editing()
	eb.WrapSaveFunc(func(in presets.SaveFunc) presets.SaveFunc {
		return func(obj interface{}, id string, ctx *web.EventContext) (err error) {
			locale, _ := l10n.IsLocalizableFromContext(ctx.R.Context())
			var localeCode string
			if p, ok := obj.(l10n.LocaleInterface); ok {
				if p.EmbedLocale().LocaleCode == "" {
					if err = reflectutils.Set(obj, "LocaleCode", locale); err != nil {
						return
					}
				}
				localeCode = p.EmbedLocale().LocaleCode
			}

			if p, ok := obj.(*Page); ok {
				if p.Slug != "" {
					p.Slug = path.Clean(p.Slug)
				}
				funcName := ctx.R.FormValue(web.EventFuncIDName)
				if funcName == publish.EventDuplicateVersion {
					var fromPage Page
					eb.Fetcher(&fromPage, ctx.Param(presets.ParamID), ctx)
					p.SEO = fromPage.SEO
				}
			}
			if err = in(obj, id, ctx); err != nil {
				return
			}

			var (
				pageID                 int
				version, parentVersion string
			)
			if id != "" {
				ctx.R.Form.Set(presets.ParamID, id)
				pageID, _, _ = b.getPrimaryColumnValuesBySlug(ctx)
			}
			if p, ok := obj.(publish.VersionInterface); ok {
				parentVersion = p.EmbedVersion().ParentVersion
				version = p.EmbedVersion().Version
			}
			err = b.db.Transaction(func(tx *gorm.DB) (inerr error) {
				if strings.Contains(ctx.R.RequestURI, publish.EventDuplicateVersion) {
					if inerr = b.copyContainersToNewPageVersion(tx, pageID, localeCode, parentVersion, version); inerr != nil {
						return
					}
					return
				}

				if v := ctx.R.FormValue(templateSelectedID); v != "" {
					var tplID int
					tplID, inerr = strconv.Atoi(v)
					if inerr != nil {
						return
					}
					if b.builder.l10n == nil {
						localeCode = ""
					}
					if inerr = b.copyContainersToAnotherPage(tx, tplID, templateVersion, localeCode, pageID, version, localeCode); inerr != nil {
						panic(inerr)
					}
				}
				if b.builder.l10n != nil && strings.Contains(ctx.R.RequestURI, l10n.DoLocalize) {
					fromID := ctx.R.Context().Value(l10n.FromID).(string)
					fromVersion := ctx.R.Context().Value(l10n.FromVersion).(string)
					fromLocale := ctx.R.Context().Value(l10n.FromLocale).(string)

					var fromIDInt int
					fromIDInt, err = strconv.Atoi(fromID)
					if err != nil {
						return
					}
					if p, ok := obj.(*Page); ok {
						if inerr = b.builder.localizeCategory(tx, p.CategoryID, fromLocale, locale); inerr != nil {
							panic(inerr)
						}
					}
					if inerr = b.localizeContainersToAnotherPage(tx, fromIDInt, fromVersion, fromLocale, pageID, version, localeCode); inerr != nil {
						panic(inerr)
					}
					return
				}
				return
			})

			return err
		}
	})
}

func (b *ModelBuilder) PreviewHTML(obj interface{}) (r string) {
	p, ok := obj.(PrimarySlugInterface)
	if !ok {
		return
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/?id=%s", p.PrimarySlug()), nil)
	b.preview.ServeHTTP(w, req)
	r = w.Body.String()
	return
}

func (b *ModelBuilder) ContextValueProvider(in context.Context) context.Context {
	return context.WithValue(in, b.name, b)
}

func (b *ModelBuilder) ExistedL10n() bool {
	return b.builder.l10n != nil
}

func (b *ModelBuilder) newContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	containers := b.renderContainersList(ctx)
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: addContainerDialogPortal,
		Body: web.Scope(
			VDialog(
				VSheet(
					VCard(
						VCardTitle(h.Text("New Element")),
						VCardText(containers),
					).Width("40%").Class("pa-4", "overflow-y-auto"),
					VCard(
						VToolbar(
							VSpacer(),
							VBtn("").Icon("mdi-close").Variant(VariantText).Attr("@click", "locals.dialog=false"),
						).Class("v-card--variant-tonal"),
						VCardText(
							web.Portal().Name(addContainerDialogContentPortal),
						).Class("px-6"),
					).Variant(VariantTonal).Width("60%").Class(H100),
				).Class("d-inline-flex"),
			).Width(665).Height(460).Attr("v-model", "locals.dialog"),
		).VSlot(`{locals}`).Init(`{dialog:true}`),
	})
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
	previewContainer, err = b.renderPreviewContainer(ctx, locale, false, true)
	if err != nil {
		return
	}
	addContainerEvent := web.Plaid().EventFunc(AddContainerEvent).
		MergeQuery(true).
		Queries(ctx.R.Form).
		Go()
	iframe := b.rendering(h.Components(previewContainer), ctx, obj, locale, false, true, true)
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: addContainerDialogContentPortal,
		Body: VCard(
			VCardText(
				h.Div(iframe).Style("pointer-events: none;"),
			).Class("pa-0"),
		).Attr("@click", addContainerEvent).
			Elevation(2).Width("100%").Height(326),
	})
	return
}
