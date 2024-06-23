package basics

import (
	"path"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	"github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var L10n = docgo.Doc(
	docgo.Markdown(`
L10n gives your models the ability to localize for different Locales.  
It can be a catalyst for the adaptation of a product, application, or document content to meet the language, cultural, and other requirements of a specific target market.
    `),
	docgo.Markdown(`
## Define a struct
Define a struct that requires embed ~l10n.Locale~.  
Also this struct must implement ~PrimarySlug() string~ and ~PrimaryColumnValuesBySlug(slug string) map[string]string~.
`),
	ch.Code(generated.L10nModelExample).Language("go"),
	docgo.Markdown(`
## Init a l10n builder
Register locales here.  
You can use ~SupportLocalesFunc~ to determine who can use which locales.
`),
	ch.Code(generated.L10nBuilderExample).Language("go"),
	docgo.Markdown(`
## Configure the model builder
Use ~l10n_view.Configure()~ func to configure l10n view.  
The ~Switch Locale~ ui will appear below the ~Brand~.  
The ~Localize~ ui will appear in the ~RowMenuItem~ under the ~Edit~ and the ~Delete~.  
~Localize~ button is used to copy a piece of data from the current locale to the other locales.
`),
	ch.Code(generated.L10nConfigureExample).Language("go"),
	docgo.Markdown(`
## Full Example
`),
	ch.Code(generated.L10nFullExample).Language("go"),
	utils.DemoWithSnippetLocation(
		"L10n",
		path.Join(examples.URLPathByFunc(examples_admin.LocalizationExample), "/l10n-models"),
		generated.L10nFullExampleLocation,
	),
).Slug("basics/l10n").Title("Localization")
