package pagebuilder

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/qor5/admin/l10n"
	"github.com/qor5/admin/presets"
	"github.com/qor5/admin/presets/actions"
	"github.com/qor5/admin/publish"
	. "github.com/qor5/ui/vuetify"
	vx "github.com/qor5/ui/vuetifyx"
	"github.com/qor5/web"
	"github.com/qor5/x/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/pat"
	"gorm.io/gorm"
)

const (
	AddContainerDialogEvent          = "page_builder_AddContainerDialogEvent"
	AddContainerEvent                = "page_builder_AddContainerEvent"
	DeleteContainerConfirmationEvent = "page_builder_DeleteContainerConfirmationEvent"
	DeleteContainerEvent             = "page_builder_DeleteContainerEvent"
	MoveContainerEvent               = "page_builder_MoveContainerEvent"
	ToggleContainerVisibilityEvent   = "page_builder_ToggleContainerVisibilityEvent"
	MarkAsSharedContainerEvent       = "page_builder_MarkAsSharedContainerEvent"
	RenameCotainerDialogEvent        = "page_builder_RenameContainerDialogEvent"
	RenameContainerEvent             = "page_builder_RenameContainerEvent"

	paramPageID          = "pageID"
	paramPageVersion     = "pageVersion"
	paramLocale          = "locale"
	paramContainerID     = "containerID"
	paramMoveResult      = "moveResult"
	paramContainerName   = "containerName"
	paramSharedContainer = "sharedContainer"
	paramModelID         = "modelID"

	Device_Phone    = "phone"
	Device_Tablet   = "tablet"
	Device_Computer = "computer"
)

func (b *Builder) Preview(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.FormValue("locale")
	ctx.Injector.HeadHTMLComponent("style", b.pageStyle, true)

	var p *Page
	r.Body, p, err = b.renderPageOrTemplate(ctx, isTpl, id, version, locale, false)
	if err != nil {
		return
	}
	r.PageTitle = p.Title
	return
}

const editorPreviewContentPortal = "editorPreviewContentPortal"

func (b *Builder) Editor(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := pat.Param(ctx.R, "id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.Form.Get("locale")
	isLocalizable := ctx.R.Form.Has("locale")
	var body h.HTMLComponent
	var containerList h.HTMLComponent
	var device string
	var p *Page
	var previewHref string
	deviceQueries := url.Values{}
	if isTpl {
		previewHref = fmt.Sprintf("/preview?id=%s&tpl=1", id)
		deviceQueries.Add("tpl", "1")
		if isLocalizable && l10nON {
			previewHref = fmt.Sprintf("/preview?id=%s&tpl=1&locale=%s", id, locale)
			deviceQueries.Add("locale", locale)
		}
	} else {
		previewHref = fmt.Sprintf("/preview?id=%s&version=%s", id, version)
		deviceQueries.Add("version", version)

		if isLocalizable && l10nON {
			previewHref = fmt.Sprintf("/preview?id=%s&version=%s&locale=%s", id, version, locale)
			deviceQueries.Add("locale", locale)
		}
	}

	body, p, err = b.renderPageOrTemplate(ctx, isTpl, id, version, locale, true)
	if err != nil {
		return
	}
	r.PageTitle = fmt.Sprintf("Editor for %s: %s", id, p.Title)
	device, _ = b.getDevice(ctx)

	containerList, err = b.renderContainersList(ctx, p.ID, p.GetVersion(), p.GetLocale(), p.GetStatus() != publish.StatusDraft)
	if err != nil {
		return
	}
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	r.Body = h.Components(
		VAppBar(
			VSpacer(),

			VBtn("").Icon(true).Children(
				VIcon("phone_iphone"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "phone").PushState(true).Go()).
				Class("mr-10").InputValue(device == "phone"),

			VBtn("").Icon(true).Children(
				VIcon("tablet_mac"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "tablet").PushState(true).Go()).
				Class("mr-10").InputValue(device == "tablet"),

			VBtn("").Icon(true).Children(
				VIcon("laptop_mac"),
			).Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "laptop").PushState(true).Go()).
				InputValue(device == "laptop"),

			VSpacer(),

			VBtn(msgr.Preview).Text(true).Href(b.prefix+previewHref).Target("_blank"),
			VAppBarNavIcon().On("click.stop", "vars.navDrawer = !vars.navDrawer"),
		).Dark(true).
			Color("primary").
			App(true),

		VMain(
			VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
				Class("mt-6").
				Fluid(true),
			VNavigationDrawer(containerList).
				App(true).
				Right(true).
				Fixed(true).
				Value(true).
				Width(420).
				Attr("v-model", "vars.navDrawer").
				Attr(web.InitContextVars, `{navDrawer: null}`),
		),
	)

	return
}

