package pagebuilder

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	vx "github.com/qor5/ui/v3/vuetifyx"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/v3/activity"
	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"
)

const (
	AddContainerDialogEvent          = "page_builder_AddContainerDialogEvent"
	AddContainerEvent                = "page_builder_AddContainerEvent"
	DeleteContainerConfirmationEvent = "page_builder_DeleteContainerConfirmationEvent"
	DeleteContainerEvent             = "page_builder_DeleteContainerEvent"
	MoveContainerEvent               = "page_builder_MoveContainerEvent"
	MoveUpDownContainerEvent         = "page_builder_MoveUpDownContainerEvent"
	ToggleContainerVisibilityEvent   = "page_builder_ToggleContainerVisibilityEvent"
	MarkAsSharedContainerEvent       = "page_builder_MarkAsSharedContainerEvent"
	RenameContainerDialogEvent       = "page_builder_RenameContainerDialogEvent"
	RenameContainerEvent             = "page_builder_RenameContainerEvent"
	ShowSortedContainerDrawerEvent   = "page_builder_ShowSortedContainerDrawerEvent"
	ReloadRenderPageOrTemplateEvent  = "page_builder_ReloadRenderPageOrTemplateEvent"
	AutoSaveContainerEvent           = "page_builder_AutoSaveContainerEvent"

	paramPageID          = "pageID"
	paramPageVersion     = "pageVersion"
	paramLocale          = "locale"
	paramStatus          = "status"
	paramContainerID     = "containerID"
	paramContainerDataID = "containerDataID"
	paramMoveResult      = "moveResult"
	paramContainerName   = "containerName"
	paramSharedContainer = "sharedContainer"
	paramModelID         = "modelID"
	paramModelName       = "modelName"
	paramMoveDirection   = "moveDirection"
	paramsTpl            = "tpl"
	paramsDevice         = "device"
	paramsDisplayName    = "DisplayName"
	paramTab             = "tab"

	DevicePhone    = "phone"
	DeviceTablet   = "tablet"
	DeviceComputer = "computer"

	EventUp          = "up"
	EventDown        = "down"
	EventDelete      = "delete"
	EventAdd         = "add"
	EventEdit        = "edit"
	iframeHeightName = "_iframeHeight"

	EditorTabElements = "Elements"
	EditorTabLayers   = "Layers"
)

func (b *Builder) Preview(ctx *web.EventContext) (r web.PageResponse, err error) {
	var p *Page
	var (
		pageID  = ctx.Param(presets.ParamID)
		version = ctx.R.FormValue(paramPageVersion)
		local   = ctx.R.FormValue(paramLocale)
	)
	r.Body, p, err = b.renderPageOrTemplate(ctx, pageID, version, local, false)
	if err != nil {
		return
	}
	r.PageTitle = p.Title
	return
}

const editorPreviewContentPortal = "editorPreviewContentPortal"

