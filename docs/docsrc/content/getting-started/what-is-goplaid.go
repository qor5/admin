package getting_started

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	. "github.com/theplant/htmlgo"
)

var WhatIsQOR5 = Doc(
	Markdown(`
QOR5 is a Go library to build web applications.
different from other MVC frameworks. the concepts in QOR5 is **Page**, **Event**, **Component**.
and doesn't include Model.

A Page composite different kinds of Components, and Components trigger Events.
A Page contains many event handlers, and renders one view, and event handlers reload the whole page,
Or update certain part of the page, Or go to a different Page.

QOR5 is opinionated in several ways:

- It prefers writing HTML in static typing Go language, rather than a certain type of template language, Not even go template.
- It try to minify the needs to write any JavaScript/Typescript for building interactive web applications
- It maximize the reusability of Components. since it uses Go to write components, You can abstract component very easy, and use component from a third party Go package is also like using normal Go packages.
- It prefers chain methods to set optional parameters of Component
- It uses [Vue](https://vuejs.org/) js under the hood. and only Vue Component can be integrated

`),
	utils.Anchor(H2(""), "Hello World"),
	Markdown(`
Here is the most sample hello world, that show the header with Hello World.
`),
	ch.Code(generated.HelloWorldSample).Language("go"),
	Markdown(`
~H1("Hello World")~ is actually a simple component. it renders h1 html tag. and been set to page body.

The above is the code you mostly writing. the following is the boilerplate code that needs to write one time.
`),
	ch.Code(generated.HelloWorldMuxSample1).Language("go"),
	ch.Code(generated.HelloWorldMuxSample2).Language("go"),
	ch.Code(generated.HelloWorldMainSample).Language("go"),
	utils.DemoWithSnippetLocation("Hello World", examples_web.HelloWorldPath, generated.HelloWorldMainSampleLocation),

	Markdown(`
If you wondering why ~H1("Hello World")~ and how this worked, Please go ahead and checkout next page
`),
).Title("What is QOR5?").
	Slug("getting-started/what-is-qor5")