func (b *Builder) getDevice(ctx *web.EventContext) (device string, style string) {
	device = ctx.R.FormValue("device")
	if len(device) == 0 {
		device = b.defaultDevice
	}

	switch device {
	case Device_Phone:
		style = "width: 414px;"
	case Device_Tablet:
		style = "width: 768px;"
		//case Device_Computer:
		//	style = "width: 1264px;"
	}

	return
}

const FreeStyleKey = "FreeStyle"

func (b *Builder) renderPageOrTemplate(ctx *web.EventContext, isTpl bool, pageOrTemplateID string, version, locale string, isEditor bool) (r h.HTMLComponent, p *Page, err error) {
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
	comps, err = b.renderContainers(ctx, p.ID, p.GetVersion(), p.GetLocale(), isEditor, isReadonly)
	if err != nil {
		return
	}
	r = h.Components(comps...)
	if b.pageLayoutFunc != nil {
		input := &PageLayoutInput{
			IsEditor:  isEditor,
			IsPreview: !isEditor,
			Page:      p,
		}

		if isEditor {
			input.EditorCss = append(input.EditorCss, h.RawHTML(`<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">`))
			input.EditorCss = append(input.EditorCss, h.Style(`
	.wrapper-shadow {
		position:absolute;
		width: 100%; 
		height: 100%;
		z-index:9999; 
		background: rgba(81, 193, 226, 0.25);
		opacity: 0;
		top: 0;
		left: 0;
	}
	.wrapper-shadow button{
		position:absolute;
		top: 0;
		right: 0;
    	line-height: 1;
		font-size: 0;
		border: 2px outset #767676;
		cursor: pointer;
	}
	.wrapper-shadow.hover {
		cursor: pointer;
		opacity: 1;
    }`))
			input.FreeStyleBottomJs = []string{`
	function scrolltoCurrentContainer(event) {
		const current = document.querySelector("div[data-container-id='"+event.data+"']");
		if (!current) {
			return;
		}
		const hover = document.querySelector(".wrapper-shadow.hover")
		if (hover) {
			hover.classList.remove('hover');
		}
		window.parent.scroll({top: current.offsetTop, behavior: "smooth"});
		current.querySelector(".wrapper-shadow").classList.add('hover');
	}
	document.querySelectorAll('.wrapper-shadow').forEach(shadow => {
		shadow.addEventListener('mouseover', event => {
			document.querySelectorAll(".wrapper-shadow.hover").forEach(item => {
				item.classList.remove('hover');
			})
			shadow.classList.add('hover');
		})
	})

	window.addEventListener("message", scrolltoCurrentContainer, false);
`}
		}
		if f := ctx.R.Context().Value(FreeStyleKey); f != nil {
			pl, ok := f.(*PageLayoutInput)
			if ok {
				input.FreeStyleCss = append(input.FreeStyleCss, pl.FreeStyleCss...)
				input.FreeStyleTopJs = append(input.FreeStyleTopJs, pl.FreeStyleTopJs...)
				input.FreeStyleBottomJs = append(input.FreeStyleBottomJs, pl.FreeStyleBottomJs...)
			}
		}
		r = b.pageLayoutFunc(h.Components(comps...), input, ctx)
		if isEditor {
			_, width := b.getDevice(ctx)
			iframeHeightName := "_iframeHeight"
			iframeHeightCookie, _ := ctx.R.Cookie(iframeHeightName)
			iframeValue := "1000px"
			if iframeHeightCookie != nil {
				iframeValue = iframeHeightCookie.Value
			}
			r = h.Div(
				h.RawHTML(fmt.Sprintf(`
						<iframe frameborder='0' scrolling='no' srcdoc="%s"
							@load='
								var height = $event.target.contentWindow.document.body.parentElement.offsetHeight+"px";
								$event.target.style.height=height;
								document.cookie="%s="+height;
							'
							style='width:100%%; display:block; border:none; padding:0; margin:0; height:%s;'></iframe>`,
					strings.ReplaceAll(
						h.MustString(r, ctx.R.Context()),
						"\"",
						"&quot;"),
					iframeHeightName,
					iframeValue,
				)),
			).Class("page-builder-container mx-auto").Attr("style", width)

		}
	}
	return
}