func (b *Builder) Editor(mb *presets.ModelBuilder) web.PageFunc {
	return func(ctx *web.EventContext) (r web.PageResponse, err error) {
		var (
			deviceToggler, versionComponent      h.HTMLComponent
			editContainerDrawer, navigatorDrawer h.HTMLComponent
			tabContent                           web.PageResponse
			activeDevice                         int
			pageAppbarContent                    []h.HTMLComponent
			page                                 *Page
			exitHref                             string

			device          = ctx.R.FormValue(paramsDevice)
			containerDataID = ctx.R.FormValue(paramContainerDataID)
		)

		switch device {
		case DeviceTablet:
			activeDevice = 1
		case DevicePhone:
			activeDevice = 2
		}

		if containerDataID != "" {
			arr := strings.Split(containerDataID, "_")
			if len(arr) >= 2 {
				editEvent := web.GET().EventFunc(actions.Edit).
					URL(fmt.Sprintf(`%s/%s`, b.prefix, arr[0])).
					Query(presets.ParamID, arr[1]).
					Query(presets.ParamPortalName, pageBuilderRightContentPortal).
					Query(presets.ParamOverlay, actions.Content).Go()
				editContainerDrawer = web.RunScript(fmt.Sprintf(`function(){%s}`, editEvent))
			}

		}
		_ = editContainerDrawer
		deviceToggler = web.Scope(
			VBtnToggle(
				VBtn("").Icon("mdi-laptop").Color(ColorPrimary).Variant(VariantText).Class("mr-4").
					Attr("@click", web.Plaid().URL(ctx.R.URL.Path).EventFunc(ReloadRenderPageOrTemplateEvent).Queries(ctx.R.Form).Query(paramsDevice, DeviceComputer).Go()),
				VBtn("").Icon("mdi-tablet").Color(ColorPrimary).Variant(VariantText).Class("mr-4").
					Attr("@click", web.Plaid().URL(ctx.R.URL.Path).EventFunc(ReloadRenderPageOrTemplateEvent).Queries(ctx.R.Form).Query(paramsDevice, DeviceTablet).Go()),
				VBtn("").Icon("mdi-cellphone").Color(ColorPrimary).Variant(VariantText).Class("mr-4").
					Attr("@click", web.Plaid().URL(ctx.R.URL.Path).EventFunc(ReloadRenderPageOrTemplateEvent).Queries(ctx.R.Form).Query(paramsDevice, DevicePhone).Go()),
			).Class("pa-2 rounded-lg ").Attr("v-model", "toggleLocals.activeDevice").Density(DensityCompact),
		).VSlot("{ locals : toggleLocals}").Init(fmt.Sprintf(`{activeDevice: %d}`, activeDevice))
		if tabContent, page, err = b.PageContent(ctx); err != nil {
			return
		}
		ctx.R.Form.Set(paramStatus, page.GetStatus())
		versionComponent = publish.DefaultVersionComponentFunc(mb, publish.VersionComponentConfig{Top: true})(page, &presets.FieldContext{ModelInfo: mb.Info()}, ctx)
		if b.mb != nil {
			exitHref = b.mb.Info().DetailingHref(ctx.Param(presets.ParamID))
		}
		pageAppbarContent = h.Components(
			h.Div(
				VIcon("mdi-exit-to-app").Class("mr-4").
					Attr("@click", web.Plaid().URL(exitHref).PushState(true).Go()),
				VAppBarTitle().Text("Page Builder"),
			).Class("d-inline-flex align-center"),
			h.Div(deviceToggler).Class("text-center  w-25 d-flex justify-space-between ml-2"),
			versionComponent,
		)
		if navigatorDrawer, err = b.renderNavigator(ctx); err != nil {
			return
		}
		r.Body = h.Components(
			VAppBar(
				h.Div(
					pageAppbarContent...,
				).Class("d-flex align-center  justify-space-between   border-b w-100").Style("height: 48px"),
			).Elevation(0).Density(DensityCompact).Class("px-6"),
			VNavigationDrawer(
				navigatorDrawer,
			).Location(LocationLeft).
				Permanent(true).
				Width(350),
			VNavigationDrawer(
				web.Portal(editContainerDrawer).Name(pageBuilderRightContentPortal),
			).Location(LocationRight).
				Permanent(true).
				Width(350),
			VMain(
				// h.Tag("vx-restore-scroll-listener"),
				vx.VXMessageListener().ListenFunc(b.generateEditorBarJsFunction(ctx)),
				tabContent.Body.(h.HTMLComponent),
			),
		)
		return
	}
}

func (b *Builder) getDevice(ctx *web.EventContext) (device string, style string) {
	device = ctx.R.FormValue(paramsDevice)
	if len(device) == 0 {
		device = b.defaultDevice
	}

	switch device {
	case DevicePhone:
		style = "414px"
	case DeviceTablet:
		style = "768px"
		// case Device_Computer:
		//	style = "width: 1264px;"
	}

	return
}

const ContainerToPageLayoutKey = "ContainerToPageLayout"

func (b *Builder) renderPageOrTemplate(ctx *web.EventContext, pageOrTemplateID, version, locale string, isEditor bool) (r h.HTMLComponent, p *Page, err error) {
	isTpl := ctx.R.FormValue(paramsTpl) != ""

	if isTpl {
		tpl := &Template{}
		err = b.db.First(tpl, "id = ? and locale_code = ?", pageOrTemplateID, locale).Error
		if err != nil {
			return
		}
		p = tpl.Page()
		version = p.Version.Version
	} else {
		err = b.db.First(&p, "id = ? and version = ? and locale_code = ?", pageOrTemplateID, version, locale).Error
		if err != nil {
			return
		}
	}

	var isReadonly bool
	if p.GetStatus() != publish.StatusDraft && isEditor {
		isReadonly = true
	}

	var comps []h.HTMLComponent
	comps, err = b.renderContainers(ctx, p, isEditor, isReadonly)
	if err != nil {
		return
	}
	r = h.Components(comps...)
	if b.pageLayoutFunc != nil {
		var seoTags h.HTMLComponent
		if b.seoBuilder != nil {
			seoTags = b.seoBuilder.Render(p, ctx.R)
		}
		input := &PageLayoutInput{
			IsEditor:  isEditor,
			IsPreview: !isEditor,
			Page:      p,
			SeoTags:   seoTags,
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
			  transition: opacity .5s ease-in-out;
			  text-align: center;
			}
			
			.editor-add div {
			  width: 100%;
			  background-color: #3E63DD;
			  transition: height .5s ease-in-out;
			  height: 3px;
			}
			
			.editor-add button {
			  width: 32px;
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
			  width: 30%;
			  height: 32px;	
			  opacity: 0;
              display: flex;
			  align-items: center;	
			  background-color: #3E63DD;
			  justify-content: space-between;
			}
   			.editor-bar-buttons{
              height: 24px;
			}
			.editor-bar button {
			  color: #FFFFFF;
			  background-color: #3E63DD; 
              height: 24px;	
			}
			
			.editor-bar h6 {
			  color: #FFFFFF;
			  margin-left: 4px;	
			}
			
			.highlight .editor-bar {
			  opacity: 1;
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

		if isEditor {
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
			r = b.pageLayoutFunc(h.Components(comps...), input, newCtx)
			newCtx.Injector.HeadHTMLComponent("style", b.pageStyle, true)
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
			_, width := b.getDevice(ctx)
			r = h.Tag("vx-scroll-iframe").Attr(
				":srcdoc", h.JSONString(h.MustString(r, ctx.R.Context())),
				"iframe-height", iframeValue,
				"iframe-height-name", iframeHeightName,
				"width", width,
				"ref", "scrollIframe")

		} else {
			r = b.pageLayoutFunc(h.Components(comps...), input, ctx)
			ctx.Injector.HeadHTMLComponent("style", b.pageStyle, true)
		}
	}

	return
}

func (b *Builder) renderContainers(ctx *web.EventContext, p *Page, isEditor bool, isReadonly bool) (r []h.HTMLComponent, err error) {
	var cons []*Container
	err = b.db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", p.ID, p.GetVersion(), p.GetLocale()).Error
	if err != nil {
		return
	}
	device, _ := b.getDevice(ctx)
	cbs := b.getContainerBuilders(cons)
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
			Page:        p,
			IsEditor:    isEditor,
			IsReadonly:  isReadonly,
			Device:      device,
			ContainerId: ec.container.PrimarySlug(),
			DisplayName: displayName,
		}
		pure := ec.builder.renderFunc(obj, &input, ctx)

		r = append(r, b.containerWrapper(pure.(*h.HTMLTagBuilder), ctx, isEditor, isReadonly, i == 0, i == len(cbs)-1,
			ec.builder.getContainerDataID(int(ec.container.ModelID)), ec.container.ModelName, &input))
	}

	return
}

