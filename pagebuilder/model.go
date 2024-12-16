package pagebuilder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	"github.com/sunfmin/reflectutils"
	h "github.com/theplant/htmlgo"
	"gorm.io/gorm"

	"github.com/qor5/admin/v3/l10n"
	"github.com/qor5/admin/v3/presets"
	"github.com/qor5/admin/v3/presets/gorm2op"
	"github.com/qor5/admin/v3/publish"
	"github.com/qor5/admin/v3/utils"
)

type (
	eventMiddlewareFunc func(in web.EventFunc) web.EventFunc

	ModelBuilder struct {
		name            string
		mb              *presets.ModelBuilder
		editor          *presets.ModelBuilder
		db              *gorm.DB
		builder         *Builder
		preview         http.Handler
		tb              *TemplateBuilder
		eventMiddleware eventMiddlewareFunc
	}
)

func (b *ModelBuilder) editorURLWithSlug(ps string) string {
	return fmt.Sprintf("%s/%s", b.editorURL(), ps)
}

func (b *ModelBuilder) editorURL() string {
	return fmt.Sprintf("%s/%s", b.builder.prefix, b.editor.Info().URIName())
}

func (b *ModelBuilder) getContainerBuilders() (cons []*ContainerBuilder) {
	pageObjName := utils.GetObjectName(&Page{})
	for _, builder := range b.builder.containerBuilders {
		if builder.onlyPages {
			if b.name == pageObjName || (b.tb != nil && pageObjName == utils.GetObjectName(b.tb.mb.NewModel())) {
				cons = append(cons, builder)
			}
		} else {
			if builder.modelBuilder == nil || (b.tb == nil && b.mb == builder.modelBuilder) || (b.tb != nil && b.tb.mb == builder.modelBuilder) {
				cons = append(cons, builder)
			}
		}
	}
	return
}

func (b *ModelBuilder) setName() {
	b.name = utils.GetObjectName(b.mb.NewModel())
}

