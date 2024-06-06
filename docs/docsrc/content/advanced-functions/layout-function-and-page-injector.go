package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var LayoutFunctionAndPageInjector = Doc(
	Markdown("Read this code first, Guess what it does."),
	ch.Code(generated.DemoLayoutSample).Language("go"),
	Markdown(`
~ctx.Injector~ is for inject html into default layout's html head, and bottom of body.
html head normally for page title, keywords etc all kinds meta data, and css styles,
javascript libraries etc. You can see we put vue.js into head, but put main.js into the bottom of body.

Next part describe about these asset references:
`),
	ch.Code(generated.ComponentsPackSample).Language("go"),

	Markdown(`
~web.JSComponentsPack~ is the production version of QOR5 core javascript code.
Created by using [@vue/cli](https://cli.vuejs.org/guide/creating-a-project.html),
It does the basic functions like render server side returned html as vue templates.
Provide basic event functions that call to server, and manage push state
(change browser address urls before or after do ajax requests). do page partial refresh etc.

the javascript or css code are packed by using [embed](https://pkg.go.dev/embed).
`),
	ch.Code(generated.PackrSample).Language("go"),
	Markdown(`
And with ~web.PacksHandler~, You can merge multiple javascript or css assets together into one url.
So that browser only need to request them one time. and cache them. The cache is set to the start
time of the process. So next time the app restarts, it invalid the cache.
`),
	utils.Anchor(H2(""), "Summary"),
	Markdown(`
For a new project:

- Use [@vue/cli](https://cli.vuejs.org/guide/creating-a-project.html) to create an asset project that manage your javascript and css. and compile them for production use
- Use [embed](https://pkg.go.dev/embed) to pack them into Go code as ~ComponentPack~, which is a string
- Use ~PacksHandler~ to mount them as available http urls
- Write Layout function to reference them inside head, or bottom of body
`),
).Title("Layout Function and Page Injector").
	Slug("basics/layout-function-and-page-injector")
