package docsrc

import (
	. "github.com/theplant/docgo"
	"github.com/theplant/htmlgo"
)

var Slug = Doc(
	htmlgo.Text("Slug provides an easy way to create a pretty URL for your model."),
	Markdown(`
## Usage

Use ~~~slug.Slug~~~ as your field type with the same name as the benefactor field, from which the slug's value should be dynamically derived, and prepended with ~~~WithSlug~~~, for example:
~~~go

type User struct {
	gorm.Model
	Name            string
	NameWithSlug    slug.Slug
}
~~~
`),
).Title("Slug")
