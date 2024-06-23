package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var TheGoHTMLBuilder = Doc(
	Markdown(`
Like at the beginning we said, That we don't use interpreted template language (eg go html/template)
to generate html page. We think they are:

- error prone without static type enforcing
- hard to refactor
- difficult to abstract out to component
- yet another tedious syntax to learn
- not flexible to use helper functions

We like to use standard Go code. the library [htmlgo](https://github.com/theplant/htmlgo) is just for that.

Although Go can't do flexible builder syntax like [Kotlin](https://kotlinlang.org/docs/reference/type-safe-builders.html) does,
But it can also do quite well.

Consider the following code:
`),
	ch.Code(generated.TypeSafeBuilderSample).Language("go"),
	Markdown(`
It's basically assembled what Kotlin can do, Also is legitimate Go code.
`),
	utils.DemoWithSnippetLocation("The Go HTML Builder", examples_web.TypeSafeBuilderSamplePath, generated.TypeSafeBuilderSampleLocation),
).Title("The Go HTML builder").
	Slug("advanced-functions/the-go-html-builder")
