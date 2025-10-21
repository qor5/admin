package pagebuilder

import (
	"errors"
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/qor5/x/v3/perm"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
)

const (
	AddContainerEvent                   = "page_builder_AddContainerEvent"
	DeleteContainerConfirmationEvent    = "page_builder_DeleteContainerConfirmationEvent"
	DeleteContainerEvent                = "page_builder_DeleteContainerEvent"
	MoveContainerEvent                  = "page_builder_MoveContainerEvent"
	MoveUpDownContainerEvent            = "page_builder_MoveUpDownContainerEvent"
	ToggleContainerVisibilityEvent      = "page_builder_ToggleContainerVisibilityEvent"
	MarkAsSharedContainerEvent          = "page_builder_MarkAsSharedContainerEvent"
	RenameContainerDialogEvent          = "page_builder_RenameContainerDialogEvent"
	RenameContainerEvent                = "page_builder_RenameContainerEvent"
	RenameContainerFromDialogEvent      = "page_builder_RenameContainerFromDialogEvent"
	ShowSortedContainerDrawerEvent      = "page_builder_ShowSortedContainerDrawerEvent"
	ReloadRenderPageOrTemplateEvent     = "page_builder_ReloadRenderPageOrTemplateEvent"
	ReloadRenderPageOrTemplateBodyEvent = "page_builder_ReloadRenderPageOrTemplateBodyEvent"
	ContainerPreviewEvent               = "page_builder_ContainerPreviewEvent"
	ReplicateContainerEvent             = "page_builder_ReplicateContainerEvent"
	EditContainerEvent                  = "page_builder_EditContainerEvent"
	UpdateContainerEvent                = "page_builder_UpdateContainerEvent"
	ReloadAddContainersListEvent        = "page_builder_ReloadAddContainersEvent"

	ParamContainerCreate = "paramContainerCreate"

	paramPageID          = "pageID"
	paramPageVersion     = "pageVersion"
	paramLocale          = "locale"
	paramStatus          = "status"
	paramContainerID     = "containerID"
	paramContainerUri    = "containerUri"
	paramContainerDataID = "containerDataID"
	paramIsUpdate        = "isUpdate"
	paramMoveResult      = "moveResult"
	paramContainerName   = "containerName"
	paramSharedContainer = "sharedContainer"
	paramModelID         = "modelID"
	paramModelName       = "modelName"
	paramMoveDirection   = "moveDirection"
	paramDevice          = "device"
	paramDisplayName     = "DisplayName"
	paramDemoContainer   = "demoContainer"

	DevicePhone    = "phone"
	DeviceTablet   = "tablet"
	DeviceComputer = "computer"

	EventUp                        = "up"
	EventDown                      = "down"
	EventDelete                    = "delete"
	EventAdd                       = "add"
	EventEdit                      = "edit"
	EventClickOutsideWrapperShadow = "clickOutsideWrapperShadow"

	paramIframeEventName  = "iframeEventName"
	changeDeviceEventName = "change_device"
	updateBodyEventName   = "update_body"

	ActionAddContainer    = "AddContainer"
	ActionDeleteContainer = "DeleteContainer"
)

const (
	editorPreviewContentPortal      = "editorPreviewContentPortal"
	addContainerDialogPortal        = "addContainerDialogPortal"
	addContainerDialogContentPortal = "addContainerDialogContentPortal"
)

func (b *Builder) emptyEdit(ctx *web.EventContext) h.HTMLComponent {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	return VLayout(
		VAppBar(
			VToolbarTitle("").Children(h.Text(msgr.Settings)),
		).Elevation(0),
		VSpacer(),
		VMain(
			VSheet(
				VCard(
					VCardText(
						h.Text(msgr.SelectElementMsg),
					),
				).Variant(VariantFlat),
			).Class("pa-2")),
	)
}

