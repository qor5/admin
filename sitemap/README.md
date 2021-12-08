# Sitemap

## Usage

- Create a sitemap

```go
sitemap := SiteMap() // will use the default file name and path as '/sitemap.xml'
sitemap := SiteMap("product") //  /product.xml
```

- Register URLs

  - Register a raw string path

    ```go
    sitemap.RegisterRawString("/product1") // path mode
    sitemap.RegisterRawString("https://qor5.dev.com/product1") //url mode
    ```

  - Register a regularURL

    ```go
    sitemap.RegisterURL(URL{Loc: "/product1"}, URL{Loc: "https://qor5.dev.com/product1"})
    ```

  - Register a contextFunc

    ```go
    sitemap.RegisterContextFunc(func(context.Context) []URL {
        // fetch and generate the urls by the context
    }
    )
    ```

  - Register a model

    ```go
    type product struct{
        ...
    }

    // model need to implement this method
    func (p product) Sitemap(ctx context.Context) []URL {
        // fetch urls from db
    }

    sitemap.RegisterModel(&product{}) // path mode
    ```

- Mount to HTTP ServeMux, will automatically fetch the host according to the request and put it in context.

  ```go
  serveMux := http.NewServeMux()
  site.MountTo(serveMux)
  ```

- Generate xml string data directly according to the host in the context

  ```go
  sitemap.EncodeToXml(WithHost("https://qor5.dev.com"))
  ```

- Ping the search engine when the new sitemap is generated

  ```go
  PingBing(sitemap,WithHost("https://qor5.dev.com"))
  PingGoogle(sitemap,WithHost("https://qor5.dev.com"))
  PingAll(sitemap,WithHost("https://qor5.dev.com"))
  ```

# Sitemap Index

```go
index := SiteMapIndex().RegisterSiteMap(SiteMap(), SiteMap("product"), SiteMap("post")) // Register multiple sitemaps

index.EncodeToXml(WithHost("https://qor5.dev.com")) // Generate xml string data directly
index.MountTo(serveMux) // MountTo Mux

```

# Robots

- Create a robots

```go
robots := Robots()
```

- Register a agent

```go
robot.Agent(GoogleAgent).Allow("/product1", "/product2") // Allow

robot.Agent(GoogleAgent).Disallow("/product1", "/product2") // Disallow

robot.Agent(GoogleAgent).AddSitemapUrl(sitemao.ToUrl(WithHost("https://qor5.dev.com")))
 // Add a sitemap

```

- Mount to HTTP ServeMux,

```go
	serveMux := http.NewServeMux()
	robot.MountTo(serveMux)
```

- Generate plain txt

```go
	robot.ToTxt()
```
