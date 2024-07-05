package advanced_functions

import (
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	examples_web "github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var WebScope = Doc(
	Markdown(`

## Initialize form and locals vue variables

Since we we don't write vue javascript code, 
It's not easy to add reactive object to that can be used in vue templates. which is returned from the server. 
reactive object are used to trigger view updates,
We pre-set the ~locals~ and ~form~ object as a reactive object, 
and then we can initialize various types of values and slot it into ~locals~. 
And the valid scopes of these values are all inside web.Scope().

For example:
`),
	ch.Code(generated.WebScopeUseLocalsSample1).Language("go"),
	utils.DemoWithSnippetLocation("Web Scope Use Locals", examples_web.WebScopeUseLocalsPath, generated.WebScopeUseLocalsSample1Location),
	Markdown(`
Use ~web.Scope()~ to determine the effective scope of the variable, then use ~.Init(...).VSlot("{ locals }")~ to initialize the variable and slot it into the ~locals~ object.

In ~VBtn("")~, you can use the ~click~ event to change the variable value in ~locals~ to achieve the effect that the page changes with the click.

In ~VBtn("Test Can Not Change Other Scope")~, values in ~locals~ will not change with the click, because the button is not in ~web.Scope()~.

Video Tutorial (<https://www.youtube.com/watch?v=UPuBvVRhUr0>)
`),

	Markdown(`

## Use form

The main use of ~form~ is to submit one form which is inside another form, 
and the two forms are completely independent forms.

In the following example, each color represents a completely separate form. The ~Material Form~ contains the ~Raw Material Form~. You can submit the ~Raw Material Form~ to the server first. After receiving it, server will save the ~Raw Material data~ and return the ~ID~.
In this way, you can submit ~Raw Material ID~ directly in the ~Material Form~.

For example:
`),
	ch.Code(generated.WebScopeUsePlaidFormSample1).Language("go"),
	utils.DemoWithSnippetLocation("Web Scope Use PlaidForm", examples_web.WebScopeUseFormPath, generated.WebScopeUsePlaidFormSample1Location),
	Markdown(`
Use ~web.Scope().VSlot("{ form }")~ to determine the scope of a form.
`),
).Title("web.Scope").
	Slug("basics/web-scope")