func (b *Builder) editorBody(ctx *web.EventContext, m *ModelBuilder) (body h.HTMLComponent) {
	var (
		deviceToggle, versionComponent       h.HTMLComponent
		editContainerDrawer, navigatorDrawer h.HTMLComponent
		tabContent                           web.PageResponse
		pageAppbarContent                    []h.HTMLComponent
		exitHref                             string
		isStag                               bool

		containerDataID = ctx.Param(paramContainerDataID)
		obj             interface{}
		msgr            = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		title           string
		err             error
		ps              = ctx.Param(presets.ParamID)
	)
	if m.mb.Info().Verifier().Do(presets.PermGet).WithReq(ctx.R).IsAllowed() != nil {
		return h.Text(perm.PermissionDenied.Error())
	}
	if obj, err = m.pageBuilderModel(ctx); err != nil {
		return
	}

	if !m.isTemplate {
		title = msgr.PageBuilder
		exitHref = m.mb.Info().DetailingHref(ctx.Param(presets.ParamID))
	} else {
		title = msgr.PageTemplate
		exitHref = m.mb.Info().ListingHref()

	}
	if containerDataID != "" {
		editEvent := web.Plaid().
			EventFunc(EditContainerEvent).
			MergeQuery(true).
			Query(paramContainerDataID, containerDataID).
			Query(presets.ParamPortalName, pageBuilderRightContentPortal).
			Query(presets.ParamOverlay, actions.Content).Go()
		editContainerDrawer = web.RunScript("function(){" + editEvent + "}")
	} else {
		editContainerDrawer = b.emptyEdit(ctx)
	}
	deviceToggle = b.deviceToggle(ctx)
	if tabContent, err = m.pageContent(ctx); err != nil {
		return
	}
	if p, ok := obj.(publish.StatusInterface); ok {
		ctx.R.Form.Set(paramStatus, p.EmbedStatus().Status)
		isStag = p.EmbedStatus().Status == publish.StatusOnline || p.EmbedStatus().Status == publish.StatusOffline
	}
	if !isStag && m.mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
		isStag = true
	}
	afterLeaveEvent := removeVirtualElement()
	addOverlay := web.Scope(
		h.Div().Style("display:none").Attr("v-on-mounted", `({watch})=>{
                    watch(() => vars.overlay, (value) => {
                        if(value){xLocals.add=false;`+web.Plaid().EventFunc(ReloadAddContainersListEvent).Query(paramStatus, ctx.Param(paramStatus)).Go()+`}
                        })
                }`),
		vx.VXOverlay(
			m.newContainerContent(ctx),
		).
			MaxWidth(665).
			Attr("ref", "overlay").
			Attr("@after-leave", fmt.Sprintf("if (!xLocals.add){%s}", afterLeaveEvent)).
			Attr("v-model", "vars.overlay"),
	).VSlot("{locals:xLocals}").Init("{add:false}")
	versionComponent = publish.DefaultVersionComponentFunc(m.mb, publish.VersionComponentConfig{
		Top: true, DisableListeners: true,
		DisableDataChangeTracking: true,
		WrapActionButtons: func(ctx *web.EventContext, obj interface{}, actionButtons []h.HTMLComponent, phraseHasPresetsDataChanged string) []h.HTMLComponent {
			if len(actionButtons) < 2 {
				return actionButtons
			}
			previewDevelopUrl := m.PreviewHref(ctx, ps)
			p, ok := obj.(publish.StatusInterface)
			if !ok {
				return actionButtons
			}
			if p.EmbedStatus().Status != publish.StatusDraft {
				return actionButtons
			}
			btn := VBtn(msgr.Preview).
				Attr(":disabled", phraseHasPresetsDataChanged).
				Href(previewDevelopUrl).
				Class("ml-2").
				Class("rounded").
				Variant(VariantElevated).Color(ColorPrimary).Height(36)
			if b.previewOpenNewTab {
				btn.Attr("target", "_blank")
			}
			return append([]h.HTMLComponent{btn}, actionButtons...)
		},
	})(obj, &presets.FieldContext{ModelInfo: m.mb.Info()}, ctx)
	pageAppbarContent = h.Components(
		h.Div(
			h.Div().Style("transform:rotateY(180deg)").Class("mr-4").Children(
				VIcon("mdi-exit-to-app").Attr("@click", fmt.Sprintf(`
						const last = vars.__history.last();
						if (last && last.url && last.url === %q) {
							$event.view.window.history.back();
							return;
						}
						%s`, exitHref, web.GET().URL(exitHref).PushState(true).Go(),
				)),
			),
			VAppBarTitle().Text(title),
		).Class("d-inline-flex align-center"),
		h.Div(deviceToggle).Class("text-center d-flex justify-space-between mx-6"),
		versionComponent,
		publish.NewListenerModelsDeleted(m.mb, ctx.Param(presets.ParamID)),
		newListenerVersionSelected(ctx, m.mb, m.editorURL(), ctx.Param(presets.ParamID)),
	)
	if navigatorDrawer, err = b.renderNavigator(ctx, m); err != nil {
		return
	}
	body = h.Components(
		h.Div().Style("display:none").
			Attr("v-on-unmounted", `({window})=>{
				vars.$pbRightDrawerOnMouseLeave = null
				vars.$pbRightDrawerOnMouseDown = null
				vars.$pbRightDrawerOnMouseMove = null
				window.removeEventListener('resize', vars.$PagebuilderResizeFn)
				vars.$PagebuilderResizeFn = null
				vars.$pbRightThrottleTimer = null
            }`).
			Attr("v-on-mounted", `({ref, window, computed})=>{
				const rightDrawerExpendMinWidth = 350 + 56 // 56 is the width of the gap, 390 is the actual size
				const leftDrawerExpendMinWidth = 350

				vars.$pbLeftDrawerFolded = window.localStorage.getItem("$pbLeftDrawerFolded") === '1'
				vars.$pbRightDrawerFolded = window.localStorage.getItem("$pbRightDrawerFolded") === '1'
				vars.$pbLeftDrawerWidth = computed(()=>vars.$pbLeftDrawerFolded ? 32 : leftDrawerExpendMinWidth)
				vars.$pbRightAdjustableWidth = +window.localStorage.getItem("$pbRightAdjustableWidth") || rightDrawerExpendMinWidth
				vars.$pbRightDrawerWidth = computed(()=>vars.$pbRightDrawerFolded ? 32 : vars.$pbRightAdjustableWidth)
				vars.$pbLeftIconName = computed(()=> vars.$pbLeftDrawerFolded ? "mdi-chevron-right": "mdi-chevron-left")
				vars.$pbRightIconName = computed(()=> vars.$pbRightDrawerFolded ? "mdi-chevron-left": "mdi-chevron-right")
				vars.$pbRightDrawerHighlight = false
				vars.$pbRightDrawerIsDragging = false
				vars.$window = window
				vars.$pbRightThrottleTimer = null

				vars.$PagebuilderResizeFn = () => {
					if(vars.$pbRightDrawerFolded) return
					
					if(vars.$pbRightThrottleTimer) return

					vars.$pbRightThrottleTimer = window.setTimeout(() => {
						const halfWindowWidth = window.innerWidth / 2
						vars.$pbRightThrottleTimer = null
						if((vars.$pbRightDrawerWidth > halfWindowWidth) && (vars.$pbRightDrawerWidth > rightDrawerExpendMinWidth)) {
						vars.$pbRightAdjustableWidth =  halfWindowWidth
					} 
						window.localStorage.setItem("$pbRightAdjustableWidth", vars.$pbRightAdjustableWidth)
					},300)
				}

				vars.$PagebuilderResizeFn()
				window.addEventListener('resize', vars.$PagebuilderResizeFn)

				// common left drawer control function
				vars.$pbLeftDrawerControl = (action = 'toggle') => {
					switch(action) {
						case 'show':
							if (vars.$pbLeftDrawerFolded) {
								vars.$pbLeftDrawerFolded = false
								vars.$window.localStorage.setItem("$pbLeftDrawerFolded", "0")
							}
							break
						case 'hide':
							if (!vars.$pbLeftDrawerFolded) {
								vars.$pbLeftDrawerFolded = true
								vars.$window.localStorage.setItem("$pbLeftDrawerFolded", "1")
							}
							break
						case 'toggle':
						default:
							vars.$pbLeftDrawerFolded = !vars.$pbLeftDrawerFolded
							vars.$window.localStorage.setItem("$pbLeftDrawerFolded", vars.$pbLeftDrawerFolded ? "1": "0")
							break
					}
				}

				// 通用的右侧drawer控制函数
				vars.$pbRightDrawerControl = (action = 'toggle') => {
					switch(action) {
						case 'show':
							if (vars.$pbRightDrawerFolded) {
								vars.$pbRightDrawerFolded = false
								vars.$window.localStorage.setItem("$pbRightDrawerFolded", "0")
							}
							break
						case 'hide':
							if (!vars.$pbRightDrawerFolded) {
								vars.$pbRightDrawerFolded = true
								vars.$window.localStorage.setItem("$pbRightDrawerFolded", "1")
							}
							break
						case 'toggle':
						default:
							vars.$pbRightDrawerFolded = !vars.$pbRightDrawerFolded
							vars.$window.localStorage.setItem("$pbRightDrawerFolded", vars.$pbRightDrawerFolded ? "1": "0")
							break
					}
				}

				function addInlineStyle(css) {
					const style = window.document.createElement('style');
					style.type = 'text/css';
					style.appendChild(window.document.createTextNode(css));
					window.document.head.appendChild(style);
				}
				
				addInlineStyle(".event-none{pointer-events:none}");

				const $body = window.document.querySelector("body")
				const borderWidth = 5
				const draggableEl = ref(null)

				vars.$pbRightDrawerRefGet = (el) => {
					if(!el) return
					draggableEl.value = el.parentElement?.parentElement || el.parentElement
				}

				function isOnLeftBorder(event) {
					const rect = draggableEl.value.getBoundingClientRect()
					const x = event.clientX - rect.left
					// console.log(event.clientX,rect.left, event.clientX - rect.left)

					return x <= borderWidth
				}
				
				function onMouseMove(event) {
					if (vars.$pbRightDrawerIsDragging) {
						if (animationFrameId) return;
						animationFrameId = window.requestAnimationFrame(() => {
							const rect = draggableEl.value.getBoundingClientRect();
							const dx = rect.right - event.clientX;
							const minWidth = rightDrawerExpendMinWidth; 
							const maxWidth = window.innerWidth / 2;

							const newWidth = Math.min(Math.max(dx, minWidth), maxWidth);
							if (vars.$pbRightAdjustableWidth !== newWidth) {
								vars.$pbRightAdjustableWidth = newWidth;
							}

							animationFrameId = null;
						});
					}
				}

				function onMouseUp () {
					vars.$pbRightDrawerIsDragging = false
					vars.$pbRightDrawerHighlight = false
					$body.classList.remove('event-none')
					window.localStorage.setItem("$pbRightAdjustableWidth", vars.$pbRightAdjustableWidth)
					window.removeEventListener("mousemove", onMouseMove);
      		window.removeEventListener("mouseup", onMouseUp);
				}

				vars.$pbRightDrawerOnMouseDown = (event) => {
					if(vars.$pbRightDrawerFolded) return

					if (isOnLeftBorder(event)) {
						vars.$pbRightDrawerIsDragging = true
						event.preventDefault()
						$body.classList.add('event-none')
						 window.addEventListener("mousemove", onMouseMove, {passive:true, capture:true});
      			 window.addEventListener("mouseup", onMouseUp, {capture:true});
					}

					if(vars.$pbRightDrawerIsDragging) {
						vars.$pbRightDrawerHighlight = true
					}
				}

				vars.$pbRightDrawerOnMouseLeave = (event) => {
					if(vars.$pbRightDrawerIsDragging) return
					vars.$pbRightDrawerHighlight = false
				}

				vars.$pbRightDrawerOnMouseMove = (event) => {
					if(vars.$pbRightDrawerFolded || vars.$pbRightDrawerIsDragging) return

					vars.$pbRightDrawerHighlight = isOnLeftBorder(event)
				}
            }`),
		VAppBar(
			h.Div(
				pageAppbarContent...,
			).Class("page-builder-edit-bar-wrap"),
		).Elevation(0).Density(DensityCompact).Height(96).Class("align-center border-b"),
		h.If(!isStag,
			VNavigationDrawer(
				h.Div(
					web.Portal(navigatorDrawer).Name(pageBuilderLayerContainerPortal),
				).Attr("v-show", "!vars.$pbLeftDrawerFolded"),
				web.Slot(
					VBtn("").
						Attr(":icon", "vars.$pbLeftIconName").
						Attr("@click.stop", "vars.$pbLeftDrawerControl('toggle')").
						Size(SizeSmall).
						Class("pb-drawer-btn drawer-btn-left")).
					Name("append"),
			).Location(LocationLeft).
				Permanent(true).
				Attr(":width", "vars.$pbLeftDrawerWidth"),
			VNavigationDrawer(
				h.Div().Style("display:none").Attr("v-on-mounted", `({el,window}) => {
							el.__handleScroll = (event) => {
								locals.__pageBuilderRightContentScrollTop = event.target.scrollTop;
							}
							el.parentElement.addEventListener('scroll', el.__handleScroll)
	
							locals.__pageBuilderRightContentKeepScroll = () => {
							const scrollTop = locals.__pageBuilderRightContentScrollTop;
							window.setTimeout(()=>{
  							el.parentElement.scrollTop = scrollTop;	
							},0)							}
                        }`).Attr("v-on-unmounted", `({el}) => {
							el.parentElement.removeEventListener('scroll', el.__handleScroll);
						}`),

				web.Slot(
					VBtn("").
						Attr("v-if", "!vars.$pbRightDrawerIsDragging").
						Attr(":icon", "vars.$pbRightIconName").
						Attr("@mousemove.stop", "()=>{vars.$pbRightDrawerHighlight=false}").
						Attr("@mousedown.stop", "()=>{vars.$pbRightDrawerHighlight=false}").
						Attr("@click", "vars.$pbRightDrawerControl('toggle')").
						Size(SizeSmall).
						Class("pb-drawer-btn drawer-btn-right")).
					Name("append"),
				h.Div(
					web.Portal(editContainerDrawer).Name(pageBuilderRightContentPortal),
				).Attr("v-show", "!vars.$pbRightDrawerFolded").Attr(":ref", "vars.$pbRightDrawerRefGet"),
			).Location(LocationRight).
				Permanent(true).
				Attr(":class", "['draggable-el',{'border-left-draggable-highlight': vars.$pbRightDrawerHighlight}]").
				Attr(":width", "vars.$pbRightDrawerWidth").
				Attr("@mousedown", "vars.$pbRightDrawerOnMouseDown").
				Attr("@mousemove", "vars.$pbRightDrawerOnMouseMove").
				Attr("@mouseleave", "vars.$pbRightDrawerOnMouseLeave"),
		),
		VMain(
			addOverlay,
			h.If(!isStag, vx.VXMessageListener().ListenFunc(b.generateEditorBarJsFunction(ctx)).Name("message")),
			tabContent.Body,
		).Attr(web.VAssign("vars", "{overlayEl:$}")...).Class("ma-2"),
	)
	return
}