type ContainerSorterItem struct {
	Index          int    `json:"index"`
	Label          string `json:"label"`
	ModelName      string `json:"model_name"`
	ModelID        string `json:"model_id"`
	DisplayName    string `json:"display_name"`
	ContainerID    string `json:"container_id"`
	URL            string `json:"url"`
	Shared         bool   `json:"shared"`
	Hidden         bool   `json:"hidden"`
	VisibilityIcon string `json:"visibility_icon"`
	ParamID        string `json:"param_id"`
	Locale         string `json:"locale"`
}

type ContainerSorter struct {
	Items []ContainerSorterItem `json:"items"`
}

func (b *Builder) renderNavigator(ctx *web.EventContext) (r h.HTMLComponent, err error) {
	tab := ctx.Param(paramTab)
	if tab == "" {
		tab = EditorTabElements
	}

	var listContainers h.HTMLComponent
	if tab == EditorTabLayers {
		if listContainers, err = b.renderContainersSortedList(ctx); err != nil {
			return
		}
	}

	r = h.Components(
		VTabs(
			VTab().Text("Elements").
				Value(EditorTabElements).Attr("@click",
				web.Plaid().PushState(true).MergeQuery(true).
					Query(paramTab, EditorTabElements).RunPushState()),
			VTab().Text("Layers").Value(EditorTabLayers).Attr("@click",
				web.Plaid().PushState(true).MergeQuery(true).
					Query(paramTab, EditorTabLayers).RunPushState()+
					";"+
					web.Plaid().
						URL(ctx.R.URL.Path).
						EventFunc(ShowSortedContainerDrawerEvent).
						Query(paramStatus, ctx.Param(paramStatus)).
						MergeQuery(true).
						Go()),
		).Attr("v-model", "vars.containerTab").FixedTabs(true),
		VTabsWindow(
			VTabsWindowItem(b.renderContainersList(ctx)).Value(EditorTabElements),
			VTabsWindowItem(web.Portal(listContainers).Name(pageBuilderLayerContainerPortal)).Value(EditorTabLayers),
		).Attr("v-model", "vars.containerTab").Attr(web.VAssign("vars", fmt.Sprintf(`{containerTab:"%s"}`, tab))...),
	)
	return
}

func (b *Builder) renderEditContainer(ctx *web.EventContext) (r h.HTMLComponent, err error) {
	var (
		modelName     = ctx.R.FormValue(paramModelName)
		containerName = ctx.R.FormValue(paramContainerName)
		modelID       = ctx.R.FormValue(paramModelID)
	)
	builder := b.ContainerByName(modelName).GetModelBuilder()
	element := builder.NewModel()
	if err = b.db.First(element, modelID).Error; err != nil {
		return
	}
	r = web.Scope(
		VLayout(
			VMain(
				h.Div(
					h.Span(containerName).Class("text-subtitle-1"),
					h.Div(
						VBtn("Save").Variant(VariantFlat).Color(ColorSecondary).Size(SizeSmall).Attr("@click", web.Plaid().
							EventFunc(actions.Update).
							URL(b.ContainerByName(modelName).mb.Info().ListingHref()).
							Query(presets.ParamID, modelID).
							Go()),
					),
				).Class("d-flex  pa-6 align-center justify-space-between"),
				VDivider(),
				h.Div(

					builder.Editing().ToComponent(builder.Info(), element, ctx),
				).Class("pa-6"),
			),
		),
	).VSlot("{ form }")
	return
}

