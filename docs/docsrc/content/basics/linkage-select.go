package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetifyx"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var LinkageSelect = Doc(
	Markdown(`
LinkageSelect is a component for multi-level linkage select.
    `),
	ch.Code(generated.VuetifyComponentsLinkageSelect).Language("go"),
	utils.DemoWithSnippetLocation("Vuetify LinkageSelect", examples_vuetifyx.VuetifyComponentsLinkageSelectPath, generated.VuetifyComponentsLinkageSelectLocation),
	Markdown(`
### Filter intergation
    `),
	ch.Code(generated.LinkageSelectFilterItem).Language("go"),
	utils.DemoWithSnippetLocation("LinkageSelect Filter Item", examples.URLPathByFunc(examples_presets.PresetsLinkageSelectFilterItem)+"/addresses", generated.LinkageSelectFilterItemLocation),
).Title("Linkage Select").
	Slug("vuetify-components/linkage-select")
