package basics

import (
	"fmt"
	"path"

	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Publish = Doc(
	Markdown(`
Publish controls the online/offline status of records. It generalizes publishing using 3 main modules:
- ~Status~: to flag a record be online/offline
- ~Schedule~: to schedule records to be online/offline automatically
- ~Version~: to allow a record to have more than one copies and chain them together

## Usage
Inject modules to the resource model.
    `),
	ch.Code(generated.PublishInjectModules).Language("go"),
	Markdown(`
Implement primary slug interfaces for passing the values of primary keys between events
    `),
	ch.Code(generated.PublishImplementSlugInterfaces).Language("go"),
	Markdown(`
Create publisher and configure Publish view for model, and remember to display Status and Schedule fields in Editing
    `),
	ch.Code(generated.PublishConfigureView).Language("go"),
	Markdown(`
Implement the publish interfaces if there is a need to publish content to storage(filesystem, AWS S3, ...)
    `),
	ch.Code(generated.PublishImplementPublishInterfaces).Language("go"),
	utils.DemoWithSnippetLocation(
		"Publish",
		path.Join(examples.URLPathByFunc(examples_admin.PublishExample), "/with-publish-products"),
		generated.PublishImplementPublishInterfacesLocation,
	),
	Markdown(fmt.Sprintf(`
## Modules
### Status
%s module stores the status of the record. 
    `, "`Status`")),
	ch.Code(generated.PublishStatus).Language("go"),
	Markdown(`
The initial status is **draft**, after publishing it becomes **online**, and after unpublishing it becomes **offline**.
    `),
	Markdown(fmt.Sprintf(`
### Schedule
%s module schedules records to be online/offline automatically with the publisher job.  
    `, "`Schedule`")),
	ch.Code(generated.PublishSchedule).Language("go"),
	Markdown(fmt.Sprintf(`
If a record has %s set, and the current time is larger than this value, the record will be published and the %s will be set to the actual published time, the %s will be cleared.  
If a record has %s set, and the current time is larger than this value, the record will be unpublished and the %s will be set to the actual unpublished time, the %s will be cleared.  
    `,
		"`ScheduledStartAt`", "`ActualStartAt`", "`ScheduledStartAt`",
		"`ScheduledEndAt`", "`ActualEndAt`", "`ScheduledEndAt`",
	)),
	Markdown(fmt.Sprintf(`
### Version
%s module allows one record to have multiple copies, with Schedule, you can even schedule different prices of a product for a whole year.
    `, "`Version`")),
	ch.Code(generated.PublishVersion).Language("go"),
	Markdown(fmt.Sprintf(`
The %s will be the primary key. By default, the %s value will be %s, e.g. %s. And you can rename a version on interface, which will modify the value of %s.   
    `, "`Version`", "`Version`", "`YYYY-MM-DD-vSeq`", "`2006-01-02-v01`", "`VersionName`")),
	Markdown(fmt.Sprintf(`
### List
%s module publishes list page of resource.
    `, "`List`")),
	ch.Code(generated.PublishList).Language("go"),
).Slug("basics/publish").Title("Publish")
