package basics

import (
	"github.com/qor5/admin/v3/docs/docsrc/examples/examples_admin"
	"github.com/qor5/admin/v3/docs/docsrc/generated"
	"github.com/qor5/admin/v3/docs/docsrc/utils"
	"github.com/qor5/web/v3/examples"
	. "github.com/theplant/docgo"
	"github.com/theplant/docgo/ch"
)

var Listing = Doc(
	Markdown(`
By the [1 Minute Quick Start](/getting-started/one-minute-quick-start.html), We get a default listing page with default columns, But default columns from database columns rarely fit the needs for any real application. Here we will introduce common customizations on the list page.

- Configure fields that displayed on the page
- Modify the display value
- Display a virtual field
- Default scope
- Extend the dot menu

There would be a runable example at the last.

## Configure fields that displayed on the page
Suppose we added a new model called ~Category~, the ~Post~ belongs to ~Category~. Then we want to display ~CategoryID~ on the list page.
`),

	ch.Code(`
type Post struct {
	ID    uint
	Title string
	Body  string

	CategoryID uint

	UpdatedAt time.Time
	CreatedAt time.Time
}

type Category struct {
	ID   uint
	Name string

	UpdatedAt time.Time
	CreatedAt time.Time
}

postModelBuilder.Listing("ID", "Title", "Body", "CategoryID")
`),

	Markdown(`
## Modify the display value
To display the category name rather than category id in the post listing page. The ~ComponentFunc~ would do the work.
The ~obj~ is the ~Post~ record, and ~field~ is the ~CategoryID~ field of this ~Post~ record. You can get the value by ~field.Value(obj)~ function.
`),

	ch.Code(`postModelBuilder.Listing().Field("CategoryID").Label("Category").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	c := models.Category{}
	cid, _ := field.Value(obj).(uint)
	if err := db.Where("id = ?", cid).Find(&c).Error; err != nil {
		// ignore err in the example
	}
	return h.Td(h.Text(c.Name))
})
`).Language("go"),

	Markdown(`
## Display virtual fields
`),
	ch.Code(`postModelBuilder.Listing("ID", "Title", "Body", "CategoryID", "VirtualField")
postModelBuilder.Listing().Field("VirtualField").ComponentFunc(func(obj interface{}, field *presets.FieldContext, ctx *web.EventContext) h.HTMLComponent {
	return h.Td(h.Text("virtual field"))
})
`),

	Markdown(`
## DefaultScope
If we want to display ~Post~ with ~disabled=false~ only. Use the ~Listing().SearcherFunc(...)~ to apply SQL conditions.
`),
	ch.Code(`postModelBuilder.Listing().SearcherFunc(func(model interface{}, params *presets.SearchParams, ctx *web.EventContext) (r interface{}, totalCount int, err error){
	qdb := db.Where("disabled != true")
	return gorm2op.DataOperator(qdb).Search(model, params, ctx)
})
`),

	Markdown(`
## Extend the dot menu
You can extend the dot menu by calling the ~RowMenuItem~ function. If you want to overwrite the default ~Edit~ and ~Delete~ link, you can pass the items you wanted to ~Listing().RowMenu()~
`),
	ch.Code(`rmn := postModelBuilder.Listing().RowMenu()
rmn.RowMenuItem("Show").ComponentFunc(func(obj interface{}, id string, ctx *web.EventContext) h.HTMLComponent {
	return h.Text("Fake Show")
})
`),

	Markdown(`
## Full Example
`),
	ch.Code(generated.PresetsListingSample).Language("go"),
	utils.DemoWithSnippetLocation("Presets Listing Customization Fields", examples.URLPathByFunc(examples_admin.ListingExample)+"/posts", generated.PresetsListingSampleLocation),
).Title("Listing").
	Slug("basics/listing")