func (b *Builder) renderContainers(ctx *web.EventContext, pageID uint, pageVersion, locale string, isEditor bool, isReadonly bool) (r []h.HTMLComponent, err error) {
	var cons []*Container
	err = b.db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", pageID, pageVersion, locale).Error
	if err != nil {
		return
	}

	cbs := b.getContainerBuilders(cons)

	device, _ := b.getDevice(ctx)
	for _, ec := range cbs {
		if ec.container.Hidden {
			continue
		}
		obj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(obj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}

		input := RenderInput{
			IsEditor:   isEditor,
			IsReadonly: isReadonly,
			Device:     device,
		}
		pure := ec.builder.renderFunc(obj, &input, ctx)
		r = append(r, pure)
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
	VisibilityIcon string `json:"visibility_icon"`
	ParamID        string `json:"param_id"`
}

type ContainerSorter struct {
	Items []ContainerSorterItem `json:"items"`
}

func (b *Builder) renderContainersList(ctx *web.EventContext, pageID uint, pageVersion, locale string, isReadonly bool) (r h.HTMLComponent, err error) {
	var cons []*Container
	err = b.db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ?", pageID, pageVersion, locale).Error
	if err != nil {
		return
	}

	var sorterData ContainerSorter
	for i, c := range cons {
		vicon := "visibility"
		if c.Hidden {
			vicon = "visibility_off"
		}
		var displayName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, c.DisplayName)

		sorterData.Items = append(sorterData.Items,
			ContainerSorterItem{
				Index:          i,
				Label:          displayName,
				ModelName:      inflection.Plural(strcase.ToKebab(c.ModelName)),
				ModelID:        strconv.Itoa(int(c.ModelID)),
				DisplayName:    displayName,
				ContainerID:    strconv.Itoa(int(c.ID)),
				URL:            b.ContainerByName(c.ModelName).mb.Info().ListingHref(),
				Shared:         c.Shared,
				VisibilityIcon: vicon,
				ParamID:        c.PrimarySlug(),
			},
		)
	}
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	r = web.Scope(
		VSheet(
			VCard(
				h.Tag("vx-draggable").
					Attr("v-model", "locals.items", "handle", ".handle", "animation", "300").
					Attr("@end", web.Plaid().
						URL(fmt.Sprintf("%s/editors", b.prefix)).
						EventFunc(MoveContainerEvent).
						FieldValue(paramMoveResult, web.Var("JSON.stringify(locals.items)")).
						Go()).Children(
					// VList(
					h.Div(
						VListItem(
							h.If(!isReadonly,
								VListItemIcon(VBtn("").Icon(true).Children(VIcon("drag_indicator"))).Class("handle my-2 ml-1 mr-1"),
							).Else(
								VListItemIcon().Class("my-2 ml-1 mr-1"),
							),
							VListItemContent(
								VListItemTitle(h.Text("{{item.label}}")).Attr(":style", "[item.shared ? {'color':'green'}:{}]"),
							),
							h.If(!isReadonly,
								VListItemIcon(VBtn("").Icon(true).Children(VIcon("edit").Small(true))).Attr("@click",
									web.Plaid().
										URL(web.Var("item.url")).
										EventFunc(actions.Edit).
										Query(presets.ParamOverlay, actions.Dialog).
										Query(presets.ParamID, web.Var("item.model_id")).
										Go(),
								).Class("my-2"),
								VListItemIcon(VBtn("").Icon(true).Children(VIcon("{{item.visibility_icon}}").Small(true))).Attr("@click",
									web.Plaid().
										URL(web.Var("item.url")).
										EventFunc(ToggleContainerVisibilityEvent).
										Query(paramContainerID, web.Var("item.param_id")).
										Go(),
								).Class("my-2"),
							),
							h.If(!isReadonly,
								VMenu(
									web.Slot(
										VBtn("").Children(
											VIcon("more_horiz"),
										).Attr("v-on", "on").Text(true).Fab(true).Small(true),
									).Name("activator").Scope("{ on }"),

									VList(
										VListItem(
											VListItemIcon(VIcon("edit_note")).Class("pl-0 mr-2"),
											VListItemTitle(h.Text("Rename")),
										).Attr("@click",
											web.Plaid().
												URL(web.Var("item.url")).
												EventFunc(RenameCotainerDialogEvent).
												Query(paramContainerID, web.Var("item.param_id")).
												Query(paramContainerName, web.Var("item.display_name")).
												Query(presets.ParamOverlay, actions.Dialog).
												Go(),
										),
										VListItem(
											VListItemIcon(VIcon("delete")).Class("pl-0 mr-2"),
											VListItemTitle(h.Text("Delete")),
										).Attr("@click", web.Plaid().
											URL(web.Var("item.url")).
											EventFunc(DeleteContainerConfirmationEvent).
											Query(paramContainerID, web.Var("item.param_id")).
											Query(paramContainerName, web.Var("item.display_name")).
											Go(),
										),
										VListItem(
											VListItemIcon(VIcon("share")).Class("pl-1 mr-2"),
											VListItemTitle(h.Text("Mark As Shared Container")),
										).Attr("@click",
											web.Plaid().
												URL(web.Var("item.url")).
												EventFunc(MarkAsSharedContainerEvent).
												Query(paramContainerID, web.Var("item.param_id")).
												Go(),
										).Attr("v-if", "!item.shared"),
									).Dense(true),
								).Left(true),
							),
						).Class("pl-0").Attr("@click", fmt.Sprintf(`document.querySelector("iframe").contentWindow.postMessage(%s+"_"+%s,"*");`, web.Var("item.model_name"), web.Var("item.model_id"))),
						VDivider().Attr("v-if", "index < locals.items.length "),
					).Attr("v-for", "(item, index) in locals.items", ":key", "item.index"),
					h.If(!isReadonly,
						VListItem(
							VListItemIcon(VIcon("add").Color("primary")).Class("ma-4"),
							VListItemTitle(VBtn(msgr.AddContainers).Color("primary").Text(true)),
						).Attr("@click",
							web.Plaid().
								URL(fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, pageID, pageVersion, locale)).
								EventFunc(AddContainerDialogEvent).
								Query(paramPageID, pageID).
								Query(paramPageVersion, pageVersion).
								Query(paramLocale, locale).
								Query(presets.ParamOverlay, actions.Dialog).
								Go(),
						),
					),
					// ).Class("py-0"),
				),
			).Outlined(true),
		).Class("pa-4 pt-2"),
	).Init(h.JSONString(sorterData)).VSlot("{ locals }")
	return
}

