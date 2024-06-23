package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var ManipulatePageURLInEventFunc = Doc(
	Markdown(`
Encode page state into query strings in url is useful. because user can paste the link to another person,
That can open the page to the exact state of the page being sent, Not the initial state of the page.

For example:
`),
	ch.Code(generated.MultiStatePageSample).Language("go"),
	utils.DemoWithSnippetLocation("Manipulate Page URL In Event Func", examples_web.MultiStatePagePath, generated.MultiStatePageSampleLocation),
	Markdown(`
This page have several state that encoded in the url:

- Page title have a default value, but if provided with a ~title~ query string, it will use that value
- The edit panel can be open, or closed based on having the ~panel~ query string or not

~web.Location(url.Values{"panel": []string{"1"}}).MergeQuery(true)~ means it will do a push state request to current page, with panel query string panel=1.
~MergeQuery~ means that it will not touch other query strings like ~title=1~ we mentioned above.

In ~update5~ event func, which is when you click the update button after open the panel, ~web.Location(url.Values{"panel": []string{""}}).MergeQuery(true)~ basically removes the query string panel=1, and won't touch any other query strings.

Don't have to be in event func to use push state query, can use a simple ~web.Bind~ to directly change the query string like:

~~~go
A().Text("change page title").Href("javascript:;").
	Attr("@click", web.POST().Queries(url.Values{"title": []string{"Hello"}}).Go()),
~~~

This don't have ~.MergeQuery(true)~, So it will replace the whole query string to only ~title=Hello~

`),
).Title("Manipulate Page URL in Event Func").
	Slug("basics/manipulate-page-url-in-event-func")
