package docsrc

import (
	"embed"

	. "github.com/theplant/docgo"
)

var Home = Doc(
	Markdown(`
## Getting Started with QOR5

This is how you start

`),
).Title("QOR5 Documentation").
	Slug("/").Tables(
	ChildrenTable(
		ContentGroup(
			Activity,
			SEO,
		).Title("Packages"),
	),
)

//go:embed assets/**.*
var Assets embed.FS