func (b *Builder) renderContainersSortedList(ctx *web.EventContext) (r h.HTMLComponent, err error) {
	var (
		cons         []*Container
		p            = new(Page)
		primarySlug  = p.PrimaryColumnValuesBySlug(ctx.Param(presets.ParamID))
		pageID       = primarySlug["id"]
		pageVersion  = primarySlug[publish.SlugVersion]
		locale       = primarySlug[l10n.SlugLocaleCode]
		status       = ctx.R.FormValue(paramStatus)
		isReadonly   = status != publish.StatusDraft
		msgr         = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		activityMsgr = i18n.MustGetModuleMessages(ctx.R, activity.I18nActivityKey, activity.Messages_en_US).(*activity.Messages)
	)

	err = b.db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", pageID, pageVersion, locale).Error
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
				Index:          i,
				Label:          inflection.Plural(strcase.ToKebab(c.ModelName)),
				ModelName:      c.ModelName,
				ModelID:        strconv.Itoa(int(c.ModelID)),
				DisplayName:    displayName,
				ContainerID:    strconv.Itoa(int(c.ID)),
				URL:            b.ContainerByName(c.ModelName).mb.Info().ListingHref(),
				Shared:         c.Shared,
				VisibilityIcon: vicon,
				ParamID:        c.PrimarySlug(),
				Locale:         locale,
				Hidden:         c.Hidden,
			},
		)
	}
	menu := VMenu(
		web.Slot(
			VBtn("").Icon("mdi-dots-horizontal").Variant(VariantText).Size(SizeSmall).Attr("v-bind", "props").Attr("v-show", "element.editShow || (isActive || isHovering)"),
		).Name("activator").Scope("{isActive,props}"),
		VList(
			VListItem(
				VBtn(msgr.Rename).PrependIcon("mdi-pencil").Attr("@click",
					"element.editShow=!element.editShow",
				),
			),
			VListItem(
				VBtn(activityMsgr.ActionDelete).PrependIcon("mdi-delete").Attr("@click",
					web.Plaid().
						URL(ctx.R.URL.Path).
						EventFunc(DeleteContainerConfirmationEvent).
						Query(paramContainerID, web.Var("element.param_id")).
						Query(paramContainerName, web.Var("element.display_name")).
						Go(),
				),
			),
		),
	)

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
															URL(fmt.Sprintf("%s/editors", b.prefix)).
															EventFunc(RenameContainerEvent).Query(paramStatus, status).Query(paramContainerID, web.Var("element.param_id")).Go()),
													VListItemTitle(h.Text("{{element.display_name}}")).Attr(":style", "[element.shared ? {'color':'green'}:{}]").Attr("v-if", "!element.editShow"),
												).VSlot("{form}").FormInit("{ DisplayName:element.display_name }"),
											),
										),
										web.Slot(
											h.If(!isReadonly,
												h.Div(
													VBtn("").Variant(VariantText).Attr(":icon", "element.visibility_icon").Size(SizeSmall).Attr("@click",
														web.Plaid().
															EventFunc(ToggleContainerVisibilityEvent).
															Query(paramContainerID, web.Var("element.param_id")).
															Query(paramStatus, status).
															Go(),
													).Attr("v-show", "element.editShow || (element.hidden || isHovering)"),

													VBtn("").Variant(VariantText).Icon("mdi-cog").Size(SizeSmall).Attr("@click",
														web.Plaid().
															URL(web.Var(fmt.Sprintf(`"%s/"+element.label`, b.prefix))).
															EventFunc(actions.Edit).
															Query(presets.ParamOverlay, actions.Content).
															Query(presets.ParamPortalName, pageBuilderRightContentPortal).
															Query(presets.ParamID, web.Var("element.model_id")).
															Go(),
													).Attr("v-show", "element.editShow || isHovering"),
													menu,
												),
											),
										).Name("append"),
									).Attr(":variant", fmt.Sprintf(` element.hidden &&!isHovering && !element.editShow?"%s":"%s"`, VariantPlain, VariantText)).
										Attr("v-bind", "props").
										Attr("@click", web.Plaid().PushState(true).MergeQuery(true).
											Query(paramContainerID, web.Var("element.param_id")).
											Query(paramContainerDataID, web.Var(`element.label+"_"+element.model_id`)).
											RunPushState()+
											";"+scrollToContainer(fmt.Sprintf(`%s+"_"+%s`, web.Var("element.label"), web.Var("element.model_id")))),
								).Name("default").Scope("{ isHovering, props }"),
							),
							VDivider(),
						),
					).Attr("#item", " { element } "),
				),
			),
		).Class("pa-4 pt-2"),
	).Init(h.JSONString(sorterData)).VSlot("{ locals:sortLocals,form }")
	return
}