func (b *Builder) AddContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	locale := ctx.R.FormValue(paramLocale)
	containerName := ctx.R.FormValue(paramContainerName)
	sharedContainer := ctx.R.FormValue(paramSharedContainer)
	modelID := ctx.QueryAsInt(paramModelID)
	var newModelID uint
	if sharedContainer == "true" {
		err = b.AddSharedContainerToPage(pageID, pageVersion, locale, containerName, uint(modelID))
		r.PushState = web.Location(url.Values{})
	} else {
		newModelID, err = b.AddContainerToPage(pageID, pageVersion, locale, containerName)
		r.VarsScript = web.Plaid().
			URL(b.ContainerByName(containerName).mb.Info().ListingHref()).
			EventFunc(actions.Edit).
			Query(presets.ParamOverlay, actions.Dialog).
			Query(presets.ParamID, fmt.Sprint(newModelID)).
			Go()
	}

	return
}

func (b *Builder) MoveContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	moveResult := ctx.R.FormValue(paramMoveResult)

	var result []ContainerSorterItem
	err = json.Unmarshal([]byte(moveResult), &result)
	if err != nil {
		return
	}
	err = b.db.Transaction(func(tx *gorm.DB) (inerr error) {
		for i, r := range result {
			if inerr = tx.Model(&Container{}).Where("id = ?", r.ContainerID).Update("display_order", i+1).Error; inerr != nil {
				return
			}
		}
		return
	})

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) ToggleContainerVisibility(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	paramID := ctx.R.FormValue(paramContainerID)
	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]

	err = b.db.Exec("UPDATE page_builder_containers SET hidden = NOT(COALESCE(hidden,FALSE)) WHERE id = ? AND locale_code = ?", containerID, locale).Error

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) DeleteContainerConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	paramID := ctx.R.FormValue(paramContainerID)

	containerName := ctx.R.FormValue(paramContainerName)

	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: presets.DeleteConfirmPortalName,
		Body: VDialog(
			VCard(
				VCardTitle(h.Text(fmt.Sprintf("Are you sure you want to delete %s?", containerName))),
				VCardActions(
					VSpacer(),
					VBtn("Cancel").
						Depressed(true).
						Class("ml-2").
						On("click", "vars.deleteConfirmation = false"),

					VBtn("Delete").
						Color("primary").
						Depressed(true).
						Dark(true).
						Attr("@click", web.Plaid().
							URL(fmt.Sprintf("%s/editors", b.prefix)).
							EventFunc(DeleteContainerEvent).
							Query(paramContainerID, paramID).
							Go()),
				),
			),
		).MaxWidth("600px").
			Attr("v-model", "vars.deleteConfirmation").
			Attr(web.InitContextVars, `{deleteConfirmation: false}`),
	})

	r.VarsScript = "setTimeout(function(){ vars.deleteConfirmation = true }, 100)"
	return
}

