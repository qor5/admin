package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var PartialRefreshWithPortal = Doc(
	Markdown(`
As said before, The results of an ~web.EventFunc~ could be:

- Go to a new page
- Reload the whole current page
- Refresh part of the current page

We have covered two. Now let's demonstrate refresh part of the current page:
`),
	ch.Code(generated.PartialUpdateSample).Language("go"),
	utils.DemoWithSnippetLocation("Partial Update", examples_web.PartialUpdatePagePath, generated.PartialUpdateSampleLocation),
	Markdown(`
~web.Portal().Name("part1")~ Place a placeholder inside you page, and append ~web.PortalUpdate~ to ~er.UpdatePortals~ to update the portal with that name.
Multiple portal can be updated at the same time.
`),
	utils.Anchor(H2(""), "Load Portal in separate AJAX request"),
	Markdown(`
With ~web.Portal~, We can also load the portal with a separate AJAX request after page load.
It is useful for the type of the content is not that important to the page, But load them are
quite heavy. Like related products of a product detail page of a ECommerce site.
`),
	ch.Code(generated.PartialReloadSample).Language("go"),
	utils.DemoWithSnippetLocation("Partial Reload", examples_web.PartialReloadPagePath, generated.PartialReloadSampleLocation),
	Markdown(`
It is not only load the portal in separate AJAX request, Also you can reload it with ease ~er.ReloadPortals = []string{"related_products"}~ in an event func.

Under the hood, We use Vue's [Dynamic & Async Components](https://vuejs.org/v2/guide/components-dynamic-async.html), to load Go generated html (vue runtime templates)
from the server and mount those vue components into the page. It works the same way for reload the whole page, push state page switch, and refresh part of the current page.
`),
).Title("Partial Refresh with Portal").
	Slug("basics/partial-refresh-with-portal")