func primaryKeys(ctx *web.EventContext) (pageID int, pageVersion string, locale string) {
	p := new(Page)
	primarySlug := p.PrimaryColumnValuesBySlug(ctx.Param(presets.ParamID))
	pageVersion = primarySlug[publish.SlugVersion]
	locale = primarySlug[l10n.SlugLocaleCode]
	pageIDi, _ := strconv.ParseInt(primarySlug["id"], 10, 64)
	pageID = int(pageIDi)
	return
}

func (b *Builder) addContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		modelName       = ctx.Param(paramModelName)
		sharedContainer = ctx.Param(paramSharedContainer)
		modelID         = ctx.ParamAsInt(paramModelID)
		containerID     = ctx.Param(paramContainerID)
		newContainerID  string
	)

	pageID, pageVersion, locale := primaryKeys(ctx)

	if sharedContainer == "true" {
		newContainerID, err = b.AddSharedContainerToPage(pageID, containerID, pageVersion, locale, modelName, uint(modelID))
	} else {
		var newModelId uint
		newModelId, newContainerID, err = b.AddContainerToPage(pageID, containerID, pageVersion, locale, modelName)
		modelID = int(newModelId)
	}
	cb := b.ContainerByName(modelName)
	r.RunScript = web.Plaid().PushState(true).MergeQuery(true).
		Query(paramContainerDataID, cb.getContainerDataID(modelID)).
		Query(paramContainerID, newContainerID).RunPushState() +
		";" + web.Plaid().
		EventFunc(ReloadRenderPageOrTemplateEvent).
		Query(paramContainerDataID, cb.getContainerDataID(modelID)).
		Query(paramContainerID, newContainerID).
		Go() + ";" +
		web.Plaid().
			URL(fmt.Sprintf(`%s/%s`, b.prefix, inflection.Plural(strcase.ToKebab(cb.name)))).
			EventFunc(actions.Edit).
			Query(presets.ParamPortalName, pageBuilderRightContentPortal).
			Query(presets.ParamOverlay, actions.Content).
			Query(presets.ParamID, modelID).
			Go()

	return
}

func (b *Builder) moveContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) moveUpDownContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) toggleContainerVisibility(ctx *web.EventContext) (r web.EventResponse, err error) {
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
			Query(paramStatus, ctx.Param(paramStatus)).
			Go()
	return
}

func (b *Builder) deleteContainerConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) deleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) AddContainerToPage(pageID int, containerID, pageVersion, locale, modelName string) (modelID uint, newContainerID string, err error) {
	model := b.ContainerByName(modelName).NewModel()
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
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ? and locale_code = ?", pageID, pageVersion, locale).Scan(&maxOrder).Error
	if err != nil {
		return
	}
	if containerID != "" {
		var lastContainer Container
		cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
		if dbErr := b.db.Where("id = ? AND locale_code = ?", cs["id"], locale).First(&lastContainer).Error; dbErr == nil {
			displayOrder = lastContainer.DisplayOrder
			if err = b.db.Model(&Container{}).Where("page_id = ? and page_version = ? and locale_code = ? and display_order > ? ", pageID, pageVersion, locale, displayOrder).
				UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; err != nil {
				return
			}
		}

	} else {
		displayOrder = maxOrder.Float64
	}
	modelID = reflectutils.MustGet(model, "ID").(uint)
	container := Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		ModelName:    modelName,
		DisplayName:  modelName,
		ModelID:      modelID,
		DisplayOrder: displayOrder + 1,
		Locale: l10n.Locale{
			LocaleCode: locale,
		},
	}
	err = b.db.Create(&container).Error
	newContainerID = container.PrimarySlug()
	return
}

func (b *Builder) AddSharedContainerToPage(pageID int, containerID, pageVersion, locale, modelName string, modelID uint) (newContainerID string, err error) {
	var c Container
	err = b.db.First(&c, "model_name = ? AND model_id = ? AND shared = true", modelName, modelID).Error
	if err != nil {
		return
	}
	var (
		maxOrder     sql.NullFloat64
		displayOrder float64
	)
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ? and locale_code = ?", pageID, pageVersion, locale).Scan(&maxOrder).Error
	if err != nil {
		return
	}
	if containerID != "" {
		var lastContainer Container
		cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
		if dbErr := b.db.Where("id = ? AND locale_code = ?", cs["id"], locale).First(&lastContainer).Error; dbErr == nil {
			displayOrder = lastContainer.DisplayOrder
			if err = b.db.Model(&Container{}).Where("page_id = ? and page_version = ? and locale_code = ? and display_order > ? ", pageID, pageVersion, locale, displayOrder).
				UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; err != nil {
				return
			}
		}

	} else {
		displayOrder = maxOrder.Float64
	}
	container := Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		ModelName:    modelName,
		DisplayName:  c.DisplayName,
		ModelID:      modelID,
		Shared:       true,
		DisplayOrder: displayOrder + 1,
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

func (b *Builder) copyContainersToNewPageVersion(db *gorm.DB, pageID int, locale, oldPageVersion, newPageVersion string) (err error) {
	return b.copyContainersToAnotherPage(db, pageID, oldPageVersion, locale, pageID, newPageVersion, locale)
}

func (b *Builder) copyContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion, locale string, toPageID int, toPageVersion, toPageLocale string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", pageID, pageVersion, locale).Error
	if err != nil {
		return
	}

	for _, c := range cons {
		newModelID := c.ModelID
		if !c.Shared {
			model := b.ContainerByName(c.ModelName).NewModel()
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
			PageID:       uint(toPageID),
			PageVersion:  toPageVersion,
			ModelName:    c.ModelName,
			DisplayName:  c.DisplayName,
			ModelID:      newModelID,
			DisplayOrder: c.DisplayOrder,
			Shared:       c.Shared,
			Locale: l10n.Locale{
				LocaleCode: toPageLocale,
			},
		}).Error; err != nil {
			return
		}
	}
	return
}

