package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var SummaryOfEventResponse = Doc(
	Markdown(`
The behaviour of ~web.EventFunc~ is controlled by it's return type ~web.EventResponse~
`),
	ch.Code(generated.EventResponseDefinition).Language("go"),
	Markdown(`
- ~PageTitle~ set the html head title, It not only set when render html page directly which is
  request the url directly from the browser. Also use javascript to set the page title when you do
  push state AJAX request to load the page
- ~Body~ is the set to ~web.PageResponse~'s body when ~Reload = true~ is set, Or set to the partial
  html component when using ~ReloadPortals~ together with ~web.Portal().EventFunc("related")~
- ~Reload~ is to reload the ~web.PageFunc~, before reload, you can set ~ctx.Flash~ object to let the
  event func render the page differently (flash message, validation errors, etc)
- ~Location~ is to change the browser url with push state, and AJAX load the page of that url
- ~RedirectURL~ is to change the browser url without AJAX, reload the whole page html includes it's
  head script, css assets
- ~ReloadPortals~ is for reload the portal that uses ~web.Portal().EventFunc("related")~
- ~UpdatePortals~ update the portal specified by the name ~web.Portal().Name("hello")~, ~pu.AfterLoaded~
  set the javascript function that execute after the portal is updated, for example:
  ~VarsScript: "setTimeout(function(){ comp.vars.drawer2 = true }, 100)"~
- ~Data~ is for any AJAX call that want pure JSON, you can set ~er.Data = myobj~ to any object that
  will marshals to JSON, and on the client side use javascript to utilize them
`),
).Title("Summary of Event Response").
	Slug("basics/summary-of-event-response")
