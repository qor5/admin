package basics

import (
	"github.com/qor5/docs/v3/docsrc/generated"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
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
	Markdown(`
By default, the activity package will use the primary key as the key to indentify the current model data. You can use ~~SetKeys~~ and ~~AddKeys~~ methods to customize it.

When diffing the modified data, the activity package will ignore the ~~ID~~, ~~CreatedAt~~, ~~UpdatedAt~~, ~~DeletedAt~~ fields. You can either use ~~AddIgnoredFields~~ to append your own fields to the default ignored fields. Or ~~SetIgnoredFields~~ method to replace the default ignored fields.

For special fields like ~~time.Time~~ or media files handled by QOR5 media_library, activity package already handled them. You can use ~~AddTypeHanders~~ method to handle your own field types.

If you want to skip the automatic recording, you can use ~~SkipCreate~~, ~~SkipUpdate~~ and ~~SkipDelete~~ methods.

The Activity package allows for displaying the activities of a record on its editing page. Simply use the ~~EnableActivityInfoTab~~ method to enable this feature. Once enabled, you can customize the format of each activity's display text using the ~~TabHeading~~ method. Additionally, you can make each activity a link to the corresponding record using the ~~SetLink~~ method.

## Record the activity log manually
If you register a preset model into the activity, the activity package will automatically record the activity log for CRUD operations. However, if you need to manually record the activity log for other operations or if you want to register a non-preset model, you can use the following sample code.`),
	ch.Code(generated.ActivityRecordLogSample).Language("go"),
).Title("Activity Log")
