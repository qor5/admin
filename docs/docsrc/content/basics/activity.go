package basics

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
)

var Activity = Doc(
	Markdown(`
QOR5 provides a built-in activity module for recording model operations that may be important for admin users of CMS. These records are designed to be easily queried and audited, and the activity module supports the following features:

* Detailed change logging functionality for model data modifications.
* Allow certain fields to be ignored when comparing modified data, such as the update time.
* Customization of the diffing process for complex field types, like time.Time.
* Customization of the keys used to identify model data.
* Support both automatic and manual CRUD operation recording.
* Provide flexibility to customize the actions other than default CRUD.
* An page for querying the activity log via QOR5 admin

## Initialize the activity package
To initialize activity package with the default configuration, you need to pass a ~presets.Builder~ instance and a database instance.
`),
	ch.Code(generated.NewActivitySample).Language("go"),
	Markdown(`
By default, the activity package uses QOR5 login package's ~~login.UserKey~~ as the default key to fetch the current user from the context. If you want to use your own key, you can use the ~~CreatorContextKey~~ function.

Same with above, the activity package uses the db instance that passed in during initialization to perform db operations. If you need another db to do the work, you can use ~~DBContextKey~~ method.

## Register the models that require activity tracking
This example demonstrates how to register ~~Product~~ into the activity. The activities on the product model will be automatically recorded when it is created, updated, or deleted.
`),
	ch.Code(generated.ActivityRegisterPresetsModelsSample).Language("go"),
).Title("Activity Log")
