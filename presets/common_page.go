package presets

import (
	. "github.com/qor5/ui/v3/vuetify"
	"github.com/qor5/web/v3"
	h "github.com/theplant/htmlgo"
)

type commonPageConfig struct {
	// TODO it should be create in defaultToPage
	formContent h.HTMLComponent

	tabPanels []TabComponentFunc
	sidePanel ObjectComponentFunc
}

// TODO set common component which in editingBuilder or DetailingBuilder
// TODO defaultToPage build a common page
func defaultToPage(config commonPageConfig, obj interface{}, ctx *web.EventContext) h.HTMLComponent {
	msgr := MustGetMessages(ctx.R)

	var asideContent h.HTMLComponent = config.formContent

	if len(config.tabPanels) != 0 {
		var tabs []h.HTMLComponent
		var contents []h.HTMLComponent
		for _, panelFunc := range config.tabPanels {
			tab, content := panelFunc(obj, ctx)
			if tab != nil {
				tabs = append(tabs, tab)
				contents = append(contents, content)
			}
		}
		if len(tabs) == 0 {
			asideContent = web.Scope(config.formContent).VSlot("{ form }")
		} else {
			asideContent = web.Scope(
				VTabs(
					VTab(h.Text(msgr.FormTitle)).Value("default"),
					h.Components(tabs...),
				).Class("v-tabs--fixed-tabs").Attr("v-model", "tab"),

				VTabsWindow(
					web.Scope(config.formContent).VSlot("{ form }"),
					h.Components(contents...),
				).Attr("v-model", "tab"),
			).VSlot("{ locals }").Init(`{tab: 'default'}`)
		}
	}

	if config.sidePanel != nil {
		sidePanel := config.sidePanel(obj, ctx)
		if sidePanel != nil {
			asideContent = VContainer(
				VRow(
					VCol(asideContent).Cols(8),
					VCol(sidePanel).Cols(4),
				),
			)
		}
	}
	return asideContent
}
