## Media Library

Media is a [Golang](http://golang.org/) library that supports the upload of *files*/*images*/*videos* to a filesystem or cloud storage as well as *linked videos* (i.e. YouTube, Vimeo, etc.). The plugin includes:

- cropping and resizing features for images.
- optional multiple sizes for each media resource.
- Accessibility helpers.


###### File Types

Media accepts any and every file type, yet it associates certain file types as *images* or *videos* so as to provide helpers supporting those media's specific needs.


    Images: .jpg, .jpeg, .png, .tif, .tiff, .bmp, .gif

    Videos: .mp4, .m4p, .m4v, .m4v, .mov, .mpeg, .webm, .avi, .ogg, .ogv


## Usage

Media depends on [GORM](https://github.com/go-gorm/gorm) models as it is using [GORM](https://github.com/go-gorm/gorm)'s callbacks to handle file processing, so you will need to register callbacks first:

```go

db, err := gorm.Open(postgres.Open(os.Getenv("DB_PARAMS")), &gorm.Config{})
media.RegisterCallbacks(db)

sess := session.Must(session.NewSession())
oss.Storage = s3.New(&s3.Config{
    Bucket:  os.Getenv("S3_Bucket"),
    Region:  os.Getenv("S3_Region"),
    Session: sess,
})


```

###  Use media box in a model
```go
type Product struct {
	gorm.Model
	HeroImage     media_library.MediaBox `sql:"type:text;"`
}

```

###  Configure media box 
```go

import (
    media_view "github.com/qor/qor5/media/views"
)
b := presets.New()
media_view.Configure(b, db)

p := b.Model(&Product{})
ed := p.Editing( "HeroImage")
ed.Field("HeroImage").WithContextValue(
    media_view.MediaBoxConfig,
    &media_library.MediaBoxConfig{
        AllowType: "image",
        Sizes: map[string]*media.Size{
            "thumb": {
                Width:  400,
                Height: 300,
            },
        },
    })

```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
