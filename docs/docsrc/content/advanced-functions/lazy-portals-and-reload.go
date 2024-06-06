package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var LazyPortalsAndReload = Doc(
	Markdown(`
Use ~web.Portal().Loader(web.POST().EventFunc("menuItems")).Name("menuContent")~ to put a portal place holder inside a part of html, and it will load specified event func's response body inside the place holder after the main page is rendered in a separate AJAX request. Later in an event func, you could also use ~r.ReloadPortals = []string{"menuContent"}~ to reload the portal.
`),
	ch.Code(generated.LazyPortalsAndReloadSample).Language("go"),
	utils.DemoWithSnippetLocation("Lazy Portals", examples_vuetify.LazyPortalsAndReloadPath, generated.LazyPortalsAndReloadSampleLocation),
).Title("Lazy Portals").
	Slug("vuetify-components/lazy-portals")
