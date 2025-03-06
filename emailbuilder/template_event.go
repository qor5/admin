package emailbuilder

import (
	"github.com/qor5/web/v3"
)

func (mb *ModelBuilder) registerFunctions() {
	mb.mb.RegisterEventFunc(ReloadSelectedTemplateEvent, mb.reloadSelectedTemplate)
}

func (mb *ModelBuilder) reloadSelectedTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectedPortalName,
		Body: mb.selectedTemplate(ctx),
	})
	return
}