func (b *Builder) localizeContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion, locale string, toPageID int, toPageVersion, toPageLocale string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", pageID, pageVersion, locale).Error
	if err != nil {
		return
	}

	for _, c := range cons {
		newModelID := c.ModelID
		newDisplayName := c.DisplayName
		if !c.Shared {
			model := b.ContainerByName(c.ModelName).NewModel()
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
			if err = db.Where("model_name = ? AND localize_from_model_id = ? AND locale_code = ? AND shared = ?", c.ModelName, c.ModelID, toPageLocale, true).First(&sharedCon).Count(&count).Error; err != nil && err != gorm.ErrRecordNotFound {
				return
			}

			if count == 0 {
				model := b.ContainerByName(c.ModelName).NewModel()
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

		if err = db.Save(&newCon).Error; err != nil {
			return
		}
	}
	return
}

func (b *Builder) localizeCategory(db *gorm.DB, fromCategoryID uint, fromLocale string, toLocale string) (err error) {
	if fromCategoryID == 0 {
		return
	}
	var category Category
	var toCategory Category
	err = db.First(&category, "id = ? AND locale_code = ?", fromCategoryID, fromLocale).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	if err != nil {
		return
	}
	err = db.First(&toCategory, "id = ? AND locale_code = ?", fromCategoryID, toLocale).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		category.LocaleCode = toLocale
		err = db.Save(&category).Error
		return
	}
	return
}

func (b *Builder) createModelAfterLocalizeDemoContainer(db *gorm.DB, c *DemoContainer) (err error) {
	model := b.ContainerByName(c.ModelName).NewModel()
	if err = db.First(model, "id = ?", c.ModelID).Error; err != nil {
		return
	}
	if err = reflectutils.Set(model, "ID", uint(0)); err != nil {
		return
	}
	if err = db.Create(model).Error; err != nil {
		return
	}

	c.ModelID = reflectutils.MustGet(model, "ID").(uint)
	return
}

func (b *Builder) markAsSharedContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) renameContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) renameContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	paramID := ctx.R.FormValue(paramContainerID)
	name := ctx.R.FormValue(paramContainerName)
	okAction := web.Plaid().
		URL(fmt.Sprintf("%s/editors", b.prefix)).
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

func (b *Builder) renderContainersList(ctx *web.EventContext) (component h.HTMLComponent) {
	var (
		p           = new(Page)
		primarySlug = p.PrimaryColumnValuesBySlug(ctx.Param(presets.ParamID))
		locale      = primarySlug["locale_code"]
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	)

	var (
		containers  []h.HTMLComponent
		groupsNames []string
	)
	sort.Slice(b.containerBuilders, func(i, j int) bool {
		return b.containerBuilders[i].group != "" && b.containerBuilders[j].group == ""
	})
	groupContainers := utils.GroupBySlice[*ContainerBuilder, string](b.containerBuilders, func(builder *ContainerBuilder) string {
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
		if b.expendContainers {
			groupsNames = append(groupsNames, groupName)
		}
		var listItems []h.HTMLComponent
		for _, builder := range group {
			cover := builder.cover
			if cover == "" {
				cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
			}
			containerName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, builder.name)
			listItems = append(listItems, VListItem(
				VListItemTitle(h.Text(containerName)),
				VListItemSubtitle(VImg().Src(cover).Height(100)),
			).Attr("@click",
				web.Plaid().EventFunc(AddContainerEvent).
					MergeQuery(true).
					Query(paramModelName, builder.name).
					Query(paramContainerName, builder.name).
					Go(),
			))
		}
		containers = append(containers, VListGroup(
			web.Slot(
				VListItem(
					VListItemTitle(h.Text(groupName)),
				).Attr("v-bind", "props").Class("bg-light-blue-lighten-5"),
			).Name("activator").Scope(" {  props }"),
			h.Components(listItems...),
		).Value(groupName))
	}

	var cons []*Container

	b.db.Select("display_name,model_name,model_id").Where("shared = true AND locale_code = ?", locale).Group("display_name,model_name,model_id").Find(&cons)
	sort.Slice(cons, func(i, j int) bool {
		return b.ContainerByName(cons[i].ModelName).group != "" && b.ContainerByName(cons[j].ModelName).group == ""
	})
	for _, group := range utils.GroupBySlice[*Container, string](cons, func(builder *Container) string {
		return b.ContainerByName(builder.ModelName).group
	}) {
		if len(group) == 0 {
			break
		}
		groupName := msgr.Shared

		if b.expendContainers {
			groupsNames = append(groupsNames, groupName)
		}
		var listItems []h.HTMLComponent
		for _, builder := range group {
			c := b.ContainerByName(builder.ModelName)
			cover := c.cover
			if cover == "" {
				cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(c.name, " ", "")+".png")
			}
			containerName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, c.name)
			listItems = append(listItems, VListItem(
				h.Div(
					VListItemTitle(h.Text(containerName)),
					VListItemSubtitle(VImg().Src(cover).Height(100)),
				).Attr("@click", web.Plaid().
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
				).Attr("v-bind", "props").Class("bg-light-blue-lighten-5"),
			).Name("activator").Scope(" {  props }"),
			h.Components(listItems...),
		).Value(groupName))
	}
	component = VList(containers...).Opened(groupsNames)
	return
}

