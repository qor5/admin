package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var VariantSubForm = Doc(
	Markdown(`
VSelect changes, the form below it will change to a new form accordingly.

By use of ~web.Portal()~ and ~VSelect~'s ~OnInput~
`),
	ch.Code(generated.VuetifyVariantSubForm).Language("go"),
	utils.DemoWithSnippetLocation("Vuetify Variant Sub Form", examples_vuetify.VuetifyVariantSubFormPath, generated.VuetifyVariantSubFormLocation),
).Title("Variant Sub Form").
	Slug("vuetify-components/variant-sub-form")
