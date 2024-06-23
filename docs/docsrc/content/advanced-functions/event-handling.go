package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var EventHandling = Doc(
	Markdown(`
We extend vue to support the following types of event handling, so you can simply use go code to implement some complex logic.

Using the ~~~Plaid()~~~ method will create an event handler that defaults to using the current ~~~vars~~~ and ~~~plaidForm~~~.
The default http request method is ~~~Post~~~, if you want to use the ~~~Get~~~ method, you can also use the ~~~Get()~~~ method directly to create an event handler
	`),

	utils.Anchor(H2(""), "URL"),
	Markdown(`Request a page.`),
	ch.Code(generated.EventHandlingURLSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=url", generated.EventHandlingURLSampleLocation),

	utils.Anchor(H2(""), "PushState"),
	Markdown(`Reqest a page and also changing the window location.`),
	ch.Code(generated.EventHandlingPushStateSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=pushstate", generated.EventHandlingPushStateSampleLocation),

	utils.Anchor(H2(""), "Reload"),
	Markdown(`Refresh page.`),
	ch.Code(generated.EventHandlingReloadSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=reload", generated.EventHandlingReloadSampleLocation),

	utils.Anchor(H2(""), "Query"),
	Markdown(`Request a page with a query.`),
	ch.Code(generated.EventHandlingQuerySample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=query", generated.EventHandlingQuerySampleLocation),

	utils.Anchor(H2(""), "MergeQuery"),
	Markdown(`Request a page with merging a query.`),
	ch.Code(generated.EventHandlingMergeQuerySample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=merge_query", generated.EventHandlingMergeQuerySampleLocation),

	utils.Anchor(H2(""), "ClearMergeQuery"),
	Markdown(`Request a page with clearing a query.`),
	ch.Code(generated.EventHandlingClearMergeQuerySample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=clear_merge_query", generated.EventHandlingClearMergeQuerySampleLocation),

	utils.Anchor(H2(""), "StringQuery"),
	Markdown(`Request a page with a query string.`),
	ch.Code(generated.EventHandlingStringQuerySample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=string_query", generated.EventHandlingStringQuerySampleLocation),

	utils.Anchor(H2(""), "Queries"),
	Markdown(`Request a page with url.Values.`),
	ch.Code(generated.EventHandlingQueriesSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=queries", generated.EventHandlingQueriesSampleLocation),

	utils.Anchor(H2(""), "PushStateURL"),
	Markdown(`Request a page with a url and also changing the window location.`),
	ch.Code(generated.EventHandlingQueriesSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=pushstateurl", generated.EventHandlingQueriesSampleLocation),

	utils.Anchor(H2(""), "Location"),
	Markdown(`Open a page with more options.`),
	ch.Code(generated.EventHandlingLocationSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=location", generated.EventHandlingLocationSampleLocation),

	utils.Anchor(H2(""), "FieldValue"),
	Markdown(`Fill in a value on form.`),
	ch.Code(generated.EventHandlingFieldValueSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=fieldvalue", generated.EventHandlingFieldValueSampleLocation),

	utils.Anchor(H2(""), "EventFunc"),
	Markdown(`Register an event func and call it when the event is triggered.`),
	ch.Code(generated.EventHandlingEventFuncSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=eventfunc", generated.EventHandlingEventFuncSampleLocation),

	utils.Anchor(H2(""), "Script"),
	Markdown(`Run a script code.`),
	ch.Code(generated.EventHandlingBeforeScriptSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=script", generated.EventHandlingBeforeScriptSampleLocation),

	utils.Anchor(H2(""), "Raw"),
	Markdown(`Directly call the js method`),
	ch.Code(generated.EventHandlingRawSample).Language("go"),
	utils.DemoWithSnippetLocation("Event Handling", examples_web.EventHandlingPagePath+"?api=raw", generated.EventHandlingRawSampleLocation),
).Title("Event Handling").Slug("basics/event-handling")
