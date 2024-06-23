package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_presets"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Filter = Doc(
	Markdown(`

Assume we have a ~status~ filed in Post. It has 2 possible values, "draft" and "online". If we want to filter posts by its status. We can add a filter like this:

`),
	ch.Code(generated.FilterSample).Language("go"),
	utils.DemoWithSnippetLocation("Basic filter", examples.URLPathByFunc(examples_presets.PresetsBasicFilter)+"/posts", generated.FilterSampleLocation),

	Markdown(`
### QOR5 now supports 7 types of filter option.

PLEASE NOTE THAT all below sample are required you to provide the ~SQLCondition~ you want to perform.

## 1. Filter by String
Set the ~ItemType~ as ~vuetifyx.ItemTypeString~. No ~Options~ needed.
Under this mode, the filter would work in 2 ways,
1. the target value equal to the input string
2. the target value contains the input string

## 2. Filter by Number
Set the ~ItemType~ as ~vuetifyx.ItemTypeNumber~. No ~Options~ needed.
Under this mode, the filter would work in 4 ways
1. the target value equal to the input number
2. the target value is between the input numbers
3. the target value is greater than the input number
4. the target value is less than the input number
`),

	Markdown(`
## 3. Filter by Date
Set the ~ItemType~ as ~vuetifyx.ItemTypeDate~. No ~Options~ needed.
Under this mode, the filter would render a date picker for users to select.
`),

	Markdown(`
## 4. Filter by Date Range
Set the ~ItemType~ as ~vuetifyx.ItemTypeDateRange~. No ~Options~ needed.
Under this mode, the filter would render 2 date pickers, "from" and "to" for users to select.
`),

	Markdown(`
## 5. Filter by Datetime Range
Set the ~ItemType~ as ~vuetifyx.ItemTypeDatetimeRange~. No ~Options~ needed.
Under this mode, the filter would render 2 *date time* pickers, "from" and "to" for users to select.
`),

	Markdown(`
## 6. Filter by Selectable Items
Set the ~ItemType~ as ~vuetifyx.ItemTypeSelect~. You need to provide ~Options~ like this. The ~Text~ is the text users can see in the selector, the ~Value~ is the value of the selector.
`),

	ch.Code(`Options: []*vuetifyx.SelectItem{
		{Text: "Active", Value: "active"},
		{Text: "Inactive", Value: "inactive"},
	},`),

	Markdown(`
## 7. Filter by Multiple Select
Set the ~ItemType~ as ~vuetifyx.ItemTypeMultipleSelect~. You need to provide ~Options~ like above "Selectable Items". But in this mode, the filter would render the options as multi-selectable checkboxes and the query of this filter becomes ~IN~ and ~NOT IN~.
`),
).Title("Filters").
	Slug("basics/filter")
