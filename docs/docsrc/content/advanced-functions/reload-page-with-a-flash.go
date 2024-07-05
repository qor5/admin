package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var ReloadPageWithAFlash = Doc(
	Markdown(`
The results of an ~web.EventFunc~ could be:

- Go to a new page
- Reload the whole current page
- Refresh part of the current page

Let's demonstrate reload the whole current page:
`),
	ch.Code(generated.ReloadWithFlashSample).Language("go"),
	utils.DemoWithSnippetLocation("Reload Page With a Flash", examples_web.ReloadWithFlashPath, generated.ReloadWithFlashSampleLocation),
	Markdown(`
~ctx.Flash~ Object is used to pass data between ~web.EventFunc~ to ~web.PageFunc~ just after the event func is executed. quite similar to [Rails's Flash](https://api.rubyonrails.org/classes/ActionDispatch/Flash.html).
Different is here you can pass in any complicated struct. as long as the page func to use that flash properly.

~er.Reload = true~ tells it will reload the whole page by running page func again, and with the result's body to replace the browser's html content. the event func and page func are executed in one AJAX request in the server.
`),
).Title("Reload Page with a Flash").
	Slug("basics/reload-page-with-a-flash")
