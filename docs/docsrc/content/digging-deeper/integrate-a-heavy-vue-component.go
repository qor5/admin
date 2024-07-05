package digging_deeper

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_vuetify"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var IntegrateAHeavyVueComponent = Doc(
	Markdown(`
We can abstract any complicated of server side render component with [htmlgo](https://github.com/theplant/htmlgo).
But a lots of components in the modern web have done many things on the client side. means there are many logic
happens before the it interact with server side.

Here is an example, a rich text editor. you have a toolbar of buttons that you can interact, most of them won't
need to communicate with server. We are going to integrate the fantastic rich text editor [tiptap](https://tiptap.dev/)
to be used as any ~htmlgo.HTMLComponent~.

**Step 1**: [Create a vue3 project](https://vuejs.org/guide/quick-start.html):

~~~
$ pnpm create vue@latest
~~~

Modify or add a separate ~vite.config.ts~ config file,

`),
	ch.Code(generated.TipTapVueConfig).Language("javascript"),

	Markdown(`
- Made ~Vue~ as externals so that it won't be packed to the dist production js file,
  Since we will be sharing one Vue.js for in one page with other libraries.
- Config svg module to inline the svg icons used by tiptap

**Step 2**: Create a vue component that use tiptap

Install ~tiptap~ and ~tiptap-extensions~ first
~~~
$ pnpm install tiptap tiptap-extensions
~~~

And write the ~editor.vue~ something like this, We omitted the template at here.

`),
	ch.Code(generated.TipTapEditorVueComponent).Language("javascript"),
	Markdown(`

**Step 3**: At ~main.js~, Use a special hook to register the component to ~web/corejs~

`),
	ch.Code(generated.TipTapRegisterVueComponent).Language("javascript"),
	Markdown(`

`),
	Markdown(`

**Step 4**: Use standard [Go embed](https://pkg.go.dev/embed) to pack the dist folder

We write a packr box inside ~tiptapjs.go~ along side the tiptapjs folder.
`),
	ch.Code(generated.TipTapPackrSample).Language("go"),
	Markdown(`
And write a ~build.sh~ to build the javascript to production version.
`),
	ch.Code(generated.TiptapBuilderSH).Language("bash"),

	Markdown(`
**Step 5**: Write a Go wrapper to wrap it to be a ~HTMLComponent~
`),
	ch.Code(generated.TipTapEditorHTMLComponent).Language("go"),

	Markdown(`
**Step 6**: Use it in your web app

To use it, first we have to mount the assets into our app
`),
	ch.Code(generated.TipTapComponentsPackSample).Language("go"),
	Markdown(`
And reference them in our layout function.
`),
	ch.Code(generated.TipTapLayoutSample).Language("go"),

	Markdown(`
And we write a page func to use it like any other component:
`),
	ch.Code(generated.HelloWorldTipTapSample).Language("go"),

	Markdown(`
And now let's check out our fruits:
`),
	utils.DemoWithSnippetLocation("Integrate a Heavy Vue Component", examples_vuetify.HelloWorldTipTapPath, generated.HelloWorldTipTapSampleLocation),
).Title("Integrate a heavy Vue Component").
	Slug("components-guide/integrate-a-heavy-vue-component")
