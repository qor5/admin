# Usage

## Overview

The seo package allows for the management and injection of dynamic data into HTML tags for the purpose of Search Engine Optimization.

## Create a SEO builder

- Create a builder and register global seo by default. `db` is an instance of gorm.DB

  ```go
  builder := seo.NewBuilder(db)
  ```

- The default global seo name is `Global SEO`, if you want to customize the name, passing the name through `WithGlobalSEOName` option to the `NewBuilder` function.

  ```go
  builder := seo.NewBuilder(db, seo.WithGlobalSEOName("My Global SEO"))
  ```

- Turn off the default inherit the upper level SEO data when the current SEO data is missing

  ```go
  builder := seo.NewBuilder(db, seo.WithInherited(false))
  ```

- The seo package supports localization. you can pass the localization code you want to support through the `WithLocales` option to the `NewBuilder` function. if no localization is set, it defaults to `en`

  ```go
  builder := seo.NewBuilder(db, seo.WithLocales('zh', 'en', 'jp'))
  ```

- **You can pass multiple options at once**

  ```go
  builder := seo.NewBuilder(db, seo.WithLocales('zh'), seo.WithInherited(false), seo.WithGlobalSEOName("My Global SEO"))
  ```

## Register and remove SEO

All registered SEO names are unique, If you have already registered a SEO named `Test`, attempting to register SEO with the same name `Test` will cause the program to panic.

```go
builder.RegisterSEO("Test")
type Test struct {
    ...
}
builder.RegisterSEO(&Test{}) // will panic
```

- Register single seo by name

  ```go
  builder.RegisterSEO("News")
  ```

- Register single seo by model

  **NOTE: the mode to be registered must embed the `Setting` struct**

  **NOTE: The name of the SEO registered through the model is obtained by reflecting the type name, which is `reflect.Type(...).Name()`**

  ```go
  type Page struct {
      ...
      seo Setting // embed `Setting`
      ...
  }
  // The seo name will be "Page"
  builder.RegisterSEO(&Page{})
  ```

- Register multiple SEOs by a list of names

  ```go
  builder.RegisterMultipleSEO("Not Fount", "Internal Server Error")
  ```

- Register multiple SEOs by a list of models

  **NOTE: the mode to be registered must embed the Setting struct**

  ```go
  // The name of the SEOs are "Product" and "Customer" respectively.
  builder.RegisterMultipleSEO(&Product{}, &Customer{})
  ```

- Remove a SEO

  **NOTE: The global seo cannot be removed**

  ```go
  builder.RemoveSEO(&Product{}).RemoveSEO("Not Found")
  ```

  When you remove an SEO, the new parent of its child SEO becomes the parent of the seo.

- The seo supports hierarchy, and you can use the `SetParent` or `AppendChild` methods of `SEO` to configure the hierarchy.

  **With `AppendChildrent` method**

  ```go
  builder.RegisterSEO("A").AppendChildren(
      builder.RegisterSEO("B"),
      builder.RegisterSEO("C"),
      builder.RegisterSEO("D"),
  )

  // The following code is equivalent to the code above.
  builder.RegisterSEO("A").AppendChildren(
      builder.RegisterMultipleSEO("B", "C", D)...
  )
  ```

  **With `SetParent` Method**

  ```go
  seoA := builder.RegisterSEO("A")
  builder.RegisterSEO("B").SetParent(seoA)
  builder.RegisterSEO("C").SetParent(seoA)
  builder.RegisterSEO("D").SetParent(seoA)
  ```

  The final seo structure is as follows:

  ```txt
  Global SEO
  ├── A
      ├── B
      └── C
      └── D
  ```

## Configure SEO

- Register setting variables

  ```go
  builder := NewBuilder(dbForTest)
  builder.RegisterSEO(&Product{}).
    RegisterSettingVariables("Type", "Place of Origin")
  ```

- Register context variables

  ```go
  seoBuilder = seo.NewBuilder(db)
  seoBuilder.RegisterSEO(&models.Post{}).
      RegisterContextVariable(
          "Title",
          func(object interface{}, _ *seo.Setting, _ *http.Request) string {
              if article, ok := object.(models.Post); ok {
                  return article.Title
              }
              return ""
          },
      )
  ```

- Register Meta Property

  ```go
  seoBuilder = seo.NewBuilder(db)
  seoBuilder.RegisterSEO(&models.Post{}).
      RegisterMetaProperty(
          "og:audio",
          func(object interface{}, _ *seo.Setting, _ *http.Request) string {
              if article, ok := object.(models.Post); ok {
                  return article.audio
              }
              return ""
          },
      )
  ```

