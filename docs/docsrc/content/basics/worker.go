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

var Worker = Doc(
	Markdown(fmt.Sprintf(`
Worker runs a single Job in the background, it can do so immediately or at a scheduled time.  
Once registered with QOR Admin, Worker will provide a Workers section in the navigation tree, containing pages for listing and managing the following aspects of Workers:

- All Jobs.
- Running: Jobs that are currently running.
- Scheduled: Jobs which have been scheduled to run at a time in the future.
- Done: finished Jobs.
- Errors: any errors output from any Workers that have been run.

## Note
- The default que GoQueQueue(https://github.com/tnclong/go-que) only supports postgres for now.
- To make a job abortable, you need to check %s channel in job handler and stop the handler func.
    `, "`ctx.Done()`")),
	Markdown(`
## Example
`),
	ch.Code(generated.WorkerExample).Language("go"),
	utils.DemoWithSnippetLocation(
		"Worker",
		path.Join(examples.URLPathByFunc(examples_admin.WorkerExample), "/workers"),
		generated.WorkerExampleLocation,
	),
	Markdown(`
## Action Worker
Action Worker is used to visualize the progress of long-running actions.
    `),
	ch.Code(generated.ActionWorkerExample).Language("go"),
	utils.DemoWithSnippetLocation(
		"Action Worker",
		path.Join(examples.URLPathByFunc(examples_admin.ActionWorkerExample), "/example-resources"),
		generated.ActionWorkerExampleLocation,
	),
).Slug("basics/worker").Title("Worker")