func (b *Builder) DeleteContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	paramID := ctx.R.FormValue(paramContainerID)
	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]

	err = b.db.Delete(&Container{}, "id = ? AND locale_code = ?", containerID, locale).Error
	if err != nil {
		return
	}
	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) AddContainerToPage(pageID int, pageVersion, locale, containerName string) (modelID uint, err error) {
	model := b.ContainerByName(containerName).NewModel()
	var dc DemoContainer
	b.db.Where("model_name = ? AND locale_code = ?", containerName, locale).First(&dc)
	if dc.ID != 0 && dc.ModelID != 0 {
		b.db.Where("id = ?", dc.ModelID).First(model)
		reflectutils.Set(model, "ID", uint(0))
	}

	err = b.db.Create(model).Error
	if err != nil {
		return
	}

	var maxOrder sql.NullFloat64
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ? and locale_code = ?", pageID, pageVersion, locale).Scan(&maxOrder).Error
	if err != nil {
		return
	}

	modelID = reflectutils.MustGet(model, "ID").(uint)
	err = b.db.Create(&Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		ModelName:    containerName,
		DisplayName:  containerName,
		ModelID:      modelID,
		DisplayOrder: maxOrder.Float64 + 1,
		Locale: l10n.Locale{
			LocaleCode: locale,
		},
	}).Error
	if err != nil {
		return
	}
	return
}

