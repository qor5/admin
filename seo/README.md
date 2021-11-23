# SEO

The SEO library allows for the management and injection of dynamic data into HTML tags for the purpose of Search Engine Optimisation.

## Usage

- Create a SEO collection

```go
// Create a collection and register global seo by default
collection := seo.NewCollection()

// Change the default global name
collection.SetGlobalName("My Global SEO")

// Change the default context db key
collection.SetDBContextKey("My DB")

// Change the default seo model setting
type MySEOSetting struct{
  QorSEOSetting
  publish
  l10n
}
collection.SetSettingModel(&MySEOSetting{})

// Turn off the default inherit the upper level SEO data when the current SEO data is missing
collection.SetInherited(false)

```

- Register and remove SEO

```go
// Register mutiple SEO by name
collection.RegisterSEOByNames("Not Found", "Internal Server Error")

// Register a SEO by model
type Product struct{
  Name  string
  Setting Setting
}
collection.RegisterSEO(&Product{})

// Remove a SEO
collection.RemoveSEO(&Product{}).RemoveSEO("Not Found")
```

- Configure SEO

```go
// Change the default SEO name when register a SEO by model

collection.RegisterSEO(&Product{}).SetName("My Product")

// Register a context Variable
collection.RegisterSEO(&Product{}).
          RegisterContextVariables("og:image", func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.image.url
					}).
          RegisterContextVariables("Name", func(obj interface{}, _ *Setting, _ *http.Request) string {
						return obj.Name
					})


// Register setting variable
collection.RegisterSEO(&Product{}).
          RegisterSettingVaribles(struct{ProductTag string}{})
```

- Render SEO html data

```go
// Render Global SEO
collection.RenderGlobal(request)

// Render SEO by a name
collection.Render("product", request)

// Render SEO by a model
collection.Render(Product{}, request)
```
