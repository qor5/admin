# Slug

Slug provides an easy way to create a pretty URL for your model.

## Usage

Use `slug.Slug` as your field type with the same name as the benefactor field, from which the slug's value should be dynamically derived, and prepended with `WithSlug`, for example:

```go

type User struct {
  gorm.Model
  Name            string
  NameWithSlug    slug.Slug
}
```

## License

Released under the MIT License.
