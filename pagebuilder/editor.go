package pagebuilder

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qor5/admin/v3/utils"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/actions"
	"github.com/qor5/admin/v3/publish"
	. "github.com/qor5/ui/v3/vuetify"
	vx "github.com/qor5/ui/v3/vuetifyx"
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"goji.io/v3/pat"
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
	ShowAddContainerDrawerEvent      = "page_builder_ShowAddContainerDrawerEvent"

	paramPageID          = "pageID"
	paramPageVersion     = "pageVersion"
	paramLocale          = "locale"
	paramContainerID     = "containerID"
	paramMoveResult      = "moveResult"
	paramContainerName   = "containerName"
	paramSharedContainer = "sharedContainer"
	paramModelID         = "modelID"
	paramMoveDirection   = "paramMoveDirection"

	DevicePhone    = "phone"
	DeviceTablet   = "tablet"
	DeviceComputer = "computer"

	EventUp     = "up"
	EventDown   = "down"
	EventDelete = "delete"
	EventAdd    = "add"
	EventEdit   = "edit"
)

func (b *Builder) Preview(ctx *web.EventContext) (r web.PageResponse, err error) {
	isTpl := ctx.R.FormValue("tpl") != ""
	id := ctx.R.FormValue("id")
	version := ctx.R.FormValue("version")
	locale := ctx.R.FormValue("locale")

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

	containerList = b.renderContainersList(ctx, p.GetStatus() != publish.StatusDraft)
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	r.Body = h.Components(
		web.Scope(
			VAppBar(
				VSpacer(),
				// icon was phone_iphone
				VBtn("").Icon("mdi-cellphone").Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "phone").PushState(true).Go()).
					Class("mr-10").Active(device == "phone"),

				// icon was tablet_mac
				VBtn("").Icon("mdi-tablet").Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "tablet").PushState(true).Go()).
					Class("mr-10").Active(device == "tablet"),
				// icon was laptop_mac
				VBtn("").Icon("mdi-laptop").Attr("@click", web.Plaid().Queries(deviceQueries).Query("device", "laptop").PushState(true).Go()).
					Active(device == "laptop"),

				VSpacer(),

				VBtn(msgr.Preview).Variant(VariantText).Href(b.prefix+previewHref).To("_blank"),
				VAppBarNavIcon().On("click.stop", "drawerLocals.navDrawer = !drawerLocals.navDrawer"),
			).Theme(ThemeDark).
				Color("primary"),
			VMain(
				VContainer(web.Portal(body).Name(editorPreviewContentPortal)).
					Class("mt-6").
					Fluid(true),
				VNavigationDrawer(containerList).
					Width(420).
					Attr("v-model", "drawerLocals.navDrawer"),
			),
		).VSlot(" { locals : drawerLocals } ").Init(`{navDrawer: null}`),
	)

	return
}

func (b *Builder) getDevice(ctx *web.EventContext) (device string, style string) {
	device = ctx.R.FormValue("device")
	if len(device) == 0 {
		device = b.defaultDevice
	}

	switch device {
	case DevicePhone:
		style = "width: 414px;"
	case DeviceTablet:
		style = "width: 768px;"
		// case Device_Computer:
		//	style = "width: 1264px;"
	}

	return
}