func (b *Builder) AddSharedContainerToPage(pageID int, pageVersion, locale, containerName string, modelID uint) (err error) {
	var c Container
	err = b.db.First(&c, "model_name = ? AND model_id = ? AND shared = true", containerName, modelID).Error
	if err != nil {
		return
	}
	var maxOrder sql.NullFloat64
	err = b.db.Model(&Container{}).Select("MAX(display_order)").Where("page_id = ? and page_version = ? and locale_code = ?", pageID, pageVersion, locale).Scan(&maxOrder).Error
	if err != nil {
		return
	}

	err = b.db.Create(&Container{
		PageID:       uint(pageID),
		PageVersion:  pageVersion,
		ModelName:    containerName,
		DisplayName:  c.DisplayName,
		ModelID:      modelID,
		Shared:       true,
		DisplayOrder: maxOrder.Float64 + 1,
		Locale: l10n.Locale{
			LocaleCode: locale,
		},
	}).Error
	if err != nil {
		return
	}
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
			var temp Container
			if err = db.Where("model_name = ? AND locale_code = ?", c.ModelName, toPageLocale).First(&temp).Count(&count).Error; err != nil && err != gorm.ErrRecordNotFound {
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
				newModelID = temp.ModelID
				newDisplayName = temp.DisplayName
			}
		}

		if err = db.Create(&Container{
			Model:        gorm.Model{ID: c.ID},
			PageID:       uint(toPageID),
			PageVersion:  toPageVersion,
			ModelName:    c.ModelName,
			DisplayName:  newDisplayName,
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

func (b *Builder) MarkAsSharedContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

func (b *Builder) RenameContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	paramID := ctx.R.FormValue(paramContainerID)
	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]
	name := ctx.R.FormValue("DisplayName")
	var c Container
	err = b.db.First(&c, "id = ? AND locale_code = ?  ", containerID, locale).Error
	if err != nil {
		return
	}
	if c.Shared {
		err = b.db.Model(&Container{}).Where("model_name = ? AND model_id = ? AND locale_code = ?", c.ModelName, c.ModelID, locale).Update("display_name", name).Error
		if err != nil {
			return
		}
	} else {
		err = b.db.Model(&Container{}).Where("id = ? AND locale_code = ?", containerID, locale).Update("display_name", name).Error
		if err != nil {
			return
		}
	}

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) RenameContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	paramID := ctx.R.FormValue(paramContainerID)
	name := ctx.R.FormValue(paramContainerName)
	okAction := web.Plaid().
		URL(fmt.Sprintf("%s/editors", b.prefix)).
		EventFunc(RenameContainerEvent).Query(paramContainerID, paramID).Go()
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: dialogPortalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(h.Text("Rename")),
					VCardText(
						VTextField().FieldName("DisplayName").Value(name),
					),
					VCardActions(
						VSpacer(),
						VBtn("Cancel").
							Depressed(true).
							Class("ml-2").
							On("click", "locals.renameDialog = false"),

						VBtn("OK").
							Color("primary").
							Depressed(true).
							Dark(true).
							Attr("@click", okAction),
					),
				),
			).MaxWidth("400px").
				Attr("v-model", "locals.renameDialog"),
		).Init("{renameDialog:true}").VSlot("{locals}"),
	})
	return
}

