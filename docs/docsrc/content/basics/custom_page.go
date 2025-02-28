package basics

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
)

var CustomPage = Doc(
	Markdown(`
Custom pages allow you to create standalone pages in your admin interface.

## Basic Custom Page

The simplest way to create a custom page is to use the `+"`HandleCustomPage`"+` method on the builder. This method takes a path and a `+"`CustomPage`"+` object that defines the page's content.

Here's how to create a basic custom page:
`),

	ch.Code(generated.PresetsCustomPageDefault).Language("go"),

	Markdown(`
## Custom Page with Parameters

You can create custom pages that accept URL parameters. This is useful for creating detail pages or any page that needs to display specific data based on an identifier.

The parameter is defined in the URL path using curly braces, and you can access it in your handler using `+"`ctx.Param()`"+`:
`),

	ch.Code(generated.PresetsCustomPageWithParams).Language("go"),

	Markdown(`
## Custom Page with Menu

By default, custom pages don't appear in the navigation menu. You can control whether a custom page appears in the menu by using the `+"`Menu`"+` method or by not providing a menu function.

### Hide from Menu

To hide a custom page from the menu, you can explicitly set the menu function to return nil:
`),

	ch.Code(generated.PresetsNewCustomPageHideMenu).Language("go"),
).Title("Custom Pages").
	Slug("basics/custom-page")
