package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var ATasteOfUsingVuetifyInGo = Doc(
	Markdown(`
[Vuetify](https://vuetifyjs.com/en/) is a really mature Vue components library for
[Material Design](https://material.io/design/). We have made the efforts to
integrate most all of it as a go package. You can use it with ease just like any
other go package.
`),
	utils.Anchor(H2(""), "Use container, toolbar, list, list item etc"),
	Markdown(`
This example is purely render, we didn't integrate any interaction (event func) to it.
`),
	ch.Code(generated.VuetifyListSample).Language("go"),
	utils.DemoWithSnippetLocation("Vuetify List", examples_vuetify.HelloVuetifyListPath, generated.VuetifyListSampleLocation),

	utils.Anchor(H2(""), "Use menu, card, list, etc"),
	Markdown(`
This example uses the menu popup, card, list component. and some interactions of clicking
buttons on the menu popup.
`),
	ch.Code(generated.VuetifyMenuSample).Language("go"),
	Markdown(`
~.VSlot("{ locals, form }").Init("{ myMenuShow: false }").FormInit(JSONString(fv))~ used to initialize the variables of ~web.Scope~.
~locals~ corresponds to the ~Init~ method and ~form~ corresponds to the ~FormInit~ method.

~.Init~ will initialize ~locals.myMenuShow~ to ~false~. So that you don't need to modify javascript code to do
the initialization. It's often useful to control dialog, popups. At this example,
We add it, So that the cancel button on the menu, could actually close the menu without
requesting server backend.

~toggleFavored~ event func did an partial update only to the favorite icon button. So that it won't close the
menu popup, but updated the button to toggle the favorite icon.
`),
	utils.DemoWithSnippetLocation("Vuetify Menu", examples_vuetify.HelloVuetifyMenuPath, generated.VuetifyMenuSampleLocation),
).Title("A Taste of using Vuetify in Go").
	Slug("vuetify-components/a-taste-of-using-vuetify-in-go")
