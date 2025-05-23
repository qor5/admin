package pagebuilder

import (
	"github.com/qor5/web/v3"

	"github.com/qor5/admin/v3/presets"
)

const (
	PageTemplateSelectionFiled = "TemplateSelection"

	ReloadSelectedTemplateEvent = "page_builder_template_ReloadSelectedTemplateEvent"

	TemplateSelectDialogPortalName = "TemplateSelectDialogPortalName"
	TemplateSelectedPortalName     = "TemplateSelectedPortalName"

	templateDialogWidth = "700"
)

func (b *TemplateBuilder) registerFunctions(mb *presets.ModelBuilder) {
	mb.RegisterEventFunc(ReloadSelectedTemplateEvent, b.reloadSelectedTemplate)
}

func (b *TemplateBuilder) reloadSelectedTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectedPortalName,
		Body: b.selectedTemplate(ctx),
	})
	return
}
