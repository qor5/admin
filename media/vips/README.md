# ENV

### dev env
```brew install vips```

export CGO_CFLAGS_ALLOW="-Xpreprocessor"


### build Dockerfile
```
FROM alpine:3.12

RUN apk add --update go gcc g++ git

RUN apk add --update build-base vips-dev
```
### build command

set CGO_ENABLED=1, eg:
```
GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -tags 'bindatafs' -a -o main main.go
```

### deploy Dockerfile

```
FROM alpine:3.12

RUN apk --update upgrade && \

    apk add ca-certificates && \
    
    apk add tzdata && \
    
    apk add build-base vips-dev && \
    
    rm -rf /var/cache/apk/*
```
 
# Usage

[Setup media library](https://github.com/qor/media#how-to-setup-a-media-library-and-use-media-box) and add below code, then it will compress jpg/png and generate webp for you.

```
import "github.com/qor/media/handlers/vips"

vips.UseVips(vips.Config{EnableGenerateWebp: true})
```

you can adjust image quality by config if you want.
```
type Config struct {
	EnableGenerateWebp bool
	PNGtoWebpQuality   int
	JPEGtoWebpQuality  int
	JPEGQuality        int
	PNGQuality         int
	PNGCompression     int
}
  ```  