func (b *Builder) AddContainerDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	locale := ctx.R.FormValue(paramLocale)
	// okAction := web.Plaid().EventFunc(RenameContainerEvent).Query(paramContainerID, containerID).Go()
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	var containers []h.HTMLComponent
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
							Text(true).
							Color("primary").Attr("@click",
							"locals.addContainerDialog = false;"+web.Plaid().
								URL(fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, pageID, pageVersion, locale)).
								EventFunc(AddContainerEvent).
								Query(paramPageID, pageID).
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
							Text(true).
							Color("primary").Attr("@click",
							"locals.addContainerDialog = false;"+web.Plaid().
								URL(fmt.Sprintf("%s/editors/%d?version=%s&locale=%s", b.prefix, pageID, pageVersion, locale)).
								EventFunc(AddContainerEvent).
								Query(paramPageID, pageID).
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
					VTab(h.Text(msgr.New)),
					VTabItem(
						VSheet(
							VContainer(
								VRow(
									containers...,
								),
							),
						),
					).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
					VTab(h.Text(msgr.Shared)),
					VTabItem(
						VSheet(
							VContainer(
								VRow(
									sharedContainers...,
								),
							),
						),
					).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
				),
			).Width("1200px").Attr("v-model", "locals.addContainerDialog"),
		).Init("{addContainerDialog:true}").VSlot("{locals}"),
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

		ctx.Injector.HeadHTML(strings.Replace(`
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto+Mono">
			<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Roboto:300,400,500">
			<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
			<link rel="stylesheet" href="{{prefix}}/assets/main.css">
			<script src='{{prefix}}/assets/vue.js'></script>

			<style>
				.page-builder-container {
					overflow: hidden;
					box-shadow: -10px 0px 13px -7px rgba(0,0,0,.3), 10px 0px 13px -7px rgba(0,0,0,.18), 5px 0px 15px 5px rgba(0,0,0,.12);	
				}
				[v-cloak] {
					display: none;
				}
			</style>
		`, "{{prefix}}", b.prefix, -1))

		b.ps.InjectExtraAssets(ctx)

		if len(os.Getenv("DEV_PRESETS")) > 0 {
			ctx.Injector.TailHTML(`
<script src='http://localhost:3080/js/chunk-vendors.js'></script>
<script src='http://localhost:3080/js/app.js'></script>
<script src='http://localhost:3100/js/chunk-vendors.js'></script>
<script src='http://localhost:3100/js/app.js'></script>
			`)

		} else {
			ctx.Injector.TailHTML(strings.Replace(`
			<script src='{{prefix}}/assets/main.js'></script>
			`, "{{prefix}}", b.prefix, -1))
		}

		var innerPr web.PageResponse
		innerPr, err = in(ctx)
		if err != nil {
			panic(err)
		}

		action := web.POST().
			EventFunc(actions.Edit).
			URL(web.Var("\""+b.prefix+"/\"+arr[0]")).
			Query(presets.ParamOverlay, actions.Dialog).
			Query(presets.ParamID, web.Var("arr[1]")).
			// Query(presets.ParamOverlayAfterUpdateScript,
			// 	web.Var(
			// 		h.JSONString(web.POST().
			// 			PushState(web.Location(url.Values{})).
			// 			MergeQuery(true).
			// 			ThenScript(`setTimeout(function(){ window.scroll({left: __scrollLeft__, top: __scrollTop__, behavior: "auto"}) }, 50)`).
			// 			Go())+".replace(\"__scrollLeft__\", scrollLeft).replace(\"__scrollTop__\", scrollTop)",
			// 	),
			// ).
			Go()
		pr.PageTitle = fmt.Sprintf("%s - %s", innerPr.PageTitle, "Page Builder")
		pr.Body = VApp(

			web.Portal().Name(presets.RightDrawerPortalName),
			web.Portal().Name(presets.DialogPortalName),
			web.Portal().Name(presets.DeleteConfirmPortalName),
			web.Portal().Name(dialogPortalName),
			h.Script(`
(function(){

	let scrollLeft = 0;
	let scrollTop = 0;
	
	function pause(duration) {
		return new Promise(res => setTimeout(res, duration));
	}
	function backoff(retries, fn, delay = 100) {
		fn().catch(err => retries > 1
			? pause(delay).then(() => backoff(retries - 1, fn, delay * 2)) 
			: Promise.reject(err));
	}

	function restoreScroll() {
		window.scroll({left: scrollLeft, top: scrollTop, behavior: "auto"});
		if (window.scrollX == scrollLeft && window.scrollY == scrollTop) {
			return Promise.resolve();
		}
		return Promise.reject();
	}

	window.addEventListener('fetchStart', (event) => {
		scrollLeft = window.scrollX;
		scrollTop = window.scrollY;
	});
	
	window.addEventListener('fetchEnd', (event) => {
		backoff(5, restoreScroll, 100);
	});
})()

`),
			vx.VXMessageListener().ListenFunc(fmt.Sprintf(`
function(e){
	if (!e.data.split) {
		return
	}
	let arr = e.data.split("_");
	if (arr.length != 2) {
		console.log(arr);
		return
	}
	%s
}`, action)),

			innerPr.Body.(h.HTMLComponent),
		).Id("vt-app").Attr(web.InitContextVars, `{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false}`)

		return
	}
}
