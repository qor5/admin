package basics

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"

	"github.com/qor5/admin/v3/docs/docsrc/generated"
)

var Redirection = Doc(
	Markdown(`
Redirection is  S3 Object Level Redirection implement

## Usage
`),
	ch.Code(generated.NewRedirectionSample).Language("go"),
).Title("Redirection").
	Slug("basics/redirection")
