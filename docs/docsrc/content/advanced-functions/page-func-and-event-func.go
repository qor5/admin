package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var PageFuncAndEventFunc = Doc(
	Markdown(`
~PageFunc~ is used to build a web page, ~EventFunc~ is called when user interact with the page, For example button or link clicks.
`),
	ch.Code(generated.PageFuncAndEventFuncDefinition).Language("go"),
	Markdown(`~web.Page(...)~ converts multiple ~EventFunc~s along with one ~PageFunc~ to a ~http.Handler~,
event func needs a name to be used by ~web.POST().EventFunc(name).Go()~ to attach to an html element that post http request to call the ~EventFunc~ when vue event like ~@click~ happens`),
	Markdown("Here is a hello world with more interactions. User click the button will reload the page with latest time"),
	ch.Code(generated.HelloWorldReloadSample).Language("go"),
	utils.DemoWithSnippetLocation("Page Func and Event Func", examples_web.HelloWorldReloadPath, generated.HelloWorldReloadSampleLocation),
	Markdown("Note that you have to mount the `web.Page(...)` instance to http.ServeMux with a path to be able to access the ~PageFunc~ in your browser, when mounting you can also wrap the ~PageFunc~ with middleware, which is ~func(in PageFunc) (out PageFunc)~ a func that take a page func and do some wrapping and return a new page func"),
	ch.Code(generated.HelloWorldReloadMuxSample1).Language("go"),
	Markdown("~wb.Page(...)~ convert any `PageFunc` into `http.Handler`, outside you can wrap any middleware that can use on Go standard `http.Handler`."),
	Markdown(`In case you don't know what is a http.Handler middleware,
It's a function that takes http.Handler as input, might also with other parameters,
And also return a new http.Handler,
[gziphandler](https://github.com/nytimes/gziphandler) is an example.`),

	Markdown(`But What the heck is ~demoLayout~ there?
Well it's a ~PageFunc~ middleware. That takes an ~PageFunc~ as input,
wrap it's ~PageResponse~ with layout html and return a new ~PageFunc~.
If you follow the code to write your own ~PageFunc~,
The button click might not work without this.
Since there is no layout to import needed javascript to make this work.
continue to next page to checkout how to add necessary javascript, css etc to make the demo work.`),
).Title("Page Func and Event Func").
	Slug("basics/page-func-and-event-func")
