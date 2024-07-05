package basics

import (
	"path"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	"github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
	h "github.com/theplant/htmlgo"
)

var I18n = docgo.Doc(
	docgo.Markdown(`
The [i18n package](https://github.com/qor5/x/tree/master/i18n) provides support for internationalization (i18n) in Go applications. 
With the package, you can support multiple languages, 
register messages for each module in each language, and serve multilingual content 
based on the user's preferences.
`),
	h.Br(),
	utils.Demo(
		"I18n",
		path.Join(examples.URLPathByFunc(examples_admin.InternationalizationExample), "/home"),
		"examples/examples_admin/internationalization.go",
	),
	docgo.Markdown(`
## Getting Started
To use the i18n package, you first need to import it into your Go application:
`),
	ch.Code(`import "github.com/qor5/x/v3/i18n"`).Language("go"),
	docgo.Markdown(`
Next, create a new ~Builder~ instance using the ~New()~ function. 
If you want to use it with QOR5, use the ~GetI18n()~ on ~presets.Builder~:
`),
	ch.Code(generated.I18nNew).Language("go"),
	docgo.Markdown(`
The ~Builder~ struct is the central point of the i18n package. 
It holds the supported languages, the messages for each module in each language, 
and the configuration for retrieving the language preference.
`),
	docgo.Markdown(`
## Adding Support Languages
To support multiple languages in your web application, you need to define the languages that you support. 
You can do this by calling the ~SupportLanguages~ function on the ~Builder~ struct:
`),
	ch.Code(generated.I18nSupportLanguages).Language("go"),
	docgo.Markdown(`
The i18n package uses English as the default language. You can add other languages by the ~SupportLanguages~ function.
`),
	docgo.Markdown(`
## Registering Module Messages
Once you have defined the languages, you need to register messages for each module. 
You can do this by the ~RegisterForModule~ function on the ~Builder~ struct:
`),
	ch.Code(generated.I18nRegisterForModule).Language("go"),
	docgo.Markdown(`
The ~RegisterForModule~ function takes three arguments: the language tag, the module key, 
and a pointer to a struct that implements the Messages interface. 
The Messages interface is an empty interface that you can use to define your own messages.

Such a struct might look like this:
`),
	ch.Code(generated.I18nMessagesExample).Language("go"),
	docgo.Markdown(`
If you want to define messages inside the system, 
you can add new variables to the message structure associated with ~presets.ModelsI18nModuleKey~, 
and the variable name definitions follow the camel case.

Such a struct might look like this:
`),
	ch.Code(generated.I18nPresetsMessagesExample).Language("go"),
	docgo.Markdown(`
The ~GetSupportLanguagesFromRequestFunc~ is a method of the ~Builder~ struct in the i18n package.
It allows you to set a function that retrieves the list of supported languages
from an HTTP request, which can be useful in scenarios where the list of supported
languages varies based on the request context.

If you create a separate page, you need to use the ~EnsureLanguage~ to get i18n to work on this page.

The ~EnsureLanguage~ function is an HTTP middleware that ensures the request's language
is properly set and stored. It does this by first checking the query parameters for
a language value, and if found, setting a cookie with that value. If no language
value is present in the query parameters, it looks for the language value in the cookie.

The middleware then determines the best-matching language from the supported languages
based on the "Accept-Language" header of the request. If no match is found,
it defaults to the first supported language. It then sets the language context for
the request, which can be retrieved later by calling the ~MustGetModuleMessages~ function.
`),
	docgo.Markdown(`
## Retrieving Messages
To retrieve module messages in your HTTP handler, you can use the ~MustGetModuleMessages~ function:
`),
	ch.Code(generated.I18nMustGetModuleMessages).Language("go"),
	docgo.Markdown(`
The ~MustGetModuleMessages~ function takes three arguments:
the HTTP request, the module key, and a pointer to a struct
that implements the Messages interface. The function retrieves the messages
for the specified module in the language set by the i18n middleware.
`),
).Slug("basics/i18n").Title("Internationalization")