func (b *Builder) addContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var containers []h.HTMLComponent
	var (
		pageVersion = ctx.R.FormValue(paramPageVersion)
		locale      = ctx.R.FormValue(paramLocale)
		msgr        = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	)

	for _, builder := range b.containerBuilders {
		cover := builder.cover
		if cover == "" {
			cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
		}
		containers = append(containers,
			VCol(
				VCard(
					VImg().Src(cover).Height(200),
					VCardActions(
						VCardTitle(h.Text(i18n.T(ctx.R, presets.ModelsI18nModuleKey, builder.name))),
						VSpacer(),
						VBtn(msgr.Select).
							Variant(VariantText).
							Color("primary").Attr("@click",
							"dialogLocals.addContainerDialog = false;"+web.Plaid().
								URL(ctx.R.URL.Path).
								EventFunc(AddContainerEvent).
								Query(paramPageVersion, pageVersion).
								Query(paramLocale, locale).
								Query(paramContainerName, builder.name).
								Go(),
						),
					),
				),
			).Cols(4),
		)
	}

	var cons []*Container
	err = b.db.Select("display_name,model_name,model_id").Where("shared = true AND locale_code = ?", locale).Group("display_name,model_name,model_id").Find(&cons).Error
	if err != nil {
		return
	}

	var sharedContainers []h.HTMLComponent
	for _, sharedC := range cons {
		c := b.ContainerByName(sharedC.ModelName)
		cover := c.cover
		if cover == "" {
			cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(c.name, " ", "")+".png")
		}
		sharedContainers = append(sharedContainers,
			VCol(
				VCard(
					VImg().Src(cover).Height(200),
					VCardActions(
						VCardTitle(h.Text(i18n.T(ctx.R, presets.ModelsI18nModuleKey, sharedC.DisplayName))),
						VSpacer(),
						VBtn(msgr.Select).
							Variant(VariantText).
							Color("primary").Attr("@click",
							"dialogLocals.addContainerDialog = false;"+web.Plaid().
								URL(ctx.R.URL.Path).
								EventFunc(AddContainerEvent).
								Query(paramPageVersion, pageVersion).
								Query(paramLocale, locale).
								Query(paramContainerName, sharedC.ModelName).
								Query(paramModelID, sharedC.ModelID).
								Query(paramSharedContainer, "true").
								Go(),
						),
					),
				),
			).Cols(4),
		)
	}

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogPortalName,
		Body: web.Scope(
			VDialog(
				VTabs(
					VTab(h.Text(msgr.New)).Value(msgr.New),
					VTab(h.Text(msgr.Shared)).Value(msgr.Shared),
				).Attr("v-model", "dialogLocals.tab"),
				VWindow(
					VWindowItem(
						VSheet(
							VContainer(
								VRow(
									containers...,
								),
							),
						),
					).Value(msgr.New).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
					VWindowItem(
						VSheet(
							VContainer(
								VRow(
									sharedContainers...,
								),
							),
						),
					).Value(msgr.Shared).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
				).Attr("v-model", "dialogLocals.tab"),
			).Width("1200px").Attr("v-model", "dialogLocals.addContainerDialog"),
		).Init(fmt.Sprintf(`{addContainerDialog:true , tab : %s } `, msgr.New)).VSlot("{locals:dialogLocals}"),
	})

	return
}

type editorContainer struct {
	builder   *ContainerBuilder
	container *Container
}

