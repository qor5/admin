package pagebuilder

import (
	"fmt"

	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	vx "github.com/qor5/x/v3/ui/vuetifyx"
	h "github.com/theplant/htmlgo"

	"github.com/qor5/admin/v3/presets"
)

const (
	PageTemplateSelectionFiled = "TemplateSelection"

	OpenTemplateDialogEvent     = "page_builder_template_OpenTemplateDialogEvent"
	ReloadSelectedTemplateEvent = "page_builder_template_ReloadSelectedTemplateEvent"
	ReloadTemplateContentEvent  = "page_builder_template_ReloadTemplateContentEvent"

	PageTemplatePortalName         = "PageTemplatePortalName"
	TemplateSelectDialogPortalName = "TemplateSelectDialogPortalName"
	TemplateSelectedPortalName     = "TemplateSelectedPortalName"

	templateDialogHeight = 620
	templateDialogWidth  = 700
)

func (b *TemplateBuilder) registerFunctions(mb *presets.ModelBuilder) {
	mb.RegisterEventFunc(ReloadTemplateContentEvent, b.reloadTemplateContent(mb))
	mb.RegisterEventFunc(OpenTemplateDialogEvent, b.openTemplateDialog(mb))
	mb.RegisterEventFunc(ReloadSelectedTemplateEvent, b.reloadSelectedTemplate)
}

func (b *TemplateBuilder) reloadTemplateContent(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: PageTemplatePortalName,
			Body: b.templateContent(ctx, mb),
		})
		return
	}
}

func (b *TemplateBuilder) openTemplateDialog(mb *presets.ModelBuilder) web.EventFunc {
	return func(ctx *web.EventContext) (r web.EventResponse, err error) {
		msgr := i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
		r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
			Name: TemplateSelectDialogPortalName,
			Body: web.Scope(
				vx.VXDialog(
					h.Div().Class("overflow-y-auto").Children(
						web.Portal(b.templateContent(ctx, mb)).Name(PageTemplatePortalName),
					).Style(fmt.Sprintf("height:%vpx", templateDialogHeight)),
				).Width(templateDialogWidth).
					Title(msgr.CreateFromTemplate).
					HideFooter(true).
					Attr("v-model", "vars.pageBuilderSelectTemplateDialog"),
			).VSlot("{ form }"),
		})
		r.RunScript = "setTimeout(function(){ vars.pageBuilderSelectTemplateDialog = true }, 100)"
		return
	}
}

func (b *TemplateBuilder) reloadSelectedTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectedPortalName,
		Body: b.selectedTemplate(ctx),
	})
	return
}
