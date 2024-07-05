package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var SwitchPagesWithPushState = Doc(
	Markdown(`Ways that page transition (between ~web.PageFunc~) in QOR5 web app:

- Use a traditional link to a new page by url
- Use a push state link to a new page that only change the current page body to new page body and browser url
- Use a button etc to trigger post to an ~web.EventFunc~ that do some logic, then go to a new page

Inside ~web.EventFunc~, two ways go to a new page:

- Use [push state](https://developer.mozilla.org/en-US/docs/Web/API/History_API#Examples) to only reload the body of the new page, This won't reload javascript and css assets.
- Use redirect url to reload the whole new page, This will reload target new page's javascript and css assets.

This example demonstrated the above:
`),
	ch.Code(generated.PageTransitionSample).Language("go"),
	utils.DemoWithSnippetLocation("Switch Pages With Push State", examples_web.Page1Path, generated.PageTransitionSampleLocation),
	Markdown(`
When running the above demo, If you check Chrome Developer Tools about Network requests,
You will see that the Location link and the Button is actually doing an AJAX request to the other page.

Look like this:
~~~
POST /samples/page_2?__execute_event__=__reload__ HTTP/1.1
~~~

The result is an JSON object with page's html inside.
~__reload__~ is another ~web.EventFunc~ that is the same as ~doAction2~,
But it is default added to every ~web.PageFunc~. So that the web page can
both respond to normal HTTP request from Browser, Search Engine, Or from
other pages in the same web app that can do push state link.
`),
	utils.Anchor(H2(""), "Summary"),
	Markdown(`
- Write once with PageFunc, you get both normal html page render, and AJAX JSON page render
- EventFunc is always called with AJAX request, and you can return to a different page, or rerender the current page,
- EventFunc is not wrapped with layout function.
- EventFunc is used to do data operations, triggered by page's html element. and it's result can be:
	1. Go to a new page
	2. Reload the whole current page
	3. Update partial of the current page

Next we will talk about how to reload the whole current page, and update partial of the current page.
`),
).Title("Switch Pages with Push State").
	Slug("basics/switch-pages-with-push-state")