func (b *Builder) getContainerBuilders(cs []*Container) (r []*editorContainer) {
	for _, c := range cs {
		for _, cb := range b.containerBuilders {
			if cb.name == c.ModelName {
				r = append(r, &editorContainer{
					builder:   cb,
					container: c,
				})
			}
		}
	}
	return
}

const (
	dialogPortalName = "pagebuilder_DialogPortalName"
)

func (b *Builder) pageEditorLayout(in web.PageFunc, config *presets.LayoutConfig) (out web.PageFunc) {
	return func(ctx *web.EventContext) (pr web.PageResponse, err error) {
		b.ps.InjectAssets(ctx)
		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}
		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, "Page Builder")
		pr.Body = VLayout(

			web.Portal().Name(presets.RightDrawerPortalName),
			web.Portal().Name(presets.DialogPortalName),
			web.Portal().Name(presets.DeleteConfirmPortalName),
			web.Portal().Name(presets.ListingDialogPortalName),
			web.Portal().Name(dialogPortalName),
			innerPr.Body.(h.HTMLComponent),
		).Attr("id", "vt-app").
			Attr(web.VAssign("vars", `{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false}`)...)
		return
	}
}

func (b *Builder) showSortedContainerDrawer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var body h.HTMLComponent
	if body, err = b.renderContainersSortedList(ctx); err != nil {
		return
	}
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: pageBuilderLayerContainerPortal, Body: body})
	return
}

func (b *Builder) reloadRenderPageOrTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	var body h.HTMLComponent
	p := new(Page)
	var (
		primarySlug     = p.PrimaryColumnValuesBySlug(ctx.Param(presets.ParamID))
		pageID          = primarySlug["id"]
		version         = primarySlug["version"]
		localeCode      = primarySlug["locale_code"]
		containerDataID = ctx.Param(paramContainerDataID)
	)
	if containerDataID != "" {
		r.RunScript = fmt.Sprintf(`setTimeout(function(){%s},100)`, scrollToContainer(fmt.Sprintf(`"%s"`, containerDataID)))
	}
	body, _, err = b.renderPageOrTemplate(ctx, pageID, version, localeCode, true)
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: editorPreviewContentPortal, Body: body.(*h.HTMLTagBuilder).Attr(web.VAssign("vars", "{el:$}")...)})
	return
}

func scrollToContainer(containerDataID interface{}) string {
	return fmt.Sprintf(`vars.el.refs.scrollIframe.scrollToCurrentContainer(%v);`, containerDataID)
}

func (b *Builder) containerWrapper(r *h.HTMLTagBuilder, ctx *web.EventContext, isEditor, isReadonly, isFirst, isEnd bool, containerDataID, modelName string, input *RenderInput) h.HTMLComponent {
	r.Attr("data-container-id", containerDataID)
	pmb := postMessageBody{
		ContainerDataID: containerDataID,
		ContainerId:     input.ContainerId,
		DisplayName:     input.DisplayName,
		ModelName:       modelName,
		isFirst:         isFirst,
		isEnd:           isEnd,
	}
	if isEditor {
		if isReadonly {
			r.AppendChildren(h.Div().Class("wrapper-shadow"))
		} else {
			r.AppendChildren(h.Div().Class("inner-shadow"))
			r = h.Div(
				r.Attr("onclick", "event.stopPropagation();document.querySelectorAll('.highlight').forEach(item=>{item.classList.remove('highlight')});this.parentElement.classList.add('highlight');"+pmb.postMessage(EventEdit)),
				h.Div(
					h.H6(input.DisplayName).Class("title"),
					h.Div(
						h.Button("").Children(h.I("arrow_upward").Class("material-icons")).Attr("onclick", pmb.postMessage(EventUp)),
						h.Button("").Children(h.I("arrow_downward").Class("material-icons")).Attr("onclick", pmb.postMessage(EventDown)),
						h.Button("").Children(h.I("delete").Class("material-icons")).Attr("onclick", pmb.postMessage(EventDelete)),
					).Class("editor-bar-buttons"),
				).Class("editor-bar"),
				h.Div(
					h.Div().Class("add"),
					h.Button("").Children(h.I("add").Class("material-icons")).Attr("onclick", pmb.postMessage(EventAdd)),
				).Class("editor-add"),
			).Class("wrapper-shadow").ClassIf("highlight", ctx.Param(paramContainerDataID) == containerDataID)
		}
	}
	return r
}

type (
	postMessageBody struct {
		MsgType         string `json:"msg_type"`
		ContainerDataID string `json:"container_data_id"`
		ContainerId     string `json:"container_id"`
		DisplayName     string `json:"display_name"`
		ModelName       string `json:"model_name"`
		isFirst         bool
		isEnd           bool
	}
)

func (b *postMessageBody) postMessage(msgType string) string {
	if msgType == EventUp && b.isFirst {
		return ""
	}
	if msgType == EventDown && b.isEnd {
		return ""
	}
	b.MsgType = msgType
	return fmt.Sprintf(`window.parent.postMessage(%s, '*')`, h.JSONString(b))
}