## Render SEO html data

- Render a single seo

  To call `Render(obj interface{}, req *http.Request)` for rendering a single seo.

  ```go
  type Post struct {
      Title   string
      Author  string
      Content string
      SEO     Setting
  }

  post := &Post{
      Title:   "TestRender",
      Author:  "iBakuman",
      Content: "Hello, Qor5 SEO",
      SEO: Setting{
          Title:            "{{Title}}",
          Description:      "post for testing",
          EnabledCustomize: true,
      },
  }
  builder := NewBuilder(dbForTest)
  builder.RegisterSEO(&Post{}).RegisterContextVariable(
      "Title",
      func(post interface{}, _ *Setting, _ *http.Request) string {
          if p, ok := post.(*Post); ok {
              return p.Title
          }
          return "No title"
      },
  ).RegisterMetaProperty("og:title",
      func(post interface{}, _ *Setting, _ *http.Request) string {
          return "Title for og:title"
      },
  )

  defaultReq, _ := http.NewRequest("POST", "http://www.demo.qor5.com", nil)
  res, err := builder.Render(post, defaultReq).MarshalHTML(context.TODO())
  if err != nil {
      panic(err)
  }
  fmt.Println(string(res))
  ```

  The output of the above code is as follows:

  ```txt
  <title>TestRender</title>

  <meta name='description' content='post for testing'>

  <meta name='keywords'>

  <meta property='og:description' name='og:description'>

  <meta property='og:url' name='og:url'>

  <meta property='og:type' name='og:type' content='website'>

  <meta property='og:image' name='og:image'>

  <meta property='og:title' name='og:title' content='Title for og:title'>
  ```

- Batch render

  To call `BatchRender(objs []interface, req *http.Request)` for batch rendering.

  **NOTE: You need to ensure that all elements in `objs` are of the same type.**

  `BatchRender` does not check if all elements in `objs` are of the same type, as this can be performance-intensive. Therefore,
  it is the responsibility of the caller to ensure that all elements in `objs` are of the same type.
  If you pass `objs` with various types of SEO, it will only take the type of the first element as the standard to obtain the SEO configuration used in the rendering process.

  ```go
  type Post struct {
      Title   string
      Author  string
      Content string
      SEO     Setting
      l10n.Locale
  }

  posts := []interface{}{
      &Post{
          Title:   "TestRenderA",
          Author:  "iBakuman",
          Content: "Hello, Qor5 SEO",
          SEO: Setting{
              Title:            "{{Title}}",
              Description:      "postA for testing",
              EnabledCustomize: true,
          },
      },
      &Post{
          Title:   "TestB",
          Author:  "iBakuman",
          Content: "Hello, Qor5 SEO",
          SEO: Setting{
              Title:            "{{Title}}",
              Description:      "postB for testing",
              EnabledCustomize: true,
          },
      },
  }
  builder := NewBuilder(dbForTest, WithLocales("en"))
  builder.RegisterSEO(&Post{}).RegisterContextVariable(
      "Title",
      func(post interface{}, _ *Setting, _ *http.Request) string {
          if p, ok := post.(*Post); ok {
              return p.Title
          }
          return "No title"
      },
  ).RegisterMetaProperty("og:title",
      func(post interface{}, _ *Setting, _ *http.Request) string {
          return "Title for og:title"
      },
  )

  defaultReq, _ := http.NewRequest("POST", "http://www.demo.qor5.com", nil)
  SEOs := builder.BatchRender(posts, defaultReq)
  for _, seo := range SEOs {
      html, _ := seo.MarshalHTML(context.TODO())
      fmt.Println(string(html))
      fmt.Println("----------------------------")
  }
  ```

  The output of the above code is as follows:

  ```txt
  <title>TestRenderA</title>

  <meta name='description' content='postA for testing'>

  <meta name='keywords'>

  <meta property='og:description' name='og:description'>

  <meta property='og:url' name='og:url'>

  <meta property='og:type' name='og:type' content='website'>

  <meta property='og:image' name='og:image'>

  <meta property='og:title' name='og:title' content='Title for og:title'>

  ----------------------------

  <title>TestB</title>

  <meta name='description' content='postB for testing'>

  <meta name='keywords'>

  <meta property='og:url' name='og:url'>

  <meta property='og:type' name='og:type' content='website'>

  <meta property='og:image' name='og:image'>

  <meta property='og:title' name='og:title' content='Title for og:title'>

  <meta property='og:description' name='og:description'>

  ----------------------------
  ```
