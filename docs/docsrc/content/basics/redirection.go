package basics

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
)

var Redirection = Doc(
	Markdown(`
Redirection is an S3 Object Level Redirection implementation
- Can control URL redirection via QOR5 Admin
- Only for static pages hosted on S3 
- If it is an internal resource, the prefix must be "/", for example, "/international/index.html", otherwise it will be considered a format error
- Can redirect internal resource A to internal resource B, but the target internal resource B must exist, otherwise the redirection will fail
- Can redirect internal resource A to external resource B, but the target external resource B must be accessible, otherwise the redirection will fail
- Only supports CSV files with the format:
`),
	ch.Code(" - Correct Example 1, redirect to external resource\n```csv\nsource,target\n/international/index.html,https://demo.qor5.com/"),
	ch.Code(" - Correct Example 2, redirect to internal resource\n```csv\nsource,target\n/international/index.html,/international/index2.html"),
	ch.Code(" - Incorrect Example 1, format error\n```csv\nsource,target\n/international/index.html,international/index2.html"),
	ch.Code(" - Incorrect Example 2, duplicate source name\n```csv\nsource,target\n/international/index3.html,/international/index1.html\n/international/index3.html,/international/index2.html"),
	ch.Code(" - Incorrect Example 3, pointing to external resource, external resource is inaccessible\n```csv\nsource,target\n/international/index3.html,https://wwwwwwww/"),
	ch.Code(generated.NewRedirectionSample).Language("go"),
).Title("Redirection").
	Slug("basics/redirection")