func (b *ModelBuilder) addSharedContainerToPage(pageID int, containerID, pageVersion, locale, modelName string, modelID uint) (newContainerID string, err error) {
	var c Container

	err = b.db.Transaction(func(tx *gorm.DB) (dbErr error) {
		if dbErr = tx.First(&c, "model_name = ? AND model_id = ? AND shared = true and page_model_name = ? ", modelName, modelID, b.name).Error; dbErr != nil {
			return
		}

		var (
			maxOrder     sql.NullFloat64
			displayOrder float64
		)
		dbErr = tx.Model(&Container{}).Select("MAX(display_order)").
			Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ?", pageID, pageVersion, locale, b.name).Scan(&maxOrder).Error
		if dbErr != nil {
			return
		}
		dbErr = tx.Model(&Container{}).Select("MAX(display_order)").
			Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ? ", pageID, pageVersion, locale, b.name).Scan(&maxOrder).Error
		if dbErr != nil {
			return
		}
		if containerID != "" {
			var lastContainer Container
			cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
			tx.Where("id = ? AND locale_code = ? and page_model_name = ? ", cs["id"], locale, b.name).First(&lastContainer)
			if lastContainer.ID > 0 {
				displayOrder = lastContainer.DisplayOrder
				if dbErr = tx.Model(&Container{}).Where("page_id = ? and page_version = ? and locale_code = ? and page_model_name = ? and display_order > ? ", pageID, pageVersion, locale, b.name, displayOrder).
					UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; dbErr != nil {
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
		if dbErr = tx.Create(&container).Error; dbErr != nil {
			return
		}
		newContainerID = container.PrimarySlug()
		return
	})
	return
}

func withLocale(builder *Builder, wh *gorm.DB, locale string) *gorm.DB {
	if builder.l10n == nil {
		return wh
	}
	return wh.Where("locale_code = ?", locale)
}

func (b *ModelBuilder) addContainerToPage(ctx *web.EventContext, pageID int, containerID, pageVersion, locale, modelName string) (modelID uint, newContainerID string, err error) {
	var (
		dc          DemoContainer
		containerMb = b.builder.ContainerByName(modelName)
		model       = containerMb.NewModel()
	)

	err = b.db.Transaction(func(tx *gorm.DB) (dbErr error) {
		tx.Where("model_name = ? AND locale_code = ?", modelName, locale).First(&dc)
		if dc.ID != 0 && dc.ModelID != 0 {
			tx.Where("id = ?", dc.ModelID).First(model)
			_ = reflectutils.Set(model, "ID", uint(0))
		}
		ctx.WithContextValue(gorm2op.CtxKeyDB{}, tx)
		defer ctx.WithContextValue(gorm2op.CtxKeyDB{}, nil)
		if dbErr = containerMb.Editing().Creating().Saver(model, "", ctx); dbErr != nil {
			return
		}

		var (
			maxOrder     sql.NullFloat64
			displayOrder float64
		)
		wh := tx.Model(&Container{}).Select("MAX(display_order)").
			Where("page_id = ? and page_version = ? and page_model_name = ? ", pageID, pageVersion, b.name)

		if dbErr = withLocale(b.builder, wh, locale).Scan(&maxOrder).Error; dbErr != nil {
			return
		}
		if containerID != "" {
			var lastContainer Container
			cs := lastContainer.PrimaryColumnValuesBySlug(containerID)
			tx.Where("id = ? AND locale_code = ? and page_model_name = ?", cs["id"], locale, b.name).First(&lastContainer)
			if lastContainer.ID > 0 {
				displayOrder = lastContainer.DisplayOrder
				if dbErr = withLocale(
					b.builder,
					tx.Model(&Container{}).
						Where("page_id = ? and page_version = ? and page_model_name = ? and display_order > ? ", pageID, pageVersion, b.name, displayOrder),
					locale,
				).
					UpdateColumn("display_order", gorm.Expr("display_order + ? ", 1)).Error; dbErr != nil {
					return
				}
			}

		} else {
			displayOrder = maxOrder.Float64
		}
		modelID = reflectutils.MustGet(model, "ID").(uint)
		displayName := modelName
		if b.builder.ps.GetI18n() != nil {
			displayName = i18n.T(ctx.R, presets.ModelsI18nModuleKey, modelName)
		}
		container := Container{
			PageID:        uint(pageID),
			PageVersion:   pageVersion,
			ModelName:     modelName,
			PageModelName: b.name,
			DisplayName:   displayName,
			ModelID:       modelID,
			DisplayOrder:  displayOrder + 1,
			Locale: l10n.Locale{
				LocaleCode: locale,
			},
		}
		err = tx.Create(&container).Error
		newContainerID = container.PrimarySlug()

		return
	})

	return
}

func (b *ModelBuilder) pageContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	var body h.HTMLComponent
	if body, err = b.renderPageOrTemplate(ctx, true, true, false); err != nil {
		return
	}
	r.Body = web.Portal(
		body,
	).Name(editorPreviewContentPortal)
	return
}

func (b *ModelBuilder) getPrimaryColumnValuesBySlug(ctx *web.EventContext) (pageID int, pageVersion string, locale string) {
	return b.primaryColumnValuesBySlug(ctx.Param(presets.ParamID))
}

func (b *ModelBuilder) primaryColumnValuesBySlug(slug string) (pageID int, pageVersion string, locale string) {
	var (
		ps map[string]string

		obj = b.mb.NewModel()
	)
	if p, ok := obj.(presets.SlugDecoder); ok {
		ps = p.PrimaryColumnValuesBySlug(slug)
	}
	pageVersion = ps[publish.SlugVersion]
	locale = ps[l10n.SlugLocaleCode]
	pageIDi, _ := strconv.ParseInt(ps["id"], 10, 64)
	pageID = int(pageIDi)
	return
}

func (b *ModelBuilder) renderPageOrTemplate(ctx *web.EventContext, isEditor, isIframe, isReloadBody bool) (r h.HTMLComponent, err error) {
	var (
		status                      = publish.StatusDraft
		obj                         interface{}
		pageID, pageVersion, locale = b.getPrimaryColumnValuesBySlug(ctx)
	)
	if pageID == 0 {
		return nil, nil
	}
	if obj, err = b.pageBuilderModel(ctx); err != nil {
		return
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
	if !isReadonly && isEditor && b.mb.Info().Verifier().Do(presets.PermUpdate).WithReq(ctx.R).IsAllowed() != nil {
		isReadonly = true
	}
	var comps []h.HTMLComponent
	comps, err = b.renderContainers(ctx, obj, pageID, pageVersion, locale, isEditor, isReadonly)
	if err != nil {
		return
	}
	r = b.rendering(comps, ctx, obj, locale, isEditor, isReadonly, isIframe, isReloadBody)
	return
}

func (b *ModelBuilder) renderScrollIframe(comps []h.HTMLComponent, ctx *web.EventContext, obj interface{}, locale string, isEditor, isIframe, isReloadBody bool) (r h.HTMLComponent) {
	r = h.Components(comps...)
	if b.builder.pageLayoutFunc == nil {
		return
	}
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
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
				display: table; /* for IE or lower versions of browers */
				display: flow-root;/* for morden browsers*/
			  position: relative;
			  width: 100%;	
			}

			.inner-shadow {
				pointer-events: none;
			  position: absolute;
			  width: 100%;
			  height: 100%;
			  opacity: 0;
			  top: 0;
			  left: 0;
			  box-shadow: 2px 2px 0 0px #3E63DD inset, -2px 2px 0 0px #3E63DD inset,2px -2px 0 0px #3E63DD inset;
				z-index:201;
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
			  height: 2px;
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
              border: 0;
              padding: 4px 0 4px 0;
			}
			.wrapper-shadow:hover {
			  cursor: pointer;
			}

			.wrapper-shadow:hover .editor-add {
			  opacity: 1;
			}
			
			.wrapper-shadow:hover .editor-add div {
			  height: 4px;
			}
			.highlight .editor-add div{
              height: 2px !important;	
			}		
			.editor-bar {
			  position: absolute;
			  z-index: 9999;
			  height: 32px;	
              width: 207px;
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
              border: 0;
              padding: 0;
			  cursor: pointer;
			  background-color: #3E63DD; 
              height: 24px;	
			}
			
			.editor-bar .title {
			  color: #FFFFFF;
			  overflow: hidden;	
			  font-size: 12px;
			  font-style: normal;
			  max-width: calc(100% - 88px);
			  font-weight: 400;
			  line-height: 16px; 
              text-overflow: ellipsis;
              white-space: nowrap;
			  letter-spacing: 0.04px;	
			}
			.highlight .editor-bar {
			  opacity: 1;
              pointer-events: auto;
			}
		
			.highlight .inner-shadow {
			  opacity: 1;
			}
`))
	}
	if f := ctx.R.Context().Value(CtxKeyContainerToPageLayout{}); f != nil {
		pl, ok := f.(*PageLayoutInput)
		if ok {
			input.FreeStyleCss = append(input.FreeStyleCss, pl.FreeStyleCss...)
			input.FreeStyleTopJs = append(input.FreeStyleTopJs, pl.FreeStyleTopJs...)
			input.FreeStyleBottomJs = append(input.FreeStyleBottomJs, pl.FreeStyleBottomJs...)
			input.Hreflang = pl.Hreflang
		}
	}

	if isIframe {
		// use newCtx to avoid inserting page head to head outside of iframe
		newCtx := &web.EventContext{
			R:        ctx.R,
			Injector: &web.PageInjector{},
		}
		r = b.builder.pageLayoutFunc(h.Components(comps...), input, newCtx)
		newCtx.Injector.HeadHTMLComponent("style", b.builder.pageStyle, true)
		body := h.Components(h.Div(
			r,
		).Id("app").Attr("v-cloak", true),
			newCtx.Injector.GetTailHTMLComponent(),
		)
		if isReloadBody {
			return body
		}
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

		scrollIframe := vx.VXScrollIframe().
			BackgroundColor(b.builder.editorBackgroundColor).
			Srcdoc(h.MustString(r, ctx.R.Context())).
			Attr("@load", "vars.__pageBuilderAddContainerBtnDisabled=false").
			Attr("v-on-mounted", fmt.Sprintf(`({el,window}) => {
							vars.__pageBuilderAddContainerBtnDisabled = true;
						}`)).
			Width(width).Attr("ref", "scrollIframe").VirtualElementText(msgr.NewContainer)
		if isEditor {
			scrollIframe.Attr(web.VAssign("vars", `{el:$}`)...)
			r = h.Components(
				scrollIframe,
				web.Listen(b.notifIframeBodyUpdated(),
					`vars.el.refs.scrollIframe.updateIframeBody(payload)`,
				),
			)
		}

	} else {
		r = b.builder.pageLayoutFunc(h.Components(comps...), input, ctx)
		ctx.Injector.HeadHTMLComponent("style", b.builder.pageStyle, true)
	}
	return
}

func (b *ModelBuilder) rendering(comps []h.HTMLComponent, ctx *web.EventContext, obj interface{}, locale string, isEditor, isReadonly, isIframe, isReloadBody bool) (r h.HTMLComponent) {
	msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)

	r = b.renderScrollIframe(comps, ctx, obj, locale, isEditor, isIframe, isReloadBody)
	if isReadonly || !isEditor || isReloadBody {
		return r
	}
	var (
		title string
		svg   string
	)
	if b.tb == nil {
		title = msgr.StartBuildingMsg
		svg = previewEmptySvg
	} else {
		title = msgr.StartBuildingTemplateMsg
		svg = previewTemplateEmptySvg
	}
	return h.Components(
		h.Div(
			VCard(
				VCardText(h.RawHTML(svg)).Class("d-flex justify-center"),
				VCardTitle(h.Text(title)).Class("d-flex justify-center"),
				VCardSubtitle(h.Text(msgr.StartBuildingSubMsg)).Class("d-flex justify-center"),
				VCardActions(
					VBtn(msgr.AddContainer).Color(ColorPrimary).Variant(VariantElevated).
						Attr("@click", appendVirtualElement()+"vars.overlay=true;vars.el.refs.overlay.showCenter()"),
				).Class("d-flex justify-center"),
			).Flat(true),
		).Attr("v-show", "vars.emptyIframe").
			Attr(web.VAssign("vars", fmt.Sprintf(`{emptyIframe: %v}`, len(comps) == 0))...).
			Style("display:flex;justify-content:center;align-items:center;flex-direction:column;height:80vh"),
		h.Div(r).Attr("v-show", "!vars.emptyIframe"),
	)
}

func (b *ModelBuilder) renderContainers(ctx *web.EventContext, obj interface{}, pageID int, pageVersion string, locale string, isEditor bool, isReadonly bool) (r []h.HTMLComponent, err error) {
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
		containerObj := ec.builder.NewModel()
		err = b.db.FirstOrCreate(containerObj, "id = ?", ec.container.ModelID).Error
		if err != nil {
			return
		}
		input := RenderInput{
			IsEditor:    isEditor,
			IsReadonly:  isReadonly,
			Device:      device,
			ContainerId: ec.container.PrimarySlug(),
			DisplayName: ec.container.DisplayName,
			Obj:         obj,
		}
		pure := ec.builder.renderFunc(containerObj, &input, ctx)

		r = append(r, b.builder.containerWrapper(pure.(*h.HTMLTagBuilder), ctx, isEditor, isReadonly, i == 0, i == len(cbs)-1,
			ec.builder.getContainerDataID(int(ec.container.ModelID)), ec.container.ModelName, &input))
	}

	return
}

func (b *ModelBuilder) renderPreviewContainer(ctx *web.EventContext, obj interface{}, locale string, isEditor, IsReadonly bool) (r h.HTMLComponent, err error) {
	var (
		modelName       = ctx.Param(paramModelName)
		sharedContainer = ctx.Param(paramSharedContainer)
		modelID         = ctx.ParamAsInt(paramModelID)
	)
	containerBuilder := b.builder.ContainerByName(modelName)

	if sharedContainer != "true" || modelID == 0 {
		var con *DemoContainer
		err = withLocale(
			b.builder,
			b.db.
				Where("model_name = ?", modelName),
			locale,
		).
			First(&con).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			b.builder.firstOrCreateDemoContainers(ctx, containerBuilder)
			err = withLocale(
				b.builder,
				b.db.
					Where("model_name = ?", modelName),
				locale,
			).
				First(&con).Error
		}
		if err != nil {
			return
		}
		modelID = int(con.ModelID)
	}

	device, _ := b.builder.getDevice(ctx)

	input := RenderInput{
		IsEditor:    isEditor,
		IsReadonly:  IsReadonly,
		Device:      device,
		ContainerId: "",
		DisplayName: modelName,
		Obj:         obj,
	}
	containerObj := containerBuilder.NewModel()
	err = b.db.FirstOrCreate(containerObj, "id = ?", modelID).Error
	if err != nil {
		return
	}
	pure := containerBuilder.renderFunc(containerObj, &input, ctx)
	r = b.builder.containerWrapper(pure.(*h.HTMLTagBuilder), ctx, isEditor, IsReadonly, false, false,
		containerBuilder.getContainerDataID(modelID), modelName, &input)
	return
}

func (b *ModelBuilder) previewContent(ctx *web.EventContext) (r web.PageResponse, err error) {
	var obj interface{}

	r.Body, err = b.renderPageOrTemplate(ctx, false, false, false)
	if err != nil {
		return
	}
	if obj, err = b.pageBuilderModel(ctx); err != nil {
		return
	}
	if b.builder.seoBuilder != nil && b.builder.seoBuilder.GetSEO(obj) != nil {
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

func (b *ModelBuilder) copyContainersToNewPageVersion(db *gorm.DB, pageID int, locale, oldPageVersion, newPageVersion, fromModelName, toModelName string) (err error) {
	return b.copyContainersToAnotherPage(db, pageID, oldPageVersion, locale, pageID, newPageVersion, locale, fromModelName, toModelName)
}

func (b *ModelBuilder) copyContainersToAnotherPage(db *gorm.DB, pageID int, pageVersion, locale string, toPageID int, toPageVersion, toPageLocale, fromModelName, toModelName string) (err error) {
	var cons []*Container
	err = db.Order("display_order ASC").Find(&cons, "page_id = ? AND page_version = ? AND locale_code = ? and page_model_name =? ", pageID, pageVersion, locale, fromModelName).Error
	if err != nil {
		return
	}
	buildeContainer := b.getContainerBuilders()
	for _, c := range cons {
		if !slices.ContainsFunc(buildeContainer, func(builder *ContainerBuilder) bool {
			return c.ModelName == builder.name
		}) {
			continue
		}
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
			PageModelName: toModelName,
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
					if inerr = b.copyContainersToNewPageVersion(tx, pageID, localeCode, parentVersion, version, b.name, b.name); inerr != nil {
						return
					}
					return
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
	p, ok := obj.(presets.SlugEncoder)
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

func (b *ModelBuilder) newContainerContent(ctx *web.EventContext) h.HTMLComponent {
	var (
		containers = b.renderContainersList(ctx)
		msgr       = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		title      string
		svg        string
	)
	if b.tb == nil {
		title = msgr.BuildYourPages
		svg = previewEmptySvg
	} else {
		title = msgr.BuildYourTemplates
		svg = previewTemplateEmptySvg

	}
	emptyContent := VCard(
		VCardText(h.RawHTML(svg)).Class("d-flex justify-center"),
		VCardTitle(h.Text(title)).Class("d-flex justify-center"),
		VCardSubtitle(h.Text(msgr.PlaceAnElementFromLibrary)).Class("d-flex justify-center"),
	).Flat(true).Tile(true).Color(ColorGreyLighten3)
	return VSheet(
		VSheet(
			VCard(
				VCardTitle(h.Text(msgr.NewElement)),
				VCardText(containers),
			).Elevation(0),
		).Class(W50).Class("pa-4", "overflow-y-auto"),
		VSheet(
			h.Div(
				VSpacer(),
				VBtn("").Icon("mdi-close").Variant(VariantText).Attr("@click", "vars.overlay=false"),
			).Class("d-flex justify-end").Style("height:40px"),
			VContainer(
				VRow(
					VCol(
						emptyContent.Attr("v-if", "!vars.containerPreview"),
						VSheet(web.Portal().Name(addContainerDialogContentPortal)).Tile(true).Attr("v-if", "vars.containerPreview"),
					),
				).Align(Center).Justify(Center).Attr("style", "height:420px"),
			).Class(W100, "py-0"),
		).Class(W50).Color(ColorGreyLighten3),
	).Class("d-inline-flex").Width(665).Height(460)
}

func (b *ModelBuilder) EventMiddleware(v eventMiddlewareFunc) *ModelBuilder {
	b.eventMiddleware = v
	return b
}

func (b *ModelBuilder) WrapEventMiddleware(w func(eventMiddlewareFunc) eventMiddlewareFunc) *ModelBuilder {
	b.eventMiddleware = w(b.eventMiddleware)
	return b
}
