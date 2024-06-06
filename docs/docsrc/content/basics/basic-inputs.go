package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var BasicInputs = Doc(
	Markdown(`
Vuetify provides many form basic inputs, and also with error messages display on fields.

Here is one example:
`),
	ch.Code(generated.VuetifyBasicInputsSample).Language("go"),
	utils.DemoWithSnippetLocation("Vuetify Basic Inputs", examples_vuetify.VuetifyBasicInputsPath, generated.VuetifyBasicInputsSampleLocation),
).Title("Basic Inputs").
	Slug("vuetify-components/basic-inputs")
