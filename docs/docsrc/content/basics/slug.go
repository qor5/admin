package basics

import (
	. "github.com/theplant/docgo"
)

var Slug = Doc(
	Markdown(`
Slug provides an easy way to create pretty URLs for your model.
## Usage

If the source field called ~~Name~~, Use ~~*WithSlug~~ which is ~~NameWithSlug~~ as the slug field name, the field type should be ~~slug.Slug~~. Then the pretty URL would be derived from ~~Name~~ automatically on editing.
~~~go

type User struct {
	gorm.Model
	Name            string
	NameWithSlug    slug.Slug
}
~~~
`),
).Title("Slug")
