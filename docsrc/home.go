package docsrc

import (
	"embed"
	"github.com/theplant/htmlgo"

	. "github.com/theplant/docgo"
)

var Home = Doc(Markdown(`
	QOR5 is the next gen of QOR3. with more flexibility.
`),
).Title("QOR5 Documentation").
	Slug("/").Tables(
	ChildrenTable(
		ContentGroup(
			Doc(htmlgo.Text("placeholder")).Title("Introduction"),
			Doc(htmlgo.Text("placeholder")).Title("Quick Start Guide"),
		).Title("Getting Started"),
		ContentGroup(
			Doc(htmlgo.Text("Form fields")).Title("Form Field"),
			Doc(htmlgo.Text("Form fields")).Title("Advanced Form Field"),
			Doc(htmlgo.Text("Form fields")).Title("Action"),
			Doc(htmlgo.Text("Form fields")).Title("Search & Filter & Scope"),
		).Title("General Admin Configuration"),
		ContentGroup(
			Activity,
			Slug,
			SEO,
		).Title("Addon Packages"),
	),
)

//go:embed assets/**.*
var Assets embed.FS
