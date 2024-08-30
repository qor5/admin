package pagebuilder

import (
	"github.com/qor5/web/v3"
	"github.com/qor5/x/v3/i18n"
	. "github.com/qor5/x/v3/ui/vuetify"
	h "github.com/theplant/htmlgo"
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
	templateDialogWidth  = 784
)

func (b *TemplateBuilder) registerFunctions() {
	b.model.mb.RegisterEventFunc(ReloadTemplateContentEvent, b.reloadTemplateContent)
	b.mb.RegisterEventFunc(OpenTemplateDialogEvent, b.openTemplateDialog)
	b.mb.RegisterEventFunc(ReloadSelectedTemplateEvent, b.reloadSelectedTemplate)

}
func (b *TemplateBuilder) reloadTemplateContent(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: PageTemplatePortalName,
		Body: b.templateContent(ctx),
	})
	return
}

func (b *TemplateBuilder) openTemplateDialog(ctx *web.EventContext) (r web.EventResponse, err error) {
	var (
		msgr = i18n.MustGetModuleMessages(ctx.R, I18nPageBuilderKey, Messages_en_US).(*Messages)
	)
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectDialogPortalName,
		Body: web.Scope(
			VDialog(
				VCard(
					VCardTitle(
						h.Span(msgr.CreateFromTemplate),
						VBtn("").Icon("mdi-close").Variant(VariantText).Attr("@click", "vars.pageBuilderSelectTemplateDialog=false"),
					).Class(W100, "d-flex justify-space-between"),
					h.Div().Class("overflow-y-auto").Children(
						web.Portal(b.templateContent(ctx)).Name(PageTemplatePortalName),
					),
				).Height(templateDialogHeight),
			).Width(templateDialogWidth).
				Attr("v-model", "vars.pageBuilderSelectTemplateDialog"),
		).VSlot("{ form }"),
	})
	r.RunScript = "setTimeout(function(){ vars.pageBuilderSelectTemplateDialog = true }, 100)"
	return
}
func (b *TemplateBuilder) reloadSelectedTemplate(ctx *web.EventContext) (r web.EventResponse, err error) {
	r.UpdatePortals = append(r.UpdatePortals, &web.PortalUpdate{
		Name: TemplateSelectedPortalName,
		Body: b.selectedTemplate(ctx),
	})
	return
}