const ContainerToPageLayoutKey = "ContainerToPageLayout"

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
	.wrapper-shadow .editor-bar{
		position:absolute;
		bottom: 0;
		left: 0;
    	line-height: 1;
		font-size: 0;
		//border: 2px outset #767676;
		cursor: pointer;
        background-color: #3E63DD;
        display: flex;
        align-items:center;
	}
	.wrapper-shadow .editor-add{
		position:absolute;
		bottom: 0;
		left: 50%;
    	line-height: 1;
		font-size: 0;
		//border: 2px outset #767676;
		cursor: pointer;
        background-color: #3E63DD;
	}
    .wrapper-shadow .bar {
      color: #FFFFFF;
      background-color: #3E63DD;
      display:inline;
     }
	 .wrapper-shadow .title{
      color: #FFFFFF;
      background-color: #3E63DD;
     margin-right:10px;
     }
	.wrapper-shadow.hover {
		cursor: pointer;
		opacity: 1;
    }`))
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
				).AttrIf("lang", newCtx.Injector.GetHTMLLang(), newCtx.Injector.GetHTMLLang() != ""),
			}
			_, width := b.getDevice(ctx)
			iframeHeightName := "_iframeHeight"
			iframeHeightCookie, _ := ctx.R.Cookie(iframeHeightName)
			iframeValue := "1000px"
			if iframeHeightCookie != nil {
				iframeValue = iframeHeightCookie.Value
			}
			r = h.Div(
				h.Tag("vx-scroll-iframe").Attr(
					":srcdoc", h.JSONString(h.MustString(r, ctx.R.Context()))).
					Attr(":iframe-height-name", h.JSONString(iframeHeightName)).
					Attr(":iframe-value", h.JSONString(iframeValue)).
					Attr("ref", "scrollIframe"),
			).Id("vx-drag-target-area").Class("page-builder-container mx-auto").Attr("style", width)

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

	cbs := b.getContainerBuilders(cons)

	device, _ := b.getDevice(ctx)
	for i, ec := range cbs {
		if ec.container.Hidden {
			continue
		}
		obj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(obj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}
		var displayName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, ec.container.DisplayName)
		input := RenderInput{
			Page:        p,
			IsEditor:    isEditor,
			IsReadonly:  isReadonly,
			Device:      device,
			ContainerId: ec.container.PrimarySlug(),
			DisplayName: displayName,
			IsFirst:     i == 0,
			IsEnd:       i == len(cbs)-1,
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
	Locale         string `json:"locale"`
}

type ContainerSorter struct {
	Items []ContainerSorterItem `json:"items"`
}

func (b *Builder) renderContainersList(ctx *web.EventContext, isReadonly bool) (r h.HTMLComponent) {
	r = VLayout(
		VAppBar().Title("Elements"),
		VMain(
			b.ContainerComponent(ctx, isReadonly),
		),
	)
	return
}

func (b *Builder) AddContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
	pageID := ctx.QueryAsInt(paramPageID)
	pageVersion := ctx.R.FormValue(paramPageVersion)
	locale := ctx.R.FormValue(paramLocale)
	containerName := ctx.R.FormValue(paramContainerName)
	sharedContainer := ctx.R.FormValue(paramSharedContainer)
	modelID := ctx.QueryAsInt(paramModelID)
	if sharedContainer == "true" {
		err = b.AddSharedContainerToPage(pageID, pageVersion, locale, containerName, uint(modelID))
	} else {
		_, err = b.AddContainerToPage(pageID, pageVersion, locale, containerName)
		//r.RunScript = web.Plaid().
		//	URL(b.ContainerByName(containerName).mb.Info().ListingHref()).
		//	EventFunc(actions.Edit).
		//	Query(presets.ParamOverlay, actions.Drawer).
		//	Query(presets.ParamID, fmt.Sprint(newModelID)).
		//	Go()

	}
	r.PushState = web.Location(url.Values{})

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
			if inerr = tx.Model(&Container{}).Where("id = ? AND locale_code = ?", r.ContainerID, r.Locale).Update("display_order", i+1).Error; inerr != nil {
				return
			}
		}
		return
	})

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) MoveUpDownContainer(ctx *web.EventContext) (r web.EventResponse, err error) {
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

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) ToggleContainerVisibility(ctx *web.EventContext) (r web.EventResponse, err error) {
	var container Container
	paramID := ctx.R.FormValue(paramContainerID)
	cs := container.PrimaryColumnValuesBySlug(paramID)
	containerID := cs["id"]
	locale := cs["locale_code"]

	err = b.db.Exec("UPDATE page_builder_containers SET hidden = NOT(coalesce(hidden,FALSE)) WHERE id = ? AND locale_code = ?", containerID, locale).Error

	r.PushState = web.Location(url.Values{})
	return
}

func (b *Builder) DeleteContainerConfirmation(ctx *web.EventContext) (r web.EventResponse, err error) {
	paramID := ctx.R.FormValue(paramContainerID)

	containerName := ctx.R.FormValue(paramContainerName)

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
							Attr("@click", "dialogLocals.deleteConfirmation = false"),

						VBtn("Delete").
							Color("primary").
							Variant(VariantFlat).
							Theme(ThemeDark).
							Attr("@click", web.Plaid().
								URL(fmt.Sprintf("%s/editors", b.prefix)).
								EventFunc(DeleteContainerEvent).
								Query(paramContainerID, paramID).
								Go()),
					),
				),
			).MaxWidth("600px").
				Attr("v-model", "dialogLocals.deleteConfirmation"),
		).VSlot(`{ locals : dialogLocals }`).Init(`{deleteConfirmation: true}`),
	})

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

func (b *Builder) ContainerComponent(ctx *web.EventContext, isReadonly bool) (component h.HTMLComponent) {

	locale := ctx.R.FormValue(paramLocale)
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	var groups []h.HTMLComponent
	var groupsNames []string
	for _, group := range utils.GroupBySlice[*ContainerBuilder, string](b.containerBuilders, func(builder *ContainerBuilder) string {
		return builder.group
	}) {
		if len(group) == 0 {
			break
		}
		var groupName = group[0].group
		groupsNames = append(groupsNames, groupName)
		var listItems []h.HTMLComponent
		for _, builder := range group {
			cover := builder.cover
			if cover == "" {
				cover = path.Join(b.prefix, b.imagesPrefix, strings.ReplaceAll(builder.name, " ", "")+".png")
			}
			containerName := i18n.T(ctx.R, presets.ModelsI18nModuleKey, builder.name)
			listItems = append(listItems, VListItem(
				h.Div(
					VListItemTitle(h.Text(containerName)),
					VListItemSubtitle(VImg().Src(cover).Height(100).Draggable(false)),
				).Attr("draggable", h.JSONString(!isReadonly), "v-bind", `{ shared : "false", modelid :0 }`).Id(builder.name)),
			)
		}
		groups = append(groups, VListGroup(
			web.Slot(
				VListItem(
					VListItemTitle(h.Text(groupName)),
					// TODO temp bg-color
				).Attr("v-bind", "props").Class("bg-primary"),
			).Name("activator").Scope(" {  props }"),
			h.Components(listItems...),
		).Value(groupName))
	}

	var cons []*Container
	var sharedGroups []h.HTMLComponent
	var sharedGroupNames []string

	b.db.Select("display_name,model_name,model_id").Where("shared = true AND locale_code = ?", locale).Group("display_name,model_name,model_id").Find(&cons)

	for _, group := range utils.GroupBySlice[*Container, string](cons, func(builder *Container) string {
		return b.ContainerByName(builder.ModelName).group
	}) {
		if len(group) == 0 {
			break
		}
		var groupName = b.ContainerByName(group[0].ModelName).group
		sharedGroupNames = append(sharedGroupNames, groupName)
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
					VListItemSubtitle(VImg().Src(cover).Height(100).Draggable(false)),
				).Attr("draggable", "true", "v-bind", fmt.Sprintf(`{ shared : "true",modelid: %v}`, builder.ModelID)).Id(c.name),
			).Value(containerName))
		}

		sharedGroups = append(sharedGroups, VListGroup(
			web.Slot(
				VListItem(
					VListItemTitle(h.Text(groupName)),
					// TODO temp bg-color
				).Attr("v-bind", "props").Class("bg-primary"),
			).Name("activator").Scope(" {  props }"),
			h.Components(listItems...),
		).Value(groupName))

	}
	component = web.Scope(
		VTabs(
			VTab(h.Text(msgr.New)).Value(msgr.New),
			VTab(h.Text(msgr.Shared)).Value(msgr.Shared),
		).Attr("v-model", "tabLocals.tab"),
		VWindow(
			VWindowItem(
				VList(groups...).Opened(groupsNames),
			).Value(msgr.New).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
			VWindowItem(
				VList(sharedGroups...).Opened(sharedGroupNames),
			).Value(msgr.Shared).Attr("style", "overflow-y: scroll; overflow-x: hidden; height: 610px;"),
		).Attr("v-model", "tabLocals.tab"),
	).Init(fmt.Sprintf(`{ tab : %s } `, msgr.New)).VSlot("{locals: tabLocals}")
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
							Variant(VariantText).
							Color("primary").Attr("@click",
							"dialogLocals.addContainerDialog = false;"+web.Plaid().
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
							Variant(VariantText).
							Color("primary").Attr("@click",
							"dialogLocals.addContainerDialog = false;"+web.Plaid().
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
			Query(presets.ParamOverlay, actions.Drawer).
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
			h.Tag("vx-restore-scroll-listener"),
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
		).Attr("id", "vt-app").
			Attr(web.VAssign("vars", `{presetsRightDrawer: false, presetsDialog: false, dialogPortalName: false}`)...)
		return
	}
}

func (b *Builder) ShowAddContainerDrawer(ctx *web.EventContext) (r web.EventResponse, err error) {
	// TODO get status
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{Name: presets.RightDrawerContentPortalName, Body: b.renderContainersList(ctx, false)})
	return
}