func (b *Builder) getDevice(ctx *web.EventContext) (device string, style string) {
	device = ctx.R.FormValue(paramDevice)
	if device == "" {
		device = b.defaultDevice
	}
	for _, d := range b.devices {
		if d.Name == device {
			style = d.Width
			return
		}
	}
	return
}

type CtxKeyContainerToPageLayout struct{}

type ContainerSorterItem struct {
	Index           int    `json:"index"`
	Label           string `json:"label"`
	ModelName       string `json:"model_name"`
	ModelID         string `json:"model_id"`
	DisplayName     string `json:"display_name"`
	ContainerID     string `json:"container_id"`
	ContainerDataID string `json:"container_data_id"`
	URL             string `json:"url"`
	Shared          bool   `json:"shared"`
	Hidden          bool   `json:"hidden"`
	VisibilityIcon  string `json:"visibility_icon"`
	ParamID         string `json:"param_id"`
	Locale          string `json:"locale"`
}

type ContainerSorter struct {
	Items []ContainerSorterItem `json:"items"`
}

func (b *Builder) renderNavigator(ctx *web.EventContext, m *ModelBuilder) (r h.HTMLComponent, err error) {
	if r, err = m.renderContainersSortedList(ctx); err != nil {
		return
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

func (b *Builder) pageEditorLayout(ctx *web.EventContext, m *ModelBuilder) h.HTMLComponent {
	return VLayout(
		web.Portal().Name(presets.RightDrawerPortalName),
		web.Portal().Name(presets.DialogPortalName),
		web.Portal().Name(presets.DeleteConfirmPortalName),
		web.Portal().Name(presets.ListingDialogPortalName),
		web.Portal().Name(dialogPortalName),
		web.Portal().Name(addContainerDialogPortal),
		b.editorBody(ctx, m),
	).Attr("id", "vt-app").
		Attr(web.VAssign("vars", fmt.Sprintf(`{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false,overlay:false,containerPreview:false,%s:{}}`, presets.VarsPresetsDataChanged))...)
}

func scrollToContainer(containerDataID interface{}) string {
	return fmt.Sprintf(`vars.el.refs.scrollIframe.scrollToCurrentContainer(%v);`, containerDataID)
}

func (b *Builder) containerWrapper(r *h.HTMLTagBuilder, ctx *web.EventContext, isEditor, isReadonly, isFirst, isEnd bool, containerDataID, modelName string, input *RenderInput) h.HTMLComponent {
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
			r = h.Div(
				h.Div().Class("inner-shadow"),
				h.Div(h.Div(r).Class("inner-container")).Attr("onclick", "event.stopPropagation();document.querySelectorAll('.highlight').forEach(item=>{item.classList.remove('highlight')});this.parentElement.classList.add('highlight');"+pmb.postMessage(EventEdit)),
				h.Div(
					h.Div(h.Text(input.DisplayName)).Class("title"),
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
	r.Attr("data-container-id", containerDataID)
	return r
}

type (
	postMessageBody struct {
		MsgType         string `json:"msg_type"`
		ContainerDataID string `json:"container_data_id"`
		ContainerId     string `json:"container_id"`
		DisplayName     string `json:"display_name"`
		ModelName       string `json:"model_name"`
		ShowRightDrawer bool   `json:"show_right_drawer"`
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
	b.ShowRightDrawer = (msgType == EventEdit)
	return fmt.Sprintf(`
const {top, left, width, height} = event.target.getBoundingClientRect();
const data= %s;
data.rect = {top, left, width, height}
window.parent.postMessage(data, '*')`, h.JSONString(b))
}

func addVirtualELeToContainer(containerDataID interface{}) string {
	return fmt.Sprintf(`vars.el.refs.scrollIframe.addVirtualElement(%v);`, containerDataID)
}

func removeVirtualElement() string {
	return "vars.el.refs.scrollIframe.removeVirtualElement();"
}

func appendVirtualElement() string {
	return "vars.el.refs.scrollIframe.appendVirtualElement();"
}

func newListenerVersionSelected(evCtx *web.EventContext, mb *presets.ModelBuilder, path, slug string) h.HTMLComponent {
	event := actions.Edit
	if mb.HasDetailing() {
		event = actions.DetailingDrawer
	}
	drawerToSlug := web.Plaid().URL(path).EventFunc(event).
		Query(presets.ParamID, web.Var("payload.slug"))
	varCurrentActive := evCtx.R.FormValue(presets.ParamVarCurrentActive)
	if varCurrentActive != "" {
		drawerToSlug.Query(presets.ParamVarCurrentActive, varCurrentActive)
	}
	return web.Listen(publish.NotifVersionSelected(mb), fmt.Sprintf(`
		if (payload.slug === %q) {
			return
		}
		if (vars.presetsRightDrawer) {
			%s
			return
		}
		%s
	`,
		slug,
		presets.CloseRightDrawerVarScript+";"+drawerToSlug.Go(),
		web.Plaid().PushState(true).URL(web.Var(fmt.Sprintf(`%q + "/" + payload.slug`, path))).Go(),
	))
}
