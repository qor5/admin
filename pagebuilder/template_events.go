package pagebuilder

import (
	"github.com/qor5/web/v3"
)

const (
	PageTemplateSelectionFiled = "TemplateSelection"

	ReloadSelectedTemplateEvent = "page_builder_template_ReloadSelectedTemplateEvent"

	TemplateSelectDialogPortalName = "TemplateSelectDialogPortalName"
	TemplateSelectedPortalName     = "TemplateSelectedPortalName"

	templateDialogWidth = "700"
)

func (b *TemplateBuilder) registerFunctions() {
	b.mb.RegisterEventFunc(ReloadSelectedTemplateEvent, b.reloadSelectedTemplate)
}

func (b *TemplateBuilder) reloadSelectedTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectedPortalName,
		Body: b.selectedTemplate(ctx),
	})
	return
}
